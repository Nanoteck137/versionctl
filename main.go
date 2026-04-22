package main

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

func main() {
	if err := ensureRepoRootOrChdir(); err != nil {
		fmt.Println("Error:", err)
		os.Exit(1)
	}

	if len(os.Args) < 2 {
		usage()
		return
	}

	switch os.Args[1] {
	case "version":
		v, err := resolveVersion()
		exitIfErr(err)
		fmt.Println(v)

	case "release":
		handleRelease(os.Args[2:])

	case "dev":
		err := setDev(false)
		exitIfErr(err)

	default:
		usage()
	}
}

func usage() {
	fmt.Println("Usage:")
	fmt.Println("  versionctl version")
	fmt.Println("  versionctl release [patch|minor|major] [--dry-run] [--label <label>] [--pre-cmd \"cmd\"]")
	fmt.Println("  versionctl dev")
}

//////////////////////////////////////////////////
// Helpers
//////////////////////////////////////////////////

func exitIfErr(err error) {
	if err != nil {
		fmt.Println("Error:", err)
		os.Exit(1)
	}
}

func run(name string, args ...string) error {
	cmd := exec.Command(name, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func output(name string, args ...string) (string, error) {
	cmd := exec.Command(name, args...)
	out, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}

//////////////////////////////////////////////////
// Git helpers
//////////////////////////////////////////////////

func getLatestTag() string {
	tag, err := output("git", "describe", "--tags", "--abbrev=0")
	if err != nil {
		return "0.0.0"
	}
	return strings.TrimPrefix(tag, "v")
}

func getCommitHash() string {
	hash, _ := output("git", "rev-parse", "--short", "HEAD")
	return hash
}

func isDirty() bool {
	err := exec.Command("git", "diff", "--quiet").Run()
	return err != nil
}

func ensureCleanTree() error {
	if isDirty() {
		return errors.New("working tree is dirty")
	}
	return nil
}

func ensureRepoRootOrChdir() error {
	root, err := output("git", "rev-parse", "--show-toplevel")
	if err != nil {
		return fmt.Errorf("not a git repository")
	}

	return os.Chdir(root)
}

//////////////////////////////////////////////////
// Version logic
//////////////////////////////////////////////////

func readVersionFile() (string, error) {
	data, err := os.ReadFile("version")
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(data)), nil
}

func writeVersionFile(v string) error {
	return os.WriteFile("version", []byte(v+"\n"), 0644)
}

func bump(version, part string) (string, error) {
	var major, minor, patch int
	_, err := fmt.Sscanf(version, "%d.%d.%d", &major, &minor, &patch)
	if err != nil {
		return "", errors.New("invalid version format")
	}

	switch part {
	case "major":
		major++
		minor = 0
		patch = 0
	case "minor":
		minor++
		patch = 0
	case "patch":
		patch++
	default:
		return "", errors.New("invalid bump type")
	}

	return fmt.Sprintf("%d.%d.%d", major, minor, patch), nil
}

func resolveVersion() (string, error) {
	v, err := readVersionFile()
	if err != nil {
		return "", err
	}

	if v != "main" {
		return v, nil
	}

	latest := getLatestTag()
	next, _ := bump(latest, "patch")

	hash := getCommitHash()

	suffix := ""
	if isDirty() {
		suffix = "-dirty"
	}

	return fmt.Sprintf("%s-dev+%s%s", next, hash, suffix), nil
}

//////////////////////////////////////////////////
// Commands
//////////////////////////////////////////////////

func handleRelease(args []string) {
	if len(args) == 0 {
		exitIfErr(errors.New("missing bump type"))
	}

	part := args[0]

	dryRun := false
	label := ""
	preCmd := ""

	// parse flags manually (simple)
	for i := 1; i < len(args); i++ {
		switch args[i] {
		case "--dry-run":
			dryRun = true
		case "--label":
			i++
			label = args[i]
		case "--pre-cmd":
			i++
			preCmd = args[i]
		}
	}

	exitIfErr(release(part, dryRun, label, preCmd))
}

func release(part string, dryRun bool, label string, preCmd string) error {
	// 1. Ensure clean repo
	if err := ensureCleanTree(); err != nil {
		return err
	}

	latest := getLatestTag()
	next, err := bump(latest, part)
	if err != nil {
		return err
	}

	if label != "" {
		next = fmt.Sprintf("%s-%s", next, label)
	}

	fmt.Println("Next version:", next)

	// 2. Run pre-command (build/test)
	if preCmd != "" {
		fmt.Println("Running pre-command:", preCmd)
		cmd := exec.Command("sh", "-c", preCmd)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			return errors.New("pre-command failed, aborting release")
		}
	}

	if dryRun {
		fmt.Println("[DRY RUN] Skipping git operations")
		return nil
	}

	// 3. Set VERSION
	if err := writeVersionFile(next); err != nil {
		return err
	}

	if err := run("git", "add", "VERSION"); err != nil {
		return err
	}
	if err := run("git", "commit", "-m", "release: "+next); err != nil {
		return err
	}

	// 4. Tag
	if err := run("git", "tag", "v"+next); err != nil {
		return err
	}

	// 5. Back to dev
	if err := writeVersionFile("main"); err != nil {
		return err
	}

	if err := run("git", "add", "VERSION"); err != nil {
		return err
	}
	if err := run("git", "commit", "-m", "chore: back to main"); err != nil {
		return err
	}

	fmt.Println("Release complete:", next)
	return nil
}

func setDev(dry bool) error {
	if dry {
		fmt.Println("[DRY RUN] Would set VERSION=main")
		return nil
	}

	if err := writeVersionFile("main"); err != nil {
		return err
	}

	if err := run("git", "add", "VERSION"); err != nil {
		return err
	}
	if err := run("git", "commit", "-m", "chore: back to main"); err != nil {
		return err
	}

	fmt.Println("Switched to main")
	return nil
}
