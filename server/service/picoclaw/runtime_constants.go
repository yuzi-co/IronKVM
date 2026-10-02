package picoclaw

import (
	"time"
)

const (
	picoclawBinaryPath      = "/usr/bin/picoclaw"
	picoclawPIDFileName     = ".picoclaw.pid"
	etcInitPicoclawScript   = "/etc/init.d/S96picoclaw"
	kvmappPicoclawScript    = "/kvmapp/system/init.d/S96picoclaw"
	picoclawStartTimeout    = 15 * time.Second
	picoclawStopTimeout     = 15 * time.Second
	picoclawOnboardTimeout  = 60 * time.Second
	picoclawInstallTimeout  = 2 * time.Minute
	picoclawStartWaitPeriod = 1500 * time.Millisecond
	picoclawStopWaitPeriod  = 500 * time.Millisecond
	picoclawReadyPollPeriod = 500 * time.Millisecond
)

// The image carries PicoClaw: ironkvm-dist builds it from a pinned commit of
// github.com/yuzi-co/picoclaw and puts it here with its sha256. Installing
// copies this file onto /data and downloads nothing. A board whose image does
// not carry it cannot install PicoClaw. Variables so the tests can move them.
var (
	picoclawBundledBinary   = "/kvmapp/picoclaw/bin/picoclaw"
	picoclawBundledChecksum = "/kvmapp/picoclaw/bin/picoclaw.sha256"
)
