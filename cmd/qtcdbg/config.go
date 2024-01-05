/*
 * qtcdbg Copyright (C) 2019-2020, 2024 Frogtoss Games, Inc.
 */

package main

import (
	"io/ioutil"
	"log"
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
)

// "github.com/BurntSushi/toml"
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
		ConfigCFlags                []string `toml:"config_cflags"`
		AdditionalIncludeSearchDirs []string `toml:"additional_include_search_dirs"`
	} `toml:"generate"`
	CompileCommands struct {
		Override bool   `toml:"override"`
		Dir      string `toml:"dir"`
	} `toml:"compile_commands"`

	// not in toml parse
	Misc struct {
		cfgPath            string
		EnvironmentId      string
		KitId              string
		ProjectRoot        string
		OriginalClangdPath string
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

	// handle relative paths
	cfg.Run.ExecutablePath = filepath.Join(cfg.Misc.ProjectRoot, cfg.Run.ExecutablePath)

	cfg.Run.WorkingDir = filepath.Join(cfg.Misc.ProjectRoot, cfg.Run.WorkingDir)

	return cfg, nil
}
