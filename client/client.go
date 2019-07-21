package client

import (
	"fmt"
	"net"
	"os"
	"bufio"
	"time"
	"../utils"
	"../events"
	"github.com/nsf/termbox-go"
)

var ToMe chan *events.Event
var ToServer chan *events.Event
var MyId int
var PeerId int
var mySnake *Snake
var peerSnake *Snake

func Start(addr string, nick string, peernick string) error {
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return err
	}

	width, height, err := InitRenderer()
	if err != nil {
		return err
	}

	ToMe = make(chan *events.Event)
	ToServer = make(chan *events.Event)

	go func(){
		reader := bufio.NewReader(conn)
		for {
			line, err := reader.ReadString('\n')
			if err != nil {
				panic(err)
			}

			event, err := events.Parse(line)
			if err != nil {
				panic(err)
			}

			ToMe <- event
		}
	}()

	go func(){
		writer := bufio.NewWriter(conn)
		for evt := range ToServer {
			data := evt.Serialize()
			_, err := writer.WriteString(data)
			if err != nil {
				panic(err)
			}

			err = writer.Flush()
			if err != nil {
				panic(err)
			}
		}
	}()

	ToServer <- &events.Event{
		events.WelcomeEvent,
		nick + "," + peernick,
		[]events.Pos{
			events.Pos{ width, height, },
		},
	}

	evt := <-ToMe
	if evt.Type == events.ErrorEvent {
		return utils.Error{ Message: evt.Message }
	}
	if evt.Type == events.StartGameEvent {
		MyId, PeerId = evt.Data[0].X, evt.Data[0].Y
		mySnake = CreateSnake(MyId, evt.Data[1].X, evt.Data[1].Y)
		peerSnake = CreateSnake(PeerId, evt.Data[2].X, evt.Data[2].Y)
		GameLoop()
		Close()
		return nil
	}

	for evt := range ToMe {
		fmt.Printf("evt: %#v\n", evt)
	}

	return nil
}

func UpdateSnake(snake *Snake, headx, heady int, dx, dy int) {
	if snake.Head.Pos.X != headx || snake.Head.Pos.Y != heady {
		panic("snake heads got out of sync")
	}

	if snake.Direction.X != 0 || snake.Direction.Y != 0 {
		if dx != -snake.Direction.X && dy != -snake.Direction.Y {
			snake.Direction.X = dx
			snake.Direction.Y = dy
		}
	} else {
	   snake.Direction.X = dx
	   snake.Direction.Y = dy
	}
}

func MoveSnake(snake *Snake) {
	if snake.Direction.X == 0 && snake.Direction.Y == 0 {
		return
	}

	sec := snake.Tail
	ClearCell(sec.Pos.X, sec.Pos.Y)

	x := snake.Head.Pos.X + snake.Direction.X
	y := snake.Head.Pos.Y + snake.Direction.Y

	if x >= width {
		x = 0
	} else if x < 0 {
		x = width - 1
	}

	if y >= height {
		y = 0
	} else if y < 0 {
		y = height - 1
	}

	sec.Pos = Vec{ x, y }
	DrawCell(x, y, snake.Color)

	snake.Tail = sec.Prev
	snake.Tail.Next = nil
	sec.Next = snake.Head
	sec.Next.Prev = sec
	sec.Prev = nil
	snake.Head = sec
}

func GameLoop() {
	dirs := make(chan Vec)
	go HandleEvents(dirs)
	go func(){
		for dir := range dirs {
			timestamp := utils.Timestamp()
			evt := &events.Event{ events.MoveEvent, "", []events.Pos{
				events.Pos{ MyId, timestamp },
				events.Pos{ mySnake.Head.Pos.X, mySnake.Head.Pos.Y },
				events.Pos{ dir.X, dir.Y },
			} }

			ToServer <- evt
		}

		Close()
	}()

	ticks := time.Tick(100 * time.Millisecond)
	for _ = range ticks {
		select {
		case evt := <-ToMe:
			if evt == nil { break }
			if evt.Type == events.MoveEvent {
				if evt.Data[0].X == MyId {
					UpdateSnake(mySnake, evt.Data[1].X, evt.Data[1].Y, evt.Data[2].X, evt.Data[2].Y)
				} else {
					UpdateSnake(peerSnake, evt.Data[1].X, evt.Data[1].Y, evt.Data[2].X, evt.Data[2].Y)
				}
			} else {
				fmt.Printf("evt: %#v\n", evt)
			}
		default:
		}

		MoveSnake(mySnake)
		MoveSnake(peerSnake)

		if err := termbox.Flush(); err != nil { panic(err) }
	}
	Close()
}

func Close() {
	termbox.Close()
	os.Exit(0)
}
