package fsac

import (
	"fmt"
	ansi "github.com/solidiquis/ansigo"
	"os"
	"testing"
)

var errors []string

func TestMain(m *testing.M) {
	// Excute tests
	exit := m.Run()

	// Clean the screen
	ansi.EraseScreen()
	ansi.CursorHome()

	// Print errors workaround..
	for _, e := range errors {
		fmt.Println(e)
	}
	os.Exit(exit)
}

// TODO: Definitely needs more thorough testing.
func TestSearch(t *testing.T) {
	fsac := fsacSetup()

	if len(fsac.Matches) > 0 {
		e("Matches should have 0 items on init.", t)
	}

	if fsac.makeSelection() != "cthulhu" {
		e("First selection should be first item in Items.", t)
	}

	fsac.Render("n")
	if fsac.makeSelection() != "nyarlathotep" {
		e("When provided char 'n', focused item should be 'nyarlathotep'.", t)
	}

	fsac.Render("<Backspace>")
	fsac.Render("sh")
	if fsac.Value == "sh" && fsac.makeSelection() != "shoggoth" {
		e("When 'sh' is Value 'shoggoth' should be top match.", t)
	}

	fsac.incSelected()
	if fsac.Selected != 1 {
		e("incSelected should increment Selected to 1,", t)
	}

	if fsac.makeSelection() != "shub-niggurath" {
		e("When Value is 'sh' and Selected is 1, selected item should be 'shub-niggurath'.", t)
	}
}

func e(err string, t *testing.T) {
	errors = append(errors, err)
	t.Error(err)
}

func fsacSetup() *search {
	prompt := "Directory: "
	done := make(chan string, 1)
	fsac := InitSearch(prompt, done)
	fsac.SetItems([]string{
		"cthulhu",
		"nyarlathotep",
		"yog-sothoth",
		"azathoth",
		"shub-niggurath",
		"shoggoth",
	})

	return fsac
}
