package main

import (
	"bufio"
	"log"
	"os"
	"strings"

	"github.com/mattn/go-runewidth"
	"github.com/nsf/termbox-go"
)

type textData_t struct {
	lines        []string
	scrollOffset int
}

// loads text data into []lines buffer
func loadText(file *os.File) (lines []string) {
	scanner := bufio.NewScanner(file)
	var returnSlice []string
	for scanner.Scan() {
		line := scanner.Text() + "\n"
		returnSlice = append(returnSlice, line)
	}
	return returnSlice
}
func drawText(textdata textData_t, maxX, maxY int) {
	x := 0 // local x and y within the function used to assign the cells
	y := 0
	termbox.Clear(termbox.ColorDefault, termbox.ColorDefault)    // resets the terminal screen
	for _, str := range textdata.lines[textdata.scrollOffset:] { // indexes the strings through the line slice
		if y == maxY-1 { // if text meets the border of it stops drawing text
			break
		}
		for _, char := range str { // loops through string and assigns the char in the cell
			termbox.SetChar(x, y, char)
			x += runewidth.RuneWidth(char)
			if char == '\n' || x == maxX { // goes to the next row on the terminal if the text is longer than the width of the screen or if the char is a newline
				x = 0
				y++
				break
			}
		}
	}
	termbox.Flush() // flush buffer
}

// draws te prompt for searching
func drawSearchBar(maxY int, searchMode *bool, eventQueue chan termbox.Event) (searchToken string) {
	*searchMode = true //set the value to true so the main loop knows how to handle the n key
	cursorX := 1
	termbox.SetChar(0, maxY, '/') //places the / char in the bottom left corner so the user knows to type input
	termbox.SetFg(0, maxY, termbox.ColorBlack)
	termbox.SetBg(0, maxY, termbox.ColorWhite)
	termbox.SetCursor(cursorX, maxY)
	termbox.Flush()
	var retToken string
	for {
		select {
		case Event := <-eventQueue:
			if Event.Ch != 0 {
				termbox.SetCell(cursorX, maxY, Event.Ch, termbox.ColorBlack, termbox.ColorWhite)
				retToken += string(Event.Ch)
				cursorX++
				termbox.SetCursor(cursorX, maxY)
				termbox.Flush()
			}
			if Event.Ch == 0 {
				if Event.Key == termbox.KeySpace {
					termbox.SetCell(cursorX, maxY, 0, termbox.ColorBlack, termbox.ColorWhite)
					retToken += " "
					cursorX++
					termbox.SetCursor(cursorX, maxY)
					termbox.Flush()
				}
				if Event.Key == termbox.KeyBackspace2 { //handle backspaces
					termbox.SetCell((cursorX - 1), maxY, 0, termbox.ColorDefault, termbox.ColorDefault)
					retToken = retToken[:len(retToken)-1]
					cursorX--
					termbox.SetCursor(cursorX, maxY)
					termbox.Flush()
				}
				if Event.Key == termbox.KeyEsc || Event.Key == termbox.KeyCtrlC || cursorX == 0 { // on escape or control c return empty string and set searchMode to false
					*searchMode = false
					termbox.Clear(termbox.ColorDefault, termbox.ColorDefault)
					termbox.HideCursor()
					return ""
				}
				if Event.Key == termbox.KeyEnter { // return token on enterkey
					termbox.HideCursor()
					return retToken
				}
			}
		}
	}
}

// finds the token in the lines slice converts every string to lower case to remove case senstivity
func findPattern(lines []string, token string, nextPattern *int) (start, end, line int) {
	tokenL := strings.ToLower(token)
	nextPaternLocal := 0 //local variable used to compare the nextPattern to this
	for index, lineData := range lines {
		lineDataL := strings.ToLower(lineData)
		tempIndex := strings.Index(lineDataL, tokenL) //finds the starting index of the pattern
		if tempIndex != -1 {
			if *nextPattern == nextPaternLocal {
				line = index
				start = tempIndex
				end = tempIndex + len(token) - 1
				*nextPattern++
				return
			}
			nextPaternLocal++
		}
	}
	return -1, -1, -1
}

func highlightPattern(start, end, line, maxX, maxY int, textdata *textData_t) {
	x := start

	if line >= maxY {
		textdata.scrollOffset = line - 3
	}
	drawText(*textdata, maxX, maxY)
	for x <= end {
		termbox.SetBg(x, (line - textdata.scrollOffset), termbox.ColorWhite)
		termbox.SetFg(x, line, termbox.ColorBlack)
		x++
	}
	termbox.Flush()
}
func printString(x, y int, str string) {
	for _, char := range str {
		termbox.SetChar(x, y, char)
		x += runewidth.RuneWidth(char)
	}
	termbox.Flush()
}
func main() {
	err := termbox.Init()
	defer termbox.Close()
	if err != nil {
		log.Fatal(err)
	}
	var textdata textData_t
	if len(os.Args) == 1 {
		textdata.lines = loadText(os.Stdin)
	} else {
		file, err := os.Open(os.Args[1])
		if err != nil {
			log.Fatal(err)
		}
		defer file.Close()
		textdata.lines = loadText(file)
	}
	maxX, maxY := termbox.Size()
	searchMode := false
	token := ""      //holds the string for the pattern to be searched by the user
	nextPattern := 0 //used for when the user presses n so the findPattern knows the skip the already found patterns
	var tokenStart, tokenEnd, tokenLine int

	drawText(textdata, maxX, maxY)
	eventQueue := make(chan termbox.Event)
	//the thread dealing  with events did this for future functionality
	go func() {
		for {
			eventQueue <- termbox.PollEvent()
		}
	}()
	for {
		Event := <-eventQueue
		if Event.Ch == '/' {
			nextPattern = 0
			token = ""
			token = drawSearchBar(maxY-1, &searchMode, eventQueue)
			tokenStart, tokenEnd, tokenLine = findPattern(textdata.lines, token, &nextPattern)
			if tokenLine == -1 {
				searchMode = false
				textdata.scrollOffset = 0
				drawText(textdata, maxX, maxY)
				printString(maxX, maxY, "Token Not found") // unsure of why this isnt working??
			} else {
				highlightPattern(tokenStart, tokenEnd, tokenLine, maxX, maxY, &textdata)
			}
		}
		if Event.Ch == 'n' && searchMode == true {
			tokenStart, tokenEnd, tokenLine = findPattern(textdata.lines, token, &nextPattern)
			highlightPattern(tokenStart, tokenStart, tokenLine, maxX, maxY, &textdata)
		}
		if Event.Type == termbox.EventKey {
			switch Event.Key {
			case termbox.KeyArrowDown:
				if textdata.scrollOffset != len(textdata.lines) {
					textdata.scrollOffset++
					drawText(textdata, maxX, maxY)
				}
			case termbox.KeyArrowUp: //scroll up
				if textdata.scrollOffset != 0 { // do nothing if its 0
					textdata.scrollOffset--
					drawText(textdata, maxX, maxY)
				}
			case termbox.KeyEsc, termbox.KeyCtrlC:
				return

			}

		}
	}
}
