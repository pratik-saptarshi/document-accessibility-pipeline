package modulepin

import (
	"runtime/debug"
	"strings"
	"testing"
)

func TestCheckPinned(t *testing.T) {
	const path = "example.com/pinned"
	tests := []struct {
		name    string
		dep     *debug.Module
		want    string
		wantErr bool
	}{
		{name: "exact version, no replacement", dep: &debug.Module{Path: path, Version: "v1.2.3"}, want: "v1.2.3"},
		{name: "wrong version", dep: &debug.Module{Path: path, Version: "v1.2.4"}, want: "v1.2.3", wantErr: true},
		{name: "empty version", dep: &debug.Module{Path: path}, want: "v1.2.3", wantErr: true},
		{name: "replaced, matching version", dep: &debug.Module{Path: path, Version: "v1.2.3", Replace: &debug.Module{Path: "example.com/other", Version: "v0.0.0"}}, want: "v1.2.3", wantErr: true},
		{name: "replaced, wrong version", dep: &debug.Module{Path: path, Version: "v9.9.9", Replace: &debug.Module{Path: "../local"}}, want: "v1.2.3", wantErr: true},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := CheckPinned(tc.dep, tc.want)
			if tc.wantErr && err == nil {
				t.Fatalf("CheckPinned(%+v, %q) = nil, want error", tc.dep, tc.want)
			}
			if !tc.wantErr && err != nil {
				t.Fatalf("CheckPinned(%+v, %q) = %v, want nil", tc.dep, tc.want, err)
			}
			if tc.wantErr && err != nil && !strings.Contains(err.Error(), path) {
				t.Fatalf("CheckPinned error %q does not name the module path %q", err, path)
			}
		})
	}
}

func TestVersion(t *testing.T) {
	const path = "example.com/pinned"
	info := &debug.BuildInfo{Deps: []*debug.Module{
		{Path: "example.com/other", Version: "v0.1.0"},
		{Path: path, Version: "v1.2.3"},
	}}
	if got, err := Version(info, true, path); err != nil || got != "v1.2.3" {
		t.Fatalf("Version(present) = %q, %v; want v1.2.3, nil", got, err)
	}
	if _, err := Version(info, true, "example.com/absent"); err == nil {
		t.Fatal("Version(absent) = nil error, want error")
	}
	if _, err := Version(nil, false, path); err == nil {
		t.Fatal("Version(no build info) = nil error, want error")
	}
	replaced := &debug.BuildInfo{Deps: []*debug.Module{
		{Path: path, Version: "v1.2.3", Replace: &debug.Module{Path: "../local"}},
	}}
	if _, err := Version(replaced, true, path); err == nil {
		t.Fatal("Version(replaced) = nil error, want error")
	}
	empty := &debug.BuildInfo{Deps: []*debug.Module{{Path: path}}}
	if _, err := Version(empty, true, path); err == nil {
		t.Fatal("Version(empty version) = nil error, want error")
	}
}
