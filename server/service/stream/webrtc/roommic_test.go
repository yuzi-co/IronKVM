package webrtc

import (
	"encoding/json"
	"path/filepath"
	"slices"
	"strings"
	"sync"
	"testing"

	"NanoKVM-Server/authn"
	"NanoKVM-Server/service/roommic"
	"NanoKVM-Server/service/stream/audio"
)

// roomCapture stands in for arecord.
type roomCapture struct {
	frames chan []byte
	once   sync.Once
}

func (c *roomCapture) Start()                {}
func (c *roomCapture) Frames() <-chan []byte { return c.frames }
func (c *roomCapture) Stop()                 { c.once.Do(func() { close(c.frames) }) }

// withRoomListeners swaps in a manager with a card and no arecord, allows the
// microphone and opens it for each user.
func withRoomListeners(t *testing.T, users ...string) {
	t.Helper()

	originalFile := roommic.SettingsFile
	roommic.SettingsFile = filepath.Join(t.TempDir(), "room-mic")
	originalManager := roomMics

	roomMics = roommic.NewManager(func() bool { return true }, func() audio.Capture {
		return &roomCapture{frames: make(chan []byte)}
	})
	t.Cleanup(func() {
		roomMics.CloseAll("the test ended")
		roomMics = originalManager
		roommic.SettingsFile = originalFile
	})

	if err := roomMics.SetSettings(roommic.Settings{Allowed: true, Gain: roommic.DefaultGain}, "admin"); err != nil {
		t.Fatal(err)
	}
	for _, user := range users {
		if _, err := roomMics.Open(user, "test"); err != nil {
			t.Fatal(err)
		}
	}
}

func roomStateFor(role authn.Role) (roomMicState, string) {
	c := NewClient(nil, nil)
	c.role = role

	c.roomMutex.Lock()
	state := c.roomStateLocked()
	c.roomMutex.Unlock()

	data, _ := json.Marshal(state)
	return state, string(data)
}

// The state pushed over the signalling socket names the listeners to an
// administrator only. Everyone else gets the live flag and the count.
func TestThePushedRoomStateNamesListenersToAnAdministratorOnly(t *testing.T) {
	withRoomListeners(t, "alice", "bob")

	admin, _ := roomStateFor(authn.RoleAdmin)
	if !admin.Live || admin.ListenerCount != 2 || !slices.Equal(admin.Listeners, []string{"alice", "bob"}) {
		t.Fatalf("the administrator was told %+v", admin)
	}

	for _, role := range []authn.Role{authn.RoleUser, authn.Role("")} {
		state, message := roomStateFor(role)
		if !state.Live || state.ListenerCount != 2 || !state.Allowed {
			t.Fatalf("role %q lost the live flag or the count: %+v", role, state)
		}
		if state.Listeners != nil || strings.Contains(message, "alice") || strings.Contains(message, "bob") {
			t.Fatalf("role %q was told who listens: %s", role, message)
		}
	}
}
