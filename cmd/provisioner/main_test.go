// main_test.go
//
// # Go Provisioner CLI Flag Tests
//
// Tests for CLI flag parsing and filtering logic in the provisioner.
//
// # Usage
//
//     go test ./cmd/provisioner -v
//
// # Tests
//   - TestProvisioner_AllFlag: --all installs all packages
//   - TestProvisioner_LazyFlag: --lazy only installs lazy packages
//
// # Example
//     go test ./cmd/provisioner -v

package main

import (
	"os"
	"os/exec"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

const testManifestYAML = `
foo:
  apt: foo
  lazy: true
bar:
  apt: bar
  lazy: false
baz:
  apt: baz
  lazy: true
`

// writeTempManifest writes the test manifest to a temp file and returns its path.
func writeTempManifest(t *testing.T) string {
	t.Helper()
	tmp, err := os.CreateTemp("", "test-manifest-*.yaml")
	if err != nil {
		t.Fatalf("failed to create temp manifest: %v", err)
	}
	if _, err := tmp.WriteString(testManifestYAML); err != nil {
		if err2 := tmp.Close(); err2 != nil {
			t.Errorf("tmp.Close failed: %v", err2)
		}
		t.Fatalf("failed to write temp manifest: %v", err)
	}
	if err := tmp.Close(); err != nil {
		t.Errorf("tmp.Close failed: %v", err)
	}
	return tmp.Name()
}

// TestProvisioner_AllFlag verifies that --all installs all packages.
func TestProvisioner_AllFlag(t *testing.T) {
	manifestPath := writeTempManifest(t)
	defer func() {
		if err := os.Remove(manifestPath); err != nil {
			t.Errorf("os.Remove failed: %v", err)
		}
	}()
	cmd := exec.Command("go", "run", ".", "--all", "--no-tui", "--manifest", manifestPath, "--dry-run")
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("provisioner --all failed: %v\nOutput: %s", err, string(out))
	}
	output := string(out)
	if !strings.Contains(output, "foo") || !strings.Contains(output, "bar") || !strings.Contains(output, "baz") {
		t.Errorf("expected all packages in output, got: %s", output)
	}
	if !strings.Contains(output, "[dry-run] Would run: apt foo") {
		t.Errorf("expected dry-run for foo, got: %s", output)
	}
	if !strings.Contains(output, "[dry-run] Would run: apt bar") {
		t.Errorf("expected dry-run for bar, got: %s", output)
	}
	if !strings.Contains(output, "[dry-run] Would run: apt baz") {
		t.Errorf("expected dry-run for baz, got: %s", output)
	}
	if !strings.Contains(output, "Provisioning complete") {
		t.Errorf("expected output to contain 'Provisioning complete', got: %s", output)
	}
}

// TestProvisioner_LazyFlag verifies that --lazy only installs lazy packages.
func TestProvisioner_LazyFlag(t *testing.T) {
	manifestPath := writeTempManifest(t)
	defer func() {
		if err := os.Remove(manifestPath); err != nil {
			t.Errorf("os.Remove failed: %v", err)
		}
	}()
	cmd := exec.Command("go", "run", ".", "--lazy", "--no-tui", "--manifest", manifestPath, "--dry-run")
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("provisioner --lazy failed: %v\nOutput: %s", err, string(out))
	}
	output := string(out)
	if !strings.Contains(output, "foo") || !strings.Contains(output, "baz") {
		t.Errorf("expected only lazy packages in output, got: %s", output)
	}
	if strings.Contains(output, "bar") {
		t.Errorf("did not expect non-lazy package 'bar' in output, got: %s", output)
	}
	if !strings.Contains(output, "[dry-run] Would run: apt foo") {
		t.Errorf("expected dry-run for foo, got: %s", output)
	}
	if !strings.Contains(output, "[dry-run] Would run: apt baz") {
		t.Errorf("expected dry-run for baz, got: %s", output)
	}
	if strings.Contains(output, "[dry-run] Would run: apt bar") {
		t.Errorf("did not expect dry-run for bar, got: %s", output)
	}
	if !strings.Contains(output, "Provisioning complete") {
		t.Errorf("expected output to contain 'Provisioning complete', got: %s", output)
	}
}

func TestModel_handleKeyMsg(t *testing.T) {
	m := initialModel()
	m.logs = make([]logEntry, 30)
	m.cursor = 10
	m.userScrolled = false

	cases := []struct {
		name             string
		key              string
		wantQuit         bool
		wantCur          int
		wantUserScrolled bool
		preCursor        int // set before test if >=0
	}{
		{"quit key", "q", true, 10, false, -1},
		{"ctrl+c", "ctrl+c", true, 10, false, -1},
		{"up", "up", false, 9, true, -1},
		{"down", "down", false, 10, false, 9}, // set cursor to 9 before test
		{"end", "end", false, len(m.logs) - logPanelHeight, false, -1},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.preCursor >= 0 {
				m.cursor = tc.preCursor
			} else {
				m.cursor = 10
			}
			m2, cmd := m.handleKeyMsg(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(tc.key), Alt: false})
			if tc.wantQuit {
				if cmd == nil {
					t.Errorf("expected tea.Quit for key %q", tc.key)
				}
			} else {
				if m2.cursor != tc.wantCur {
					t.Errorf("cursor: got %d, want %d", m2.cursor, tc.wantCur)
				}
				if m2.userScrolled != tc.wantUserScrolled {
					t.Errorf("userScrolled: got %v, want %v", m2.userScrolled, tc.wantUserScrolled)
				}
			}
		})
	}
}

//revive:disable:var-naming
func SkipTestModel_handleLogMsg(t *testing.T) {
	//revive:enable:var-naming
	cases := []struct {
		name       string
		msg        logMsg
		wantStatus string
		wantCursor int
	}{
		{"success", logMsg{Level: "success", Text: "Provisioning complete"}, "Done", 0},
		{"error", logMsg{Level: "error", Text: "fail"}, "fail", 0},
		{"info", logMsg{Level: "info", Text: "info msg"}, "", 0},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			m := initialModel()
			m.logs = []logEntry{}
			m.status = ""
			m.cursor = 0
			m.userScrolled = false
			m2 := m.handleLogMsg(tc.msg)
			if m2.status != tc.wantStatus {
				t.Errorf("status: got %q, want %q", m2.status, tc.wantStatus)
			}
			if m2.cursor != tc.wantCursor {
				t.Errorf("cursor: got %d, want %d", m2.cursor, tc.wantCursor)
			}
			if len(m2.logs) != 1 {
				t.Errorf("logs: got %d, want 1", len(m2.logs))
			}
		})
	}
}
