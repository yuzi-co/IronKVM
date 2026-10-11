package middleware

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

const testLimit = 1024

// bodyEngine answers POST /echo and POST /upload with the length of the body
// the handler could read, and the bind error if it could not.
func bodyEngine() *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(LimitBody(testLimit, map[string]int64{"POST /upload": NoBodyLimit}))

	readAll := func(c *gin.Context) {
		data, err := io.ReadAll(c.Request.Body)
		if err != nil {
			c.JSON(http.StatusOK, gin.H{"code": -1, "msg": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"code": 0, "data": len(data)})
	}
	r.POST("/echo", readAll)
	r.POST("/upload", readAll)
	r.POST("/json", func(c *gin.Context) {
		var req struct {
			Text string `json:"text"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusOK, gin.H{"code": -1, "msg": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"code": 0, "data": len(req.Text)})
	})
	return r
}

type reply struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
	Data int    `json:"data"`
}

func post(t *testing.T, r http.Handler, path string, body io.Reader, length int64) (int, reply) {
	t.Helper()
	req := httptest.NewRequest(http.MethodPost, path, body)
	req.ContentLength = length
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	var rsp reply
	if err := json.Unmarshal(w.Body.Bytes(), &rsp); err != nil {
		t.Fatalf("POST %s answered %d with a body that is not the envelope: %q", path, w.Code, w.Body.String())
	}
	return w.Code, rsp
}

// onlyReader hides everything but Read, so httptest cannot work out a length.
type onlyReader struct{ io.Reader }

func TestABodyAtTheCapIsUnaffected(t *testing.T) {
	r := bodyEngine()
	body := strings.Repeat("a", testLimit)

	code, rsp := post(t, r, "/echo", strings.NewReader(body), testLimit)
	if code != http.StatusOK || rsp.Code != 0 || rsp.Data != testLimit {
		t.Fatalf("a body of exactly the cap answered %d %+v", code, rsp)
	}
}

func TestADeclaredBodyOneByteOverTheCapIsRefused(t *testing.T) {
	r := bodyEngine()
	body := strings.Repeat("a", testLimit+1)

	code, rsp := post(t, r, "/echo", strings.NewReader(body), testLimit+1)
	if code != http.StatusRequestEntityTooLarge {
		t.Fatalf("expected 413, got %d %+v", code, rsp)
	}
	if rsp.Code == 0 || rsp.Msg == "" {
		t.Fatalf("the refusal should carry a non-zero code and a reason, got %+v", rsp)
	}
}

func TestAChunkedBodyIsCappedToo(t *testing.T) {
	r := bodyEngine()

	code, rsp := post(t, r, "/echo", onlyReader{strings.NewReader(strings.Repeat("a", testLimit+1))}, -1)
	if code != http.StatusRequestEntityTooLarge || rsp.Code == 0 {
		t.Fatalf("a chunked body over the cap answered %d %+v", code, rsp)
	}

	code, rsp = post(t, r, "/echo", onlyReader{strings.NewReader(strings.Repeat("a", testLimit))}, -1)
	if code != http.StatusOK || rsp.Code != 0 || rsp.Data != testLimit {
		t.Fatalf("a chunked body at the cap should reach the handler whole, got %d %+v", code, rsp)
	}
}

func TestJSONBindingSeesTheBodyUnderTheCap(t *testing.T) {
	r := bodyEngine()
	body := `{"text":"` + strings.Repeat("x", 100) + `"}`

	code, rsp := post(t, r, "/json", strings.NewReader(body), int64(len(body)))
	if code != http.StatusOK || rsp.Code != 0 || rsp.Data != 100 {
		t.Fatalf("a small JSON body answered %d %+v", code, rsp)
	}

	big := `{"text":"` + strings.Repeat("x", testLimit) + `"}`
	code, _ = post(t, r, "/json", strings.NewReader(big), int64(len(big)))
	if code != http.StatusRequestEntityTooLarge {
		t.Fatalf("a JSON body over the cap answered %d", code)
	}
}

func TestARouteWithoutACapTakesALargeBody(t *testing.T) {
	r := bodyEngine()
	size := 64 * testLimit

	code, rsp := post(t, r, "/upload", strings.NewReader(strings.Repeat("a", size)), int64(size))
	if code != http.StatusOK || rsp.Code != 0 || rsp.Data != size {
		t.Fatalf("the uncapped route answered %d %+v", code, rsp)
	}
}

func TestAnUnknownRouteIsCappedAtTheDefault(t *testing.T) {
	r := bodyEngine()

	code, _ := post(t, r, "/nowhere", strings.NewReader(strings.Repeat("a", testLimit+1)), testLimit+1)
	if code != http.StatusRequestEntityTooLarge {
		t.Fatalf("a POST to an unknown route over the cap answered %d", code)
	}
}

// The HID, terminal and stream websockets upgrade a GET, which carries no
// body; the test says the cap leaves them alone rather than the theory.
func TestAWebsocketStillConnectsThroughTheCap(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(LimitBody(testLimit, nil))
	upgrader := websocket.Upgrader{}
	r.GET("/ws", func(c *gin.Context) {
		conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
		if err != nil {
			return
		}
		defer conn.Close()
		kind, data, err := conn.ReadMessage()
		if err != nil {
			return
		}
		_ = conn.WriteMessage(kind, data)
	})

	server := httptest.NewServer(r)
	defer server.Close()

	conn, _, err := websocket.DefaultDialer.Dial("ws"+strings.TrimPrefix(server.URL, "http")+"/ws", nil)
	if err != nil {
		t.Fatalf("the websocket did not connect: %s", err)
	}
	defer conn.Close()

	// A message larger than the body cap: the cap is on request bodies, not
	// on what a socket carries once it is open.
	message := strings.Repeat("m", 4*testLimit)
	if err := conn.WriteMessage(websocket.TextMessage, []byte(message)); err != nil {
		t.Fatalf("write: %s", err)
	}
	_, data, err := conn.ReadMessage()
	if err != nil {
		t.Fatalf("read: %s", err)
	}
	if string(data) != message {
		t.Fatalf("the echo came back %d bytes, want %d", len(data), len(message))
	}
}
