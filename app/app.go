package app

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/charmbracelet/huh"
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
	// TODO(patrik): This is wrong, we should just get the current version
	// and not bump it
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

func askContinue(prompt string) bool {
	var val bool

	form := huh.NewConfirm().
		Title(prompt).
		Value(&val)

	err := form.Run()
	if err != nil {
		return false
	}

	return val
}

func askPart(currentVersion string) (string, error) {
	part := ""

	patchVersion, _ := bump(currentVersion, "patch")
	minorVersion, _ := bump(currentVersion, "minor")
	majorVersion, _ := bump(currentVersion, "major")

	form := huh.NewSelect[string]().
		Title("What new version").
		Options(
			huh.NewOption("Patch ("+patchVersion+")", "patch"),
			huh.NewOption("Minor ("+minorVersion+")", "minor"),
			huh.NewOption("Major ("+majorVersion+")", "major"),
			huh.NewOption("Quit", "quit"),
		).
		Value(&part)

	err := form.Run()
	if err != nil {
		return "", err
	}

	if part == "quit" {
		return "", errors.New("quitting")
	}

	return part, nil
}

func Release(conf *config.Config, version, label string) error {
	next := ""
	if version != "" {
		// TODO(patrik): Check version for correct format
		next = version
	} else {
		var err error

		latest := getLatestTag()

		part, err := askPart(latest)
		if err != nil {
			return err
		}

		next, err = bump(latest, part)
		if err != nil {
			return err
		}
	}

	if label != "" {
		next = fmt.Sprintf("%s-%s", next, label)
	}

	if isDirty() {
		return errors.New("working tree is dirty")
	}

	// Run pre cmd
	if conf.PreCmd != "" {
		fmt.Printf("Running pre-command '%s'\n", conf.PreCmd)

		err := run("sh", "-c", conf.PreCmd)
		if err != nil {
			return errors.New("pre-command failed, aborting release")
		}
	}

	fmt.Println("Next version:", next)

	if !askContinue("Do you want to continue") {
		return errors.New("abort")
	}

	// Release new version
	fmt.Println("Writing version to 'version' file")
	err := writeVersionFile(next)
	if err != nil {
		return err
	}

	err = run("git", "add", "version")
	if err != nil {
		return err
	}

	fmt.Println("Commiting version file")
	err = run("git", "commit", "-m", "release: version "+next)
	if err != nil {
		return err
	}

	fmt.Println("Tagging the new version")
	err = run("git", "tag", next)
	if err != nil {
		return err
	}

	// Go back to main
	fmt.Println("Writing 0.0.0 to 'version' file")
	err = writeVersionFile("0.0.0")
	if err != nil {
		return err
	}

	err = run("git", "add", "version")
	if err != nil {
		return err
	}

	fmt.Println("Commiting version file back to 0.0.0")
	err = run("git", "commit", "-m", "chore: back to 0.0.0")
	if err != nil {
		return err
	}

	// TODO(patrik): Ask the user to push
	if askContinue("Do you want to push the commits and the new tag?") {
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
