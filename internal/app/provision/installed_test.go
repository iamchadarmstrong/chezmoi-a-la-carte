package provision

import (
	"strings"
	"testing"
)

type fakeOutputRunner struct {
	outputs map[string][]byte
}

func (f *fakeOutputRunner) Run(cmd string, args ...string) error { return nil }
func (f *fakeOutputRunner) Output(cmd string, args ...string) ([]byte, error) {
	key := cmd
	if len(args) > 0 {
		key += " " + strings.Join(args, " ")
	}
	if out, ok := f.outputs[key]; ok {
		return out, nil
	}
	return nil, nil
}

func TestGetInstalledPackages(t *testing.T) {
	runner := &fakeOutputRunner{outputs: map[string][]byte{
		"dpkg -l": []byte(`
ii  foo    1.0 all some package
rc  bar    2.0 all removed config
`),
		"brew list":    []byte("bat\nfd\n"),
		"brew list -1": []byte("bat\nfd\n"),
		"pipx list": []byte(`venvs are in /home/user/.local/pipx/venvs
  - black
  - isort
`),
		"cargo install": []byte(`bat v0.23.0:
fd-find v8.2.1:
`),
		"cargo install --list": []byte(`bat v0.23.0:
fd-find v8.2.1:
`),
		"npm list":    []byte(``),
		"npm list -g": []byte(``),
		"npm list -g --depth=0": []byte(`
/home/user/.nvm/versions/node/v18.16.1/lib
├── npm@8.19.2
├── zx@7.2.3
└── cowsay@1.5.0
`),
	}}
	got := GetInstalledPackages(runner)
	want := map[string]bool{
		"foo":     true,
		"bat":     true,
		"fd":      true,
		"black":   true,
		"isort":   true,
		"fd-find": true,
		"npm":     true,
		"zx":      true,
		"cowsay":  true,
	}
	for k := range want {
		if !got[k] {
			t.Errorf("expected %s to be detected as installed", k)
		}
	}
	// Should not include 'bar' (rc state in dpkg)
	if got["bar"] {
		t.Errorf("did not expect 'bar' to be detected as installed")
	}
}
