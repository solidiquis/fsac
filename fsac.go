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
// Struct is unexported, meant to be instantiated with InitSearch.
type search struct {
	// Text preceding user input.
	Prompt string

	// User input value.
	Value string

	// Item in Items that is most similar to Value.
	Guess string

	// List of items to search through.
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

// Clears window, records window dimensions, prints prompt, and instantiates search.
func InitSearch(promptTxt string, done chan string) *search {
	ansi.EraseScreen()
	ansi.CursorHome()

	winCol, winRow, err := ansi.TerminalDimensions()
	if err != nil {
		log.Fatalln("Unable to retrieve window dimensions.")
	}

	prompt := fmt.Sprintf("%s: ", promptTxt)

	fmt.Print(ansi.Bright(prompt))

	return &search{
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
func (s *search) Render(key string) {
	var scrolling bool

	switch key {
	case "<Enter>":
		s.doneChan <- s.makeSelection()
		return
	case "<Backspace>":
		s.backspace()
	case "<Up>":
		s.decSelected()
		scrolling = true
	case "<Down>":
		s.incSelected()
		scrolling = true
	case "\t":
		s.autocomplete()
	case "<Left>", "<Right>", "<ESC>":
		return
	default:
		s.printChar(key)
	}

	// If items not set, no autocomplete and list render.
	if len(s.Items) < 1 {
		return
	}

	// Autocomplete goodness.
	var prediction string

	if match := fuzzy.Find(s.Value, s.Items); len(match) > 0 {
		prediction = match[0].Str
		s.Matches = match
	} else {
		s.Matches = nil
	}

	if !scrolling {
		s.Selected = 0
	}

	if len(prediction) > 0 {
		s.printGuess(prediction)
	} else {
		s.Guess = ""
	}

	s.RenderMatches()
}

// Simple setter for setting list of items. Items are not set in the InitSearch
// func in case the preparing of said items is a slow operation. By setting the
// items separately, say, in a goroutine, this allows the application to display
// the prompt immediately without being blocked.
func (s *search) SetItems(items []string) {
	s.Items = items
}

// Renders a list of all the potential items that match user input, taking into
// account the window dimesions, using a sliding window to handle overflow.
func (s *search) RenderMatches() {
	ansi.CursorSavePos()
	ansi.CursorHome()
	ansi.CursorDownStartLn(2)

	frameLen := s.WinRow + s.Selected - WIN_MARGIN_BTM

	if len(s.Matches) > 0 {
		for i := s.Selected; i < frameLen; i++ {
			if i >= len(s.Matches) {
				ansi.EraseLine()
				ansi.CursorDownStartLn(1)
				continue
			}

			item := s.Matches[i].Str

			if len(item)+len(LIST_OFFSET) >= s.WinCol {
				item = s.truncate(item)
			}

			ansi.EraseLine()

			if i == s.Selected {
				fmt.Println(s.focus(item))
				continue
			}
			fmt.Println(LIST_OFFSET, item)
		}

		ansi.CursorRestorePos()
		return
	}

	for i := s.Selected; i < frameLen; i++ {
		if i >= len(s.Items) {
			ansi.EraseLine()
			ansi.CursorDownStartLn(1)
			continue
		}

		item := s.Items[i]

		if len(item)+len(LIST_OFFSET) >= s.WinCol {
			item = s.truncate(item)
		}

		ansi.EraseLine()

		if i == s.Selected {
			fmt.Println(s.focus(item))
			continue
		}
		fmt.Println("  ", item)
	}

	ansi.CursorRestorePos()
}

/* PRIVATE */

// Truncates and prepends '...' to matches that overflow to next line.
func (s *search) truncate(item string) string {
	overflowAmnt := len(LIST_OFFSET) + len(item) - s.WinCol
	truncatedStr := item[overflowAmnt+4:]
	return fmt.Sprintf("...%s", truncatedStr)

}

// Increments Selected field for proper RenderMatches behavior.
func (s *search) incSelected() {
	if len(s.Matches) > 0 {
		if s.Selected+1 >= len(s.Matches) {
			return
		}

		s.Selected++
	} else {
		if s.Selected+1 >= len(s.Items) {
			return
		}

		s.Selected++
	}
}

// Decrements Selected field for proper RenderMatches behavior.
func (s *search) decSelected() {
	if len(s.Matches) > 0 {
		if s.Selected-1 < 0 {
			return
		}

		s.Selected--
	} else {
		if s.Selected-1 < 0 {
			return
		}

		s.Selected--
	}
}

// Returns a highlighted string to represent that the item is focused on.
func (s *search) focus(item string) string {
	return fmt.Sprintf(" %s %s", ansi.FgBlue("\u25B6"), ansi.FgMagenta(item))
}

// Used to determine whether or not user input with/without autocomplete
// will overflow onto the next line.
func (s *search) lnOverflow() bool {
	return len(s.Prompt)+len(s.Guess) > s.WinCol || len(s.Prompt)+len(s.Value) > s.WinCol
}

// Handler for Tab press, prints the predicted text and set the Value to Guess.
func (s *search) autocomplete() {
	if len(s.Guess) < 1 {
		return
	}
	ansi.CursorSetPos(s.StartRow, s.StartCol)
	fmt.Print(s.Guess)
	ansi.EraseToEndln()
	s.Value = s.Guess
}

// Handler for backpacing.
func (s *search) backspace() {
	if len(s.Value) < 1 {
		return
	}

	if s.lnOverflow() {
		ansi.CursorSavePos()
		ansi.CursorDown(1)
		ansi.EraseLine()
		ansi.CursorRestorePos()
	}
	ansi.EraseToEndln()
	ansi.Backspace()
	s.Value = s.Value[:len(s.Value)-1]
}

// Handler for printing normal input.
func (s *search) printChar(ch string) {
	s.Value += ch
	fmt.Print(ch)
}

// Prints the predicted text as dim text.
func (s *search) printGuess(prediction string) {
	s.Guess = prediction
	ansi.CursorSavePos()
	ansi.EraseToEndln()
	fmt.Print(ansi.Dim(prediction[len(s.Value):]))
	ansi.CursorRestorePos()
}

// Handler for when user presses "ENTER" or "RETURN", returns focused item.
func (s *search) makeSelection() string {
	if len(s.Matches) > 0 {
		return s.Matches[s.Selected].Str
	} else {
		return s.Items[s.Selected]
	}
}
