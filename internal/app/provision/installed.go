package provision

import (
	"bufio"
	"regexp"
	"strings"
)

// GetInstalledPackages queries the system for installed packages for supported managers.
// It returns a map of package names (keys) that are installed.
// Uses the provided ExecRunner for testability.
func GetInstalledPackages(runner ExecRunner) map[string]bool {
	installed := make(map[string]bool)

	merge := func(pkgs map[string]bool) {
		for k := range pkgs {
			installed[k] = true
		}
	}

	merge(getAptInstalled(runner))
	merge(getBrewInstalled(runner))
	merge(getPipxInstalled(runner))
	merge(getCargoInstalled(runner))
	merge(getNpmInstalled(runner))

	return installed
}

func getAptInstalled(runner ExecRunner) map[string]bool {
	pkgs := make(map[string]bool)
	out, err := runner.Output("dpkg", "-l")
	if err != nil {
		return pkgs
	}
	scan := bufio.NewScanner(strings.NewReader(string(out)))
	for scan.Scan() {
		line := scan.Text()
		if strings.HasPrefix(line, "ii ") {
			fields := strings.Fields(line)
			if len(fields) > 1 {
				pkgs[fields[1]] = true
			}
		}
	}
	return pkgs
}

func getBrewInstalled(runner ExecRunner) map[string]bool {
	pkgs := make(map[string]bool)
	out, err := runner.Output("brew", "list", "-1")
	if err != nil {
		return pkgs
	}
	scan := bufio.NewScanner(strings.NewReader(string(out)))
	for scan.Scan() {
		name := strings.TrimSpace(scan.Text())
		if name != "" {
			pkgs[name] = true
		}
	}
	return pkgs
}

func getPipxInstalled(runner ExecRunner) map[string]bool {
	pkgs := make(map[string]bool)
	out, err := runner.Output("pipx", "list")
	if err != nil {
		return pkgs
	}
	scan := bufio.NewScanner(strings.NewReader(string(out)))
	for scan.Scan() {
		line := scan.Text()
		if strings.HasPrefix(line, "  - ") {
			name := strings.TrimSpace(strings.TrimPrefix(line, "  - "))
			if name != "" {
				pkgs[name] = true
			}
		}
	}
	return pkgs
}

func getCargoInstalled(runner ExecRunner) map[string]bool {
	pkgs := make(map[string]bool)
	out, err := runner.Output("cargo", "install", "--list")
	if err != nil {
		return pkgs
	}
	scan := bufio.NewScanner(strings.NewReader(string(out)))
	for scan.Scan() {
		line := scan.Text()
		if line != "" && !strings.HasPrefix(line, " ") && strings.Contains(line, " ") {
			fields := strings.Fields(line)
			if len(fields) > 0 {
				pkgs[fields[0]] = true
			}
		}
	}
	return pkgs
}

func getNpmInstalled(runner ExecRunner) map[string]bool {
	pkgs := make(map[string]bool)
	out, err := runner.Output("npm", "list", "-g", "--depth=0")
	if err != nil {
		return pkgs
	}
	scan := bufio.NewScanner(strings.NewReader(string(out)))
	pkgRe := regexp.MustCompile(`([a-zA-Z0-9._-]+)@`)
	for scan.Scan() {
		line := scan.Text()
		if strings.Contains(line, "@") {
			matches := pkgRe.FindStringSubmatch(line)
			if len(matches) > 1 {
				pkgs[matches[1]] = true
			}
		}
	}
	return pkgs
}
