package vm

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"

	"NanoKVM-Server/config"
)

// usePowerLEDSetting sets the power LED setting for one test and records what
// the handler saves instead of writing /etc/kvm/server.yaml.
func usePowerLEDSetting(t *testing.T, connected bool, saveErr error) *[]bool {
	t.Helper()

	conf := config.GetInstance()
	original := conf.HardwareSettings
	conf.HardwareSettings.PowerLED = connected

	var saved []bool
	originalSave := savePowerLEDSetting
	savePowerLEDSetting = func(connected bool) error {
		saved = append(saved, connected)
		return saveErr
	}

	t.Cleanup(func() {
		conf.HardwareSettings = original
		savePowerLEDSetting = originalSave
	})
	return &saved
}

func gpioEngine() *gin.Engine {
	gin.SetMode(gin.TestMode)
	service := NewService()
	r := gin.New()
	r.GET("/api/vm/gpio", service.GetGpio)
	r.GET("/api/vm/gpio/power-led", service.GetPowerLED)
	r.POST("/api/vm/gpio/power-led", service.SetPowerLED)
	return r
}

func serve(r *gin.Engine, method, path, body string) map[string]any {
	request := httptest.NewRequest(method, path, bytes.NewBufferString(body))
	if body != "" {
		request.Header.Set("Content-Type", "application/json")
	}
	w := httptest.NewRecorder()
	r.ServeHTTP(w, request)

	var out map[string]any
	_ = json.Unmarshal(w.Body.Bytes(), &out)
	return out
}

func TestPowerLEDConnectedFollowsTheSetting(t *testing.T) {
	usePowerLEDSetting(t, false, nil)
	if PowerLEDConnected() {
		t.Fatal("PowerLEDConnected() with the setting off")
	}

	config.GetInstance().HardwareSettings.PowerLED = true
	if !PowerLEDConnected() {
		t.Fatal("not PowerLEDConnected() with the setting on")
	}
}

func TestSetPowerLEDSavesAndAppliesTheSetting(t *testing.T) {
	saved := usePowerLEDSetting(t, false, nil)
	r := gpioEngine()

	rsp := serve(r, http.MethodPost, "/api/vm/gpio/power-led", `{"connected":true}`)
	if rsp["code"] != float64(0) {
		t.Fatalf("POST answered %v", rsp)
	}
	if len(*saved) != 1 || !(*saved)[0] {
		t.Fatalf("saved %v, want [true]", *saved)
	}
	if !PowerLEDConnected() {
		t.Fatal("the setting was saved but not applied")
	}

	rsp = serve(r, http.MethodGet, "/api/vm/gpio/power-led", "")
	data, _ := rsp["data"].(map[string]any)
	if data["connected"] != true {
		t.Fatalf("GET answered %v", rsp)
	}
}

func TestSetPowerLEDThatFailsToSaveChangesNothing(t *testing.T) {
	usePowerLEDSetting(t, false, errors.New("disk full"))

	rsp := serve(gpioEngine(), http.MethodPost, "/api/vm/gpio/power-led", `{"connected":true}`)
	if rsp["code"] == float64(0) {
		t.Fatalf("POST answered %v, want an error", rsp)
	}
	if PowerLEDConnected() {
		t.Fatal("the setting changed although it was not saved")
	}
}

// The UI shows an unknown LED rather than an unlit one when it is not wired,
// and a board without the line at all is not an error then.
func TestGetGpioSaysWhetherTheLEDIsConnected(t *testing.T) {
	_, _, led := useButtons(t, "0\n")
	// Only the alpha board has an HDD LED to read as well.
	config.GetInstance().Hardware.Version = config.HWVersionBeta
	usePowerLEDSetting(t, true, nil)
	r := gpioEngine()

	rsp := serve(r, http.MethodGet, "/api/vm/gpio", "")
	data, _ := rsp["data"].(map[string]any)
	if rsp["code"] != float64(0) || data["ledConnected"] != true || data["pwr"] != true {
		t.Fatalf("connected: GET answered %v", rsp)
	}

	config.GetInstance().HardwareSettings.PowerLED = false
	config.GetInstance().Hardware.GPIOPowerLED = led + ".missing"
	rsp = serve(r, http.MethodGet, "/api/vm/gpio", "")
	data, _ = rsp["data"].(map[string]any)
	if rsp["code"] != float64(0) || data["ledConnected"] != false || data["pwr"] != false {
		t.Fatalf("not connected: GET answered %v", rsp)
	}
}
