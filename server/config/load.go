package config

import (
	"errors"
	"reflect"
	"strconv"
	"strings"

	"gopkg.in/yaml.v3"
)

// server.yaml was read at start through viper until it was dropped to shrink the
// binary. Viper matched keys in any case and converted a scalar of the wrong type
// where it could ("8080" for a port, "true" for a switch, one address where a
// list was due). Files on devices may depend on either, so parseConfig and
// decodeConfig keep both. The tests in load_test.go hold the results to what
// viper produced for the same files.

// parseConfig parses server.yaml content. It fails where viper's read failed,
// which made the server start from the defaults: content that is not YAML, a
// document that is not a mapping, and a key given twice.
func parseConfig(data []byte) (*yaml.Node, error) {
	// Viper read the file into a map first, which is what refuses a repeated
	// key and a document that is not a mapping.
	var m map[string]any
	if err := yaml.Unmarshal(data, &m); err != nil {
		return nil, err
	}

	var doc yaml.Node
	if err := yaml.Unmarshal(data, &doc); err != nil {
		return nil, err
	}
	if doc.Kind == 0 {
		// An empty file.
		return &yaml.Node{Kind: yaml.MappingNode, Tag: "!!map"}, nil
	}
	if doc.Kind != yaml.DocumentNode || len(doc.Content) != 1 || doc.Content[0].Kind != yaml.MappingNode {
		return nil, errors.New("server.yaml is not a mapping")
	}

	root := doc.Content[0]
	normalize(root, reflect.TypeOf(Config{}))
	return root, nil
}

// defaultRoot is the default configuration as parseConfig would return it.
func defaultRoot() (*yaml.Node, error) {
	data, err := yaml.Marshal(defaultConfig)
	if err != nil {
		return nil, err
	}
	return parseConfig(data)
}

// hasPorts reports whether the configuration names both ports. A file written
// by an early version may not, and is then replaced by the defaults.
func hasPorts(root *yaml.Node) bool {
	// A port that does not decode counts as missing, so the error is not
	// needed: yaml.v3 decodes what it can before it reports one.
	var p struct {
		Port Port `yaml:"port"`
	}
	_ = root.Decode(&p)

	return p.Port.Http > 0 && p.Port.Https > 0
}

// decodeConfig decodes what parseConfig returned.
func decodeConfig(root *yaml.Node, conf *Config) error {
	return root.Decode(conf)
}

// normalize rewrites n, which is to be decoded into a value of type t, so that
// yaml.v3 decodes it as viper did: every key known to t in the casing of its
// yaml tag, the later of two keys that differ only in case, and scalars
// converted to the field's type by the rules of viper's weak decoding.
func normalize(n *yaml.Node, t reflect.Type) {
	for t.Kind() == reflect.Pointer {
		t = t.Elem()
	}
	if n.Kind == yaml.AliasNode {
		n = n.Alias
	}

	switch t.Kind() {
	case reflect.Struct:
		if n.Kind == yaml.MappingNode {
			normalizeMapping(n, t)
		}
	case reflect.Slice:
		switch {
		case n.Kind == yaml.SequenceNode:
			for _, c := range n.Content {
				normalize(c, t.Elem())
			}
		case n.Kind == yaml.ScalarNode && n.ShortTag() != "!!null" && t.Elem().Kind() == reflect.String:
			scalarToSequence(n)
		}
	case reflect.String, reflect.Bool,
		reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		if n.Kind == yaml.ScalarNode {
			convertScalar(n, t.Kind())
		}
	}
}

func normalizeMapping(n *yaml.Node, t reflect.Type) {
	type field struct {
		name string
		typ  reflect.Type
	}
	fields := make(map[string]field)
	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)
		name, _, _ := strings.Cut(f.Tag.Get("yaml"), ",")
		if name == "-" || !f.IsExported() {
			continue
		}
		if name == "" {
			name = strings.ToLower(f.Name)
		}
		fields[strings.ToLower(name)] = field{name, f.Type}
	}

	// at is where each field's pair went in out, and given the key as the
	// file wrote it.
	at := make(map[string]int)
	given := make(map[string]string)
	out := n.Content[:0:0]
	for i := 0; i+1 < len(n.Content); i += 2 {
		k, v := n.Content[i], n.Content[i+1]
		f, ok := fields[strings.ToLower(k.Value)]
		if !ok || k.Kind != yaml.ScalarNode {
			out = append(out, k, v)
			continue
		}
		normalize(v, f.typ)

		// The same key twice is left for the decoder to refuse, as viper
		// did. The same key in two casings is one key to viper, and the later
		// one counts.
		if j, seen := at[f.name]; seen && given[f.name] != k.Value {
			given[f.name] = k.Value
			out[j+1] = v
			continue
		}
		at[f.name] = len(out)
		given[f.name] = k.Value
		k.Value = f.name
		out = append(out, k, v)
	}
	n.Content = out
}

// scalarToSequence turns a scalar into the list viper made of it for a list of
// strings: a string split at commas, an empty string an empty list, and
// anything else a list of itself.
func scalarToSequence(n *yaml.Node) {
	var v any
	if err := n.Decode(&v); err != nil {
		return
	}

	var items []*yaml.Node
	if s, ok := v.(string); ok {
		if s != "" {
			for _, part := range strings.Split(s, ",") {
				items = append(items, &yaml.Node{Kind: yaml.ScalarNode, Tag: "!!str", Value: part})
			}
		}
	} else {
		item := *n
		convertScalar(&item, reflect.String)
		items = append(items, &item)
	}

	*n = yaml.Node{Kind: yaml.SequenceNode, Tag: "!!seq", Content: items, Line: n.Line, Column: n.Column}
}

// convertScalar converts n to the kind of field it is decoded into, by the
// rules of viper's weak decoding. A value those rules cannot convert is left
// for the decoder to refuse.
func convertScalar(n *yaml.Node, kind reflect.Kind) {
	if n.ShortTag() == "!!null" {
		return
	}
	var v any
	if err := n.Decode(&v); err != nil {
		return
	}

	switch kind {
	case reflect.String:
		var s string
		switch x := v.(type) {
		case bool:
			s = "0"
			if x {
				s = "1"
			}
		case int:
			s = strconv.Itoa(x)
		case uint64:
			s = strconv.FormatUint(x, 10)
		case float64:
			s = strconv.FormatFloat(x, 'f', -1, 64)
		default:
			return
		}
		n.Tag, n.Value, n.Style = "!!str", s, 0

	case reflect.Bool:
		var b bool
		switch x := v.(type) {
		case bool:
			return
		case int:
			b = x != 0
		case uint64:
			b = x != 0
		case float64:
			b = x != 0
		case string:
			if x != "" {
				var err error
				if b, err = strconv.ParseBool(x); err != nil {
					return
				}
			}
		default:
			return
		}
		n.Tag, n.Value, n.Style = "!!bool", strconv.FormatBool(b), 0

	default:
		// An integer kind.
		var s string
		switch x := v.(type) {
		case bool:
			s = "0"
			if x {
				s = "1"
			}
		case int, uint64:
			return
		case float64:
			if kind >= reflect.Uint && kind <= reflect.Uint64 {
				s = strconv.FormatUint(uint64(x), 10)
			} else {
				s = strconv.FormatInt(int64(x), 10)
			}
		case string:
			if x == "" {
				s = "0"
			} else if kind >= reflect.Uint && kind <= reflect.Uint64 {
				u, err := strconv.ParseUint(x, 0, 64)
				if err != nil {
					return
				}
				s = strconv.FormatUint(u, 10)
			} else {
				i, err := strconv.ParseInt(x, 0, 64)
				if err != nil {
					return
				}
				s = strconv.FormatInt(i, 10)
			}
		default:
			return
		}
		n.Tag, n.Value, n.Style = "!!int", s, 0
	}
}
