# QtCreator Debug Launcher #

**Need**: You want a powerful graphical debugger for your non-Qt project.

**Problem**: Maintaining and distributing a QtCreator project file for a Makefile project _just for debugging_ is a time consuming challenge.

**Solution**: QtCreator Debug Launcher ("`qtcdbg`") discovers your project, generates an adhoc QtCreator project and launches QtCreator for you.

## Usage ##

 1. [Download QtCreator](https://download.qt.io/official_releases/qtcreator/4.11/4.11.0/).  4.11.0 is recommended.  Older versions do not always work.
 1. Create `qtcdb.toml` file in your project from examples.
 2. `cd` to project root
 3. Type `qtcdbg`

## Building ##

Building requires Go.  When the project matures, binaries will be available in the releases tab.

    go get -u github.com/mlabbe/qtcdbg/cmd/qtcdbg
    
## Installation ##

    mv $GOPATH/bin/qtcdbg /usr/local/bin
    # ensure qtcreator binary is in your path 

## Project Status ##

Project has been successfully used by the original developer.  There is one report of another user having success.

The version of QtCreator matters.  Up until now, QtCreator 4.11.0 is the only tested version.

This works on Linux and while it could work on MacOS, it has not been tested.  It won't work on Windows in its current state.

## FAQs and Troubleshooting ##

 - **Generating a QtCreator project on every launch is fine, but it wipes my breakpoints**.  Actually, no it doesn't.  You can use [QtCreator Sessions](https://doc.qt.io/qtcreator/creator-project-managing-sessions.html) to locally store bookmarks.  They seem to work even if the project is regenerated.
 - **Should I check in the files that QtcDbg generates?** No, add them to your gitignore (or similar). They contain local paths and data and cannot be shared.  Only the .toml file should be checked in.

### Future Features ###

 - Support other QtCreator versions than 4.11.0
 - Release binary
 - Support cleaning builds inside QtCreator
 - Support for scalar variants in the toml config file, ie: `make config=debug_$ARCH`
 - Default to putting generated files in a system temp dir
 - Macos and Windows support
