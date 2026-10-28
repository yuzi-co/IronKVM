package install

import (
	"archive/tar"
	"compress/gzip"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"sort"
	"strings"
	"time"

	"NanoKVM-Server/service/extensions/vpn"
)

// Variables rather than constants so the tests can stand in for the tools.
var (
	// XzcatPath decompresses an xz member. The image has xzcat.
	XzcatPath = "xzcat"
	// ApkPath checks and unpacks an Alpine package. The image has apk.
	ApkPath = "apk"
)

const toolTimeout = 10 * time.Minute

// ExtractTarGz takes the listed members out of a .tar.gz into dir, under
// their saved names, each checked against its own sum, and returns the sums,
// keyed by saved name. Licence files go to dir/licenses. Nothing else in the
// archive is written, so an archive's path never becomes a path on the board,
// and a member the archive lacks is an error.
func ExtractTarGz(archive string, members []Member, dir string) (map[string]string, error) {
	if err := validateMembers(members); err != nil {
		return nil, err
	}
	f, err := os.Open(archive)
	if err != nil {
		return nil, err
	}
	defer func() { _ = f.Close() }()
	gz, err := gzip.NewReader(f)
	if err != nil {
		return nil, err
	}
	defer func() { _ = gz.Close() }()

	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, err
	}
	// The names are the manifest's, not the archive's, but dir is on /data
	// and a symbolic link there must not carry a write elsewhere.
	root, err := os.OpenRoot(dir)
	if err != nil {
		return nil, err
	}
	defer func() { _ = root.Close() }()

	want := map[string]Member{}
	var licences []string
	for _, m := range members {
		if m.Licences {
			licences = append(licences, m.Path)
			continue
		}
		want[m.Path] = m
	}

	sums := map[string]string{}
	found := map[string]bool{}
	tr := tar.NewReader(gz)
	for {
		hdr, err := tr.Next()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return nil, err
		}
		if hdr.Typeflag != tar.TypeReg {
			continue
		}
		if m, ok := want[hdr.Name]; ok {
			if found[hdr.Name] {
				return nil, fmt.Errorf("the archive holds %s twice", hdr.Name)
			}
			sum, err := writeMember(root, tr, hdr.Size, m)
			if err != nil {
				return nil, err
			}
			sums[m.Name] = sum
			found[hdr.Name] = true
			continue
		}
		if name, ok := licenceName(licences, hdr.Name); ok {
			if hdr.Size > licenceMax {
				return nil, fmt.Errorf("licence file %s is %d bytes, larger than a licence", hdr.Name, hdr.Size)
			}
			if err := root.MkdirAll(LicencesDir, 0o755); err != nil {
				return nil, err
			}
			sum, err := writeMember(root, tr, hdr.Size, Member{Path: hdr.Name, Name: LicencesDir + "/" + name, Max: licenceMax})
			if err != nil {
				return nil, err
			}
			sums[LicencesDir+"/"+name] = sum
		}
	}

	var missing []string
	for p := range want {
		if !found[p] {
			missing = append(missing, p)
		}
	}
	if len(missing) > 0 {
		sort.Strings(missing)
		return nil, fmt.Errorf("the archive has no %s", strings.Join(missing, ", "))
	}
	return sums, nil
}

// licenceName is the file's name when path is a regular file directly in one
// of the licence directories.
func licenceName(dirs []string, path string) (string, bool) {
	for _, d := range dirs {
		if rest, ok := strings.CutPrefix(path, d); ok && plainName(rest) {
			return rest, true
		}
	}
	return "", false
}

// writeMember writes r, size bytes, to m.Name in root, decompressed when m is
// xz, and keeps it only when its sum is m's, if m has one. It returns the sum.
func writeMember(root *os.Root, r io.Reader, size int64, m Member) (string, error) {
	if !m.XZ && size > m.max() {
		return "", fmt.Errorf("%s is %d bytes, larger than the %d it may have", m.Path, size, m.max())
	}
	tmp := m.Name + ".part"
	out, err := root.OpenFile(tmp, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, m.mode())
	if err != nil {
		return "", err
	}
	defer func() { _ = root.Remove(tmp) }()

	hasher := sha256.New()
	w := io.MultiWriter(out, hasher)
	var copyErr error
	if m.XZ {
		ctx, cancel := context.WithTimeout(context.Background(), toolTimeout)
		defer cancel()
		cmd := exec.CommandContext(ctx, XzcatPath)
		cmd.Stdin = io.LimitReader(r, size)
		cmd.Stdout = &limitWriter{w: w, left: m.max()}
		var stderr limitBuffer
		cmd.Stderr = &stderr
		if err := cmd.Run(); err != nil {
			copyErr = fmt.Errorf("decompress %s: %w: %s", m.Path, err, stderr.String())
		}
	} else {
		_, copyErr = io.Copy(w, io.LimitReader(r, size))
	}
	closeErr := out.Close()
	if copyErr != nil {
		return "", copyErr
	}
	if closeErr != nil {
		return "", closeErr
	}
	got := hex.EncodeToString(hasher.Sum(nil))
	if m.SHA256 != "" && got != strings.ToLower(m.SHA256) {
		return "", fmt.Errorf("sha256 mismatch: %s is %s, want %s", m.Path, got, m.SHA256)
	}
	// exFAT keeps no mode, so the mode is chmod'ed where it can be and
	// recorded in the image by Pack in any case.
	_ = root.Chmod(tmp, m.mode())
	// A rename would replace a symbolic link rather than follow it, but one
	// where a member goes means somebody has been at /data: refuse it.
	if fi, err := root.Lstat(m.Name); err == nil && fi.Mode()&os.ModeSymlink != 0 {
		return "", fmt.Errorf("refusing to replace the symbolic link %s", m.Name)
	}
	if err := root.Rename(tmp, m.Name); err != nil {
		return "", err
	}
	return got, nil
}

// ExtractApk has apk check the package's signature, unpacks it into a
// scratch directory beside dir, and takes the listed members out of it into
// dir as ExtractTarGz does. apk extract checks the signature again and reads
// both of Alpine's package formats; --no-chown because /data is exFAT, which
// has no owners.
func ExtractApk(pkg string, members []Member, dir, apkPath string) (map[string]string, error) {
	if err := validateMembers(members); err != nil {
		return nil, err
	}
	if apkPath == "" {
		apkPath = ApkPath
	}
	if err := VerifyApk(pkg, apkPath); err != nil {
		return nil, err
	}
	unpacked := dir + ".apk"
	_ = os.RemoveAll(unpacked)
	if err := os.MkdirAll(unpacked, 0o755); err != nil {
		return nil, err
	}
	defer func() { _ = os.RemoveAll(unpacked) }()
	if _, err := runTool(apkPath, "extract", "--no-chown", "--destination", unpacked, pkg); err != nil {
		return nil, err
	}

	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, err
	}
	src, err := os.OpenRoot(unpacked)
	if err != nil {
		return nil, err
	}
	defer func() { _ = src.Close() }()
	dst, err := os.OpenRoot(dir)
	if err != nil {
		return nil, err
	}
	defer func() { _ = dst.Close() }()

	sums := map[string]string{}
	for _, m := range members {
		if m.Licences {
			continue
		}
		p := strings.TrimPrefix(m.Path, "./")
		fi, err := src.Lstat(p)
		if err != nil || !fi.Mode().IsRegular() {
			return nil, fmt.Errorf("the package has no %s", m.Path)
		}
		in, err := src.Open(p)
		if err != nil {
			return nil, err
		}
		sum, err := writeMember(dst, in, fi.Size(), m)
		_ = in.Close()
		if err != nil {
			return nil, err
		}
		sums[m.Name] = sum
	}
	return sums, nil
}

// VerifyApk has apk check Alpine's signature on a package or an index
// (APKINDEX.tar.gz), against the keys in /etc/apk/keys, which on the image
// are alpine-keys' riscv64 keys. apkPath is ApkPath when empty.
func VerifyApk(file, apkPath string) error {
	if apkPath == "" {
		apkPath = ApkPath
	}
	_, err := runTool(apkPath, "verify", file)
	return err
}

// runTool runs a tool under its own deadline. WaitDelay closes the pipes if
// the tool leaves a child holding them, so the deadline holds. A failure
// carries the output's last lines, which vpn.Message shows the owner.
func runTool(name string, args ...string) ([]byte, error) {
	ctx, cancel := context.WithTimeout(context.Background(), toolTimeout)
	defer cancel()
	cmd := exec.CommandContext(ctx, name, args...)
	cmd.WaitDelay = 2 * time.Second
	return vpn.Run(cmd)
}

// limitWriter fails a write past its limit, so a member that decompresses
// to more than it may have stops early.
type limitWriter struct {
	w    io.Writer
	left int64
}

func (l *limitWriter) Write(p []byte) (int, error) {
	if int64(len(p)) > l.left {
		return 0, fmt.Errorf("larger than its bound")
	}
	l.left -= int64(len(p))
	return l.w.Write(p)
}

// limitBuffer keeps the first 4 KB of a command's error output.
type limitBuffer struct{ buf []byte }

func (b *limitBuffer) Write(p []byte) (int, error) {
	if room := 4096 - len(b.buf); room > 0 {
		b.buf = append(b.buf, p[:min(room, len(p))]...)
	}
	return len(p), nil
}

func (b *limitBuffer) String() string { return string(b.buf) }
