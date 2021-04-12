package main

import (
	"fmt"
	"io/fs"
	"log"
	"os/exec"
	"path/filepath"

	ansi "github.com/solidiquis/ansigo"
	"github.com/solidiquis/fsac"
)

func must(err error) {
	if err != nil {
		log.Fatalln(err)
	}
}

func getDirs() []string {
	var dirs []string
	filepath.WalkDir(".", func(path string, dir fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if dir.IsDir() {
			dirs = append(dirs, path)
		}

		return nil
	})

	return dirs
}

func pbcopy(input string) error {
	cmd := exec.Command("pbcopy")
	subStdin, err := cmd.StdinPipe()
	must(err)

	err = cmd.Start()
	must(err)

	_, err = subStdin.Write([]byte(input))
	must(err)

	subStdin.Close()

	return cmd.Wait()
}

func main() {
	stdin := make(chan string, 1)
	go ansi.GetChar(stdin)

	done := make(chan string, 1)
	search := fsac.InitSearch("Directory", done)

	itemsSet := make(chan bool, 1)
	go func() {
		search.SetItems(getDirs())
		itemsSet <- true
	}()

	for {
		select {
		case key := <-stdin:
			search.Render(key)
		case <-itemsSet:
			search.RenderMatches()
		case value := <-done:
			ansi.EraseScreen()
			ansi.CursorHome()
			err := pbcopy(value)
			if err == nil {
				fmt.Printf("%s copied to clipboard.\n", ansi.FgRed(value))
			} else {
				log.Fatalln(err)
			}
			return
		}
	}
}
