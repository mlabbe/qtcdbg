# QtCreator Debug Launcher #

**Need**: You want a powerful graphical debugger for your non-Qt project.

**Problem**: Maintaining and distributing a QtCreator project file for a Makefile project _just for debugging_ is a time consuming challenge.

**Solution**: QtCreator Debug Launcher ("`qtcdbg`") discovers your project, generates an adhoc QtCreator project and launches QtCreator for you.

## Usage ##

 1. Create `qtcdb.toml` file in your project from examples
 2. `cd` to project root
 3. Type `qtcdbg`

## Building ##

    go get -u github.com/mlabbe/qtcdbg/cmd/qtcdbg
    
## Installation ##

    mv $GOPATH/bin/qtcdbg /usr/local/bin
    # ensure qtcreator binary is in your path 

## Project Status ##

Usable first Alpha. Tested with QtCreator 4.11.0 on Linux.  Could work on MacOS, won't work on Windows.

### Future Features ###

 - Support other QtCreator versions than 4.11.0
 - Release binary
 - Support cleaning builds inside QtCreator
 - Support for scalar variants in the toml config file, ie: `make config=debug_$ARCH`
 - Default to putting generated files in a system temp dir
 - Macos and Windows support
