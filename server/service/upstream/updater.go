package upstream

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"

	"NanoKVM-Server/proto"
)

// updateTimeout bounds a whole update, the downloads included.
const updateTimeout = 20 * time.Minute

// Component is what an Updater needs to know about the thing it updates.
type Component struct {
	// Name is what the logs call it, such as "netboot.xyz boot files".
	Name string
	// Repo is the GitHub repository, as "owner/name".
	Repo string
	// Pinned is the release the server installs by default.
	Pinned string
	// ChecksumFile names the checksum file in a release of version v.
	ChecksumFile func(v string) string
	// Assets names the files an update to version v downloads; each must
	// be in the release and in its checksum file.
	Assets func(v string) []string
	// Installed returns the version in place, and false when the component
	// is not installed, which offers no update.
	Installed func() (string, bool)
	// InUse returns why the component cannot be replaced now, or nil.
	InUse func() error
	// Install puts the release in place. It is given the verified sums of
	// the release's assets, and reports its progress as a percentage.
	Install func(ctx context.Context, rel Release, sums Checksums, progress func(int)) error
}

// Job states.
const (
	JobIdle    = "idle"
	JobRunning = "running"
	JobDone    = "done"
	JobFailed  = "failed"
)

// Job is the last update the owner asked for.
type Job struct {
	State    string `json:"state"`
	Version  string `json:"version"`
	Progress int    `json:"progress"`
	Error    string `json:"error"`
}

// Status is what a page shows about a component's updates.
type Status struct {
	Installed bool   `json:"installed"`
	Version   string `json:"version"`
	Pinned    string `json:"pinned"`
	// Latest is the newest release the last check found, or "".
	Latest     string     `json:"latest"`
	CheckedAt  *time.Time `json:"checkedAt"`
	CheckError string     `json:"checkError"`
	// UpdateAvailable is true when Latest is newer than Version and can
	// be verified.
	UpdateAvailable bool `json:"updateAvailable"`
	// Unverifiable says why a newer release is not offered, such as a
	// release with no checksum file.
	Unverifiable string `json:"unverifiable"`
	// InUse says why an update would be refused now.
	InUse string `json:"inUse"`
	Job   Job    `json:"job"`
}

// Updater checks for and runs the updates of one component.
type Updater struct {
	c      Component
	client *Client

	mu  sync.Mutex
	job Job
}

func NewUpdater(c Component, client *Client) *Updater {
	if client == nil {
		client = Default
	}
	return &Updater{c: c, client: client, job: Job{State: JobIdle}}
}

// Status reports the component's updates from the last check, without
// asking GitHub.
func (u *Updater) Status() Status {
	st := Status{Pinned: u.c.Pinned}
	st.Version, st.Installed = u.c.Installed()
	if err := u.c.InUse(); err != nil {
		st.InUse = err.Error()
	}

	if rel, at, err, ok := u.client.Cached(u.c.Repo); ok {
		st.CheckedAt = &at
		if err != nil {
			st.CheckError = err.Error()
		} else {
			st.Latest = rel.Version
			if st.Installed && CompareVersions(rel.Version, st.Version) > 0 {
				if err := u.verifiable(rel); err != nil {
					st.Unverifiable = err.Error()
				} else {
					st.UpdateAvailable = true
				}
			}
		}
	}

	u.mu.Lock()
	st.Job = u.job
	u.mu.Unlock()
	return st
}

// verifiable says why a release cannot be checked, or nil.
func (u *Updater) verifiable(rel Release) error {
	sumsFile := u.c.ChecksumFile(rel.Version)
	if _, ok := rel.Assets[sumsFile]; !ok {
		return fmt.Errorf("release %s publishes no checksum file (%s), so it cannot be verified", rel.Version, sumsFile)
	}
	for _, name := range u.c.Assets(rel.Version) {
		if _, ok := rel.Assets[name]; !ok {
			return fmt.Errorf("release %s has no %s", rel.Version, name)
		}
	}
	return nil
}

// Check asks GitHub for the latest release, from the cache unless force.
func (u *Updater) Check(ctx context.Context, force bool) Status {
	if _, err := u.client.Latest(ctx, u.c.Repo, force); err != nil {
		log.Warnf("%s: check for updates: %s", u.c.Name, err)
	}
	return u.Status()
}

var (
	ErrRunning      = errors.New("an update is already running")
	ErrNotInstalled = errors.New("install it first: only an installed component is updated")
	ErrUpToDate     = errors.New("already the latest release")
)

// Start begins the update to the latest release in the background. It
// refuses at once what it can tell will fail: another update, a component
// that is not installed or in use, and a release that cannot be verified.
func (u *Updater) Start() error {
	u.mu.Lock()
	if u.job.State == JobRunning {
		u.mu.Unlock()
		return ErrRunning
	}
	u.mu.Unlock()

	current, ok := u.c.Installed()
	if !ok {
		return ErrNotInstalled
	}
	if err := u.c.InUse(); err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(context.Background(), updateTimeout)
	rel, err := u.client.Latest(ctx, u.c.Repo, false)
	if err != nil {
		cancel()
		return err
	}
	if CompareVersions(rel.Version, current) <= 0 {
		cancel()
		return fmt.Errorf("%w (%s)", ErrUpToDate, current)
	}
	if err := u.verifiable(rel); err != nil {
		cancel()
		return err
	}

	u.mu.Lock()
	if u.job.State == JobRunning {
		u.mu.Unlock()
		cancel()
		return ErrRunning
	}
	u.job = Job{State: JobRunning, Version: rel.Version}
	u.mu.Unlock()

	go func() {
		defer cancel()
		err := u.run(ctx, rel)

		u.mu.Lock()
		defer u.mu.Unlock()
		if err != nil {
			log.Errorf("%s: update to %s: %s", u.c.Name, rel.Version, err)
			u.job = Job{State: JobFailed, Version: rel.Version, Error: err.Error()}
			return
		}
		log.Infof("%s: updated to %s", u.c.Name, rel.Version)
		u.job = Job{State: JobDone, Version: rel.Version, Progress: 100}
	}()
	return nil
}

func (u *Updater) run(ctx context.Context, rel Release) error {
	sums, err := u.client.FetchChecksums(ctx, rel.Assets[u.c.ChecksumFile(rel.Version)])
	if err != nil {
		return fmt.Errorf("read the checksum file: %w", err)
	}
	verified := Checksums{}
	for _, name := range u.c.Assets(rel.Version) {
		sum, ok := sums[name]
		if !ok {
			return fmt.Errorf("the checksum file of %s lists no %s", rel.Version, name)
		}
		verified[name] = sum
	}
	return u.c.Install(ctx, rel, verified, func(p int) {
		u.mu.Lock()
		if u.job.State == JobRunning {
			u.job.Progress = min(max(p, 0), 99)
		}
		u.mu.Unlock()
	})
}

// Download fetches one of the release's assets, checked against its sum,
// and reports progress as a share of the whole update: from is where this
// file's progress starts, and span how much of it this file takes.
func (u *Updater) Download(ctx context.Context, rel Release, sums Checksums, name string, limit int64, dst string, progress func(int), from, span int) error {
	return u.client.Download(ctx, rel.Assets[name], sums[name], limit, dst, func(done, total int64) {
		if progress != nil && total > 0 {
			progress(from + int(done*int64(span)/total))
		}
	})
}

// The handlers answer with the web UI's envelope: -2 for an update refused
// because of the board's state, -3 for one that failed.

type checkRequest struct {
	Force bool `json:"force" form:"force"`
}

func (u *Updater) GetStatus(c *gin.Context) {
	var rsp proto.Response
	rsp.OkRspWithData(c, u.Status())
}

func (u *Updater) CheckHandler(c *gin.Context) {
	var rsp proto.Response
	var req checkRequest
	_ = c.ShouldBind(&req)

	st := u.Check(c.Request.Context(), req.Force)
	rsp.OkRspWithData(c, st)
}

func (u *Updater) UpdateHandler(c *gin.Context) {
	var rsp proto.Response

	if err := u.Start(); err != nil {
		code := -3
		if errors.Is(err, ErrRunning) || errors.Is(err, ErrNotInstalled) || errors.Is(err, ErrUpToDate) || u.c.InUse() != nil {
			code = -2
		}
		rsp.ErrRsp(c, code, err.Error())
		return
	}
	rsp.OkRspWithData(c, u.Status())
}

// Routes registers GET <path>, POST <path>/check and POST <path> on group.
func (u *Updater) Routes(group gin.IRoutes, path string) {
	group.GET(path, u.GetStatus)
	group.POST(path+"/check", u.CheckHandler)
	group.POST(path, u.UpdateHandler)
}
