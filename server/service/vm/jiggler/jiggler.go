package jiggler

import (
	"context"
	"os"
	"strings"
	"sync"
	"time"

	"NanoKVM-Server/service/controlmode"
	"NanoKVM-Server/service/hid"
	"NanoKVM-Server/service/inputcontrol"
)

// ConfigFile records whether the mouse jiggler is on, and its contents are the
// mode. A variable so tests can point it somewhere writable.
var ConfigFile = "/etc/kvm/mouse-jiggler"

// DefaultInterval is how long the target may sit idle before the jiggler moves
// the mouse or presses its key.
const DefaultInterval = 15 * time.Second

var (
	jiggler Jiggler
	once    sync.Once
)

// Jiggler runs two independent keep-awake actions on one idle schedule: the
// mouse jiggler moves the pointer and back, the key jiggler presses and
// releases a key the host ignores. Either, both or neither can be on. They
// share the loop and the idle clock, so real input holds both off, and a tick
// that finds the target idle runs every action that is on.
type Jiggler struct {
	mutex sync.Mutex
	// enabled and mode are the mouse jiggler.
	enabled bool
	mode    string
	// keyEnabled and key are the key jiggler. key is one of the Key constants
	// and is kept while the key jiggler is off.
	keyEnabled bool
	key        string
	running    bool
	// closed is set by Shutdown, after which nothing starts the loop again.
	closed      bool
	lastUpdated time.Time

	// actionMutex is held from the moment the loop decides to act until the
	// key and the pointer are back where they started, so Shutdown can wait
	// for a press to be released before the process exits.
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
			key:         DefaultKey,
			lastUpdated: time.Now(),
			interval:    DefaultInterval,
			move:        move,
			press:       press,
		}

		jiggler.keyEnabled, jiggler.key = readKeyConfig()

		content, err := os.ReadFile(ConfigFile)
		if err != nil {
			return
		}

		mode := strings.ReplaceAll(string(content), "\n", "")
		if mode != "" {
			jiggler.mode = mode
		}

		jiggler.enabled = true
	})

	return &jiggler
}

// Enable turns the mouse jiggler on and remembers the choice across a restart.
// The file is written first, so a jiggler that reports itself on is one the
// next boot also finds on.
func (j *Jiggler) Enable(mode string) error {
	if err := os.WriteFile(ConfigFile, []byte(mode), 0644); err != nil {
		return err
	}

	j.mutex.Lock()
	j.enabled = true
	j.mode = mode
	j.mutex.Unlock()

	// Outside the lock, because Run takes it.
	j.Run()

	return nil
}

func (j *Jiggler) Disable() error {
	if err := os.Remove(ConfigFile); err != nil {
		return err
	}

	j.mutex.Lock()
	defer j.mutex.Unlock()

	j.enabled = false
	j.mode = "relative"

	return nil
}

// Shutdown stops the loop for good and waits for a press or move in flight to
// finish, so the process never exits between a key's press and its release.
// The config files are left alone: the jigglers are still on at the next boot.
func (j *Jiggler) Shutdown() {
	j.mutex.Lock()
	j.enabled = false
	j.keyEnabled = false
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
	j.keyEnabled = false
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
	if !(j.enabled || j.keyEnabled) || j.running || j.closed {
		j.mutex.Unlock()
		return false
	}

	j.running = true
	j.lastUpdated = time.Now()
	j.mutex.Unlock()

	go j.loop()

	return true
}

// loop moves the mouse and presses the key, whichever are on, whenever the
// target has been idle for the interval.
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
// decision and the actions share actionMutex, so a Shutdown that lands between
// them waits for the actions instead of returning while a key is down.
func (j *Jiggler) step() bool {
	j.actionMutex.Lock()
	defer j.actionMutex.Unlock()

	a, running := j.tick()
	if a.key != "" {
		j.press(a.key)
	}
	if a.move {
		j.move(a.mode)
	}
	return running
}

// action is what one tick does: move the mouse in mode, press key, both or
// neither.
type action struct {
	move bool
	mode string
	key  string
}

// tick decides what the loop does next and whether it should keep running at
// all.
//
// The decision and the state it reads live together under the lock, because the
// websocket read loop writes lastUpdated once per HID event. A move or a press
// counts as activity, so tick records it here rather than making the loop
// reacquire the lock to say so.
func (j *Jiggler) tick() (a action, running bool) {
	j.mutex.Lock()
	defer j.mutex.Unlock()

	if !j.enabled && !j.keyEnabled {
		j.running = false
		return action{}, false
	}

	if time.Since(j.lastUpdated) <= j.interval {
		return action{}, true
	}

	j.lastUpdated = time.Now()

	a.move = j.enabled
	a.mode = j.mode
	if j.keyEnabled {
		a.key = j.key
	}
	return a, true
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
