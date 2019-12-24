/*
 * qtcdbg Copyright (C) 2019 Frogtoss Games, Inc.
 */

package main

import (
	"bufio"
	"errors"
	"fmt"
	"os/exec"
	"path/filepath"
	"regexp"

	"os"

	"github.com/alecthomas/kingpin"
)

var (
	debug      = kingpin.Flag("debug", "Debug mode").Bool()
	configPath = kingpin.Arg("config", "Path to config file").Default("").String()
)

const ConfigDefault = "qtcdbg.toml"

const VersionMajor = 0
const VersionMinor = 1

// find the user's config file
func findConfig(userConfig string) (string, error) {
	// user requested configs must be in the current dir
	if userConfig != "" {
		return userConfig, nil
	}

	// check for default filename in current directory
	info, err := os.Stat(ConfigDefault)
	if !os.IsNotExist(err) && !info.IsDir() {
		return ConfigDefault, nil
	}

	// search for file recursively from the launch location
	var foundPath *string
	err = filepath.Walk(".",
		func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}

			if !info.IsDir() && info.Name() == ConfigDefault {
				foundPath = &path
			}

			return nil
		})
	if err != nil {
		return "", nil
	}

	if foundPath != nil {
		fmt.Printf("Launching with config %s\n", *foundPath)
		return *foundPath, nil
	}

	return "", errors.New("Could not find " + ConfigDefault)
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

func handleGenerationError(err error) {
	fmt.Fprintf(os.Stderr, "Generation error: %v\n", err)
	os.Exit(1)
}

func LaunchQtCreator(projectPath string) error {
	exePath, err := exec.LookPath("qtcreator.sh")
	if err != nil {
		return err
	}

	cmd := exec.Command(exePath, projectPath)
	err = cmd.Run()
	if err != nil {
		return err
	}

	return nil
}

func main() {
	kingpin.Parse()

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
	if *debug {
		fmt.Printf("EnvironmentId: %s\n", cfg.Misc.EnvironmentId)
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

	creatorPath := getGeneratorPath(&cfg, cfg.Project.Name+".creator")
	err = LaunchQtCreator(creatorPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to launch qtcreator: %s", err)
		os.Exit(1)
	}
}
