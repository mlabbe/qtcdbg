package main

import (
	"io/ioutil"
	"log"
	"os"

	"github.com/BurntSushi/toml"
)

//	"github.com/BurntSushi/toml"

type TomlConfig struct {
	Project struct {
		Name         string `toml:"name"`
		RelativeRoot string `toml:"relative_root"`
	} `toml:"project"`
	Build struct {
		BuildDir  string `toml:"build_dir"`
		BuildStep string `toml:"build_step"`
		CleanStep string `toml:"clean_step"`
	} `toml:"build"`
	Run struct {
		RunPath       string `toml:"run_path"`
		RunWorkingDir string `toml:"run_working_dir"`
		RunArguments  string `toml:"run_arguments"`
	} `toml:"run"`
	Generate struct {
		ConfigDefines []string `toml:"config_defines"`
	} `toml:"generate"`

	// not in toml parse
	misc struct {
		cfgPath string
	}
}

func parseConfig(path string) (TomlConfig, error) {
	var cfg TomlConfig

	file, err := os.Open(path)
	if err != nil {
		log.Fatal(err)
	}

	tomlBytes, err := ioutil.ReadAll(file)
	if err != nil {
		log.Fatal(err)
	}

	if err := toml.Unmarshal(tomlBytes, &cfg); err != nil {
		log.Fatal(err)
	}

	cfg.misc.cfgPath = path

	return cfg, nil
}
