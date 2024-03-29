/*
 * qtcdbg Copyright (C) 2019-2020, 2024 Frogtoss Games, Inc.
 */

package main

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"text/template"

	"github.com/chzyer/readline"
)

var tmplToml = `#
# qtcdbg file -- everything necessary to launch QtCreator te debug this project
#
# see https://github.com/mlabbe/qtcdbg
#

# this file was generated with "qtcdbg init" and should be checked in to 
# source control.

[project]

# project name
name = "{{ .Project.Name }}"

# project root relative to this config file
relative_root = "{{ .Project.RelativeRoot }}"

[build]

# directory to run build command in
working_dir = "{{ .Build.WorkingDir }}"

command = "{{ .Build.Command }}"

arguments = "{{ .Build.Arguments }}"

[run]

# cwd while running the program
working_dir = "{{ .Run.WorkingDir }}"

# path including filename of executable to debug
executable_path = "{{ .Run.ExecutablePath }}"

# arguments to run with
arguments = ""

# whether qtcreator should pop up a terminal
run_in_terminal = true

[generate]
# qtcreator's syntax highlighting dims proprocessor paths not generated.
# this specifies additional defines for qtcreator
config_defines = [
]

# qtcreator config flags for c
config_cflags = [
]

additional_include_search_dirs = [
]

[compile_commands]
# if qtcreator is set up to use clangd, it will create its own .qtc_clangd/compile_commands.json
# which is not a good match.  Set this to true to override compile_commands.json with
# one supplied for your own project.  This helps warnings, defines and other compile flags
# match your main development ones.
override = false

# dir to compile_commands.json, relative to working dir
path = "src/"
`

var Separator = fmt.Sprintf("%c", filepath.Separator)

func escape(s *string) string {
	s2 := strings.ReplaceAll(*s, `\`, `\\`)
	*s = s2

	return s2
}

func askYesNo(rl *readline.Instance, question string) bool {
	fmt.Println(question + " (y/N)")

	var line string
	var err error
	for {
		line, err = rl.Readline()
		if err != nil {
			panic(err)
		}

		if len(line) > 0 {
			break
		}
	}

	if strings.ToLower(line)[0] == 'y' {
		return true
	}

	return false
}

func askString(rl *readline.Instance, question string, def *string) string {
	fmt.Println(question)

	var line string
	var err error
	if def == nil {
		line, err = rl.Readline()
	} else {
		line, err = rl.ReadlineWithDefault(*def)
	}
	if err != nil {
		panic(err)
	}

	return line
}

func Init() {
	rl, err := readline.New("> ")
	if err != nil {
		panic(err)
	}
	defer rl.Close()

	fmt.Println("This initialization process asks a few questions about your project and generates a qtcdbg.toml file.")
	fmt.Println("This toml file is then used on subsequent launches.\n\n")

	//
	// create file befoer asking the user a ton of questions in case there's an error
	//
	if _, err := os.Stat(defaultConfig()); err == nil {
		if !askYesNo(rl, defaultConfig()+" already exists.  Overwrite your config?") {
			fmt.Println("No changes made.")
			os.Exit(1)
		}
	}
	outFile, err := os.Create(defaultConfig())
	if err != nil {
		panic(err)
	}
	defer outFile.Close()

	//
	// ask questions
	//
	var cfg TomlConfig

	// get relative root
	if askYesNo(rl, "Did you just launch qtcdbg from the project repo root?") == false {
		fmt.Println("Re-run \"qtcdbg init\" from your project root")
		os.Exit(1)
	}

	workingDir, err := os.Getwd()
	cfg.Project.RelativeRoot = workingDir
	if err != nil {
		panic(err)
	}

	defaultProjectName := filepath.Base(workingDir)
	cfg.Project.Name = askString(rl, "What is the name of your project?", &defaultProjectName)

	if !askYesNo(rl, "When you launch your compiled program, do you do it from the project root?") {
		cfg.Run.WorkingDir = askString(rl, "What is the working directory, relative to project root, that the debugged executable runs in? (eg: bin/)", nil)
	} else {
		cfg.Run.WorkingDir = "." + Separator
	}

	candidateExecutablePath := cfg.Run.WorkingDir + Separator + cfg.Project.Name
	if runtime.GOOS == "windows" {
		candidateExecutablePath += ".exe"
	}
	cfg.Run.ExecutablePath = askString(rl, "What is the path and filename of the debug executable?", &candidateExecutablePath)

	cfg.Run.Arguments = askString(rl, "Which command line arguments would you like to launch it with when debugging?", nil)
	cfg.Run.RunInTerminal = true

	if askYesNo(rl, "Would you like to be able to build your program inside QtCreator, too?") {
		cfg.Build.WorkingDir = askString(rl, "What is the directory, relative to project root, that your build command runs in? (eg: build/)", nil)
		cfg.Build.Command = askString(rl, "What is the build command?", nil)
		cfg.Build.Arguments = askString(rl, "What are the build command arguments?", nil)
	} else {
		// just launch qtcdbg.exe --help to get it to dummy return 0
		cfg.Build.WorkingDir, _ = filepath.Abs(filepath.Dir(os.Args[0]))
		cfg.Build.Command = os.Args[0]
		cfg.Build.Arguments = "--help"
	}

	//
	// escape slashes in paths
	//
	escape(&cfg.Project.RelativeRoot)
	escape(&cfg.Run.WorkingDir)
	escape(&cfg.Build.WorkingDir)
	escape(&cfg.Run.ExecutablePath)
	escape(&cfg.Build.Command)

	//
	// render template
	//
	tmpl, err := template.New("config").Parse(tmplToml)
	if err != nil {
		panic(err)
	}

	err = tmpl.Execute(outFile, cfg)
	if err != nil {
		fmt.Printf("config:\n%+v\n", tmplToml)
		panic(err)
	}

	fmt.Printf("%s was successfully written with your preferences!", defaultConfig())
	fmt.Println("Feel free to check this file in to source control. It should work for all users.\n")
	fmt.Println("There are a couple options you may want to edit, even after this init procedure:")
	fmt.Println(" - config_defines lets you specify defines that alter QtCreator's source gray-out")
	fmt.Println(" - run_in_terminal can disable the terminal pop-up when debugging if it is not needed\n")
	fmt.Println("Running qtcdbg without arguments is usually enough to launch QtCreator at this point.")
	os.Exit(0)
}
