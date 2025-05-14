package app

import (
	"os"
	"testing"
)

const sampleYAML = `testapp:
  _bin: testapp
  _desc: Test app description
  _docs: https://example.com/docs
  _github: https://github.com/example/testapp
  _home: https://example.com
  _name: TestApp
  _short: A test app
  brew: testapp
  apt: testapp
  pacman: testapp
  choco: testapp
  go: github.com/example/testapp@
  snap: testapp
`

func TestLoadManifest(t *testing.T) {
	f, err := os.CreateTemp("", "manifest-*.yml")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	defer os.Remove(f.Name())
	if _, err := f.WriteString(sampleYAML); err != nil {
		t.Fatalf("failed to write sample YAML: %v", err)
	}
	f.Close()

	manifest, err := LoadManifest(f.Name())
	if err != nil {
		t.Fatalf("LoadManifest failed: %v", err)
	}
	entry, ok := manifest["testapp"]
	if !ok {
		t.Fatalf("expected 'testapp' entry in manifest")
	}
	if len(entry.Bin) != 1 || entry.Bin[0] != "testapp" || entry.Name != "TestApp" {
		t.Errorf("unexpected entry values: %+v", entry)
	}
}
