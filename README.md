# fsac
<a href="https://pkg.go.dev/github.com/solidiquis/fsac">
  <img src="https://godoc.org/github.com/golang/gddo?status.svg" alt="GoDoc">
</a>

<a href="https://github.com/solidiquis/fsac/actions">
  <img src="https://github.com/solidiquis/fsac/workflows/Go/badge.svg" alt="Build Status">
</a>

Lightning fast scrollable **f**uzzy **s**earch utility for the terminal, the result of which can be easily provided as an **a**rgument to a custom **c**ommand. This repo is intended to contain a collection of applications that make use of **fsac**.

<img height="auto" width="75%" src="https://github.com/solidiquis/solidiquis/blob/master/assets/fsac_demo.gif">

## How to install an fsac program

Make sure you have <a href="https://golang.org/dl/">Go</a> installed.

**Option 1**:
`git clone` this repo and from the project's root directory:
```
$ make compile_dirsearch dest=path/to/my/bin_name
```
Make sure that the executable is in your `$PATH`.

**Option 2**:
Make sure you `$GOPATH` is set, then run: 
```
$ go get github.com/solidiquis/fsac/cmd/dirf
``` 

## Wish to contribute by making an fsac program?
Again, this repo will contain a collection of programs which can be compiled independently. Here are the steps to follow if you wish to contribute:
1. Make a new directory in `cmd/` named after your program.
2. Make your program and give it a short catchy name, kind of like `grep`, which is short for `global regular expression print`.
3. Add something like this to the top of your `main.go`:
```
// Installation:
// go get github.com/solidiquis/fsac/cmd/dirf

// Name:
// dirf -> (dir)ectory (f)ind

// Utility:
// Searches through directory tree from working directory and
// copies selection to clipboard.
```
4. Add three new targets to the `Makefile`, using `dirf` as a template: `run`, `debug`, & `compile`.

## Libraries used
- https://github.com/sahilm/fuzzy: Awesome string matching algorithm.
- https://github.com/solidiquis/ansigo: Convenient Go-wrapper for various ANSI escapes + keypress detection.

## License
<a href="https://github.com/solidiquis/fsac/blob/master/LICENSE">MIT</a>
