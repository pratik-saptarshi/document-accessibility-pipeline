package cloudtools

import (
	"archive/tar"
	"compress/gzip"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"time"
)

const maxDownload = 128 << 20

type Options struct {
	Root     string
	Cache    string
	GoBinary string
	Client   *http.Client
	Output   io.Writer
	LockWait time.Duration
}

type InstalledStatus struct {
	Name    string
	Version string
	Path    string
	Status  string
	OK      bool
}

// VerifyInstalled performs read-only version/digest checks for every local tool.
func VerifyInstalled(ctx context.Context, m Manifest, o Options) ([]InstalledStatus, error) {
	if err := ctx.Err(); err != nil {
		return nil, fmt.Errorf("verification canceled: %w", err)
	}
	if err := m.Validate(); err != nil {
		return nil, err
	}
	if runtime.GOOS+"/"+runtime.GOARCH != m.Platform {
		return nil, fmt.Errorf("unsupported host %s/%s; manifest supports %s", runtime.GOOS, runtime.GOARCH, m.Platform)
	}
	if o.Root == "" || o.GoBinary == "" {
		return nil, errors.New("root and verified Go binary paths are required")
	}
	var results []InstalledStatus
	for _, tool := range m.Tools {
		if err := ctx.Err(); err != nil {
			return nil, fmt.Errorf("verification canceled: %w", err)
		}
		result := InstalledStatus{Name: tool.Name, Version: tool.Version}
		path := filepath.Join(o.Root, "bin", tool.Binary)
		if tool.Kind == "bootstrap-toolchain" {
			path = o.GoBinary
		}
		result.Path = path
		if ok, reason := installedOK(ctx, tool, path, o); ok {
			result.Status = "verified version/digest"
			result.OK = true
		} else {
			if err := ctx.Err(); err != nil {
				return nil, fmt.Errorf("verification canceled: %w", err)
			}
			result.Status = "missing or invalid: " + reason
		}
		results = append(results, result)
	}
	if err := ctx.Err(); err != nil {
		return nil, fmt.Errorf("verification canceled: %w", err)
	}
	return results, nil
}

func Bootstrap(ctx context.Context, m Manifest, o Options) error {
	if err := m.Validate(); err != nil {
		return err
	}
	if runtime.GOOS+"/"+runtime.GOARCH != m.Platform {
		return fmt.Errorf("unsupported host %s/%s; manifest supports %s", runtime.GOOS, runtime.GOARCH, m.Platform)
	}
	if o.Root == "" || o.Cache == "" || o.GoBinary == "" {
		return errors.New("root, cache, and verified Go binary paths are required")
	}
	goBinary, err := filepath.Abs(o.GoBinary)
	if err != nil {
		return fmt.Errorf("resolve verified Go binary path: %w", err)
	}
	o.GoBinary = goBinary
	for _, p := range []string{o.Root, o.Cache} {
		if err := os.MkdirAll(p, 0700); err != nil {
			return fmt.Errorf("create isolated directory %s: %w", p, err)
		}
	}
	unlock, err := acquireLock(ctx, filepath.Join(o.Root, ".bootstrap.lock"), o.LockWait)
	if err != nil {
		return err
	}
	defer unlock()
	if o.Client == nil {
		o.Client = &http.Client{Timeout: 2 * time.Minute}
	}
	if o.Output == nil {
		o.Output = io.Discard
	}
	var failures []string
	for _, tool := range m.Tools {
		if err := ctx.Err(); err != nil {
			return fmt.Errorf("bootstrap canceled: %w", err)
		}
		if tool.Kind == "bootstrap-toolchain" {
			if ok, reason := installedOK(ctx, tool, o.GoBinary, o); !ok {
				return fmt.Errorf("verified Go toolchain is invalid: %s", reason)
			}
			fmt.Fprintf(o.Output, "verified bootstrap toolchain %s %s\n", tool.Name, tool.Version)
			continue
		}
		binPath := filepath.Join(o.Root, "bin", tool.Binary)
		if ok, reason := installedOK(ctx, tool, binPath, o); ok {
			fmt.Fprintf(o.Output, "verified %s %s\n", tool.Name, tool.Version)
			continue
		} else if _, err := os.Stat(binPath); err == nil {
			fmt.Fprintf(o.Output, "repair %s: %s\n", tool.Name, reason)
		}
		if err := installOne(ctx, tool, binPath, o); err != nil {
			if ctx.Err() != nil {
				return fmt.Errorf("bootstrap canceled: %w", ctx.Err())
			}
			failures = append(failures, fmt.Sprintf("%s %s: %s", tool.Name, tool.Version, err))
			continue
		}
		if err := ctx.Err(); err != nil {
			return fmt.Errorf("bootstrap canceled: %w", err)
		}
		if ok, reason := installedOK(ctx, tool, binPath, o); !ok {
			if ctx.Err() != nil {
				return fmt.Errorf("bootstrap canceled: %w", ctx.Err())
			}
			failures = append(failures, fmt.Sprintf("verify installed %s %s: %s", tool.Name, tool.Version, reason))
			continue
		}
		fmt.Fprintf(o.Output, "installed %s %s\n", tool.Name, tool.Version)
	}
	if err := ctx.Err(); err != nil {
		return fmt.Errorf("bootstrap canceled: %w", err)
	}
	if err := writeActivation(filepath.Join(o.Root, "activate.sh"), o.GoBinary); err != nil {
		return err
	}
	if len(failures) != 0 {
		return fmt.Errorf("bootstrap completed with blocked tools: %s", strings.Join(failures, "; "))
	}
	return nil
}

func writeActivation(path, goBinary string) error {
	if goBinary == "" || !filepath.IsAbs(goBinary) {
		return errors.New("activation requires an absolute verified Go binary path")
	}
	goCommandBin := ""
	if filepath.Base(goBinary) != "go" {
		goCommandBin = filepath.Join(filepath.Dir(path), ".activation", "bin")
		if err := writeGoLauncher(filepath.Join(goCommandBin, "go"), goBinary); err != nil {
			return fmt.Errorf("write selected Go command launcher: %w", err)
		}
	}
	// Escape the shell's %% parameter expansion for fmt.Sprintf.
	script := fmt.Sprintf(`# Source this file to use the isolated cloud build tools in the current shell.
_cloud_tools_bin="$(CDPATH= cd -- "$(dirname -- "${BASH_SOURCE[0]}")/bin" && pwd)" || return
_cloud_go_bin="$(CDPATH= cd -- "$(dirname -- %s)" && pwd)" || return
_cloud_go_command_bin=%s
_cloud_new_path="${_cloud_go_bin}:${_cloud_tools_bin}"
if [[ -n "${_cloud_go_command_bin}" ]]; then
  _cloud_new_path="${_cloud_go_command_bin}:${_cloud_new_path}"
fi
_cloud_path="${PATH-}"
if [[ -n "${_cloud_path}" ]]; then
  while :; do
    if [[ "${_cloud_path}" == *:* ]]; then
      _cloud_path_entry="${_cloud_path%%%%:*}"
      _cloud_path="${_cloud_path#*:}"
      _cloud_path_has_more=1
    else
      _cloud_path_entry="${_cloud_path}"
      _cloud_path_has_more=0
    fi
    _cloud_skip_path=0
    if [[ "${_cloud_path_entry}" == "${_cloud_go_bin}" || "${_cloud_path_entry}" == "${_cloud_tools_bin}" ]]; then
      _cloud_skip_path=1
    fi
    if [[ -n "${_cloud_go_command_bin}" && "${_cloud_path_entry}" == "${_cloud_go_command_bin}" ]]; then
      _cloud_skip_path=1
    fi
    if [[ "${_cloud_skip_path}" == 0 ]]; then
      _cloud_new_path="${_cloud_new_path}:${_cloud_path_entry}"
    fi
    [[ "${_cloud_path_has_more}" == 1 ]] || break
  done
fi
PATH="${_cloud_new_path}"
export PATH
unset _cloud_go_command_bin
unset _cloud_path
unset _cloud_new_path
unset _cloud_path_entry
unset _cloud_path_has_more
unset _cloud_skip_path
unset _cloud_go_bin
unset _cloud_tools_bin
`, shellQuote(goBinary), shellQuote(goCommandBin))
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, []byte(script), 0600); err != nil {
		return err
	}
	if err := os.Rename(tmp, path); err != nil {
		_ = os.Remove(tmp)
		return err
	}
	return nil
}

const goLauncherHeader = "#!/bin/sh\n# Managed by cloud-tools-bootstrap; do not edit.\n"

func writeGoLauncher(path, goBinary string) error {
	if goBinary == "" || !filepath.IsAbs(goBinary) {
		return errors.New("Go command launcher requires an absolute selected binary path")
	}
	selectedInfo, err := os.Lstat(goBinary)
	if err != nil {
		return fmt.Errorf("inspect selected Go binary for launcher: %w", err)
	}
	if !selectedInfo.Mode().IsRegular() || selectedInfo.Mode()&0111 == 0 {
		return errors.New("selected Go binary for launcher is not a regular executable file")
	}

	activationDir := filepath.Dir(filepath.Dir(path))
	if err := ensurePrivateDir(activationDir); err != nil {
		return err
	}
	binDir := filepath.Dir(path)
	if err := ensurePrivateDir(binDir); err != nil {
		return err
	}

	launcher := goLauncherHeader + "exec " + shellQuote(goBinary) + " \"$@\"\n"
	if info, err := os.Lstat(path); err == nil {
		if !info.Mode().IsRegular() || info.Mode()&0111 == 0 {
			return errors.New("refusing to replace a non-regular selected Go launcher")
		}
		contents, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("read existing selected Go launcher: %w", err)
		}
		if !strings.HasPrefix(string(contents), goLauncherHeader) {
			return errors.New("refusing to replace an unmanaged file at the selected Go launcher path")
		}
	} else if !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("inspect existing selected Go launcher: %w", err)
	}

	tmp, err := os.CreateTemp(binDir, ".go-launcher-*")
	if err != nil {
		return fmt.Errorf("create selected Go launcher temporary file: %w", err)
	}
	tmpPath := tmp.Name()
	defer os.Remove(tmpPath)
	if _, err := io.WriteString(tmp, launcher); err != nil {
		_ = tmp.Close()
		return fmt.Errorf("write selected Go launcher: %w", err)
	}
	if err := tmp.Chmod(0700); err != nil {
		_ = tmp.Close()
		return fmt.Errorf("secure selected Go launcher: %w", err)
	}
	if err := tmp.Close(); err != nil {
		return fmt.Errorf("close selected Go launcher: %w", err)
	}
	if err := os.Rename(tmpPath, path); err != nil {
		return fmt.Errorf("publish selected Go launcher: %w", err)
	}
	return nil
}

func ensurePrivateDir(path string) error {
	if err := os.Mkdir(path, 0700); err != nil && !errors.Is(err, os.ErrExist) {
		return fmt.Errorf("create selected Go activation directory %s: %w", path, err)
	}
	info, err := os.Lstat(path)
	if err != nil {
		return fmt.Errorf("inspect selected Go activation directory %s: %w", path, err)
	}
	if !info.IsDir() {
		return fmt.Errorf("selected Go activation path %s is not a directory", path)
	}
	if info.Mode().Perm() != 0700 {
		if err := os.Chmod(path, 0700); err != nil {
			return fmt.Errorf("secure selected Go activation directory %s: %w", path, err)
		}
	}
	return nil
}

func shellQuote(value string) string {
	return "'" + strings.ReplaceAll(value, "'", "'\\''") + "'"
}

func acquireLock(ctx context.Context, path string, maxWait time.Duration) (func(), error) {
	if maxWait <= 0 {
		maxWait = 2 * time.Minute
	}
	deadline := time.NewTimer(maxWait)
	defer deadline.Stop()
	tick := time.NewTicker(100 * time.Millisecond)
	defer tick.Stop()
	for {
		err := os.Mkdir(path, 0700)
		if err == nil {
			return func() { _ = os.Remove(path) }, nil
		}
		if !errors.Is(err, os.ErrExist) {
			return nil, fmt.Errorf("create bootstrap lock %s: %w", path, err)
		}
		select {
		case <-ctx.Done():
			return nil, fmt.Errorf("waiting for concurrent bootstrap: %w", ctx.Err())
		case <-deadline.C:
			return nil, fmt.Errorf("timed out waiting for bootstrap lock %s; inspect and remove it only if no bootstrap is running", path)
		case <-tick.C:
		}
	}
}

func installOne(ctx context.Context, t Tool, target string, o Options) error {
	if err := os.MkdirAll(filepath.Dir(target), 0700); err != nil {
		return err
	}
	stageDir, err := os.MkdirTemp(filepath.Dir(target), ".stage-"+t.Name+"-")
	if err != nil {
		return err
	}
	defer os.RemoveAll(stageDir)
	stageBinary := filepath.Join(stageDir, t.Binary)
	if t.Kind == "go-install" {
		if err := installGoModule(ctx, t, stageDir, o); err != nil {
			return err
		}
		stageBinary = filepath.Join(stageDir, t.Binary)
	} else {
		asset, err := verifiedAsset(ctx, t, o)
		if err != nil {
			return err
		}
		if t.Kind == "raw" {
			if err := copyFile(asset, stageBinary, 0755); err != nil {
				return err
			}
		} else if err := extractTarBinary(asset, t.ArchiveMember, stageBinary); err != nil {
			return err
		}
	}
	if err := os.Chmod(stageBinary, 0755); err != nil {
		return err
	}
	if t.BinarySHA256 != "" {
		if ok, err := fileMatchesSHA(stageBinary, t.BinarySHA256); err != nil || !ok {
			return fmt.Errorf("installed binary SHA-256 does not match the pinned value")
		}
	}
	if ok, reason := installedOK(ctx, t, stageBinary, o); !ok {
		return fmt.Errorf("staged binary failed version validation: %s", reason)
	}
	if err := os.Rename(stageBinary, target); err != nil {
		return fmt.Errorf("atomically replace %s: %w", target, err)
	}
	return nil
}

func verifiedAsset(ctx context.Context, t Tool, o Options) (string, error) {
	dir := filepath.Join(o.Cache, "downloads")
	if err := os.MkdirAll(dir, 0700); err != nil {
		return "", err
	}
	filename := filepath.Base(strings.Split(t.Source, "?")[0])
	stable := filepath.Join(dir, t.Name+"-"+t.Version+"-"+filename)
	if ok, _ := fileMatchesSHA(stable, t.SHA256); ok {
		return stable, nil
	}
	_ = os.Remove(stable)
	tmp, err := os.CreateTemp(dir, ".download-"+t.Name+"-")
	if err != nil {
		return "", err
	}
	tmpName := tmp.Name()
	defer os.Remove(tmpName)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, t.Source, nil)
	if err != nil {
		_ = tmp.Close()
		return "", err
	}
	resp, err := o.Client.Do(req)
	if err != nil {
		_ = tmp.Close()
		return "", errors.New("HTTPS request failed; URL details withheld; check connectivity to the official manifest host")
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		_ = tmp.Close()
		return "", fmt.Errorf("download returned HTTP %s", resp.Status)
	}
	h := sha256.New()
	n, err := io.Copy(io.MultiWriter(tmp, h), io.LimitReader(resp.Body, maxDownload+1))
	closeErr := tmp.Close()
	if err != nil {
		return "", err
	}
	if closeErr != nil {
		return "", closeErr
	}
	if n > maxDownload {
		return "", fmt.Errorf("download exceeds %d-byte limit", maxDownload)
	}
	actual := hex.EncodeToString(h.Sum(nil))
	if !strings.EqualFold(actual, t.SHA256) {
		return "", fmt.Errorf("SHA-256 mismatch: expected %s, got %s; downloaded file was discarded", t.SHA256, actual)
	}
	if err := os.Rename(tmpName, stable); err != nil {
		return "", err
	}
	return stable, nil
}

func installGoModule(ctx context.Context, t Tool, stageDir string, o Options) error {
	ctx, cancel := context.WithTimeout(ctx, 4*time.Minute)
	defer cancel()
	cache := filepath.Join(o.Cache, "go")
	for _, sub := range []string{"mod", "build", "gopath"} {
		if err := os.MkdirAll(filepath.Join(cache, sub), 0700); err != nil {
			return err
		}
	}
	cmd := commandContext(ctx, o.GoBinary, "install", t.Module)
	cmd.Env = cleanEnv(map[string]string{
		"GOBIN":       stageDir,
		"GOPATH":      filepath.Join(cache, "gopath"),
		"GOMODCACHE":  filepath.Join(cache, "mod"),
		"GOCACHE":     filepath.Join(cache, "build"),
		"GOTOOLCHAIN": "local",
		"GOENV":       "off",
		// Use only the approved public proxy. Do not add a direct fallback.
		"GOPROXY":             "https://proxy.golang.org",
		"GONOPROXY":           "",
		"GOSUMDB":             "sum.golang.org",
		"GOPRIVATE":           "",
		"GONOSUMDB":           "",
		"GOINSECURE":          "",
		"GOWORK":              "off",
		"GOFLAGS":             "-mod=readonly",
		"GIT_TERMINAL_PROMPT": "0",
	})
	out, err := cmd.CombinedOutput()
	if err != nil {
		_ = out
		return fmt.Errorf("pinned go install for %s failed (%v); subprocess output was withheld to avoid leaking proxy credentials; check official source connectivity and checksum-database access", t.Module, err)
	}
	return nil
}

func cleanEnv(replace map[string]string) []string {
	values := make(map[string]string, len(replace)+1)
	for k, v := range replace {
		values[k] = v
	}
	if _, ok := values["PATH"]; !ok {
		values["PATH"] = "/usr/local/bin:/usr/bin:/bin"
	}
	keys := map[string]bool{}
	for k := range values {
		keys[k] = true
	}
	env := make([]string, 0, len(os.Environ())+len(values))
	for _, item := range os.Environ() {
		k, _, _ := strings.Cut(item, "=")
		if !keys[k] {
			env = append(env, item)
		}
	}
	for k, v := range values {
		env = append(env, k+"="+v)
	}
	return env
}

func extractTarBinary(archive, member, target string) error {
	f, err := os.Open(archive)
	if err != nil {
		return err
	}
	defer f.Close()
	gz, err := gzip.NewReader(f)
	if err != nil {
		return err
	}
	defer gz.Close()
	tr := tar.NewReader(gz)
	for {
		h, err := tr.Next()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return err
		}
		clean := filepath.Clean(filepath.FromSlash(h.Name))
		if filepath.IsAbs(clean) || clean == "." || clean == ".." || strings.HasPrefix(clean, ".."+string(filepath.Separator)) {
			return fmt.Errorf("unsafe archive path %q", h.Name)
		}
		if filepath.Base(clean) != member {
			continue
		}
		if h.Typeflag != tar.TypeReg && h.Typeflag != tar.TypeRegA {
			return fmt.Errorf("archive member %q is not a regular file", h.Name)
		}
		out, err := os.OpenFile(target, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0700)
		if err != nil {
			return err
		}
		_, cpErr := io.Copy(out, io.LimitReader(tr, 256<<20))
		clErr := out.Close()
		if cpErr != nil {
			return cpErr
		}
		if clErr != nil {
			return clErr
		}
		return nil
	}
	return fmt.Errorf("archive member %q not found", member)
}

func copyFile(src, dst string, mode os.FileMode) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()
	out, err := os.OpenFile(dst, os.O_CREATE|os.O_EXCL|os.O_WRONLY, mode)
	if err != nil {
		return err
	}
	_, cpErr := io.Copy(out, io.LimitReader(in, maxDownload+1))
	clErr := out.Close()
	if cpErr != nil {
		return cpErr
	}
	return clErr
}

func fileMatchesSHA(path, expected string) (bool, error) {
	f, err := os.Open(path)
	if err != nil {
		return false, err
	}
	defer f.Close()
	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return false, err
	}
	return strings.EqualFold(hex.EncodeToString(h.Sum(nil)), expected), nil
}

func installedOK(ctx context.Context, t Tool, path string, o Options) (bool, string) {
	info, err := os.Lstat(path)
	if err != nil {
		return false, "binary is missing"
	}
	if !info.Mode().IsRegular() || info.Mode()&0111 == 0 {
		return false, "binary is not a regular executable file"
	}
	if t.BinarySHA256 != "" {
		if ok, err := fileMatchesSHA(path, t.BinarySHA256); err != nil || !ok {
			return false, "binary digest does not match the pinned executable"
		}
	}
	if t.VersionSource != "" {
		ok, err := fileMatchesSHA(path, t.SHA256)
		if err != nil || !ok {
			return false, "binary digest does not match the pinned release asset"
		}
		return true, ""
	}
	ctx, cancel := context.WithTimeout(ctx, 20*time.Second)
	defer cancel()
	cmd := commandContext(ctx, path, t.VersionArgs...)
	cmd.Env = cleanEnv(nil)
	out, runErr := cmd.CombinedOutput()
	if runErr == nil && regexp.MustCompile(t.VersionPattern).Match(out) {
		return true, ""
	}
	if t.VersionFallback != "" {
		fallback := commandContext(ctx, o.GoBinary, "version", "-m", path)
		fallback.Env = cleanEnv(nil)
		meta, err := fallback.CombinedOutput()
		wantModule := strings.SplitN(t.Module, "@", 2)[0]
		if err == nil && matchesModuleVersion(string(meta), wantModule, t.Version) {
			return true, ""
		}
		return false, "CLI output did not report the pinned version; inspect the binary and version command"
	}
	return false, "CLI output did not report the pinned version; inspect the binary and version command"
}

func matchesModuleVersion(buildInfo, module, version string) bool {
	for _, line := range strings.Split(buildInfo, "\n") {
		fields := strings.Fields(line)
		if len(fields) >= 3 && fields[0] == "mod" && fields[1] == module && strings.TrimPrefix(fields[2], "v") == version {
			return true
		}
	}
	return false
}
