package main

import (
	"bufio"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/mattn/go-runewidth"
	"github.com/nsf/termbox-go"
)

/*
*	main.go: main file for goless
*	This file contains the entire program
* - geogory tibisov
* */
const GOTOEXIT int = -1

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

// draws the goto bar and handles all the logic
func drawGotoBar(maxY int) (line int) {
	cursorX := 5
	printString(0, maxY, "goto:")
	termbox.SetCursor(cursorX, maxY)
	termbox.Flush()
	userInput := ""
	for {
		Event := termbox.PollEvent()
		if Event.Ch != 0 {
			if Event.Ch >= '0' && Event.Ch <= '9' {
				termbox.SetCell(cursorX, maxY, Event.Ch, termbox.ColorBlack, termbox.ColorWhite)
				userInput += string(Event.Ch)
				cursorX++
				termbox.SetCursor(cursorX, maxY)
				termbox.Flush()
			}
		}

		if Event.Key == termbox.KeyBackspace2 {
			if len(userInput) != 0 {
				termbox.SetCell(cursorX-1, maxY, 0, termbox.ColorDefault, termbox.ColorDefault)
				cursorX--
				termbox.SetCursor(cursorX, maxY)
				userInput = userInput[:len(userInput)-1]
				termbox.Flush()
			} else {
				termbox.HideCursor()
				termbox.Flush()
				return GOTOEXIT
			}
		}
		if Event.Key == termbox.KeyEnter {
			if len(userInput) != 0 {
				gotoLine, _ := strconv.Atoi(userInput)
				termbox.HideCursor()
				termbox.Flush()
				return gotoLine
			} else {
				termbox.HideCursor()
				termbox.Flush()
				return GOTOEXIT
			}
		}
		if Event.Key == termbox.KeyEsc || Event.Ch == 'q' {
			termbox.HideCursor()
			termbox.Flush()
			return GOTOEXIT
		}

	}

}

// draws te prompt for searching and handles all logic
func drawSearchBar(maxY int, searchMode *bool) (searchToken string) {
	*searchMode = true //set the value to true so the main loop knows how to handle the n key
	cursorX := 1
	termbox.SetChar(0, maxY, '/') //places the / char in the bottom left corner so the user knows to type input
	termbox.SetFg(0, maxY, termbox.ColorBlack)
	termbox.SetBg(0, maxY, termbox.ColorWhite)
	termbox.SetCursor(cursorX, maxY)
	termbox.Flush()
	var retToken string
	for {
		Event := termbox.PollEvent()
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
				if cursorX != 1 {
					termbox.SetCell((cursorX - 1), maxY, 0, termbox.ColorDefault, termbox.ColorDefault)
					retToken = retToken[:len(retToken)-1]
					cursorX--
					termbox.SetCursor(cursorX, maxY)
					termbox.Flush()
				}
			}
			if Event.Key == termbox.KeyEsc || Event.Key == termbox.KeyCtrlC || cursorX == 0 { // on escape or control c return empty string and set searchMode to false
				*searchMode = false
				termbox.Clear(termbox.ColorDefault, termbox.ColorDefault)
				termbox.HideCursor()
				return ""
			}
			if Event.Key == termbox.KeyEnter { // return token on enterkey
				if len(retToken) == 0 {
					*searchMode = false
					return ""
				}
				termbox.HideCursor()
				return retToken
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
				*nextPattern++
				return tempIndex, (tempIndex + len(token) - 1), index
			}
			nextPaternLocal++
		}
	}
	return -1, -1, -1
}
func highlightLine(line, maxX, maxY int, textdata *textData_t) {
	y := line
	x := 0
	if y > maxY {
		textdata.scrollOffset = line
		y = 0
	}
	if line < textdata.scrollOffset {
		textdata.scrollOffset = line
		y = 0
	}
	drawText(*textdata, maxX, maxY)
	for x <= len(textdata.lines[line]) {
		termbox.SetBg(x, y, termbox.ColorWhite)
		termbox.SetFg(x, y, termbox.ColorBlack)
		x++
	}
	termbox.Flush()
}
func highlightPattern(start, end, line, maxX, maxY int, textdata *textData_t) {
	x := start

	if line >= maxY {
		textdata.scrollOffset = line - 3
	}
	drawText(*textdata, maxX, maxY)
	for x <= end {
		termbox.SetBg(x, (line - textdata.scrollOffset), termbox.ColorWhite)
		termbox.SetFg(x, (line - textdata.scrollOffset), termbox.ColorBlack)
		x++
	}
	termbox.Flush()
}

// prints a string
func printString(x, y int, str string) {
	for _, char := range str {
		termbox.SetCell(x, y, char, termbox.ColorBlack, termbox.ColorWhite)
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
		stat, _ := os.Stdin.Stat()
		if stat.Mode()&os.ModeCharDevice == 0 {
			textdata.lines = loadText(os.Stdin)
		} else {
			panic("No file specfied or text piped to program")
		}
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
	for {
		Event := termbox.PollEvent()
		if Event.Ch == 'g' && searchMode == false {
			gotoLine := drawGotoBar(maxY - 1)
			if gotoLine < 1 || gotoLine > len(textdata.lines) {
				drawText(textdata, maxX, maxY)
				printString(0, maxY-1, "Invaild line: "+("'"+strconv.Itoa(gotoLine)+"'"))
			} else {
				highlightLine(gotoLine-1, maxX, maxY-1, &textdata)
			}
		}
		if Event.Ch == '/' {
			drawText(textdata, maxX, maxY)
			nextPattern = 0
			token = ""
			token = drawSearchBar(maxY-1, &searchMode)
			tokenStart, tokenEnd, tokenLine = findPattern(textdata.lines, token, &nextPattern)
			if tokenLine == -1 {
				searchMode = false
				textdata.scrollOffset = 0
				drawText(textdata, maxX, maxY)
				printString(0, maxY-1, "Pattern Not found: "+("'"+token+"'"))
				token = ""
			} else {
				highlightPattern(tokenStart, tokenEnd, tokenLine, maxX, maxY, &textdata)
			}
		}
		if Event.Ch == 'n' && searchMode == true {
			tokenStart, tokenEnd, tokenLine = findPattern(textdata.lines, token, &nextPattern)
			if tokenLine == -1 {
				printString(0, maxY-1, "Last instance of pattern: "+("'"+token+"'"))

			} else {
				highlightPattern(tokenStart, tokenEnd, tokenLine, maxX, maxY, &textdata)
			}
		}
		// not working currently for some reason?
		if Event.Ch == 'p' && searchMode == true {
			nextPattern--
			if nextPattern < 0 {
				nextPattern = 0
				searchMode = false
			}
			tokenStart, tokenEnd, tokenLine = findPattern(textdata.lines, token, &nextPattern)
			highlightPattern(tokenStart, tokenEnd, tokenLine, maxX, maxY, &textdata)
			if nextPattern == 0 {
				printString(0, maxY-1, "First Instance of pattern: "+("'"+token+"'"))
			}
		}
		//I need to impliment text wrapping for this to ever be usefull
		if Event.Type == termbox.EventResize {
			maxX, maxY = termbox.Size()
			if searchMode == true {
				highlightPattern(tokenStart, tokenEnd, tokenLine, maxX, maxY, &textdata)
			} else {
				drawText(textdata, maxX, maxY)
			}
		}
		if Event.Type == termbox.EventKey {
			switch Event.Key {
			case termbox.KeyPgup:
				textdata.scrollOffset -= (maxY - 1)
				if textdata.scrollOffset < 0 {
					textdata.scrollOffset = 0
				}
				drawText(textdata, maxX, maxY)
			case termbox.KeyPgdn:
				textdata.scrollOffset += (maxY - 1)
				if textdata.scrollOffset > len(textdata.lines) {
					textdata.scrollOffset = len(textdata.lines)
				}
				drawText(textdata, maxX, maxY)
				if textdata.scrollOffset == len(textdata.lines) {
					printString(0, maxY-1, "End Of Buffer")
				}
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
				if searchMode == true {
					searchMode = false
					drawText(textdata, maxX, maxY)
				} else {
					return
				}
			}

		}
	}
}
