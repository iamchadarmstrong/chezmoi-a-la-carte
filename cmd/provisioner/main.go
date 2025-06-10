package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"time"

	"a-la-carte/internal/app"
	"a-la-carte/internal/app/provision"
	"a-la-carte/internal/ui/core" // Changed from "a-la-carte/internal/ui"

	"flag"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

const logPanelHeight = 20

// logEntry represents a single log line with a level.
type logEntry struct {
	Level string // "info", "success", "error"
	Text  string
}

type logMsg logEntry

type doneMsg struct{}

type quitNowMsg struct{}

// Add spinner to model
type model struct {
	logs         []logEntry
	status       string
	cursor       int // for scrolling
	logChan      chan tea.Msg
	ready        bool
	userScrolled bool // track if user has scrolled up
	spinner      spinner.Model
	// For summary
	attempted  int
	succeeded  int
	failed     int
	failedPkgs []string
	// CLI flags for provisioning
	all      bool
	lazy     bool
	manifest string
	dryRun   bool
	groups   []string
	only     []string
}

func initialModel() *model {
	sp := spinner.New()
	sp.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("#7dcfff"))
	return &model{
		logs:    []logEntry{},
		status:  "Ready to provision...",
		cursor:  0,
		logChan: make(chan tea.Msg, 100),
		ready:   false,
		spinner: sp,
	}
}

// tuiExecRunner implements provision.ExecRunner and sends logs as tea.Msgs.
type tuiExecRunner struct {
	dispatch func(logMsg)
}

// Utility to strip ANSI codes
func stripANSI(input string) string {
	ansi := regexp.MustCompile(`\x1b\[[0-9;]*[a-zA-Z]`)
	return ansi.ReplaceAllString(input, "")
}

// Helper to construct exec.Cmd and log message for a given command
func buildExecCmd(cmd string, args ...string) (c *exec.Cmd, logMsgStr string) {
	switch cmd {
	case "apt":
		aptArgs := []string{"-o", "DPkg::Options::=--force-confdef", "install", "-y", "--no-install-recommends", "--ignore-missing"}
		aptArgs = append(aptArgs, args...)
		fullCmd := append([]string{"env", "DEBIAN_FRONTEND=noninteractive", "apt-get"}, aptArgs...)
		logMsgStr = "sudo " + strings.Join(fullCmd, " ")
		c = exec.Command("sudo", fullCmd...)
	case "apk":
		apkArgs := append([]string{"add", "--no-cache"}, args...)
		logMsgStr = "sudo apk " + strings.Join(apkArgs, " ")
		c = exec.Command("sudo", append([]string{"apk"}, apkArgs...)...)
	case "dnf", "yum":
		pmArgs := append([]string{"install", "-y", "--setopt=skip_if_unavailable=True", "--setopt=skip_missing_names_on_install=True"}, args...)
		logMsgStr = "sudo " + cmd + " " + strings.Join(pmArgs, " ")
		c = exec.Command("sudo", append([]string{cmd}, pmArgs...)...)
	case "zypper":
		zypperArgs := append([]string{"--non-interactive", "install", "-y"}, args...)
		logMsgStr = "sudo zypper " + strings.Join(zypperArgs, " ")
		c = exec.Command("sudo", append([]string{"zypper"}, zypperArgs...)...)
	default:
		logMsgStr = cmd + " " + strings.Join(args, " ")
		c = exec.Command(cmd, args...)
	}
	return c, logMsgStr
}

// Helper to stream output from stdout/stderr and dispatch log messages
func streamOutput(stdout, stderr io.ReadCloser, dispatch func(logMsg)) {
	done := make(chan struct{}, 2)
	go func() {
		scanner := bufio.NewScanner(stdout)
		for scanner.Scan() {
			line := stripANSI(scanner.Text())
			if strings.TrimSpace(line) != "" {
				dispatch(logMsg{Level: "info", Text: line})
			}
		}
		done <- struct{}{}
	}()
	go func() {
		scanner := bufio.NewScanner(stderr)
		for scanner.Scan() {
			line := stripANSI(scanner.Text())
			if strings.TrimSpace(line) != "" {
				dispatch(logMsg{Level: "info2", Text: line})
			}
		}
		done <- struct{}{}
	}()
	<-done
	<-done
}

func (r *tuiExecRunner) Run(cmd string, args ...string) error {
	if cmd == "section" && len(args) > 0 {
		r.dispatch(logMsg{Level: "section", Text: args[0]})
		return nil
	}
	if cmd == "info" && len(args) > 0 {
		r.dispatch(logMsg{Level: "info", Text: args[0]})
		return nil
	}

	c, logMsgStr := buildExecCmd(cmd, args...)
	r.dispatch(logMsg{Level: "info", Text: logMsgStr})

	stdout, err := c.StdoutPipe()
	if err != nil {
		r.dispatch(logMsg{Level: "error", Text: "Failed to get stdout: " + err.Error()})
		return err
	}
	stderr, err := c.StderrPipe()
	if err != nil {
		r.dispatch(logMsg{Level: "error", Text: "Failed to get stderr: " + err.Error()})
		return err
	}
	if startErr := c.Start(); startErr != nil {
		r.dispatch(logMsg{Level: "error", Text: "Failed to start command: " + startErr.Error()})
		return startErr
	}
	streamOutput(stdout, stderr, r.dispatch)
	err = c.Wait()
	if err != nil {
		r.dispatch(logMsg{Level: "error", Text: fmt.Sprintf("Error: %s: %v", logMsgStr, err)})
		return err
	}
	r.dispatch(logMsg{Level: "success", Text: fmt.Sprintf("Success: %s", logMsgStr)})
	return nil
}

func (r *tuiExecRunner) Output(cmd string, args ...string) ([]byte, error) {
	msg := fmt.Sprintf("Output: %s %s", cmd, strings.Join(args, " "))
	r.dispatch(logMsg{Level: "info", Text: msg})
	return []byte("output"), nil
}

// realSystemRunner implements provision.ExecRunner using os/exec (no logging, real output)
type realSystemRunner struct{}

func (r *realSystemRunner) Run(cmd string, args ...string) error {
	if cmd == "section" || cmd == "info" {
		return nil
	}
	if cmd == "script" && len(args) > 0 {
		script := args[0]
		tmpRaw, err := os.CreateTemp("", "provision-script-raw-*.sh")
		if err != nil {
			return err
		}
		defer func() {
			_ = os.Remove(tmpRaw.Name())
		}()
		if _, err2 := tmpRaw.WriteString(script); err2 != nil {
			_ = tmpRaw.Close()
			return err2
		}
		if err2 := tmpRaw.Close(); err2 != nil {
			return err2
		}

		tmpTmpl, err := os.CreateTemp("", "provision-script-tmpl-*.sh")
		if err != nil {
			return err
		}
		defer func() {
			_ = os.Remove(tmpTmpl.Name())
		}()

		// Process through chezmoi execute-template
		chezCmd := exec.Command("chezmoi", "execute-template", tmpRaw.Name())
		out, err := chezCmd.Output()
		if err != nil {
			return err
		}
		if _, err2 := tmpTmpl.Write(out); err2 != nil {
			_ = tmpTmpl.Close()
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

// getInstalledPackages returns a map of installed package keys. For now, returns an empty map (stub).
// func getInstalledPackages() map[string]bool {
// 	// TODO: Implement real detection logic for installed packages
// 	return map[string]bool{}
// }

func initialModelWithFlags(all, lazy bool, manifestPath string, dryRun bool, groups, only []string) *model {
	m := initialModel()
	m.all = all
	m.lazy = lazy
	m.manifest = manifestPath
	m.dryRun = dryRun
	m.groups = groups
	m.only = only
	return m
}

type tickMsg time.Time

func (m *model) Init() tea.Cmd {
	// Start the provisioning goroutine
	go func() {
		manifest, err := app.LoadManifest(m.manifest)
		if err != nil {
			m.logChan <- logMsg{Level: "error", Text: fmt.Sprintf("Failed to load manifest: %v", err)}
			m.logChan <- doneMsg{}
			return
		}
		var keys []string
		switch {
		case len(m.only) > 0:
			keys = m.only
		case len(m.groups) > 0:
			for k := range manifest {
				entry := manifest[k]
				entryPtr := &entry
				for _, g := range entryPtr.Groups {
					for _, want := range m.groups {
						if g == want {
							keys = append(keys, k)
							break
						}
					}
				}
			}
		default:
			for k := range manifest {
				keys = append(keys, k)
			}
		}
		var runner provision.ExecRunner
		if m.dryRun {
			runner = &dryRunRunner{}
		} else {
			runner = &realSystemRunner{}
		}
		installed := provision.GetInstalledPackages(runner)
		dispatch := func(msg logMsg) { m.logChan <- msg }
		prov := provision.NewProvisioner(nil, manifest, &tuiExecRunner{dispatch: dispatch})
		prov.LazyOnly = m.lazy
		dispatch(logMsg{Level: "info", Text: "Starting provisioning..."})
		dispatch(logMsg{Level: "info", Text: "Planning..."})
		plan, err := prov.PlanProvision(keys, installed)
		if err != nil {
			dispatch(logMsg{Level: "error", Text: fmt.Sprintf("Failed to plan provision: %v", err)})
			m.logChan <- doneMsg{}
			return
		}
		if len(plan) == 0 {
			dispatch(logMsg{Level: "info", Text: "Nothing to install. All requested packages are already installed or filtered out."})
		}
		dispatch(logMsg{Level: "info", Text: "Installing..."})
		err = prov.ExecutePlan(plan)
		if err != nil {
			dispatch(logMsg{Level: "error", Text: fmt.Sprintf("Provisioning failed: %v", err)})
		} else {
			dispatch(logMsg{Level: "success", Text: "Provisioning complete"})
		}
		m.logChan <- doneMsg{}
	}()
	// Start the ticker for live updates
	return tea.Tick(50*time.Millisecond, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

func (m *model) handleKeyMsg(msg tea.KeyMsg) (*model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c", "q":
		return m, tea.Quit
	case "up", "k":
		if m.cursor > 0 {
			m.cursor--
			m.userScrolled = true
		}
	case "down", "j":
		if m.cursor < len(m.logs)-logPanelHeight {
			m.cursor++
			if m.cursor >= len(m.logs)-logPanelHeight {
				m.userScrolled = false
			}
		}
	case "end":
		m.cursor = len(m.logs) - logPanelHeight
		if m.cursor < 0 {
			m.cursor = 0
		}
		m.userScrolled = false
	}
	return m, nil
}

func (m *model) handleLogMsg(msg logMsg) *model {
	m.logs = append(m.logs, logEntry(msg))
	if msg.Text == "Planning..." || msg.Text == "Installing..." {
		m.status = msg.Text
	}
	switch msg.Level {
	case "section":
		// No-op for summary
	case "success":
		if strings.Contains(msg.Text, "Installed") {
			m.succeeded++
			m.attempted++
		}
	case "error":
		if strings.Contains(msg.Text, "Failed to install") {
			m.failed++
			m.attempted++
			// Extract package name from msg.Text if possible
			parts := strings.Fields(msg.Text)
			if len(parts) > 3 {
				m.failedPkgs = append(m.failedPkgs, parts[3])
			}
		}
	}
	if !m.userScrolled {
		m.cursor = len(m.logs) - logPanelHeight
		if m.cursor < 0 {
			m.cursor = 0
		}
	}
	return m
}

func (m *model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		newModel, _ := m.handleKeyMsg(msg)
		return newModel, nil
	case logMsg:
		newModel := m.handleLogMsg(msg)
		return newModel, nil
	case tickMsg:
		cmds := []tea.Cmd{}
		var spinnerCmd tea.Cmd
		m.spinner, spinnerCmd = m.spinner.Update(nil)
		cmds = append(cmds, spinnerCmd)
		select {
		case lm := <-m.logChan:
			switch lm := lm.(type) {
			case logMsg:
				newModel := m.handleLogMsg(lm)
				return newModel, tea.Batch(append(cmds, tea.Tick(50*time.Millisecond, func(t time.Time) tea.Msg { return tickMsg(t) }))...)
			case doneMsg:
				return m, tea.Batch(append(cmds, tea.Tick(2*time.Second, func(time.Time) tea.Msg { return quitNowMsg{} }))...)
			default:
				return m, tea.Batch(append(cmds, tea.Tick(50*time.Millisecond, func(t time.Time) tea.Msg { return tickMsg(t) }))...)
			}
		default:
			return m, tea.Batch(append(cmds, tea.Tick(50*time.Millisecond, func(t time.Time) tea.Msg { return tickMsg(t) }))...)
		}
	case doneMsg:
		return m, tea.Tick(2*time.Second, func(time.Time) tea.Msg { return quitNowMsg{} })
	case quitNowMsg:
		return m, tea.Quit
	default:
		return m, nil
	}
}

// Helper to render log lines
func renderLogLines(logs []logEntry, start, end int) string {
	var b strings.Builder
	currentStyles := core.CurrentStyles() // Added
	currentTheme := core.CurrentTheme()   // Added

	for _, entry := range logs[start:end] {
		var style lipgloss.Style
		var prefix string
		switch entry.Level {
		case "section":
			// Check if section headers should be shown and if the current entry is not "Complete"
			if currentTheme.ShowSectionHeaders() && entry.Text != "Complete" { // Changed ui.CurrentTheme() to currentTheme
				style = currentStyles.HeaderStyle.Bold(true).Underline(true).Align(lipgloss.Left) // Changed ui.CurrentStyles() to currentStyles
				prefix = ""
				b.WriteString(style.Render(entry.Text) + "\\n")
			}
			continue
		case "error":
			style = currentStyles.ErrorStyle // Changed ui.ErrorStyle() to currentStyles.ErrorStyle
			prefix = "✖ "
		case "success":
			style = currentStyles.ItemStyle.Foreground(currentTheme.Accent()) // Changed ui.SuccessStyle() to use currentStyles and currentTheme
			prefix = "✔ "
		case "info2":
			style = currentStyles.ItemStyle.Foreground(currentTheme.TextMuted()) // Changed ui.InfoStyle() to use currentStyles and currentTheme
			prefix = "ℹ️  "
		case "info":
			style = currentStyles.ItemStyle.Foreground(currentTheme.TextMuted()) // Changed ui.InfoStyle() to use currentStyles and currentTheme
			prefix = "  "                                                        // two spaces for emoji alignment
		default:
			style = currentStyles.DimStyle // Changed ui.MutedTextStyle() to currentStyles.DimStyle
			prefix = "  "
		}
		b.WriteString(style.Render(prefix+entry.Text) + "\\n")
	}
	return b.String()
}

// Helper to render the status bar
func renderStatusBar(m *model) string {
	var statusBar strings.Builder
	currentStyles := core.CurrentStyles() // Added
	currentTheme := core.CurrentTheme()   // Added

	switch {
	case m.status == "Done":
		statusBar.WriteString(currentStyles.FooterStyle.Foreground(currentTheme.Accent()).Render("✔ Provisioning complete!")) // Changed
		statusBar.WriteString("\\n")
		statusBar.WriteString(currentStyles.FooterStyle.Render( // Changed
			fmt.Sprintf("Attempted: %d  Succeeded: %d  Failed: %d", m.attempted, m.succeeded, m.failed)))
		if m.failed > 0 {
			statusBar.WriteString("\\n" + currentStyles.FooterStyle.Foreground(currentTheme.Secondary()).Render("Failed packages: ")) // Changed
			statusBar.WriteString(strings.Join(m.failedPkgs, ", "))
		}
	case strings.Contains(m.status, "Failed") || strings.Contains(m.status, "error"):
		statusBar.WriteString(currentStyles.FooterStyle.Foreground(currentTheme.Secondary()).Render("✖ Provisioning failed!")) // Changed
		statusBar.WriteString("\\n" + currentStyles.FooterStyle.Render(m.status))                                              // Changed
		if m.failed > 0 {
			statusBar.WriteString("\\n" + currentStyles.FooterStyle.Foreground(currentTheme.Secondary()).Render("Failed packages: ")) // Changed
			statusBar.WriteString(strings.Join(m.failedPkgs, ", "))
		}
	default:
		// Animated spinner during provisioning
		statusBar.WriteString(currentStyles.FooterStyle.Render(m.spinner.View() + " " + m.status)) // Changed
	}
	// Keyboard shortcut help (only show when not done)
	if m.status != "Done" && !strings.Contains(m.status, "Failed") && !strings.Contains(m.status, "error") {
		statusBar.WriteString("\\n[q] quit  [↑/↓] scroll")
	}
	return statusBar.String()
}

func (m *model) View() string {
	var b strings.Builder
	maxLines := logPanelHeight
	start := m.cursor
	if start > len(m.logs)-maxLines {
		start = len(m.logs) - maxLines
	}
	if start < 0 {
		start = 0
	}
	end := start + maxLines
	if end > len(m.logs) {
		end = len(m.logs)
	}
	b.WriteString(renderLogLines(m.logs, start, end))
	// Pad with empty lines if not enough logs
	for i := end - start; i < maxLines; i++ {
		b.WriteString("\n")
	}
	b.WriteString("\n" + renderStatusBar(m))
	return b.String()
}

// ensureSudo prompts for sudo password up front and caches credentials.
func ensureSudo() {
	cmd := exec.Command("sudo", "-v")
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	_ = cmd.Run()
}

func main() {
	core.RegisterTheme("default", core.DefaultTheme{}) // Changed ui.RegisterTheme and ui.DefaultTheme
	ensureSudo()
	// CLI flag parsing
	allFlag := flag.Bool("all", false, "Install all packages (ignores selection)")
	allFlagShort := flag.Bool("a", false, "Alias for --all")
	lazyFlag := flag.Bool("lazy", false, "Only install packages with lazy=true")
	lazyFlagShort := flag.Bool("l", false, "Alias for --lazy")
	noTUIFlag := flag.Bool("no-tui", false, "Run in headless mode (no TUI, just logs to stdout)")
	manifestFlag := flag.String("manifest", "data/package_manifest.yaml", "Path to the manifest YAML file")
	dryRunFlag := flag.Bool("dry-run", false, "Print commands instead of running them (safe for tests)")
	groupFlag := flag.String("group", "", "Only install packages in this group (comma-separated, e.g. dev,ops)")
	onlyFlag := flag.String("only", "", "Only install the specified packages (comma-separated, e.g. foo,bar)")
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s [--all|-a] [--lazy|-l] [--no-tui] [--manifest <file>] [--dry-run] [--group <name>[,<name2>...]] [--only <pkg1>[,<pkg2>...]]\n", os.Args[0])
		flag.PrintDefaults()
	}
	flag.Parse()

	all := *allFlag || *allFlagShort
	lazy := *lazyFlag || *lazyFlagShort
	noTUI := *noTUIFlag
	manifestPath := *manifestFlag
	dryRun := *dryRunFlag

	// Parse group/only flags
	var groups []string
	if *groupFlag != "" {
		for _, g := range strings.Split(*groupFlag, ",") {
			g = strings.TrimSpace(g)
			if g != "" {
				groups = append(groups, g)
			}
		}
	}
	var only []string
	if *onlyFlag != "" {
		for _, o := range strings.Split(*onlyFlag, ",") {
			o = strings.TrimSpace(o)
			if o != "" {
				only = append(only, o)
			}
		}
	}

	if noTUI {
		headlessMain(lazy, manifestPath, dryRun, groups, only)
		return
	}

	p := tea.NewProgram(initialModelWithFlags(all, lazy, manifestPath, dryRun, groups, only))
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error running provision TUI: %v\n", err)
		os.Exit(1)
	}
}

// dryRunRunner implements provision.ExecRunner and just prints/logs commands.
type dryRunRunner struct{}

func (r *dryRunRunner) Run(cmd string, args ...string) error {
	if cmd == "section" || cmd == "info" {
		return nil
	}
	fmt.Printf("[dry-run] Would run: %s %s\n", cmd, strings.Join(args, " "))
	return nil
}
func (r *dryRunRunner) Output(cmd string, args ...string) ([]byte, error) {
	out := fmt.Sprintf("[dry-run] Would output: %s %s", cmd, strings.Join(args, " "))
	return []byte(out), nil
}

// headlessMain runs the provisioner logic without the TUI, printing logs to stdout.
func headlessMain(lazy bool, manifestPath string, dryRun bool, groups, only []string) {
	manifest, err := app.LoadManifest(manifestPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to load manifest: %v\n", err)
		os.Exit(1)
	}
	var keys []string
	switch {
	case len(only) > 0:
		keys = only
	case len(groups) > 0:
		for k := range manifest {
			entry := manifest[k]
			entryPtr := &entry
			for _, g := range entryPtr.Groups {
				for _, want := range groups {
					if g == want {
						keys = append(keys, k)
						break
					}
				}
			}
		}
	default:
		for k := range manifest {
			keys = append(keys, k)
		}
	}
	var runner provision.ExecRunner
	if dryRun {
		runner = &dryRunRunner{}
	} else {
		runner = &realSystemRunner{}
	}
	installed := provision.GetInstalledPackages(runner)
	prov := provision.NewProvisioner(nil, manifest, runner)
	prov.LazyOnly = lazy
	fmt.Println("Starting provisioning...")
	plan, err := prov.PlanProvision(keys, installed)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to plan provision: %v\n", err)
		os.Exit(1)
	}
	if len(plan) == 0 {
		fmt.Println("Nothing to install. All requested packages are already installed or filtered out.")
	}
	err = prov.ExecutePlan(plan)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Provisioning failed: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("Provisioning complete")
}
