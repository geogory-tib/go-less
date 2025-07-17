package main

import (
	"bufio"
	"log"
	"os"

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
		if y == maxY { // if text meets the border of it stops drawing text
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
