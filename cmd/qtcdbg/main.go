/*
 *
 */

package main

import (
	"bufio"
	"errors"
	"fmt"
	"regexp"

	"os"

	"github.com/alecthomas/kingpin"
)

var (
	debug      = kingpin.Flag("debug", "Debug mode").Bool()
	configPath = kingpin.Arg("config", "Path to config file").Default("qtcdbg.toml").String()
)

// Read the Environment Id from QtCreator ini file.
func GetEnvironmentId() (string, error) {
	IniLocations := []string{"/home/mlabbe/.config/QtProject/QtCreator.ini"}

	var ini *os.File
	for _, iniLocation := range IniLocations {
		ini, _ = os.Open(iniLocation)
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
	fmt.Fprintf(os.Stderr, "Generation error: %v", err)
	os.Exit(1)
}

func main() {
	kingpin.Parse()

	environmentId, err := GetEnvironmentId()
	if err != nil {
		fmt.Fprintf(os.Stderr, `Did not find the environmentId in the QtCreator config file.\n
Running QtCreator once should generate this file.\n`)
		os.Exit(1)
	}

	if *debug {
		fmt.Printf("EnvironmentId: %s\n", environmentId)
	}

	cfg, err := parseConfig(*configPath)
	if err != nil {
		fmt.Printf("Error loading config: %v", err)
		os.Exit(1)
	}

	err = GenerateCflags(&cfg)
	if err != nil {
		handleGenerationError(err)
	}

	err = GenerateConfig(&cfg)
	if err != nil {
		handleGenerationError(err)
	}
}
