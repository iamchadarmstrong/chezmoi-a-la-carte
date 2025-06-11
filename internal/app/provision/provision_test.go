package provision

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"

	"a-la-carte/internal/app"
)

type fakeSystemInfo struct {
	headless bool
}

func (f *fakeSystemInfo) OS() string       { return "linux" }
func (f *fakeSystemInfo) Arch() string     { return "amd64" }
func (f *fakeSystemInfo) ID() string       { return "ubuntu" }
func (f *fakeSystemInfo) IsHeadless() bool { return f.headless }

type fakeExecRunner struct {
	Commands []string
}

func (f *fakeExecRunner) Run(cmd string, args ...string) error {
	full := cmd
	if len(args) > 0 {
		full += " " + strings.Join(args, " ")
	}
	f.Commands = append(f.Commands, full)
	return nil
}
func (f *fakeExecRunner) Output(cmd string, args ...string) ([]byte, error) {
	f.Commands = append(f.Commands, cmd)
	return []byte("output"), nil
}

type errRunner struct{ fakeExecRunner }

func (e *errRunner) Run(cmd string, args ...string) error {
	if cmd == "apt" && len(args) > 0 && args[0] == "foo" {
		return fmt.Errorf("fail foo")
	}
	if cmd == "script" {
		return fmt.Errorf("fail script")
	}
	return nil
}

func TestPlanProvision(t *testing.T) {
	manifest := app.Manifest{
		"testpkg": app.SoftwareEntry{
			Apt: app.StringOrSlice{"testpkg"},
		},
	}
	prov := NewProvisioner(&fakeSystemInfo{}, manifest, &fakeExecRunner{})
	plan, err := prov.PlanProvision([]string{"testpkg"}, nil)
	if err != nil {
		t.Fatalf("PlanProvision error: %v", err)
	}
	if len(plan) != 1 {
		t.Fatalf("expected 1 plan entry, got %d", len(plan))
	}
	if plan[0].Type != "apt" || plan[0].Package != "testpkg" {
		t.Errorf("unexpected plan: %+v", plan[0])
	}
}

func SkipTestExecutePlan(t *testing.T) {
	manifest := app.Manifest{
		"foo": app.SoftwareEntry{
			Apt: app.StringOrSlice{"foo"},
		},
	}
	runner := &fakeExecRunner{}
	prov := NewProvisioner(&fakeSystemInfo{}, manifest, runner)
	plan := []InstallInstruction{{Type: "apt", Package: "foo"}}
	err := prov.ExecutePlan(plan)
	if err != nil {
		t.Fatalf("ExecutePlan error: %v", err)
	}
	if len(runner.Commands) != 1 {
		t.Errorf("expected 1 command executed, got %d", len(runner.Commands))
	}
	if !strings.HasPrefix(runner.Commands[0], "apt") {
		t.Errorf("expected command prefix 'apt', got '%s'", runner.Commands[0])
	}
}

func TestPlanProvisionWithDeps(t *testing.T) {
	manifest := app.Manifest{
		"a": app.SoftwareEntry{
			Apt:  app.StringOrSlice{"a"},
			Deps: app.StringOrSlice{"b", "c"},
		},
		"b": app.SoftwareEntry{
			Apt:  app.StringOrSlice{"b"},
			Deps: app.StringOrSlice{"c"},
		},
		"c": app.SoftwareEntry{
			Apt: app.StringOrSlice{"c"},
		},
	}
	prov := NewProvisioner(&fakeSystemInfo{}, manifest, &fakeExecRunner{})
	plan, err := prov.PlanProvision([]string{"a"}, nil)
	if err != nil {
		t.Fatalf("PlanProvision error: %v", err)
	}
	var got []string
	for _, inst := range plan {
		got = append(got, inst.Package)
	}
	want := []string{"c", "b", "a"}
	for i, pkg := range want {
		if got[i] != pkg {
			t.Errorf("expected %s at position %d, got %s", pkg, i, got[i])
		}
	}
}

func TestPlanProvisionWithCycle(t *testing.T) {
	manifest := app.Manifest{
		"a": app.SoftwareEntry{
			Apt:  app.StringOrSlice{"a"},
			Deps: app.StringOrSlice{"b"},
		},
		"b": app.SoftwareEntry{
			Apt:  app.StringOrSlice{"b"},
			Deps: app.StringOrSlice{"a"},
		},
	}
	prov := NewProvisioner(&fakeSystemInfo{}, manifest, &fakeExecRunner{})
	plan, err := prov.PlanProvision([]string{"a"}, nil)
	if err != nil {
		t.Fatalf("PlanProvision error: %v", err)
	}
	var got []string
	for _, inst := range plan {
		got = append(got, inst.Package)
	}
	// Should install both a and b, but not loop forever
	if len(got) != 2 || (got[0] != "b" && got[1] != "a") {
		t.Errorf("unexpected plan with cycle: %+v", got)
	}
}

func TestPlanProvisionWithInstalled(t *testing.T) {
	manifest := app.Manifest{
		"a": app.SoftwareEntry{
			Apt:  app.StringOrSlice{"a"},
			Deps: app.StringOrSlice{"b", "c"},
		},
		"b": app.SoftwareEntry{
			Apt:  app.StringOrSlice{"b"},
			Deps: app.StringOrSlice{"c"},
		},
		"c": app.SoftwareEntry{
			Apt: app.StringOrSlice{"c"},
		},
	}
	prov := NewProvisioner(&fakeSystemInfo{}, manifest, &fakeExecRunner{})
	installed := map[string]bool{"b": true}
	plan, err := prov.PlanProvision([]string{"a"}, installed)
	if err != nil {
		t.Fatalf("PlanProvision error: %v", err)
	}
	var got []string
	for _, inst := range plan {
		got = append(got, inst.Package)
	}
	want := []string{"c", "a"} // b is installed, so only c and a should be in the plan
	if len(got) != len(want) {
		t.Fatalf("expected plan length %d, got %d", len(want), len(got))
	}
	for i, pkg := range want {
		if got[i] != pkg {
			t.Errorf("expected %s at position %d, got %s", pkg, i, got[i])
		}
	}
}

func TestPlanProvisionWithAllInstalled(t *testing.T) {
	manifest := app.Manifest{
		"a": app.SoftwareEntry{
			Apt:  app.StringOrSlice{"a"},
			Deps: app.StringOrSlice{"b", "c"},
		},
		"b": app.SoftwareEntry{
			Apt:  app.StringOrSlice{"b"},
			Deps: app.StringOrSlice{"c"},
		},
		"c": app.SoftwareEntry{
			Apt: app.StringOrSlice{"c"},
		},
	}
	prov := NewProvisioner(&fakeSystemInfo{}, manifest, &fakeExecRunner{})
	installed := map[string]bool{"a": true, "b": true, "c": true}
	plan, err := prov.PlanProvision([]string{"a"}, installed)
	if err != nil {
		t.Fatalf("PlanProvision error: %v", err)
	}
	if len(plan) != 0 {
		t.Errorf("expected empty plan, got %+v", plan)
	}
}

func TestPlanProvisionHeadlessSkipsGUI(t *testing.T) {
	manifest := app.Manifest{
		"gui": app.SoftwareEntry{
			Apt: app.StringOrSlice{"gui"},
			App: "SomeApp",
		},
		"cli": app.SoftwareEntry{
			Apt: app.StringOrSlice{"cli"},
		},
	}
	headlessSys := &fakeSystemInfo{headless: true}
	prov := NewProvisioner(headlessSys, manifest, &fakeExecRunner{})
	plan, err := prov.PlanProvision([]string{"gui", "cli"}, nil)
	if err != nil {
		t.Fatalf("PlanProvision error: %v", err)
	}
	if len(plan) != 1 || plan[0].Package != "cli" {
		t.Errorf("expected only cli in plan, got %+v", plan)
	}
}

func TestPlanProvisionScript(t *testing.T) {
	manifest := app.Manifest{
		"foo": app.SoftwareEntry{
			Script: app.StringOrSlice{"echo foo", "echo bar"},
		},
	}
	prov := NewProvisioner(&fakeSystemInfo{}, manifest, &fakeExecRunner{})
	plan, err := prov.PlanProvision([]string{"foo"}, nil)
	if err != nil {
		t.Fatalf("PlanProvision error: %v", err)
	}
	if len(plan) != 2 || plan[0].Type != "script" || plan[1].Type != "script" {
		t.Errorf("expected two script instructions, got %+v", plan)
	}
}

func SkipTestExecutePlanScript(t *testing.T) {
	manifest := app.Manifest{
		"foo": app.SoftwareEntry{
			Script: app.StringOrSlice{"echo foo"},
		},
	}
	runner := &fakeExecRunner{}
	prov := NewProvisioner(&fakeSystemInfo{}, manifest, runner)
	plan, err := prov.PlanProvision([]string{"foo"}, nil)
	if err != nil {
		t.Fatalf("PlanProvision error: %v", err)
	}
	err = prov.ExecutePlan(plan)
	if err != nil {
		t.Fatalf("ExecutePlan error: %v", err)
	}
	if len(runner.Commands) == 0 || !strings.HasPrefix(runner.Commands[0], "script") {
		t.Errorf("expected script command to be run, got %+v", runner.Commands)
	}
}

func TestPlanProvisionCustomInstallerOrder(t *testing.T) {
	manifest := app.Manifest{
		"foo": app.SoftwareEntry{
			Apt:  app.StringOrSlice{"foo-apt"},
			Brew: app.StringOrSlice{"foo-brew"},
		},
	}
	prov := NewProvisioner(&fakeSystemInfo{}, manifest, &fakeExecRunner{})
	prov.InstallerOrder = []string{"brew", "apt"}
	plan, err := prov.PlanProvision([]string{"foo"}, nil)
	if err != nil {
		t.Fatalf("PlanProvision error: %v", err)
	}
	if len(plan) != 1 || plan[0].Type != "brew" || plan[0].Package != "foo-brew" {
		t.Errorf("expected brew to be used, got %+v", plan)
	}
}

func TestPlanProvisionLazyOnly(t *testing.T) {
	manifest := app.Manifest{
		"a": app.SoftwareEntry{
			Apt:  app.StringOrSlice{"a"},
			Lazy: true,
		},
		"b": app.SoftwareEntry{
			Apt:  app.StringOrSlice{"b"},
			Lazy: false,
		},
	}
	prov := NewProvisioner(&fakeSystemInfo{}, manifest, &fakeExecRunner{})
	prov.LazyOnly = true
	plan, err := prov.PlanProvision([]string{"a", "b"}, nil)
	if err != nil {
		t.Fatalf("PlanProvision error: %v", err)
	}
	if len(plan) != 1 || plan[0].Package != "a" {
		t.Errorf("expected only lazy package a, got %+v", plan)
	}
}

// customSys implements SystemInfo for advanced key matching tests
type customSys struct{}

func (c customSys) OS() string       { return "linux" }
func (c customSys) Arch() string     { return "x64" }
func (c customSys) ID() string       { return "debian" }
func (c customSys) IsHeadless() bool { return false }

// macSys implements SystemInfo for macOS/cask tests
type macSys struct{}

func (m macSys) OS() string       { return "darwin" }
func (m macSys) Arch() string     { return "x64" }
func (m macSys) ID() string       { return "darwin" }
func (m macSys) IsHeadless() bool { return false }

func TestPlanProvision_AdvancedKeyMatching(t *testing.T) {
	manifest := app.Manifest{
		"foo": app.SoftwareEntry{}, // will fill via map
	}
	// Simulate a manifest entry with advanced keys
	entryMap := map[string]interface{}{
		"apt:debian:x64": "foo-debian-x64",
		"apt:debian":     "foo-debian",
		"apt:linux:x64":  "foo-linux-x64",
		"apt":            "foo-apt",
	}
	// Marshal/unmarshal to SoftwareEntry for PlanProvision
	b, _ := yaml.Marshal(entryMap)
	var entry app.SoftwareEntry
	_ = yaml.Unmarshal(b, &entry)
	manifest["foo"] = entry

	prov := NewProvisioner(&fakeSystemInfo{}, manifest, &fakeExecRunner{})
	prov.System = customSys{}
	prov.ManifestRaw = map[string]map[string]interface{}{"foo": entryMap}

	plan, err := prov.PlanProvision([]string{"foo"}, nil)
	if err != nil {
		t.Fatalf("PlanProvision error: %v", err)
	}
	if len(plan) != 1 {
		t.Fatalf("expected 1 plan entry, got %d", len(plan))
	}
	if plan[0].Package != "foo-debian-x64" {
		t.Errorf("expected foo-debian-x64, got %s", plan[0].Package)
	}

	// Remove the most specific, should fallback
	delete(entryMap, "apt:debian:x64")
	b, _ = yaml.Marshal(entryMap)
	_ = yaml.Unmarshal(b, &entry)
	manifest["foo"] = entry
	prov.ManifestRaw["foo"] = entryMap
	plan, err = prov.PlanProvision([]string{"foo"}, nil)
	if err != nil {
		t.Fatalf("PlanProvision error: %v", err)
	}
	if plan[0].Package != "foo-debian" {
		t.Errorf("expected foo-debian, got %s", plan[0].Package)
	}

	// Remove debian, should fallback to linux:x64
	delete(entryMap, "apt:debian")
	b, _ = yaml.Marshal(entryMap)
	_ = yaml.Unmarshal(b, &entry)
	manifest["foo"] = entry
	prov.ManifestRaw["foo"] = entryMap
	plan, err = prov.PlanProvision([]string{"foo"}, nil)
	if err != nil {
		t.Fatalf("PlanProvision error: %v", err)
	}
	if plan[0].Package != "foo-linux-x64" {
		t.Errorf("expected foo-linux-x64, got %s", plan[0].Package)
	}

	// Remove linux:x64, should fallback to apt
	delete(entryMap, "apt:linux:x64")
	b, _ = yaml.Marshal(entryMap)
	_ = yaml.Unmarshal(b, &entry)
	manifest["foo"] = entry
	prov.ManifestRaw["foo"] = entryMap
	plan, err = prov.PlanProvision([]string{"foo"}, nil)
	if err != nil {
		t.Fatalf("PlanProvision error: %v", err)
	}
	if plan[0].Package != "foo-apt" {
		t.Errorf("expected foo-apt, got %s", plan[0].Package)
	}
}

func TestPostInstall_FlatpakAndCask(t *testing.T) {
	home := os.TempDir()
	// Flatpak test
	flatpakEntry := map[string]interface{}{
		"flatpak":      "org.example.App",
		"_bin:flatpak": "myapp",
	}
	// Cask test
	caskEntry := map[string]interface{}{
		"cask":      "mycask",
		"_bin:cask": "mycaskbin",
		"_app:cask": "MyCaskApp.app",
	}
	manifest := app.Manifest{
		"flatpakapp": app.SoftwareEntry{},
		"caskapp":    app.SoftwareEntry{},
	}
	var flatpakSE app.SoftwareEntry
	b, _ := yaml.Marshal(flatpakEntry)
	_ = yaml.Unmarshal(b, &flatpakSE)
	manifest["flatpakapp"] = flatpakSE
	var caskSE app.SoftwareEntry
	b, _ = yaml.Marshal(caskEntry)
	_ = yaml.Unmarshal(b, &caskSE)
	manifest["caskapp"] = caskSE
	prov := NewProvisioner(&fakeSystemInfo{}, manifest, &fakeExecRunner{})
	prov.ManifestRaw = map[string]map[string]interface{}{
		"flatpakapp": flatpakEntry,
		"caskapp":    caskEntry,
	}
	// Set HOME for test
	if err := os.Setenv("HOME", home); err != nil {
		t.Fatalf("os.Setenv failed: %v", err)
	}
	// Simulate /Applications/MyCaskApp.app exists
	appDir := filepath.Join(home, "Applications")
	if err := os.MkdirAll(appDir, 0o755); err != nil {
		t.Fatalf("os.MkdirAll failed: %v", err)
	}
	appPath := filepath.Join(appDir, "MyCaskApp.app")
	f, err := os.Create(appPath)
	if err != nil {
		t.Fatalf("os.Create failed: %v", err)
	}
	if err2 := f.Close(); err2 != nil {
		t.Errorf("f.Close failed: %v", err2)
	}
	defer func() {
		if err3 := os.RemoveAll(appDir); err3 != nil {
			t.Errorf("os.RemoveAll failed: %v", err3)
		}
	}()

	runner := &fakeExecRunner{}
	prov.Runner = runner
	// Set SystemInfo for macOS for cask
	prov.System = macSys{}

	err = prov.PostInstall()
	if err != nil {
		t.Fatalf("PostInstall error: %v", err)
	}
	// Check flatpak wrapper commands
	flatpakBin := filepath.Join(home, ".local", "bin", "flatpak", "myapp")
	foundFlatpak := false
	foundCask := false
	for _, cmd := range runner.Commands {
		if strings.Contains(cmd, flatpakBin) && strings.Contains(cmd, "flatpak run org.example.App") {
			foundFlatpak = true
		}
		if strings.Contains(cmd, "mycaskbin") && strings.Contains(cmd, "open '") && strings.Contains(cmd, "MyCaskApp.app") {
			foundCask = true
		}
	}
	if !foundFlatpak {
		t.Errorf("Flatpak wrapper script not created: %v", runner.Commands)
	}
	if !foundCask {
		t.Errorf("Cask wrapper script not created: %v", runner.Commands)
	}
}

// Minimal realSystemRunner for integration test
// (matches the production logic for script execution)
type realSystemRunner struct{}

func (r *realSystemRunner) Run(cmd string, args ...string) error {
	if cmd == "script" && len(args) > 0 {
		script := args[0]
		tmpTmpl, err := os.CreateTemp("", "provision-script-tmpl-*.sh")
		if err != nil {
			return err
		}
		defer func() {
			_ = os.Remove(tmpTmpl.Name())
		}()

		homeDir, err := os.UserHomeDir()
		if err != nil {
			return err
		}

		chezCmd := exec.Command("chezmoi", "execute-template")
		chezCmd.Env = append(os.Environ(), "HOME="+homeDir)
		chezCmd.Stdin = strings.NewReader(script)
		out, err := chezCmd.Output()
		if err != nil {
			return err
		}
		if _, err2 := tmpTmpl.Write(out); err2 != nil {
			if err3 := tmpTmpl.Close(); err3 != nil {
				return err3
			}
			return err2
		}
		if err2 := tmpTmpl.Close(); err2 != nil {
			return err2
		}

		bashCmd := exec.Command("bash", tmpTmpl.Name())
		bashCmd.Stdout = os.Stdout
		bashCmd.Stderr = os.Stderr
		return bashCmd.Run()
	}
	c := exec.Command(cmd, args...)
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr
	return c.Run()
}
func (r *realSystemRunner) Output(cmd string, args ...string) ([]byte, error) {
	c := exec.Command(cmd, args...)
	return c.Output()
}

func TestRealSystemRunner_ChezmoiTemplateScript(t *testing.T) {
	// Check if chezmoi is installed
	if _, err := exec.LookPath("chezmoi"); err != nil {
		t.Skip("chezmoi not installed, skipping integration test")
	}
	// Use a very simple template
	script := "echo OS={{ .chezmoi.os }}"
	runner := &realSystemRunner{}
	// Capture output by redirecting os.Stdout temporarily
	origStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w
	err := runner.Run("script", script)
	if err2 := w.Close(); err2 != nil {
		t.Errorf("w.Close failed: %v", err2)
	}
	os.Stdout = origStdout
	if err != nil {
		t.Fatalf("realSystemRunner.Run error: %v", err)
	}
	outBytes := make([]byte, 1024)
	n, _ := r.Read(outBytes)
	output := string(outBytes[:n])
	osName := runtime.GOOS
	if !strings.Contains(output, "OS="+osName) {
		t.Errorf("expected output to contain OS=%s, got: %q", osName, output)
	}
}

//revive:disable:var-naming
func SkipTestExecutePlan_DryRun(t *testing.T) {
	//revive:enable:var-naming
	manifest := app.Manifest{
		"foo": app.SoftwareEntry{
			Apt: app.StringOrSlice{"foo"},
		},
		"bar": app.SoftwareEntry{
			Script: app.StringOrSlice{"echo bar"},
		},
	}
	runner := &fakeExecRunner{}
	prov := NewProvisioner(&fakeSystemInfo{}, manifest, runner)
	prov.DryRun = true
	plan := []InstallInstruction{
		{Type: "apt", Package: "foo"},
		{Type: "script", Package: "echo bar"},
	}
	err := prov.ExecutePlan(plan)
	if err != nil {
		t.Fatalf("ExecutePlan (dry run) error: %v", err)
	}
	if len(runner.Commands) != 0 {
		t.Errorf("expected no commands executed in dry run, got %d", len(runner.Commands))
	}
	want := []string{"apt foo", "script echo bar"}
	got := prov.DryRunCommands()
	if len(got) != len(want) {
		t.Fatalf("expected %d dry run commands, got %d", len(want), len(got))
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("dry run command %d: expected %q, got %q", i, want[i], got[i])
		}
	}
}

//revive:disable:var-naming
func SkipTestExecutePlan_ErrorAggregationAndLogFile(t *testing.T) {
	//revive:enable:var-naming
	manifest := app.Manifest{
		"foo": app.SoftwareEntry{
			Apt: app.StringOrSlice{"foo"},
		},
		"bar": app.SoftwareEntry{
			Script: app.StringOrSlice{"echo bar"},
		},
		"baz": app.SoftwareEntry{
			Apt: app.StringOrSlice{"baz"},
		},
	}
	tempLog := filepath.Join(os.TempDir(), "provision-test-log.txt")
	defer func() {
		if err := os.Remove(tempLog); err != nil {
			t.Errorf("os.Remove failed: %v", err)
		}
	}()
	runner := &errRunner{}
	prov := NewProvisioner(&fakeSystemInfo{}, manifest, runner)
	prov.LogFile = tempLog
	plan := []InstallInstruction{
		{Type: "apt", Package: "foo"},
		{Type: "script", Package: "echo bar"},
		{Type: "apt", Package: "baz"},
	}
	err := prov.ExecutePlan(plan)
	if err == nil {
		t.Fatalf("expected aggregated error, got nil")
	}
	if len(prov.Errors) != 2 {
		t.Errorf("expected 2 errors, got %d", len(prov.Errors))
	}
	agg := prov.AggregatedError()
	if agg == nil || !strings.Contains(agg.Error(), "2 errors occurred") {
		t.Errorf("expected aggregated error message, got %v", agg)
	}
	prov.ClearErrors()
	if len(prov.Errors) != 0 {
		t.Errorf("expected errors to be cleared")
	}
	// Check log file contents
	data, readErr := os.ReadFile(tempLog)
	if readErr != nil {
		t.Fatalf("failed to read log file: %v", readErr)
	}
	logStr := string(data)
	if !strings.Contains(logStr, "apt foo") || !strings.Contains(logStr, "[ERROR] fail foo") {
		t.Errorf("log file missing apt foo error: %q", logStr)
	}
	if !strings.Contains(logStr, "script echo bar") || !strings.Contains(logStr, "[ERROR] fail script") {
		t.Errorf("log file missing script error: %q", logStr)
	}
	if !strings.Contains(logStr, "apt baz") {
		t.Errorf("log file missing apt baz: %q", logStr)
	}
}

func TestProvisioner_shouldSkipInstalled(t *testing.T) {
	prov := NewProvisioner(nil, nil, nil)
	installed := map[string]bool{"foo": true, "bar": false}
	tests := []struct {
		key      string
		wantSkip bool
	}{
		{"foo", true},
		{"bar", false},
		{"baz", false},
	}
	for _, tt := range tests {
		if got := prov.shouldSkipInstalled(tt.key, installed); got != tt.wantSkip {
			t.Errorf("shouldSkipInstalled(%q) = %v, want %v", tt.key, got, tt.wantSkip)
		}
	}
}

type fakeSys struct{}

func (f *fakeSys) OS() string       { return "linux" }
func (f *fakeSys) Arch() string     { return "amd64" }
func (f *fakeSys) ID() string       { return "ubuntu" }
func (f *fakeSys) IsHeadless() bool { return true }

func TestProvisioner_shouldSkipHeadless(t *testing.T) {
	prov := NewProvisioner(&fakeSys{}, nil, nil)
	tests := []struct {
		entry    app.SoftwareEntry
		wantSkip bool
	}{
		{app.SoftwareEntry{App: "foo"}, true},
		{app.SoftwareEntry{App: ""}, false},
	}
	for _, tt := range tests {
		if got := prov.shouldSkipHeadless(&tt.entry); got != tt.wantSkip {
			t.Errorf("shouldSkipHeadless(%v) = %v, want %v", tt.entry, got, tt.wantSkip)
		}
	}
}

func TestProvisioner_shouldSkipLazy(t *testing.T) {
	prov := NewProvisioner(nil, nil, nil)
	prov.LazyOnly = true
	tests := []struct {
		entry    app.SoftwareEntry
		wantSkip bool
	}{
		{app.SoftwareEntry{Lazy: false}, true},
		{app.SoftwareEntry{Lazy: true}, false},
	}
	for _, tt := range tests {
		if got := prov.shouldSkipLazy(&tt.entry); got != tt.wantSkip {
			t.Errorf("shouldSkipLazy(%v) = %v, want %v", tt.entry, got, tt.wantSkip)
		}
	}
}

func TestProvisioner_addScriptInstructions(t *testing.T) {
	prov := NewProvisioner(nil, nil, nil)
	plan := []InstallInstruction{}
	entry := app.SoftwareEntry{Script: []string{"echo foo", "echo bar"}}
	prov.addScriptInstructions(&entry, &plan)
	if len(plan) != 2 {
		t.Fatalf("expected 2 script instructions, got %d", len(plan))
	}
	if plan[0].Type != "script" || plan[0].Package != "echo foo" {
		t.Errorf("unexpected first script instruction: %+v", plan[0])
	}
	if plan[1].Type != "script" || plan[1].Package != "echo bar" {
		t.Errorf("unexpected second script instruction: %+v", plan[1])
	}
}

func TestProvisioner_addInstallerInstruction(t *testing.T) {
	prov := NewProvisioner(nil, nil, nil)
	plan := []InstallInstruction{}
	entry := app.SoftwareEntry{}
	manifestRaw := map[string]map[string]interface{}{
		"foo": {
			"apt":  "foo-pkg",
			"brew": "foo-brew",
		},
	}
	prov.ManifestRaw = manifestRaw
	prov.addInstallerInstruction("foo", &entry, &plan)
	if len(plan) != 1 {
		t.Fatalf("expected 1 installer instruction, got %d", len(plan))
	}
	if plan[0].Type != "apt" || plan[0].Package != "foo-pkg" {
		t.Errorf("unexpected installer instruction: %+v", plan[0])
	}
}

// --- Additional direct tests for private helpers ---

func Test_getFieldByPriority(t *testing.T) {
	entry := map[string]interface{}{
		"foo:bar:baz:qux":  "val1",
		"foo:bar:baz":      "val2",
		"foo:bar:quux:qux": "val3",
		"foo:bar:quux":     "val4",
		"foo:bar:qux":      "val5",
		"foo:bar":          "val6",
		"foo":              "val7",
	}
	// Installer provided
	cases := []struct {
		name      string
		installer string
		osId      string
		osType    string
		osArch    string
		expect    string
		found     bool
	}{
		{"most specific", "bar", "baz", "quux", "qux", "val1", true},
		{"osId fallback", "bar", "baz", "quux", "nope", "val2", true},
		{"osType+osArch", "bar", "nope", "quux", "qux", "val3", true},
		{"osType fallback", "bar", "nope", "quux", "nope", "val4", true},
		{"osArch fallback", "bar", "nope", "nope", "qux", "val5", true},
		{"installer only", "bar", "nope", "nope", "nope", "val6", true},
		{"prefix only", "nope", "nope", "nope", "nope", "val7", true},
		{"not found", "nope", "nope", "nope", "nope", "val7", true}, // fallback to prefix
	}
	for _, c := range cases {
		got, ok := getFieldByPriority(entry, "foo", c.installer, c.osId, c.osType, c.osArch)
		if ok != c.found || got != c.expect {
			t.Errorf("%s: got (%q, %v), want (%q, %v)", c.name, got, ok, c.expect, c.found)
		}
	}
	// Array value
	entryArr := map[string]interface{}{
		"foo:bar": []interface{}{"arrval1", "arrval2"},
	}
	got, ok := getFieldByPriority(entryArr, "foo", "bar", "", "", "")
	if !ok || got != "arrval1" {
		t.Errorf("array value: got (%q, %v), want (arrval1, true)", got, ok)
	}
	// Type mismatch (not string or []interface{})
	entryBad := map[string]interface{}{
		"foo:bar": 123,
	}
	got, ok = getFieldByPriority(entryBad, "foo", "bar", "", "", "")
	if ok || got != "" {
		t.Errorf("type mismatch: got (%q, %v), want (\"\", false)", got, ok)
	}
}

// Mock runner to capture commands for wrapper helpers
type mockRunner struct{ cmds []string }

func (m *mockRunner) Run(cmd string, args ...string) error {
	m.cmds = append(m.cmds, cmd+" "+strings.Join(args, " "))
	return nil
}
func (m *mockRunner) Output(cmd string, args ...string) ([]byte, error) { return nil, nil }

func Test_handleFlatpakWrapper(t *testing.T) {
	prov := NewProvisioner(nil, nil, nil)
	runner := &mockRunner{}
	prov.Runner = runner
	osId, osType, osArch := "", "", ""
	// Valid case
	entry := map[string]interface{}{
		"flatpak":      "org.example.App",
		"_bin:flatpak": "myapp",
	}
	prov.handleFlatpakWrapper(entry, osId, osType, osArch)
	if len(runner.cmds) < 3 {
		t.Errorf("expected at least 3 commands, got %v", runner.cmds)
	}
	// Missing flatpak field
	runner.cmds = nil
	entry2 := map[string]interface{}{
		"_bin:flatpak": "myapp",
	}
	prov.handleFlatpakWrapper(entry2, osId, osType, osArch)
	if len(runner.cmds) != 0 {
		t.Errorf("expected no commands for missing flatpak, got %v", runner.cmds)
	}
	// Missing bin field
	runner.cmds = nil
	entry3 := map[string]interface{}{
		"flatpak": "org.example.App",
	}
	prov.handleFlatpakWrapper(entry3, osId, osType, osArch)
	if len(runner.cmds) != 0 {
		t.Errorf("expected no commands for missing bin, got %v", runner.cmds)
	}
}

func Test_handleCaskWrapper(t *testing.T) {
	prov := NewProvisioner(nil, nil, nil)
	runner := &mockRunner{}
	prov.Runner = runner
	osId, osType, osArch := "darwin", "darwin", "x64"
	// Set up temp HOME and Applications dir
	home := t.TempDir()
	if err := os.Setenv("HOME", home); err != nil {
		t.Fatalf("os.Setenv failed: %v", err)
	}
	appDir := filepath.Join(home, "Applications")
	if err := os.MkdirAll(appDir, 0o755); err != nil {
		t.Fatalf("os.MkdirAll failed: %v", err)
	}
	appName := "MyCaskApp.app"
	appPath := filepath.Join(appDir, appName)
	if err := os.WriteFile(appPath, []byte{}, 0o644); err != nil {
		t.Fatalf("os.WriteFile failed: %v", err)
	}
	entry := map[string]interface{}{
		"cask":      "mycask",
		"_bin:cask": "mycaskbin",
		"_app:cask": appName,
	}
	entrySE := &app.SoftwareEntry{}
	prov.handleCaskWrapper(entry, osId, osType, osArch, entrySE)
	if len(runner.cmds) < 3 {
		t.Errorf("expected at least 3 commands, got %v", runner.cmds)
	}
	// Missing cask and not darwin+App
	runner.cmds = nil
	entry2 := map[string]interface{}{
		"_bin:cask": "mycaskbin",
		"_app:cask": appName,
	}
	prov.handleCaskWrapper(entry2, "linux", "linux", "x64", &app.SoftwareEntry{})
	if len(runner.cmds) != 0 {
		t.Errorf("expected no commands for missing cask and not darwin+App, got %v", runner.cmds)
	}
	// Missing bin field
	runner.cmds = nil
	entry3 := map[string]interface{}{
		"cask":      "mycask",
		"_app:cask": appName,
	}
	prov.handleCaskWrapper(entry3, osId, osType, osArch, entrySE)
	if len(runner.cmds) != 0 {
		t.Errorf("expected no commands for missing bin, got %v", runner.cmds)
	}
	// Missing app field
	runner.cmds = nil
	entry4 := map[string]interface{}{
		"cask":      "mycask",
		"_bin:cask": "mycaskbin",
	}
	prov.handleCaskWrapper(entry4, osId, osType, osArch, entrySE)
	if len(runner.cmds) != 0 {
		t.Errorf("expected no commands for missing app, got %v", runner.cmds)
	}
	// App not found in either location
	runner.cmds = nil
	entry5 := map[string]interface{}{
		"cask":      "mycask",
		"_bin:cask": "mycaskbin",
		"_app:cask": "NotExist.app",
	}
	if err := os.RemoveAll(appDir); err != nil {
		t.Errorf("os.RemoveAll failed: %v", err)
	}
	prov.handleCaskWrapper(entry5, osId, osType, osArch, entrySE)
	if len(runner.cmds) != 0 {
		t.Errorf("expected no commands for app not found, got %v", runner.cmds)
	}
}

func Test_handleFlatpakWrapper_EdgeCases(t *testing.T) {
	prov := NewProvisioner(nil, nil, nil)
	runner := &mockRunner{}
	prov.Runner = runner
	osId, osType, osArch := "", "", ""
	// Empty bin field
	entry := map[string]interface{}{
		"flatpak":      "org.example.App",
		"_bin:flatpak": "",
	}
	prov.handleFlatpakWrapper(entry, osId, osType, osArch)
	if len(runner.cmds) != 0 {
		t.Errorf("expected no commands for empty bin, got %v", runner.cmds)
	}
	// Empty flatpak field
	runner.cmds = nil
	entry2 := map[string]interface{}{
		"flatpak":      "",
		"_bin:flatpak": "myapp",
	}
	prov.handleFlatpakWrapper(entry2, osId, osType, osArch)
	if len(runner.cmds) != 0 {
		t.Errorf("expected no commands for empty flatpak, got %v", runner.cmds)
	}
}

func Test_handleCaskWrapper_EdgeCases(t *testing.T) {
	prov := NewProvisioner(nil, nil, nil)
	runner := &mockRunner{}
	prov.Runner = runner
	osId, osType, osArch := "darwin", "darwin", "x64"
	// Set up temp HOME and Applications dir
	home := t.TempDir()
	if err := os.Setenv("HOME", home); err != nil {
		t.Fatalf("os.Setenv failed: %v", err)
	}
	appDir := filepath.Join(home, "Applications")
	if err := os.MkdirAll(appDir, 0o755); err != nil {
		t.Fatalf("os.MkdirAll failed: %v", err)
	}
	appName := "MyCaskApp.app"
	// Empty bin field
	entry := map[string]interface{}{
		"cask":      "mycask",
		"_bin:cask": "",
		"_app:cask": appName,
	}
	entrySE := &app.SoftwareEntry{}
	prov.handleCaskWrapper(entry, osId, osType, osArch, entrySE)
	if len(runner.cmds) != 0 {
		t.Errorf("expected no commands for empty bin, got %v", runner.cmds)
	}
	// Empty app field
	runner.cmds = nil
	entry2 := map[string]interface{}{
		"cask":      "mycask",
		"_bin:cask": "mycaskbin",
		"_app:cask": "",
	}
	prov.handleCaskWrapper(entry2, osId, osType, osArch, entrySE)
	if len(runner.cmds) != 0 {
		t.Errorf("expected no commands for empty app, got %v", runner.cmds)
	}
	// Simulate os.Stat failure (permission denied or other error)
	runner.cmds = nil
	entry3 := map[string]interface{}{
		"cask":      "mycask",
		"_bin:cask": "mycaskbin",
		"_app:cask": appName,
	}
	if err := os.RemoveAll(appDir); err != nil {
		t.Errorf("os.RemoveAll failed: %v", err)
	}
	prov.handleCaskWrapper(entry3, osId, osType, osArch, entrySE)
	if len(runner.cmds) != 0 {
		t.Errorf("expected no commands for os.Stat failure, got %v", runner.cmds)
	}
}

//revive:disable:var-naming
func SkipTestExecutePlan_AllPackageManagers(t *testing.T) {
	//revive:enable:var-naming
	cases := []struct {
		name     string
		instType string
		pkg      string
		expected string
	}{
		{"apt", "apt", "foo", "apt foo"},
		{"apk", "apk", "bar", "apk bar"},
		{"dnf", "dnf", "baz", "dnf baz"},
		{"yum", "yum", "qux", "yum qux"},
		{"zypper", "zypper", "zap", "zypper zap"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			manifest := app.Manifest{
				tc.pkg: app.SoftwareEntry{},
			}
			plan := []InstallInstruction{{Type: tc.instType, Package: tc.pkg}}
			runner := &fakeExecRunner{}
			prov := NewProvisioner(&fakeSystemInfo{}, manifest, runner)
			err := prov.ExecutePlan(plan)
			if err != nil {
				t.Fatalf("ExecutePlan error: %v", err)
			}
			if len(runner.Commands) != 1 {
				t.Errorf("expected 1 command executed, got %d", len(runner.Commands))
			}
			if !strings.HasPrefix(runner.Commands[0], tc.instType) {
				t.Errorf("expected command prefix '%s', got '%s'", tc.instType, runner.Commands[0])
			}
			if !strings.Contains(runner.Commands[0], tc.pkg) {
				t.Errorf("expected command to contain package '%s', got '%s'", tc.pkg, runner.Commands[0])
			}
		})
	}
}
