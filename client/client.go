package client

import (
	"fmt"
	"net"
	"bufio"
	"../events"
	"github.com/nsf/termbox-go"
)

var ToMe chan *events.Event
var ToServer chan *events.Event

func Start(addr string, nick string, peernick string) error {
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return err
	}

	if err := termbox.Init(); err != nil {
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

	width, height := termbox.Size()
	ToServer <- &events.Event{
		events.WelcomeEvent,
		nick + "," + peernick,
		[]events.Pos{
			events.Pos{ width, height, },
		},
	}

	for evt := range ToMe {
		fmt.Printf("evt: %#v\n", evt)
	}

	return nil
}
