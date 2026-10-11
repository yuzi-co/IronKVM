package notify

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime"
	"mime/multipart"
	"net/http"
	"net/textproto"
	"net/url"
	"sort"
	"strings"
	"time"

	"NanoKVM-Server/service/events"
)

// Channel names, as the page and the test route use them.
const (
	ChannelWebhook  = "webhook"
	ChannelNtfy     = "ntfy"
	ChannelTelegram = "telegram"
)

const (
	sendTimeout = 20 * time.Second
	// maxReply bounds how much of an answer is read.
	maxReply = 4096

	telegramAPI = "https://api.telegram.org"
)

// message is what a channel sends: the event, its device, and the picture
// when there is one.
type message struct {
	ev     events.Event
	device string
	jpeg   []byte
}

// text is the plain-text form: the title, the message and the fields, one per
// line.
func (m message) text() string {
	var b strings.Builder
	b.WriteString(m.ev.Title)
	if m.ev.Message != "" {
		b.WriteString("\n")
		b.WriteString(m.ev.Message)
	}
	keys := make([]string, 0, len(m.ev.Fields))
	for k := range m.ev.Fields {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		fmt.Fprintf(&b, "\n%s: %s", k, m.ev.Fields[k])
	}
	return b.String()
}

// scrub turns a transport error into one that is safe to log and to show: the
// request URL (which is a secret for a webhook and for Telegram) is dropped,
// and any secret that still appears is masked.
func scrub(err error, secrets ...string) error {
	if err == nil {
		return nil
	}
	var ue *url.Error
	text := err.Error()
	if errors.As(err, &ue) {
		text = ue.Err.Error()
	}
	if errors.Is(err, context.DeadlineExceeded) {
		text = "timed out"
	}
	for _, s := range secrets {
		if s != "" {
			text = strings.ReplaceAll(text, s, "***")
		}
	}
	return errors.New(text)
}

func drain(resp *http.Response) []byte {
	body, _ := io.ReadAll(io.LimitReader(resp.Body, maxReply))
	_ = resp.Body.Close()
	return body
}

// do sends the request and checks for a 2xx answer. detail reads a reason
// from the answer's body for a failure, or returns "".
func do(client *http.Client, req *http.Request, detail func([]byte) string, secrets ...string) error {
	resp, err := client.Do(req)
	if err != nil {
		return scrub(err, secrets...)
	}
	body := drain(resp)
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		if detail != nil {
			if d := detail(body); d != "" {
				return fmt.Errorf("answered %d: %s", resp.StatusCode, scrub(errors.New(d), secrets...))
			}
		}
		return fmt.Errorf("answered %d", resp.StatusCode)
	}
	return nil
}

func newRequest(ctx context.Context, method, target string, body io.Reader) (*http.Request, error) {
	req, err := http.NewRequestWithContext(ctx, method, target, body)
	if err != nil {
		return nil, scrub(err, target)
	}
	req.Header.Set("User-Agent", "IronKVM")
	return req, nil
}

// sendWebhook posts the event as JSON. The picture is in it as base64 only
// when the owner asked for that.
func sendWebhook(ctx context.Context, client *http.Client, w Webhook, m message) error {
	body := map[string]any{
		"type":    m.ev.Type,
		"time":    m.ev.Time.UTC().Format(time.RFC3339),
		"device":  m.device,
		"title":   m.ev.Title,
		"message": m.ev.Message,
		"fields":  m.ev.Fields,
	}
	if w.IncludeSnapshot && len(m.jpeg) > 0 {
		body["snapshot"] = base64.StdEncoding.EncodeToString(m.jpeg)
	}
	data, err := json.Marshal(body)
	if err != nil {
		return err
	}

	req, err := newRequest(ctx, http.MethodPost, w.URL, bytes.NewReader(data))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	return do(client, req, nil, w.URL)
}

// ntfyPriority and ntfyTags give each event type its urgency and its icon.
var ntfyStyle = map[string]struct {
	priority string
	tags     string
}{
	events.SignalLost:      {"4", "warning,tv"},
	events.SignalRestored:  {"3", "white_check_mark,tv"},
	events.WatchdogAction:  {"4", "rotating_light"},
	events.LoginFailures:   {"4", "lock"},
	events.ATXPress:        {"3", "electric_plug"},
	events.FailsafeEntered: {"5", "rotating_light"},
	events.HostLogMatch:    {"5", "rotating_light,scroll"},
}

// header makes s safe for an HTTP header: one line, and RFC 2047 when it is
// not ASCII, which ntfy reads.
func header(s string) string {
	s = strings.NewReplacer("\r", " ", "\n", " ").Replace(s)
	if len(s) > 250 {
		s = s[:250]
	}
	for i := 0; i < len(s); i++ {
		if s[i] >= 0x80 {
			return mime.BEncoding.Encode("UTF-8", strings.ToValidUTF8(s, ""))
		}
	}
	return s
}

// sendNtfy publishes to the topic. A picture goes as the body, an attachment
// ntfy shows inline, and the text moves to a header.
func sendNtfy(ctx context.Context, client *http.Client, n Ntfy, m message) error {
	target := strings.TrimRight(n.Server, "/") + "/" + url.PathEscape(n.Topic)

	var body io.Reader
	text := m.text()
	hdr := http.Header{}
	hdr.Set("Title", header(m.ev.Title))
	if style, ok := ntfyStyle[m.ev.Type]; ok {
		hdr.Set("Priority", style.priority)
		hdr.Set("Tags", style.tags)
	}
	if len(m.jpeg) > 0 {
		body = bytes.NewReader(m.jpeg)
		hdr.Set("Filename", "snapshot.jpg")
		hdr.Set("Content-Type", "image/jpeg")
		hdr.Set("Message", header(strings.TrimPrefix(text, m.ev.Title+"\n")))
	} else {
		body = strings.NewReader(strings.TrimPrefix(text, m.ev.Title+"\n"))
		hdr.Set("Content-Type", "text/plain; charset=utf-8")
	}

	req, err := newRequest(ctx, http.MethodPost, target, body)
	if err != nil {
		return err
	}
	for k, v := range hdr {
		req.Header[k] = v
	}
	if n.Token != "" {
		req.Header.Set("Authorization", "Bearer "+n.Token)
	}
	return do(client, req, ntfyReason, n.Token)
}

func ntfyReason(body []byte) string {
	var r struct {
		Error string `json:"error"`
	}
	if json.Unmarshal(body, &r) == nil {
		return r.Error
	}
	return ""
}

// telegramBase is a variable so a test can point the bot at a local server.
var telegramBase = telegramAPI

const (
	telegramCaptionMax = 1024
	telegramTextMax    = 4096
)

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return strings.ToValidUTF8(s[:n-3], "") + "..."
}

// sendTelegram sends a photo with the text as its caption, or the text alone.
func sendTelegram(ctx context.Context, client *http.Client, t Telegram, m message) error {
	endpoint := telegramBase + "/bot" + t.BotToken + "/"

	var (
		req *http.Request
		err error
	)
	if len(m.jpeg) > 0 {
		var buf bytes.Buffer
		w := multipart.NewWriter(&buf)
		_ = w.WriteField("chat_id", t.ChatID)
		_ = w.WriteField("caption", truncate(m.text(), telegramCaptionMax))
		part, perr := w.CreatePart(textproto.MIMEHeader{
			"Content-Disposition": {`form-data; name="photo"; filename="snapshot.jpg"`},
			"Content-Type":        {"image/jpeg"},
		})
		if perr != nil {
			return perr
		}
		_, _ = part.Write(m.jpeg)
		if err := w.Close(); err != nil {
			return err
		}
		req, err = newRequest(ctx, http.MethodPost, endpoint+"sendPhoto", &buf)
		if err != nil {
			return scrub(err, t.BotToken)
		}
		req.Header.Set("Content-Type", w.FormDataContentType())
	} else {
		data, _ := json.Marshal(map[string]string{
			"chat_id": t.ChatID,
			"text":    truncate(m.text(), telegramTextMax),
		})
		req, err = newRequest(ctx, http.MethodPost, endpoint+"sendMessage", bytes.NewReader(data))
		if err != nil {
			return scrub(err, t.BotToken)
		}
		req.Header.Set("Content-Type", "application/json")
	}
	return do(client, req, telegramReason, t.BotToken)
}

func telegramReason(body []byte) string {
	var r struct {
		Description string `json:"description"`
	}
	if json.Unmarshal(body, &r) == nil {
		return r.Description
	}
	return ""
}
