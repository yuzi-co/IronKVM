package audit

import (
	"fmt"
	"strings"

	"github.com/gin-gonic/gin"

	"NanoKVM-Server/proto"
)

// TiersFile is a variable so tests can point it at a directory.
var TiersFile = "/etc/kvm/agent-tiers.json"

// Tier is a class of what an MCP client or the on-board agent may do.
type Tier string

const (
	// TierObserve reads: screenshots and status.
	TierObserve Tier = "observe"
	// TierInput sends keyboard and mouse input to the host.
	TierInput Tier = "input"
	// TierPower presses the ATX power and reset buttons.
	TierPower Tier = "power"
	// TierMedia inserts and ejects virtual media.
	TierMedia Tier = "media"
	// TierAdmin changes the board's own settings. No tool needs it yet; the
	// tier is there so that one added later is off until an owner opts in.
	TierAdmin Tier = "admin"
)

// AllTiers lists the tiers in the order the UI shows them.
var AllTiers = []Tier{TierObserve, TierInput, TierPower, TierMedia, TierAdmin}

// Subjects of the tiers. They are the two ways an AI reaches the board: an
// external client through /api/mcp, and the agent that runs on the board.
const (
	SubjectMCP   = "mcp"
	SubjectAgent = "agent"
)

// Tiers is the set of tiers each subject has.
type Tiers struct {
	MCP   []Tier `json:"mcp"`
	Agent []Tier `json:"agent"`
}

// DefaultTiers is what a board that was never configured allows: the
// screenshots and the keyboard and mouse that MCP and the agent could always
// use. Power, media and admin are off until an owner turns them on.
func DefaultTiers() Tiers {
	return Tiers{MCP: []Tier{TierObserve, TierInput}, Agent: []Tier{TierObserve, TierInput}}
}

var (
	tiers       Tiers
	tiersLoaded bool
)

// CurrentTiers returns the tiers in force.
func CurrentTiers() Tiers {
	fileMu.Lock()
	defer fileMu.Unlock()
	if !tiersLoaded {
		tiers = DefaultTiers()
		var stored struct {
			MCP   *[]Tier `json:"mcp"`
			Agent *[]Tier `json:"agent"`
		}
		if readJSON(TiersFile, &stored) == nil {
			if stored.MCP != nil {
				tiers.MCP = normalizeTiers(*stored.MCP)
			}
			if stored.Agent != nil {
				tiers.Agent = normalizeTiers(*stored.Agent)
			}
		}
		tiersLoaded = true
	}
	return Tiers{MCP: append([]Tier(nil), tiers.MCP...), Agent: append([]Tier(nil), tiers.Agent...)}
}

// SaveTiers validates and stores new tiers.
func SaveTiers(t Tiers) (Tiers, error) {
	for _, list := range [][]Tier{t.MCP, t.Agent} {
		for _, tier := range list {
			if !validTier(tier) {
				return Tiers{}, fmt.Errorf("unknown tier %q", tier)
			}
		}
	}
	t = Tiers{MCP: normalizeTiers(t.MCP), Agent: normalizeTiers(t.Agent)}

	fileMu.Lock()
	defer fileMu.Unlock()
	if err := writeJSON(TiersFile, t); err != nil {
		return Tiers{}, err
	}
	tiers, tiersLoaded = t, true
	return Tiers{MCP: append([]Tier(nil), t.MCP...), Agent: append([]Tier(nil), t.Agent...)}, nil
}

// Allowed reports whether the subject has the tier.
func Allowed(subject string, tier Tier) bool {
	t := CurrentTiers()
	list := t.MCP
	if subject == SubjectAgent {
		list = t.Agent
	}
	for _, have := range list {
		if have == tier {
			return true
		}
	}
	return false
}

// Reset forgets what was read, so the next call reads the files again. Tests
// use it after they point the files elsewhere.
func Reset() {
	fileMu.Lock()
	defer fileMu.Unlock()
	loaded, tiersLoaded = false, false
	settings = Settings{}
}

func validTier(t Tier) bool {
	for _, known := range AllTiers {
		if known == t {
			return true
		}
	}
	return false
}

// normalizeTiers drops unknown and repeated names and puts the rest in the
// order of AllTiers. A name a newer slot wrote and this one does not know is
// ignored rather than failing the read.
func normalizeTiers(in []Tier) []Tier {
	out := make([]Tier, 0, len(in))
	for _, tier := range AllTiers {
		for _, have := range in {
			if have == tier {
				out = append(out, tier)
				break
			}
		}
	}
	return out
}

// ResetForTest forgets what was read, so the next call reads the files again.
func ResetForTest() {
	fileMu.Lock()
	defer fileMu.Unlock()
	tiersLoaded, loaded = false, false
	settings = Settings{}
}

type tiersRsp struct {
	Tiers
	// Available lists the tiers there are, in the order to show them.
	Available []Tier `json:"available"`
}

func currentTiersRsp() tiersRsp {
	return tiersRsp{Tiers: CurrentTiers(), Available: AllTiers}
}

// GetTiers answers GET /api/agent/tiers.
func (s *Service) GetTiers(c *gin.Context) {
	var rsp proto.Response
	rsp.OkRspWithData(c, currentTiersRsp())
}

// SetTiers answers POST /api/agent/tiers.
func (s *Service) SetTiers(c *gin.Context) {
	var req struct {
		MCP   *[]Tier `json:"mcp"`
		Agent *[]Tier `json:"agent"`
	}
	var rsp proto.Response

	if err := c.ShouldBindJSON(&req); err != nil || req.MCP == nil || req.Agent == nil {
		rsp.ErrRsp(c, -1, "invalid arguments")
		return
	}
	saved, err := SaveTiers(Tiers{MCP: *req.MCP, Agent: *req.Agent})
	if err != nil {
		rsp.ErrRsp(c, -1, err.Error())
		return
	}
	SetDetail(c, fmt.Sprintf("mcp=%s agent=%s", joinTiers(saved.MCP), joinTiers(saved.Agent)))
	rsp.OkRspWithData(c, currentTiersRsp())
}

func joinTiers(list []Tier) string {
	if len(list) == 0 {
		return "none"
	}
	names := make([]string, len(list))
	for i, t := range list {
		names[i] = string(t)
	}
	return strings.Join(names, ",")
}
