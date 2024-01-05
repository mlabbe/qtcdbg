/*
 * qtcdbg Copyright (C) 2019-2020, 2024 Frogtoss Games, Inc.
 */

package main

import (
	"errors"
	"fmt"
	"gopkg.in/ini.v1"
	"os"
	"os/exec"
	"path/filepath"
	"text/template"
)

var tmplClangWrapperSh = `#!/bin/sh
{{ .ClangdPath }} $@ --compile-commands-dir={{ .CompileCommandsDir }}
`

// set the clangd path, rewriting the .ini file
func SetClangdPath(clangdPath string) error {
	// clangd path is in QtCreator.ini's ClangdSettings.ClangdPath

	iniPath, err := GetIniPath()
	if err != nil {
		return err
	}

	cfg, err := ini.Load(iniPath)
	if err != nil {
		return err
	}

	cfg.Section("ClangdSettings").Key("ClangdPath").SetValue(clangdPath)
	cfg.SaveTo(iniPath)

	return nil
}

func GetClangdPath() (string, error) {
	// clangd path is in QtCreator.ini's ClangdSettings.ClangdPath

	iniPath, err := GetIniPath()
	if err != nil {
		return "", err
	}

	cfg, err := ini.Load(iniPath)
	if err != nil {
		return "", err
	}

	//
	// first attempt: get it from the .ini file
	//
	clangdPath := cfg.Section("ClangdSettings").Key("ClangdPath").String()

	// clangdPath is explicit in the ini
	if clangdPath != "" {

		if *debug {
			fmt.Printf("clangdPath: %s\n", clangdPath)
		}

		return clangdPath, nil
	}

	//
	// second attempt: get the qtcreator bundled clangd
	//
	qtCreatorPath, err := exec.LookPath("qtcreator")
	if err != nil {
		return "", err
	}
	qtCreatorDir := filepath.Dir(qtCreatorPath)
	clangdPath, err = filepath.Abs(qtCreatorDir + "/../libexec/qtcreator/clang/bin/clangd")
	fmt.Printf("path: %v", clangdPath)
	if _, err := os.Stat(clangdPath); err == nil {
		if *debug {
			fmt.Printf("clangdPath: %s\n", clangdPath)
		}

		return clangdPath, nil
	}

	//
	// final attempt: get clangd from system path
	//
	clangdPath, err = exec.LookPath("clangd")
	if err != nil {
		return "", err
	}
	if _, err := os.Stat(clangdPath); err == nil {
		if *debug {
			fmt.Printf("clangdPath: %s\n", clangdPath)
		}

		return clangdPath, nil
	}

	return "", errors.New("Could not find clangdPath")
}

// the Clangd wrapper script calls the original clangd program, but forces '--compile-commands-json=' onto the end,
// overriding the one specified in the QtCreator.ini.
//
// It is created in a temp file, and is cleaned up at the end of the program, and the QtCreator.ini is returned
// to its original setting.
func WriteClangdWrapper(clangdPath, compileCommandsDir string) (string, error) {

	tempFile, err := os.CreateTemp("", ".qtcdbg-clangd-wrapper")
	if err != nil {
		return "", err
	}

	defer tempFile.Close()

	err = os.Chmod(tempFile.Name(), 0700)
	if err != nil {
		return "", fmt.Errorf("Error chmodding clangd wrapper script: %+v", err)
	}

	//
	// render template
	//
	tmpl, err := template.New("wrapper").Parse(tmplClangWrapperSh)
	if err != nil {
		return "", err
	}

	err = tmpl.Execute(tempFile, struct {
		ClangdPath         string
		CompileCommandsDir string
	}{
		ClangdPath:         clangdPath,
		CompileCommandsDir: compileCommandsDir,
	})
	if err != nil {
		return "", err
	}

	if *debug {
		fmt.Printf("wrote clangd wrapper at '%s'\n", tempFile.Name())
	}

	return tempFile.Name(), nil
}
