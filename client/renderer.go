package client

import (
	"fmt"
	"os"
	"github.com/nsf/termbox-go"
)

const (
	BOXRUNE = '█'
	BOX1 = '╔'
	BOX2 = '╗'
	BOX3 = '╚'
	BOX4 = '╝'
	BOX5 = '═'
	BOX6 = '║'
)

var width, height int
var termwidth, termheight int

type Vec struct {
	X, Y int
}

type Snake struct {
	Direction Vec
	Color termbox.Attribute
	Head, Tail *SnakeSection
}

type SnakeSection struct {
	Pos Vec
	Prev, Next *SnakeSection
}

func InitRenderer() (int, int, error) {
	if err := termbox.Init(); err != nil {
		return 0, 0, err
	}

	termwidth, termheight = termbox.Size()
	if termwidth < 20 || termheight < 20 {
		fmt.Fprintf(os.Stderr, "terminal too small\n")
		os.Exit(1)
	}
	if termwidth % 2 == 1 { termwidth -= 1 }
	width, height = (termwidth - 2) / 2, termheight - 3


	termbox.Clear(termbox.ColorDefault, termbox.ColorDefault)

	// Draw Frame
	termbox.SetCell(0, 1, BOX1, termbox.ColorDefault, termbox.ColorDefault)
	termbox.SetCell(termwidth - 1, 1, BOX2, termbox.ColorDefault, termbox.ColorDefault)
	termbox.SetCell(0, termheight - 1, BOX3, termbox.ColorDefault, termbox.ColorDefault)
	termbox.SetCell(termwidth - 1, termheight - 1, BOX4, termbox.ColorDefault, termbox.ColorDefault)

	for i := 1; i < termwidth - 1; i++ {
		termbox.SetCell(i, 1, BOX5, termbox.ColorDefault, termbox.ColorDefault)
		termbox.SetCell(i, termheight - 1, BOX5, termbox.ColorDefault, termbox.ColorDefault)
	}

	for i := 2; i < termheight - 1; i++ {
		termbox.SetCell(0, i, BOX6, termbox.ColorDefault, termbox.ColorDefault)
		termbox.SetCell(termwidth - 1, i, BOX6, termbox.ColorDefault, termbox.ColorDefault)
	}

	return width, height, nil
}

func (snake *Snake) Grow() {
    sec := new(SnakeSection)
    sec.Pos = snake.Tail.Pos
    sec.Prev = snake.Tail
    snake.Tail.Next = sec
    snake.Tail = sec
}

func DrawCell(x, y int, color termbox.Attribute) {
    termbox.SetCell((x * 2) + 1, y + 2, BOXRUNE, color, termbox.ColorDefault)
    termbox.SetCell((x * 2) + 2, y + 2, BOXRUNE, color, termbox.ColorDefault)
}

func ClearCell(x, y int) {
    termbox.SetCell((x * 2) + 1, y + 2, ' ', termbox.ColorDefault, termbox.ColorDefault)
    termbox.SetCell((x * 2) + 2, y + 2, ' ', termbox.ColorDefault, termbox.ColorDefault)
}

func CreateSnake(id, x, y int) *Snake {
	// Build Snake
	snake := &Snake{ Vec{0, 0}, 0, nil, nil }
	if id % 2 == 0 {
		snake.Color = termbox.ColorGreen
	} else {
		snake.Color = termbox.ColorRed
	}

	head := new(SnakeSection)
	head.Pos = Vec{ x, y }
	tail := new(SnakeSection)
	tail.Prev = head
	tail.Pos = Vec{ x, y }
	head.Next = tail
	snake.Head = head
	snake.Tail = tail

	for i := 0; i < 5; i += 1 {
		snake.Grow()
	}

	DrawCell(snake.Head.Pos.X, snake.Head.Pos.Y, snake.Color)

	if err := termbox.Flush(); err != nil {
		panic(err)
	}

	return snake
}

func HandleEvents(dirchan chan Vec) {
	for {
		if event := termbox.PollEvent(); event.Type == termbox.EventKey {
			switch event.Key {
			case termbox.KeyArrowUp:
				dirchan <- Vec{ X:  0, Y: -1 }
			case termbox.KeyArrowDown:
				dirchan <- Vec{ X:  0, Y:  1 }
			case termbox.KeyArrowLeft:
				dirchan <- Vec{ X: -1, Y:  0 }
			case termbox.KeyArrowRight:
				dirchan <- Vec{ X:  1, Y:  0 }
				case termbox.KeyCtrlC: fallthrough
			case termbox.KeyEsc:
				close(dirchan)
				return
			}
		}
	}
}
