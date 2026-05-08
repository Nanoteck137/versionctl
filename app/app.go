package app

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/nanoteck137/versionctl/config"
)

//////////////////////////////////////////////////
// Helpers
//////////////////////////////////////////////////

func run(name string, args ...string) error {
	cmd := exec.Command(name, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err := cmd.Run()
	if err != nil {
		return err
	}

	return nil
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

	return tag
}

func getCommitHash() string {
	hash, _ := output("git", "rev-parse", "--short", "HEAD")
	return hash
}

func isDirty() bool {
	err := exec.Command("git", "diff", "--quiet").Run()
	return err != nil
}

func EnsureRepoRootOrChdir() error {
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

func ResolveVersion() (string, error) {
	v, err := readVersionFile()
	if err != nil {
		return "", err
	}

	if v != "main" {
		return v, nil
	}

	latest := getLatestTag()
	// TODO(patrik): Handle error
	next, err := bump(latest, "patch")
	if err != nil {
		return "", err
	}

	hash := getCommitHash()

	suffix := ""
	if isDirty() {
		suffix = "-dirty"
	}

	return fmt.Sprintf("%s-dev+%s%s", next, hash, suffix), nil
}

func Release(conf *config.Config, label string) error {
	part := "patch"

	if isDirty() {
		return errors.New("working tree is dirty")
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

	// Run pre cmd
	if conf.PreCmd != "" {
		fmt.Println("Running pre-command:", conf.PreCmd)

		cmd := exec.Command("sh", "-c", conf.PreCmd)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr

		err := cmd.Run()
		if err != nil {
			return errors.New("pre-command failed, aborting release")
		}
	}

	// Release new version
	err = writeVersionFile(next)
	if err != nil {
		return err
	}

	err = run("git", "add", "version")
	if err != nil {
		return err
	}

	err = run("git", "commit", "-m", "release: version "+next)
	if err != nil {
		return err
	}

	err = run("git", "tag", next)
	if err != nil {
		return err
	}

	// Go back to main
	err = writeVersionFile("0.0.0")
	if err != nil {
		return err
	}

	err = run("git", "add", "version")
	if err != nil {
		return err
	}

	err = run("git", "commit", "-m", "chore: back to 0.0.0")
	if err != nil {
		return err
	}

	// Push
	if conf.Push {
		fmt.Println("Running git push")

		err = run("git", "push")
		if err != nil {
			return err
		}

		// TODO(patrik): Configure origin?
		err = run("git", "push", "origin", next)
		if err != nil {
			return err
		}
	}

	fmt.Println("Release complete:", next)

	return nil
}

// func setDev(dry bool) error {
// 	if dry {
// 		fmt.Println("[DRY RUN] Would set VERSION=main")
// 		return nil
// 	}
//
// 	if err := writeVersionFile("main"); err != nil {
// 		return err
// 	}
//
// 	if err := run("git", "add", "VERSION"); err != nil {
// 		return err
// 	}
// 	if err := run("git", "commit", "-m", "chore: back to main"); err != nil {
// 		return err
// 	}
//
// 	fmt.Println("Switched to main")
// 	return nil
// }
