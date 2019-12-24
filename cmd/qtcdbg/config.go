/*
 * qtcdbg Copyright (C) 2019 Frogtoss Games, Inc.
 */

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
		WorkingDir string `toml:"working_dir"`
		Command    string `toml:"command"`
		Arguments  string `toml:"arguments"`
	} `toml:"build"`
	Run struct {
		WorkingDir     string `toml:"working_dir"`
		ExecutablePath string `toml:"executable_path"`
		Arguments      string `toml:"arguments"`
		RunInTerminal  bool   `toml:"run_in_terminal"`
	} `toml:"run"`
	Generate struct {
		ConfigDefines               []string `toml:"config_defines"`
		AdditionalIncludeSearchDirs []string `toml:"additional_include_search_dirs"`
	} `toml:"generate"`

	// not in toml parse
	Misc struct {
		cfgPath       string
		EnvironmentId string
		ProjectRoot   string
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

	cfg.Misc.cfgPath = path
	cfg.Misc.ProjectRoot = getProjectRoot(&cfg)

	return cfg, nil
}
