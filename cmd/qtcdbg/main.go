/*
 * qtcdbg Copyright (C) 2019-2020 Frogtoss Games, Inc.
 */

package main

import (
	"bufio"
	"errors"
	"fmt"
	"os/exec"
	"runtime"
	"path/filepath"
	"regexp"
	"strings"

	"os"

	"github.com/alecthomas/kingpin"
)

var (
	app = kingpin.New("qtcdbg", "QtCreator debugger launcher")

	// common
	debug   = app.Flag("debug", "Verbose debug qtcdbg").Bool()
	version = kingpin.Flag("version", "Show version and exit").Short('v').Bool()

	// launch (default command)
	launchCmd  = app.Command("launch", "Launch QtCreator as a debugger").Default()
	configPath = launchCmd.Arg("config", "Path to config file").Default("").String()
	noRun      = launchCmd.Flag("no-run", "Do not run QtCreator -- just generate project files").Bool()

	// init
	initCmd = app.Command("init", "Create toml config for your project")
)

const VersionMajor = 0
const VersionMinor = 7

func defaultConfig() string {
	return "qtcdbg." + runtime.GOOS + ".toml"
}

// find the user's config file
func findConfig(userConfig string) (string, error) {
	// user requested configs must be in the current dir
	if userConfig != "" {
		return userConfig, nil
	}

	// check for default filename in current directory
	info, err := os.Stat(defaultConfig())
	if !os.IsNotExist(err) && !info.IsDir() {
		return defaultConfig(), nil
	}

	// search for file recursively from the launch location
	var foundPath *string
	err = filepath.Walk(".",
		func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}

			if !info.IsDir() && info.Name() == defaultConfig() {
				foundPath = &path
			}

			return nil
		})
	if err != nil {
		return "", nil
	}

	if foundPath != nil {
		fmt.Printf("Launching with found config %s\n", *foundPath)
		return *foundPath, nil
	}

	return "", errors.New("Could not find " + defaultConfig() + "\n")
}

// Read the Environment Id from QtCreator ini file.
func GetEnvironmentId() (string, error) {
	home := os.Getenv("HOME")

	// todo: support windows
	IniLocations := []string{
		// common and linux
		home + "/.config/QtProject/QtCreator.ini",
		home + "/.local/share/data/QtProject/qtcreator/QtCreator.ini",

		// macos
		home + "/.Library/Application Support/QtProject/Qt Creator/QtCreator.ini"}

	var ini *os.File
	for _, iniLocation := range IniLocations {

		ini, _ = os.Open(filepath.Clean(iniLocation))
		if ini != nil {
			break
		}
		defer ini.Close()
	}

	if ini == nil {
		return "", errors.New("Could not find QtCreator.ini")
	}

	scanner := bufio.NewScanner(ini)
	re := regexp.MustCompile("Settings\\\\EnvironmentId=@ByteArray\\(\\{(.*)\\}\\)")

	for scanner.Scan() {
		match := re.FindStringSubmatch(scanner.Text())
		if len(match) != 0 {
			return match[1], nil
		}
	}

	return "", errors.New("Could not find QtCreator.ini")
}

// Read the kit id
func GetKitId() (string, error) {
	home := os.Getenv("HOME")

	ProfileLocations := []string{
		home + "/.config/QtProject/qtcreator/profiles.xml",
	}

	var xml *os.File
	for _, xmlLocation := range ProfileLocations {
		xml, _ = os.Open(filepath.Clean(xmlLocation))
		if xml != nil {
			break
		}
		defer xml.Close()
	}

	if xml == nil {
		return "", errors.New("Could not find qtcreator/profiles.xml")
	}

	scanner := bufio.NewScanner(xml)

	// first guid in file after Profile.Default variable is a match

	for scanner.Scan() {
		if strings.Contains(scanner.Text(), "<variable>Profile.Default</variable>") {
			break
		}
	}

	re := regexp.MustCompile(`\{(\w{8}-\w{4}-\w{4}-\w{4}-\w{12})\}`)

	for scanner.Scan() {
		match := re.FindStringSubmatch(scanner.Text())
		if len(match) != 0 {
			return match[1], nil
		}
	}

	return "", errors.New("Could not find profile in profiles.xml")
}

func handleGenerationError(err error) {
	fmt.Fprintf(os.Stderr, "Generation error: %v\n", err)
	os.Exit(1)
}

func LaunchQtCreator(projectPath string) error {
	// try to find it in path
	exePath, err := exec.LookPath("qtcreator")
	if err != nil {
		if runtime.GOOS != "darwin" {
			return err
		}
	}

	if runtime.GOOS == "darwin" {
		exePath = "/Applications/Qt Creator.app/Contents/MacOS/Qt Creator"
	}

	cmd := exec.Command(exePath, projectPath, "-lastsession")
	err = cmd.Run()
	if err != nil {
		return err
	}

	return nil
}

func main() {
	switch kingpin.MustParse(app.Parse(os.Args[1:])) {
	case initCmd.FullCommand():
		Init()
		os.Exit(0)
	}

	if *version {
		fmt.Printf("qtcdbg %d.%d\n", VersionMajor, VersionMinor)
		os.Exit(0)
	}

	actualConfigPath, err := findConfig(*configPath)
	if err != nil {
		fmt.Printf("Could not find config: %v", err)
		os.Exit(1)
	}

	cfg, err := parseConfig(actualConfigPath)
	if err != nil {
		fmt.Printf("Error loading config: %v", err)
		os.Exit(1)
	}

	environmentId, err := GetEnvironmentId()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Did not find the environmentId in the QtCreator config file.\n")
		fmt.Fprintf(os.Stderr, "Running QtCreator once should generate this.\n")
		os.Exit(1)
	}

	cfg.Misc.EnvironmentId = environmentId
	kitId, err := GetKitId()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Did not find the kit id.\n")
		os.Exit(1)
	}
	cfg.Misc.KitId = kitId

	if *debug {
		fmt.Printf("EnvironmentId: %s\n", cfg.Misc.EnvironmentId)
		fmt.Printf("KitId: %s\n", cfg.Misc.KitId)
	}

	err = GenerateCflags(&cfg)
	if err != nil {
		handleGenerationError(err)
	}

	err = GenerateConfig(&cfg)
	if err != nil {
		handleGenerationError(err)
	}

	err = GenerateCreator(&cfg)
	if err != nil {
		handleGenerationError(err)
	}

	err = GenerateFiles(&cfg)
	if err != nil {
		handleGenerationError(err)
	}

	err = GenerateIncludes(&cfg)
	if err != nil {
		handleGenerationError(err)
	}

	err = GenerateCreatorUser(&cfg)
	if err != nil {
		handleGenerationError(err)
	}

	if *noRun {
		os.Exit(0)
	}

	creatorPath := getGeneratorPath(&cfg, cfg.Project.Name+".creator")
	err = LaunchQtCreator(creatorPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to launch qtcreator: %s\n", err)
		fmt.Println("qtcdbg requires qtcreator to be in the system path.  See \"usage\" in README for details.")
		os.Exit(1)
	}
}
