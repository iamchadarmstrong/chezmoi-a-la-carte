package provision

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"errors"

	"gopkg.in/yaml.v3"

	"a-la-carte/internal/app"
)

// SystemInfo abstracts OS and environment detection for testability.
//
// # Usage
//
//	sys := &RealSystemInfo{}
//	os := sys.OS()
type SystemInfo interface {
	OS() string
	Arch() string
	ID() string
	IsHeadless() bool
}

// ExecRunner abstracts command execution for testability.
//
// # Usage
//
//	runner := &RealExecRunner{}
//	err := runner.Run("echo", "hello")
type ExecRunner interface {
	Run(cmd string, args ...string) error
	Output(cmd string, args ...string) ([]byte, error)
}

// Provisioner is the main struct for provisioning logic.
//
// # Fields
//   - System:   Provides system/OS info
//   - Manifest: The loaded software manifest
//   - ManifestRaw: The raw manifest map for advanced key matching (optional)
//   - Runner:   Executes system commands
//   - InstallerOrder: Preferred order of installer types (overrides default)
//   - LazyOnly: If true, only install packages with Lazy=true
//   - DryRun:   If true, do not actually run commands, just log them
//   - DryRunLog: Stores dry run log entries
//   - Errors:   Aggregated errors from last ExecutePlan
//   - LogFile:  If set, logs all command attempts and errors to this file
type Provisioner struct {
	System         SystemInfo
	Manifest       app.Manifest
	ManifestRaw    map[string]map[string]interface{} // Raw manifest for advanced key matching
	Runner         ExecRunner
	InstallerOrder []string // Preferred order of installer types
	LazyOnly       bool     // Only install packages with Lazy=true
	DryRun         bool     // If true, do not actually run commands, just log them
	DryRunLog      []string // Stores dry run log entries
	Errors         []error  // Aggregated errors from last ExecutePlan
	LogFile        string   // If set, logs all command attempts and errors to this file
}

// InstallInstruction represents a single install/provision action.
//
// # Fields
//   - Type:    The installer type (e.g., "apt", "brew")
//   - Package: The package name to install
type InstallInstruction struct {
	Type    string // e.g. "apt", "brew", etc.
	Package string
}

// NewProvisioner creates a new Provisioner with the given dependencies.
//
// # Parameters
//   - sys:      SystemInfo implementation
//   - manifest: The loaded manifest
//   - runner:   ExecRunner implementation
//
// # Returns
//   - *Provisioner: The new provisioner instance
func NewProvisioner(sys SystemInfo, manifest app.Manifest, runner ExecRunner) *Provisioner {
	return &Provisioner{
		System:   sys,
		Manifest: manifest,
		Runner:   runner,
	}
}

// getFieldByPriority returns the value for a manifest field with advanced key matching.
// It supports keys like prefix:installer:osId:osArch, etc, with fallback order as in installx.js.
func getFieldByPriority(entry map[string]interface{}, prefix, installer, osId, osType, osArch string) (string, bool) {
	if installer != "" {
		keys := []string{
			prefix + ":" + installer + ":" + osId + ":" + osArch,
			prefix + ":" + installer + ":" + osId,
			prefix + ":" + installer + ":" + osType + ":" + osArch,
			prefix + ":" + installer + ":" + osType,
			prefix + ":" + installer + ":" + osArch,
			prefix + ":" + installer,
			prefix,
		}
		for _, k := range keys {
			if v, ok := entry[k]; ok {
				if s, ok := v.(string); ok {
					return s, true
				}
				if arr, ok := v.([]interface{}); ok && len(arr) > 0 {
					if s, ok := arr[0].(string); ok {
						return s, true
					}
				}
			}
		}
	} else {
		keys := []string{
			prefix + ":" + osId + ":" + osArch,
			prefix + ":" + osId,
			prefix + ":" + osType + ":" + osArch,
			prefix + ":" + osType,
			prefix + ":" + osArch,
			prefix,
		}
		for _, k := range keys {
			if v, ok := entry[k]; ok {
				if s, ok := v.(string); ok {
					return s, true
				}
				if arr, ok := v.([]interface{}); ok && len(arr) > 0 {
					if s, ok := arr[0].(string); ok {
						return s, true
					}
				}
			}
		}
	}
	return "", false
}

func (p *Provisioner) shouldSkipInstalled(key string, installed map[string]bool) bool {
	return installed != nil && installed[key]
}

func (p *Provisioner) shouldSkipHeadless(entry *app.SoftwareEntry) bool {
	return p.System != nil && p.System.IsHeadless() && entry.App != ""
}

func (p *Provisioner) shouldSkipLazy(entry *app.SoftwareEntry) bool {
	return p.LazyOnly && !entry.Lazy
}

func (p *Provisioner) addScriptInstructions(entry *app.SoftwareEntry, plan *[]InstallInstruction) {
	for _, script := range entry.Script {
		*plan = append(*plan, InstallInstruction{
			Type:    "script",
			Package: script,
		})
	}
}

func (p *Provisioner) addInstallerInstruction(key string, entry *app.SoftwareEntry, plan *[]InstallInstruction) {
	installerOrder := p.InstallerOrder
	if len(installerOrder) == 0 {
		installerOrder = []string{
			"apt", "brew", "pacman", "apk", "dnf", "zypper", "scoop", "choco", "go", "cargo", "pipx", "cask", "flatpak", "snap", "port", "yay", "pkg", "emerge", "nix", "mas", "xbps", "binary:darwin", "binary:linux", "binary:windows",
		}
	}
	var entryMap map[string]interface{}
	if p.ManifestRaw != nil {
		entryMap = p.ManifestRaw[key]
	} else {
		entryMap = make(map[string]interface{})
		b, _ := yaml.Marshal(entry)
		_ = yaml.Unmarshal(b, &entryMap)
	}
	for _, instType := range installerOrder {
		osId, osType, osArch := "", "", ""
		if p.System != nil {
			osId = p.System.ID()
			osType = p.System.OS()
			osArch = p.System.Arch()
		}
		if val, ok := getFieldByPriority(entryMap, instType, "", osId, osType, osArch); ok {
			// Patch: For apt and similar, only use the last word if value contains spaces
			pkg := val
			if (instType == "apt" || instType == "apk" || instType == "dnf" || instType == "zypper" || instType == "yum") && strings.Contains(val, " ") {
				fields := strings.Fields(val)
				pkg = fields[len(fields)-1]
			}
			*plan = append(*plan, InstallInstruction{
				Type:    instType,
				Package: pkg,
			})
			break
		}
	}
}

// expandDeps recursively expands dependencies for the given keys.
func (p *Provisioner) expandDeps(keys []string, visited map[string]bool) ([]string, error) {
	var result []string
	for _, key := range keys {
		if visited[key] {
			continue
		}
		visited[key] = true
		entry, ok := p.Manifest[key]
		if !ok {
			return nil, fmt.Errorf("manifest key not found: %s", key)
		}
		if len(entry.Deps) > 0 {
			depsExpanded, err := p.expandDeps(entry.Deps, visited)
			if err != nil {
				return nil, err
			}
			result = append(result, depsExpanded...)
		}
		result = append(result, key)
	}
	return result, nil
}

// planForKey adds install instructions for a single key if not skipped.
func (p *Provisioner) planForKey(key string, installed map[string]bool, plan *[]InstallInstruction) error {
	entry, ok := p.Manifest[key]
	if !ok {
		return fmt.Errorf("manifest key not found: %s", key)
	}
	if p.shouldSkipInstalled(key, installed) {
		if p.Runner != nil {
			_ = p.Runner.Run("info", fmt.Sprintf("Skipping %s: already installed", key))
		}
		return nil
	}
	if p.shouldSkipHeadless(&entry) {
		if p.Runner != nil {
			_ = p.Runner.Run("info", fmt.Sprintf("Skipping %s: headless mode", key))
		}
		return nil
	}
	if p.shouldSkipLazy(&entry) {
		if p.Runner != nil {
			_ = p.Runner.Run("info", fmt.Sprintf("Skipping %s: not marked lazy", key))
		}
		return nil
	}
	p.addScriptInstructions(&entry, plan)
	p.addInstallerInstruction(key, &entry, plan)
	return nil
}

func (p *Provisioner) PlanProvision(keys []string, installed map[string]bool) ([]InstallInstruction, error) {
	if p.Runner != nil {
		_ = p.Runner.Run("section", "Planning")
	}
	var plan []InstallInstruction
	visited := make(map[string]bool)
	expandedKeys, err := p.expandDeps(keys, visited)
	if err != nil {
		return nil, err
	}
	for _, key := range expandedKeys {
		err := p.planForKey(key, installed, &plan)
		if err != nil {
			return nil, err
		}
	}
	// Log planned installs
	if p.Runner != nil {
		for _, inst := range plan {
			_ = p.Runner.Run("info", fmt.Sprintf("Will install: %s %s", inst.Type, inst.Package))
		}
	}
	return plan, nil
}

// ExecutePlan executes the given install/provision instructions.
//
// # Parameters
//   - plan: The list of install instructions to execute
//
// # Returns
//   - error: If any error occurs (aggregated)
func (p *Provisioner) ExecutePlan(plan []InstallInstruction) error {
	if len(plan) == 0 {
		return nil
	}
	// Section header: Installing
	if p.Runner != nil {
		_ = p.Runner.Run("section", "Installing")
	}
	var errs []error
	for _, inst := range plan {
		logLine := inst.Type + " " + inst.Package
		if p.DryRun {
			p.DryRunLog = append(p.DryRunLog, logLine)
			continue
		}
		var err error
		if inst.Type == "script" {
			err = p.Runner.Run("script", inst.Package)
		} else {
			switch inst.Type {
			case "apt", "apk", "dnf", "zypper", "yum":
				err = p.Runner.Run(inst.Type, inst.Package)
			case "brew":
				err = p.Runner.Run("brew", "install", inst.Package)
			case "go":
				err = p.Runner.Run("go", "install", inst.Package)
			default:
				err = p.Runner.Run(inst.Type, inst.Package)
			}
		}
		if err != nil {
			errs = append(errs, err)
		}
	}
	// Section header: Complete
	if p.Runner != nil {
		_ = p.Runner.Run("section", "Complete")
	}
	if len(errs) > 0 {
		return errors.Join(errs...)
	}
	return nil
}

// AggregatedError returns a single error representing all errors from last ExecutePlan, or nil.
func (p *Provisioner) AggregatedError() error {
	if len(p.Errors) == 0 {
		return nil
	}
	msg := ""
	for i, err := range p.Errors {
		msg += fmt.Sprintf("[%d] %v\n", i+1, err)
	}
	return fmt.Errorf("%d errors occurred:\n%s", len(p.Errors), msg)
}

// ClearErrors clears the error list.
func (p *Provisioner) ClearErrors() {
	p.Errors = nil
}

// DryRunCommands returns the list of commands that would be run in dry run mode.
func (p *Provisioner) DryRunCommands() []string {
	return append([]string(nil), p.DryRunLog...)
}

// PostInstall performs post-install hooks (e.g., flatpak/cask symlinks/wrappers).
// For flatpak: creates ~/.local/bin/flatpak/<bin> wrappers that run flatpak run <app-id> $*
// For cask: creates ~/.local/bin/cask/<bin> wrappers that run open <app-path> $*
func (p *Provisioner) PostInstall() error {
	osId, osType, osArch := "", "", ""
	if p.System != nil {
		osId = p.System.ID()
		osType = p.System.OS()
		osArch = p.System.Arch()
	}
	for key := range p.Manifest {
		entry := p.Manifest[key]
		entryPtr := &entry
		var entryMap map[string]interface{}
		if p.ManifestRaw != nil {
			entryMap = p.ManifestRaw[key]
		} else {
			entryMap = make(map[string]interface{})
			b, _ := yaml.Marshal(entryPtr)
			_ = yaml.Unmarshal(b, &entryMap)
		}
		p.handleFlatpakWrapper(entryMap, osId, osType, osArch)
		p.handleCaskWrapper(entryMap, osId, osType, osArch, entryPtr)
	}
	return nil
}

func (p *Provisioner) handleFlatpakWrapper(entryMap map[string]interface{}, osId, osType, osArch string) {
	val, ok := getFieldByPriority(entryMap, "flatpak", "", osId, osType, osArch)
	if !ok || val == "" {
		return
	}
	bin, ok := getFieldByPriority(entryMap, "_bin", "flatpak", osId, osType, osArch)
	if !ok || bin == "" {
		return
	}
	appId := val
	binDir := filepath.Join(os.Getenv("HOME"), ".local", "bin", "flatpak")
	binPath := filepath.Join(binDir, bin)
	_ = p.Runner.Run("mkdir", "-p", binDir)
	cmd := "echo '#!/usr/bin/env bash\\nflatpak run " + appId + " $*' > '" + binPath + "'"
	_ = p.Runner.Run("sh", "-c", cmd)
	_ = p.Runner.Run("chmod", "+x", binPath)
}

func (p *Provisioner) handleCaskWrapper(entryMap map[string]interface{}, osId, osType, osArch string, entry *app.SoftwareEntry) {
	if _, ok := getFieldByPriority(entryMap, "cask", "", osId, osType, osArch); !ok && !(osId == "darwin" && entry.App != "") {
		return
	}
	bin, ok := getFieldByPriority(entryMap, "_bin", "cask", osId, osType, osArch)
	if !ok || bin == "" {
		return
	}
	appName, ok := getFieldByPriority(entryMap, "_app", "cask", osId, osType, osArch)
	if !ok || appName == "" {
		return
	}
	binDir := filepath.Join(os.Getenv("HOME"), ".local", "bin", "cask")
	binPath := filepath.Join(binDir, bin)
	appPath := "/Applications/" + appName
	if _, err := os.Stat(appPath); os.IsNotExist(err) {
		homeAppPath := filepath.Join(os.Getenv("HOME"), "Applications", appName)
		if _, err := os.Stat(homeAppPath); err == nil {
			appPath = homeAppPath
		} else {
			return
		}
	}
	_ = p.Runner.Run("mkdir", "-p", binDir)
	cmd := "echo '#!/usr/bin/env bash\\nopen '" + appPath + "' $*' > '" + binPath + "'"
	_ = p.Runner.Run("sh", "-c", cmd)
	_ = p.Runner.Run("chmod", "+x", binPath)
}
