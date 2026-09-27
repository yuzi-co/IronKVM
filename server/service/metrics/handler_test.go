package metrics

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestHandlerServesTheTextFormat(t *testing.T) {
	setVar(t, &sections, []section{sectionWriting("a", 1)})

	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.GET("/api/metrics", Handler)

	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/api/metrics", nil))

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", w.Code)
	}
	if got := w.Header().Get("Content-Type"); got != ContentType {
		t.Fatalf("Content-Type = %q, want %q", got, ContentType)
	}
	assertText(t, w.Body.String(), "# HELP ironkvm_a Section a.\n# TYPE ironkvm_a gauge\nironkvm_a 1\n")
}
