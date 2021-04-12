package fsac

import (
	"fmt"
	"github.com/sahilm/fuzzy"
	"log"

	ansi "github.com/solidiquis/ansigo"
)

const (
	// Minimum space between bottom of list to bottom of window.
	WIN_MARGIN_BTM = 4

	// Spacing between left of window to items.
	LIST_OFFSET = "  "
)

// Struct that encapsulates everything necessary to render fsac.
// Struct is unexported, meant to be instantiated with InitFsac.
type fsac struct {
	// Text preceding user input.
	Prompt string

	// User input value.
	Value string

	// Item in Items that is most similar to Value.
	Guess string

	// List of items to fsac through.
	Items []string

	// Match objects ordered by most similar to Value to least similar.
	Matches []fuzzy.Match

	// Cursor column and row position for where user types input.
	StartCol int
	StartRow int

	// Index representing the focused item from the list of Matches/Items.
	Selected int

	// Dimensions of the terminal window running fsac.
	WinCol int
	WinRow int

	// Channel to send selection through.
	doneChan chan string
}

// Clears window, records window dimensions, prints prompt, and instantiates fsac.
func InitFsac(promptTxt string, done chan string) *fsac {
	ansi.EraseScreen()
	ansi.CursorHome()

	winCol, winRow, err := ansi.TerminalDimensions()
	if err != nil {
		log.Fatalln("Unable to retrieve window dimensions.")
	}

	prompt := fmt.Sprintf("%s: ", promptTxt)

	fmt.Print(ansi.Bright(prompt))

	return &fsac{
		Prompt:   prompt,
		Value:    "",
		Guess:    "",
		StartCol: len(prompt) + 1,
		StartRow: 1,
		Selected: 0,
		WinCol:   winCol,
		WinRow:   winRow,
		doneChan: done,
	}
}

// Main handler for rendering behavior.
func (f *fsac) Render(key string) {
	var scrolling bool

	switch key {
	case "<Enter>":
		f.doneChan <- f.makeSelection()
		return
	case "<Backspace>":
		f.backspace()
	case "<Up>":
		f.decSelected()
		scrolling = true
	case "<Down>":
		f.incSelected()
		scrolling = true
	case "\t":
		f.autocomplete()
	case "<Left>", "<Right>", "<ESC>":
		return
	default:
		f.printChar(key)
	}

	// If items not set, no autocomplete and list render.
	if len(f.Items) < 1 {
		return
	}

	// Autocomplete goodnesf.
	var prediction string

	if match := fuzzy.Find(f.Value, f.Items); len(match) > 0 {
		prediction = match[0].Str
		f.Matches = match
	} else {
		f.Matches = nil
	}

	if !scrolling {
		f.Selected = 0
	}

	if len(prediction) > 0 {
		f.printGuess(prediction)
	} else {
		f.Guess = ""
	}

	f.RenderMatches()
}

// Simple setter for setting list of items. Items are not set in the InitFsac
// func in case the preparing of said items is a slow operation. By setting the
// items separately, say, in a goroutine, this allows the application to display
// the prompt immediately without being blocked.
func (f *fsac) SetItems(items []string) {
	f.Items = items
}

// Renders a list of all the potential items that match user input, taking into
// account the window dimesions, using a sliding window to handle overflow.
func (f *fsac) RenderMatches() {
	ansi.CursorSavePos()
	ansi.CursorHome()
	ansi.CursorDownStartLn(2)

	frameLen := f.WinRow + f.Selected - WIN_MARGIN_BTM

	if len(f.Matches) > 0 {
		for i := f.Selected; i < frameLen; i++ {
			if i >= len(f.Matches) {
				ansi.EraseLine()
				ansi.CursorDownStartLn(1)
				continue
			}

			item := f.Matches[i].Str

			if len(item)+len(LIST_OFFSET) >= f.WinCol {
				item = f.truncate(item)
			}

			ansi.EraseLine()

			if i == f.Selected {
				fmt.Println(f.focus(item))
				continue
			}
			fmt.Println(LIST_OFFSET, item)
		}

		ansi.CursorRestorePos()
		return
	}

	for i := f.Selected; i < frameLen; i++ {
		if i >= len(f.Items) {
			ansi.EraseLine()
			ansi.CursorDownStartLn(1)
			continue
		}

		item := f.Items[i]

		if len(item)+len(LIST_OFFSET) >= f.WinCol {
			item = f.truncate(item)
		}

		ansi.EraseLine()

		if i == f.Selected {
			fmt.Println(f.focus(item))
			continue
		}
		fmt.Println("  ", item)
	}

	ansi.CursorRestorePos()
}

/* PRIVATE */

// Truncates and prepends '...' to matches that overflow to next line.
func (f *fsac) truncate(item string) string {
	overflowAmnt := len(LIST_OFFSET) + len(item) - f.WinCol
	truncatedStr := item[overflowAmnt+4:]
	return fmt.Sprintf("...%s", truncatedStr)

}

// Increments Selected field for proper RenderMatches behavior.
func (f *fsac) incSelected() {
	if len(f.Matches) > 0 {
		if f.Selected+1 >= len(f.Matches) {
			return
		}

		f.Selected++
	} else {
		if f.Selected+1 >= len(f.Items) {
			return
		}

		f.Selected++
	}
}

// Decrements Selected field for proper RenderMatches behavior.
func (f *fsac) decSelected() {
	if len(f.Matches) > 0 {
		if f.Selected-1 < 0 {
			return
		}

		f.Selected--
	} else {
		if f.Selected-1 < 0 {
			return
		}

		f.Selected--
	}
}

// Returns a highlighted string to represent that the item is focused on.
func (f *fsac) focus(item string) string {
	return fmt.Sprintf(" %s %s", ansi.FgBlue("\u25B6"), ansi.FgMagenta(item))
}

// Used to determine whether or not user input with/without autocomplete
// will overflow onto the next line.
func (f *fsac) lnOverflow() bool {
	return len(f.Prompt)+len(f.Guess) > f.WinCol || len(f.Prompt)+len(f.Value) > f.WinCol
}

// Handler for Tab press, prints the predicted text and set the Value to Guesf.
func (f *fsac) autocomplete() {
	if len(f.Guess) < 1 {
		return
	}
	ansi.CursorSetPos(f.StartRow, f.StartCol)
	fmt.Print(f.Guess)
	ansi.EraseToEndln()
	f.Value = f.Guess
}

// Handler for backpacing.
func (f *fsac) backspace() {
	if len(f.Value) < 1 {
		return
	}

	if f.lnOverflow() {
		ansi.CursorSavePos()
		ansi.CursorDown(1)
		ansi.EraseLine()
		ansi.CursorRestorePos()
	}
	ansi.EraseToEndln()
	ansi.Backspace()
	f.Value = f.Value[:len(f.Value)-1]
}

// Handler for printing normal input.
func (f *fsac) printChar(ch string) {
	f.Value += ch
	fmt.Print(ch)
}

// Prints the predicted text as dim text.
func (f *fsac) printGuess(prediction string) {
	f.Guess = prediction
	ansi.CursorSavePos()
	ansi.EraseToEndln()
	fmt.Print(ansi.Dim(prediction[len(f.Value):]))
	ansi.CursorRestorePos()
}

// Handler for when user presses "ENTER" or "RETURN", returns focused item.
func (f *fsac) makeSelection() string {
	if len(f.Matches) > 0 {
		return f.Matches[f.Selected].Str
	} else {
		return f.Items[f.Selected]
	}
}
