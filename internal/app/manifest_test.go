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
	f, err := os.CreateTemp("", "test-*.yml")
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if closeErr := f.Close(); closeErr != nil {
			t.Error(closeErr)
		}
		if removeErr := os.Remove(f.Name()); removeErr != nil {
			t.Error(removeErr)
		}
	}()

	if _, writeErr := f.WriteString(sampleYAML); writeErr != nil {
		t.Fatal(writeErr)
	}

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
