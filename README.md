# QtCreator Debug Launcher #

**Need**: QtCreator has a powerful graphical debugger.  You want a powerful graphical debugger for your non-Qt project.

**Problem**: Maintaining a QtCreator project file for a Makefile project _just for debugging_ is a time consuming challenge.

**Solution**: QtCreator Debug Launcher ("`qtcdbg`") discovers your project, generates an adhoc QtCreator project and launches QtCreator for you.

## Usage ##

 1. Create `qtcdb.toml` file in your project from examples
 2. `cd` to project root
 3. Type `qtcdbg`

## Building ##

    go get github.com/mlabbe/qtcdbg
    
## Installation ##

    mv qtcdbg /usr/local/bin
    # ensure QtCreator.sh is in your path 

## Project Status ##

Usable first Alpha. Tested with QtCreator 4.11.0 on Linux.  Could work on MacOS, won't work on Windows.

### Future Features ###

 - Support cleaning builds inside QtCreator
 - Support for scalar variants in the toml config file, ie: `make config=debug_$ARCH`
 - Default to putting generated files in a system temp dir
 - Macos and Windows support
 
