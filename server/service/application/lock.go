package application

// AcquireUpdateLock takes the lock an application update holds, for another
// installer that writes under AppDir: a UI install must not run while an
// application update moves the directory it writes to. It reports false when
// an update is running.
func AcquireUpdateLock() bool { return acquireUpdateLock() }

// ReleaseUpdateLock releases what AcquireUpdateLock took.
func ReleaseUpdateLock() { releaseUpdateLock() }
