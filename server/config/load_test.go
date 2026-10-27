package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

// goldenResult is what the server started with for one server.yaml: where the
// configuration came from ("file", "default" when the file could not be read,
// "reset" when it lacked the ports and was replaced), whether decoding it
// failed, which stops the server, and the configuration when it did not.
type goldenResult struct {
	Source string
	Error  bool
	Config *Config `json:",omitempty"`
}

// loadForTest runs content through the steps initialize takes, without the
// files it writes.
func loadForTest(data []byte) goldenResult {
	source := "file"
	root, err := parseConfig(data)
	if err != nil {
		source = "default"
		if root, err = defaultRoot(); err != nil {
			panic(err)
		}
	}
	if !hasPorts(root) {
		source = "reset"
		if root, err = defaultRoot(); err != nil {
			panic(err)
		}
	}

	var conf Config
	if err := decodeConfig(root, &conf); err != nil {
		return goldenResult{Source: source, Error: true}
	}
	return goldenResult{Source: source, Config: &conf}
}

// Each file in testdata/serveryaml is a server.yaml the server may meet: the
// defaults this fork and upstream NanoKVM write, the examples in the READMEs,
// an early upstream file, and files with keys in other casings, values of
// other types and broken content. Its .want.json is what viper made of it,
// recorded before viper was removed; the server must start from the same
// configuration, or stop where viper stopped it.
//
// readme-en.yaml and readme-zh.yaml stop the server: their security block has
// a comma after each number.
func TestServerYAMLLoadsAsViperLoadedIt(t *testing.T) {
	files, err := filepath.Glob("testdata/serveryaml/*.yaml")
	if err != nil {
		t.Fatal(err)
	}
	if len(files) == 0 {
		t.Fatal("no fixtures")
	}

	for _, f := range files {
		name := strings.TrimSuffix(filepath.Base(f), ".yaml")
		t.Run(name, func(t *testing.T) {
			data, err := os.ReadFile(f)
			if err != nil {
				t.Fatal(err)
			}
			wantJSON, err := os.ReadFile(strings.TrimSuffix(f, ".yaml") + ".want.json")
			if err != nil {
				t.Fatal(err)
			}
			var want goldenResult
			if err := json.Unmarshal(wantJSON, &want); err != nil {
				t.Fatal(err)
			}

			got := loadForTest(data)
			if !reflect.DeepEqual(got, want) {
				gotJSON, _ := json.MarshalIndent(got, "", "\t")
				t.Fatalf("got\n%s\nwant\n%s", gotJSON, wantJSON)
			}
		})
	}
}

// The default configuration the server writes must load as itself.
func TestTheDefaultConfigurationRoundTrips(t *testing.T) {
	root, err := defaultRoot()
	if err != nil {
		t.Fatal(err)
	}
	var conf Config
	if err := decodeConfig(root, &conf); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(conf, *defaultConfig) {
		t.Fatalf("got %+v, want %+v", conf, *defaultConfig)
	}
}

// Saving a setting reads server.yaml through Read, which must see a key in
// another casing the way the server did at start, or the save would drop it.
func TestReadMatchesKeysInAnyCase(t *testing.T) {
	data, err := os.ReadFile("testdata/serveryaml/quirk-case.yaml")
	if err != nil {
		t.Fatal(err)
	}
	root, err := parseConfig(data)
	if err != nil {
		t.Fatal(err)
	}
	var conf Config
	if err := decodeConfig(root, &conf); err != nil {
		t.Fatal(err)
	}
	if conf.JWT.SecretKey != "abc" || !conf.HTTP2 || !conf.HardwareSettings.PowerLED {
		t.Fatalf("keys in other casings were lost: %+v", conf)
	}
}
