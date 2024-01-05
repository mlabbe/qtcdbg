/*
 * qtcdbg Copyright (C) 2019-2020, 2024 Frogtoss Games, Inc.
 */

package main

import (
	"bufio"
	"errors"
	"fmt"
	"gopkg.in/ini.v1"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"

	"os"

	"github.com/alecthomas/kingpin"
)

var (
	app = kingpin.New("qtcdbg", "QtCreator debugger launcher")

	// common
	debug   = app.Flag("debug", "Verbose debug qtcdbg").Bool()
	version = app.Flag("version", "Show version and exit").Short('v').Bool()

	// launch (default command)
	launchCmd  = app.Command("launch", "Launch QtCreator as a debugger").Default()
	configPath = launchCmd.Arg("config", "Path to config file").Default("").String()
	noRun      = launchCmd.Flag("no-run", "Do not run QtCreator -- just generate project files").Bool()

	// init
	initCmd = app.Command("init", "Create toml config for your project")
)

const VersionMajor = 1
const VersionMinor = 1

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

func GetIniPath() (string, error) {
	home := os.Getenv("HOME")

	IniLocations := []string{
		// common and linux
		home + "/.config/QtProject/QtCreator.ini",
		home + "/.local/share/data/QtProject/qtcreator/QtCreator.ini",

		// macos
		home + "/.Library/Application Support/QtProject/Qt Creator/QtCreator.ini"}

	if runtime.GOOS == "windows" {
		appData := os.Getenv("APPDATA")
		IniLocations = []string{
			filepath.Join(appData, "QtProject", "QtCreator.ini"),
		}
	}

	for _, iniLocation := range IniLocations {

		if *debug {
			fmt.Printf("Trying ini location %s\n", iniLocation)
		}

		if _, err := os.Stat(iniLocation); errors.Is(err, os.ErrNotExist) {
			continue
		}

		if *debug {
			fmt.Printf("Found ini at %s\n", iniLocation)
			return iniLocation, nil
		}
	}

	return "", errors.New("Could not find QtCreator.ini")
}

// Read the Environment Id from QtCreator ini file.
func GetEnvironmentId() (string, error) {
	iniPath, err := GetIniPath()
	if err != nil {
		return "", err
	}

	cfg, err := ini.Load(iniPath)
	if err != nil {
		return "", fmt.Errorf("Failed to parse ini: %+v", err)
	}
	environmentIdKey := cfg.Section("ProjectExplorer").Key("Settings\\EnvironmentId").String()

	// key is like @ByteArray({a70246c0-282b-4255-87b8-41e5cc6b85ff}), and we just want the guid
	reFindGuid := regexp.MustCompile(`\{(.+)\}`)
	environmentId := reFindGuid.FindStringSubmatch(environmentIdKey)

	if len(environmentId) != 2 {
		return "", fmt.Errorf("No 'ProjectExplorer.Settings\\EnvironmentId' key found in %s", iniPath)
	}

	if *debug {
		fmt.Printf("Found environmentId %s\n", environmentId[1])
	}

	return environmentId[1], nil
}

// Read the kit id
func GetKitId() (string, error) {
	home := os.Getenv("HOME")

	ProfileLocations := []string{
		home + "/.config/QtProject/qtcreator/profiles.xml",
	}

	if runtime.GOOS == "windows" {
		appData := os.Getenv("APPDATA")
		ProfileLocations = []string{
			filepath.Join(appData, "QtProject", "qtcreator", "profiles.xml"),
		}
	}

	var xml *os.File
	for _, xmlLocation := range ProfileLocations {
		xml, _ = os.Open(filepath.Clean(xmlLocation))
		if xml != nil {
			defer xml.Close()

			break
		}
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
}

func LaunchQtCreator(projectPath string) error {
	// try to find it in path
	exePath, err := exec.LookPath("qtcreator")
	if err != nil {
		if runtime.GOOS == "linux" {
			return err
		}
	}

	if runtime.GOOS == "darwin" {
		exePath = "/Applications/Qt Creator.app/Contents/MacOS/Qt Creator"
	}

	if runtime.GOOS == "windows" && err != nil {
		// I TOLD you we only supported one version for now
		exePath = "c:\\Qt\\qtcreator-4.11.0\\bin\\qtcreator.exe"
	}

	cmd := exec.Command(exePath, projectPath, "-lastsession")
	err = cmd.Run()
	if err != nil {
		return err
	}

	return nil
}

func cleanupPath(path string) {
	if path == "" {
		return
	}

	err := os.Remove(path)

	if !*debug {
		if err != nil {
			panic("could not cleanup path on exit")
		}
	}
}

func main() {
	os.Exit(RealMain())
}

func RealMain() int {
	switch kingpin.MustParse(app.Parse(os.Args[1:])) {
	case initCmd.FullCommand():
		Init()
		return 0
	}

	if *version {
		fmt.Printf("qtcdbg %d.%d\n", VersionMajor, VersionMinor)
		return 0
	}

	actualConfigPath, err := findConfig(*configPath)
	if err != nil {
		fmt.Printf("Could not find config: %v", err)
		return 0
	}

	cfg, err := parseConfig(actualConfigPath)
	if err != nil {
		fmt.Printf("Error loading config: %v", err)
		return 1
	}

	if cfg.CompileCommands.Override {
		compileCommandsPath := cfg.CompileCommands.Dir + "/compile_commands.json"
		_, err := os.Stat(compileCommandsPath)
		if errors.Is(err, os.ErrNotExist) {
			fmt.Printf("Compile commands file '%s' does not exist\n", compileCommandsPath)
			return 1
		}
	}

	environmentId, err := GetEnvironmentId()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Did not find the environmentId in the QtCreator config file.\n")
		fmt.Fprintf(os.Stderr, "Running QtCreator once should generate this.\n")
		return 1
	}

	var clangdWrapperPath string
	if cfg.CompileCommands.Override {
		cfg.Misc.OriginalClangdPath, err = GetClangdPath()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Clangd path not found: %+v\n", cfg.Misc.OriginalClangdPath)
			return 1
		}

		clangdWrapperPath, err = WriteClangdWrapper(cfg.Misc.OriginalClangdPath, cfg.CompileCommands.Dir)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to write clangd wrapper script: %+v\n", err)
			return 1
		}

		err = SetClangdPath(clangdWrapperPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to write clangd wrapper to path: %+v\n", err)
			return 1
		}
	}
	defer cleanupPath(clangdWrapperPath)
	defer SetClangdPath(cfg.Misc.OriginalClangdPath)

	cfg.Misc.EnvironmentId = environmentId
	kitId, err := GetKitId()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Did not find the kit id.\n")
		return 1
	}
	cfg.Misc.KitId = kitId

	if *debug {
		fmt.Printf("EnvironmentId: %s\n", cfg.Misc.EnvironmentId)
		fmt.Printf("KitId: %s\n", cfg.Misc.KitId)
	}

	defer CleanupGeneratedFiles(&cfg, *noRun)

	//
	// begin generation
	//
	err = GenerateFlags(&cfg)
	if err != nil {
		handleGenerationError(err)
		return 1
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
		return 0
	}

	creatorPath := getGeneratorPath(&cfg, cfg.Project.Name+".creator")
	err = LaunchQtCreator(creatorPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to launch qtcreator: %s\n", err)
		fmt.Println("qtcdbg requires qtcreator to be in the system path.  See \"usage\" in README for details.")
		return 1
	}

	return 0
}
