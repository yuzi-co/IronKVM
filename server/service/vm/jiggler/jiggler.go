package jiggler

import (
	"context"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	"NanoKVM-Server/service/controlmode"
	"NanoKVM-Server/service/hid"
	"NanoKVM-Server/service/inputcontrol"
)

// ConfigFile records whether the jiggler is on. Its first line is the mouse
// mode; a second line, when present, names the key the jiggler presses instead
// of moving the mouse. Files written before keys existed hold only the mode and
// read as the mouse method. A variable so tests can point it somewhere writable.
var ConfigFile = "/etc/kvm/mouse-jiggler"

// DefaultInterval is how long the target may sit idle before the jiggler moves
// the mouse or presses its key.
const DefaultInterval = 15 * time.Second

var (
	jiggler Jiggler
	once    sync.Once
)

type Jiggler struct {
	mutex   sync.Mutex
	enabled bool
	running bool
	// closed is set by Shutdown, after which nothing starts the loop again.
	closed bool
	mode   string
	// key is the key the jiggler presses, one of the Key constants, or empty
	// to move the mouse.
	key         string
	lastUpdated time.Time

	// actionMutex is held from the moment the loop decides to act until the
	// key or the pointer is back where it started, so Shutdown can wait for a
	// press to be released before the process exits.
	actionMutex sync.Mutex

	// interval, move and press belong to the instance rather than the package
	// so a test can drive the loop quickly and without a HID gadget behind it.
	// All are set when the value is built and never written again, so the loop
	// reads them without the lock.
	interval time.Duration
	move     func(string)
	press    func(string)
}

func GetJiggler() *Jiggler {
	once.Do(func() {
		jiggler = Jiggler{
			mutex:       sync.Mutex{},
			enabled:     false,
			running:     false,
			mode:        "relative",
			lastUpdated: time.Now(),
			interval:    DefaultInterval,
			move:        move,
			press:       press,
		}

		content, err := os.ReadFile(ConfigFile)
		if err != nil {
			return
		}

		mode, key := parseConfig(string(content))
		if mode != "" {
			jiggler.mode = mode
		}
		jiggler.key = key

		jiggler.enabled = true
	})

	return &jiggler
}

// parseConfig reads the config file's contents: the mode on the first line and
// the key, if any, on the second. A key this build does not know reads as the
// mouse method, so a damaged file still keeps the host awake.
func parseConfig(content string) (mode string, key string) {
	lines := strings.Split(strings.ReplaceAll(content, "\r", ""), "\n")
	mode = strings.TrimSpace(lines[0])
	if len(lines) > 1 {
		key = strings.TrimSpace(lines[1])
		if !ValidKey(key) {
			key = ""
		}
	}
	return mode, key
}

// formatConfig is parseConfig's inverse. The mouse method writes the mode
// alone, the format older builds wrote, so going back to one keeps working.
func formatConfig(mode string, key string) string {
	if key == "" {
		return mode
	}
	return mode + "\n" + key
}

// Enable turns the jiggler on and remembers the choice across a restart. The
// file is written first, so a jiggler that reports itself on is one the next
// boot also finds on. key is one of the Key constants, or empty to move the
// mouse.
func (j *Jiggler) Enable(mode string, key string) error {
	if key != "" && !ValidKey(key) {
		return fmt.Errorf("unknown jiggler key %q", key)
	}

	if err := os.WriteFile(ConfigFile, []byte(formatConfig(mode, key)), 0644); err != nil {
		return err
	}

	j.mutex.Lock()
	j.enabled = true
	j.mode = mode
	j.key = key
	j.mutex.Unlock()

	// Outside the lock, because Run takes it.
	j.Run()

	return nil
}

// Disable turns the jiggler off. The key is kept, so the method shown while the
// jiggler is off is the one it resumes with. Disabling a jiggler that is
// already off is not an error.
func (j *Jiggler) Disable() error {
	if err := os.Remove(ConfigFile); err != nil && !os.IsNotExist(err) {
		return err
	}

	j.mutex.Lock()
	defer j.mutex.Unlock()

	j.enabled = false
	j.mode = "relative"

	return nil
}

// SetKey records the key to press without turning the jiggler on, so the
// method can be chosen while it is off. It lives in memory until the next
// Enable writes it to the file.
func (j *Jiggler) SetKey(key string) error {
	if key != "" && !ValidKey(key) {
		return fmt.Errorf("unknown jiggler key %q", key)
	}

	j.mutex.Lock()
	defer j.mutex.Unlock()

	j.key = key
	return nil
}

// Shutdown stops the loop for good and waits for a press or move in flight to
// finish, so the process never exits between a key's press and its release.
// The config file is left alone: the jiggler is still on at the next boot.
func (j *Jiggler) Shutdown() {
	j.mutex.Lock()
	j.enabled = false
	j.closed = true
	j.mutex.Unlock()

	j.actionMutex.Lock()
	defer j.actionMutex.Unlock()
}

// stop asks the loop to finish. It returns before the loop has noticed.
func (j *Jiggler) stop() {
	j.mutex.Lock()
	defer j.mutex.Unlock()

	j.enabled = false
}

// Run starts the loop and reports whether this call is the one that started it.
// Boot calls it once and every Enable calls it again, so it has to be safe to
// call at any time from anywhere.
//
// The check and the claim happen under one lock. Reading the flag and then
// setting it as two steps let two callers both find it clear, and two loops
// move the mouse twice as often as asked.
func (j *Jiggler) Run() bool {
	j.mutex.Lock()
	if !j.enabled || j.running || j.closed {
		j.mutex.Unlock()
		return false
	}

	j.running = true
	j.lastUpdated = time.Now()
	j.mutex.Unlock()

	go j.loop()

	return true
}

// loop moves the mouse or presses the key whenever the target has been idle
// for the interval.
func (j *Jiggler) loop() {
	ticker := time.NewTicker(j.interval)
	defer ticker.Stop()

	for range ticker.C {
		if !j.step() {
			return
		}
	}
}

// step runs one tick and reports whether the loop should keep going. The
// decision and the action share actionMutex, so a Shutdown that lands between
// them waits for the action instead of returning while a key is down.
func (j *Jiggler) step() bool {
	j.actionMutex.Lock()
	defer j.actionMutex.Unlock()

	mode, key, due, running := j.tick()
	if due {
		if key != "" {
			j.press(key)
		} else {
			j.move(mode)
		}
	}
	return running
}

// tick decides what the loop does next: the mode to move in or the key to
// press, whether an action is due, and whether the loop should keep running.
//
// The decision and the state it reads live together under the lock, because the
// websocket read loop writes lastUpdated once per HID event. A move counts as
// activity, so tick records it here rather than making the loop reacquire the
// lock to say so.
func (j *Jiggler) tick() (mode string, key string, due bool, running bool) {
	j.mutex.Lock()
	defer j.mutex.Unlock()

	if !j.enabled {
		j.running = false
		return "", "", false, false
	}

	if time.Since(j.lastUpdated) <= j.interval {
		return "", "", false, true
	}

	j.lastUpdated = time.Now()

	return j.mode, j.key, true, true
}

// Update records that the target saw real input, which holds the jiggler off.
// It runs once per HID event on the websocket read loop.
func (j *Jiggler) Update() {
	j.mutex.Lock()
	defer j.mutex.Unlock()

	if j.running {
		j.lastUpdated = time.Now()
	}
}

func (j *Jiggler) IsEnabled() bool {
	j.mutex.Lock()
	defer j.mutex.Unlock()

	return j.enabled
}

func (j *Jiggler) GetMode() string {
	j.mutex.Lock()
	defer j.mutex.Unlock()

	return j.mode
}

// GetKey returns the key the jiggler presses, or empty when it moves the mouse.
func (j *Jiggler) GetKey() string {
	j.mutex.Lock()
	defer j.mutex.Unlock()

	return j.key
}

func move(mode string) {
	_, releaseMode, err := controlmode.GetManager().AcquireStable()
	if err != nil {
		return
	}
	defer releaseMode()

	ctx, release, err := inputcontrol.GetCoordinator().BeginBackground(context.Background())
	if err != nil {
		return
	}
	defer release()

	h := hid.GetHid()

	if mode == "absolute" {
		if err := h.WriteAbsoluteMouseReport([]byte{0x00, 0x00, 0x3f, 0x00, 0x3f, 0x00}); err != nil {
			return
		}
		defer func() {
			_ = h.WriteAbsoluteMouseReport([]byte{0x00, 0xff, 0x3f, 0xff, 0x3f, 0x00})
		}()
		_ = waitMove(ctx, 100*time.Millisecond)
	} else {
		if err := h.WriteRelativeMouseReport([]byte{0x00, 0xa, 0xa, 0x00, 0x00}); err != nil {
			return
		}
		defer func() {
			_ = h.WriteRelativeMouseReport([]byte{0x00, 0xf6, 0xf6, 0x00, 0x00})
		}()
		_ = waitMove(ctx, 100*time.Millisecond)
	}
}

func waitMove(ctx context.Context, delay time.Duration) bool {
	timer := time.NewTimer(delay)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return false
	case <-timer.C:
		return true
	}
}
