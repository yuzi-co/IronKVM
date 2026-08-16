package application

const (
	// StableURL is the IronKVM feed. It serves latest.json alone: the manifest
	// names the package on GitHub Releases, because a release asset lives under
	// a per-tag path that no fixed base URL can reach.
	StableURL = "https://yuzi-co.github.io/IronKVM"

	// SipeedURL is the official feed, kept so that leaving this firmware stays
	// one click rather than a research project. The shallow rename exists for
	// the same reason.
	SipeedURL = "https://cdn.sipeed.com/nanokvm"

	PreviewURL = "https://cdn.sipeed.com/nanokvm/preview"

	CacheDir = "/root/.kvmcache"

	updateWorkspacePrefix = "nanokvm-update-"
	cacheDirMode          = 0o700
	maxPackageSize        = uint64(1 << 30)
	maxExpandedSize       = uint64(2 << 30)
	maxArchiveEntries     = 100_000
	minFreeReserve        = uint64(128 << 20)
	freeReservePercent    = uint64(5)
)

// AppDir and BackupDir are variables rather than constants so a test can point
// them at a scratch tree. Installing a package moves whole directories about,
// and a test that could only run against the real /kvmapp would not be run.
// This package already treats versionFile the same way.
var (
	AppDir    = "/kvmapp"
	BackupDir = "/root/old"
)

type Service struct{}

func NewService() *Service {
	return &Service{}
}
