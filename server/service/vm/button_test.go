package vm

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/gin-gonic/gin"

	"NanoKVM-Server/config"
)

// useButtons points the power, reset and LED lines at files in a temporary
// directory for one test. The buttons start released ("0"), and the LED reads
// ledValue.
func useButtons(t *testing.T, ledValue string) (power, reset, led string) {
	t.Helper()

	dir := t.TempDir()
	power = filepath.Join(dir, "power")
	reset = filepath.Join(dir, "reset")
	led = filepath.Join(dir, "led")

	for _, path := range []string{power, reset} {
		if err := os.WriteFile(path, []byte("0"), 0o600); err != nil {
			t.Fatalf("failed to seed %s: %s", path, err)
		}
	}
	if ledValue != "" {
		if err := os.WriteFile(led, []byte(ledValue), 0o600); err != nil {
			t.Fatalf("failed to seed the led: %s", err)
		}
	}

	conf := config.GetInstance()
	original := conf.Hardware
	conf.Hardware.GPIOPower = power
	conf.Hardware.GPIOReset = reset
	conf.Hardware.GPIOPowerLED = led
	t.Cleanup(func() { conf.Hardware = original })

	return power, reset, led
}

func readLine(t *testing.T, path string) string {
	t.Helper()

	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read %s: %s", path, err)
	}
	return string(content)
}

func TestPressButtonHoldsOnlyItsOwnLine(t *testing.T) {
	power, reset, _ := useButtons(t, "")
	if err := os.WriteFile(reset, []byte("untouched"), 0o600); err != nil {
		t.Fatal(err)
	}

	start := time.Now()
	if err := PressButton(ButtonPower, 50*time.Millisecond); err != nil {
		t.Fatalf("PressButton failed: %s", err)
	}

	if elapsed := time.Since(start); elapsed < 50*time.Millisecond {
		t.Fatalf("press took %s, want at least 50ms", elapsed)
	}
	if got := readLine(t, power); got != "0" {
		t.Fatalf("power line left at %q, want %q", got, "0")
	}
	if got := readLine(t, reset); got != "untouched" {
		t.Fatalf("reset line was written: %q", got)
	}
}

func TestPressButtonRefusesAnUnknownButton(t *testing.T) {
	power, reset, _ := useButtons(t, "")

	err := PressButton("hdd", 10*time.Millisecond)
	if !errors.Is(err, errUnknownButton) {
		t.Fatalf("PressButton(hdd) = %v, want errUnknownButton", err)
	}
	if readLine(t, power) != "0" || readLine(t, reset) != "0" {
		t.Fatal("a refused press wrote a line")
	}
}

func TestPressButtonRefusesADurationOutOfRange(t *testing.T) {
	useButtons(t, "")

	for _, d := range []time.Duration{0, -time.Second, maxPressDuration + time.Millisecond} {
		if err := PressButton(ButtonReset, d); !errors.Is(err, errPressOutOfRange) {
			t.Fatalf("PressButton(reset, %s) = %v, want errPressOutOfRange", d, err)
		}
	}
}

// A reset must not land while power is still held, so presses of different
// buttons wait for each other here, unlike bare writeGpio calls.
func TestPressButtonSerializesPowerAndReset(t *testing.T) {
	useButtons(t, "")

	const press = 150 * time.Millisecond

	start := time.Now()

	var wg sync.WaitGroup
	for _, kind := range []string{ButtonPower, ButtonReset} {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if err := PressButton(kind, press); err != nil {
				t.Errorf("PressButton(%s) failed: %s", kind, err)
			}
		}()
	}
	wg.Wait()

	if elapsed := time.Since(start); elapsed < 2*press {
		t.Fatalf("two presses took %s, want at least %s: they overlapped", elapsed, 2*press)
	}
}

// The LED line is active low.
func TestPowerLEDReadsTheLine(t *testing.T) {
	_, _, led := useButtons(t, "0\n")

	on, err := PowerLED()
	if err != nil || !on {
		t.Fatalf("PowerLED() with the line at 0 = %v, %v, want true, nil", on, err)
	}

	if err := os.WriteFile(led, []byte("1\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	on, err = PowerLED()
	if err != nil || on {
		t.Fatalf("PowerLED() with the line at 1 = %v, %v, want false, nil", on, err)
	}
}

func TestPowerLEDFailsWithoutTheLine(t *testing.T) {
	useButtons(t, "")

	if _, err := PowerLED(); err == nil {
		t.Fatal("PowerLED() with no LED line succeeded")
	}
}

// The UI's press goes through the same lock as Redfish's: while the lock is
// held, SetGpio waits.
func TestSetGpioWaitsForTheButtonLock(t *testing.T) {
	power, _, _ := useButtons(t, "")
	gin.SetMode(gin.TestMode)

	r := gin.New()
	r.POST("/api/vm/gpio", NewService().SetGpio)

	body, _ := json.Marshal(map[string]any{"type": "power", "duration": 20})
	request := httptest.NewRequest(http.MethodPost, "/api/vm/gpio", bytes.NewReader(body))
	request.Header.Set("Content-Type", "application/json")

	buttonMu.Lock()
	done := make(chan *httptest.ResponseRecorder)
	go func() {
		w := httptest.NewRecorder()
		r.ServeHTTP(w, request)
		done <- w
	}()

	select {
	case <-done:
		buttonMu.Unlock()
		t.Fatal("SetGpio pressed while another press held the lock")
	case <-time.After(100 * time.Millisecond):
	}

	buttonMu.Unlock()
	w := <-done
	if w.Code != http.StatusOK || !bytes.Contains(w.Body.Bytes(), []byte(`"code":0`)) {
		t.Fatalf("SetGpio answered %d %s", w.Code, w.Body.String())
	}
	if got := readLine(t, power); got != "0" {
		t.Fatalf("power line left at %q, want %q", got, "0")
	}
}

func TestFirmwareVersionNamesTheApplicationAndTheImage(t *testing.T) {
	useVersionFile(t, "2.3.0\n")
	useImageFile(t, "2026-06-10-1_4_3.img\n")
	original := ownImageFile
	ownImageFile = filepath.Join(t.TempDir(), "missing")
	t.Cleanup(func() { ownImageFile = original })

	if got, want := FirmwareVersion(), "2.3.0 (image v1.4.3)"; got != want {
		t.Fatalf("FirmwareVersion() = %q, want %q", got, want)
	}
}

func TestFirmwareVersionWithoutAnImageIsTheApplication(t *testing.T) {
	useVersionFile(t, "2.3.0\n")
	original := imageVersionFile
	imageVersionFile = filepath.Join(t.TempDir(), "missing")
	t.Cleanup(func() { imageVersionFile = original })

	if got, want := FirmwareVersion(), "2.3.0"; got != want {
		t.Fatalf("FirmwareVersion() = %q, want %q", got, want)
	}
}
