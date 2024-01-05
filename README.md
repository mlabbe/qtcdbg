# QtCreator Debug Launcher #

**Need**: You want a powerful graphical debugger for your non-Qt project.

**Problem**: Maintaining and distributing a QtCreator project file for a Makefile project _just for debugging_ is a time consuming challenge.

**Solution**: QtCreator Debug Launcher ("`qtcdbg`") discovers your project, generates an adhoc QtCreator project and launches QtCreator for you.

## Usage ##

 1. [Download QtCreator](https://download.qt.io/official_releases/qtcreator/11.0/11.0.2/).  11.0.2 is recommended.  Older versions do not always work.
 2. Linux: Copy QtCreator to your path.  MacOs: Run QtCreator once to clear notarization warning. 
 3. `cd` to project root
 4. Run `qtcdbg init` and answer questions about your project to create a config file.
 5. Type `qtcdbg` to launch QtCreator.

## Downloading ##

See the releases tab on the official github page.

## Installation ##

    mv $GOPATH/bin/qtcdbg /usr/local/bin
    # ensure qtcreator binary is in your path 

## Building ##

Building requires Go.

    go install github.com/mlabbe/qtcdbg/cmd/qtcdbg@latest

## Project Status ##

Project has been successfully used by the original developer.  There is one report of another user having success.

The version of QtCreator matters.  Up until now, QtCreator 11.0.2 and 4.11.0 are the only tested versions.  Use of 11.0.2 is recommended.

This works and is tested on Linux, MacOS and Windows, tested on amd64.

QtCreator hangs on the developer's machine on Windows when debugging but it doesn't seem to be related to qtcdbg.  Have fun!

## FAQs and Troubleshooting ##

 - **I need specific environment variables to be set when I debug my program.**  By default, QtCreator uses environment variables it inherits at its launch when debugging.  Simply launch like this: `ENV_VAR=VALUE qtcdbg` and `ENV_VAR` will be passed along.
 - **Should I check in the files that qtcdbg generates?** No, add them to your gitignore (or similar). They contain local paths and data and cannot be shared.  Only the .toml file should be checked in.
 - **I ran qtcdbg on two projects and they both show up in the file bar in QtCreator. What is happening?**  Qtcdbg launches qtcreator with `-lastsession` in order to maintain breakpoints and open files across launches.  If you find this disagreeable, use QtCreator's session manager to create a new, named session, and then relaunch qtcdbg.
 - **Where are the generated project files?** qtcdbg deletes them after running QtCreator.  If you want to generate them and keep them around, use `qtcdbg launch --no-run`.
 - **Why does my hardware accelerated window not come up when debugging?**  Set `run_in_terminal = false` in the toml file.
 - **I am using the new compile commands override, but it doesn't seem to be working.** Check `preferences->C++->clangd` to ensure "Use clangd" is set.  The path to the executable should be a shell script under a temp directory.

### Future Features ###

 - Support cleaning builds inside QtCreator

### Override compile_commands.json ###

One of the things that has always been a problem is that QtCreator attempts to parse the codebase, imposing its own warnings and defines. It’s very hard to silence and is incongruous with the usual build settings in the project when using the primary text editor.

QtCreator uses clangd on the backend, generating a `compile_commands.json` in a hidden directory.  By overriding that `compile_commands.json` with a real, pre-existing one for the project, the error messages are on par with any LSP editor.

A `compile_commands.json` file can be produced with numerous tools, including `ninja -t compdb`.
