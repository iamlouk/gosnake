package server

import (
	"net"
	"sync"
	"bufio"
	"strings"
	"fmt"
	"../events"
	"../utils"
)

type StateType int8
const (
	WaitingState StateType = iota
	InGameState
	ClosedState
)

type Game struct {
	P1, P2 *Client
}

func NewGame(p1 *Client, p2 *Client) *Game {
	game := &Game { p1, p2 }
	width, height := p1.Width, p1.Height

	p1.State = InGameState
	p1.Id = 1
	p2.State = InGameState
	p2.Id = 2
	fmt.Printf("server: new game: `%s` vs. `%s`\n", p1.Nick, p2.Nick)

	p1.ToMe <- &events.Event{ events.StartGameEvent, "Welcome to the Party!", []events.Pos{ events.Pos{ 1, 2 }, events.Pos{ width / 3, height / 2 } } }
	p2.ToMe <- &events.Event{ events.StartGameEvent, "Welcome to the Party!", []events.Pos{ events.Pos{ 2, 1 }, events.Pos{ (width / 3) * 2, height / 2 } } }

	for {
		from := 0
		var evt *events.Event = nil
		select {
		case _evt := <-p1.FromMe:
			from = 1
			evt = _evt
		case _evt := <-p2.FromMe:
			from = 2
			evt = _evt
		}
		if evt == nil {
			game.Close(nil)
			return nil
		}

		fmt.Printf("server: msg by #%d: `%#v`\n", from, evt)
	}

	return game
}

func (this *Game) Close(err error) {
	this.P1.Close(err)
	this.P2.Close(err)
}

type Client struct {
	Nick string
	WaitingFor string
	Game *Game
	Width, Height, Id int
	ToMe, FromMe chan *events.Event
	State StateType
	conn net.Conn
}

func (this *Client) Close(err error) {
	if this.State == ClosedState {
		return
	}
	this.State = ClosedState
	if this.Game != nil {
		this.Game.Close(err)
	}
	this.State = ClosedState
	close(this.ToMe)
	close(this.FromMe)
	this.conn.Close()
	if this.Nick != "" {
		clients.Delete(this.Nick)
	}

	fmt.Printf("server: client `%s` closed\n", this.Nick)
}

var clients sync.Map
var games []*Game

func handleConnection(conn net.Conn) {
	client := new(Client)
	client.ToMe = make(chan *events.Event)
	client.FromMe = make(chan *events.Event)
	client.conn = conn

	go func(){
		reader := bufio.NewReader(conn)
		for client.State != ClosedState {
			line, err := reader.ReadString('\n')
			if err != nil {
				client.Close(err)
				return
			}

			event, err := events.Parse(line)
			if err != nil {
				client.Close(err)
				return
			}

			client.FromMe <- event
		}
	}()

	go func(){
		writer := bufio.NewWriter(conn)
		for event := range client.ToMe {
			if client.State == ClosedState {
				return
			}

			data := event.Serialize()
			_, err := writer.WriteString(data)
			if err != nil {
				client.Close(err)
				return
			}

			err = writer.Flush()
			if err != nil {
				client.Close(err)
				return
			}
		}
	}()

	welcomeEvt := <-client.FromMe
	nicks := strings.Split(welcomeEvt.Message, ",")
	if welcomeEvt.Type != events.WelcomeEvent || len(nicks) != 2 || len(welcomeEvt.Data) != 1 {
		client.ToMe <- &events.Event{ events.ErrorEvent, "expected events.WelcomeEvent", []events.Pos{} }
		return
	}

	client.Nick = nicks[0]
	client.WaitingFor = nicks[1]
	client.Width = welcomeEvt.Data[0].X
	client.Height = welcomeEvt.Data[0].Y

	if !utils.IsNickValid(client.Nick) || !utils.IsNickValid(client.WaitingFor) {
		client.ToMe <- &events.Event{ events.ErrorEvent, "expected valid nicknames, not `" + client.Nick + "` and `" + client.WaitingFor + "`", []events.Pos{} }
		return
	}

	if _, exists := clients.LoadOrStore(client.Nick, client); exists {
		client.ToMe <- &events.Event{ events.ErrorEvent, "nickname `" + client.Nick + "` already in use", []events.Pos{} }
		return
	}

	fmt.Printf("server: client `%s` wants to play with `%s`\n", client.Nick, client.WaitingFor)

	if peer, exists := clients.Load(client.WaitingFor); exists {
		peer := peer.(*Client)
		if peer.WaitingFor != client.Nick {
			client.ToMe <- &events.Event{ events.ErrorEvent, "`" + client.WaitingFor + "` does not want to play with you", []events.Pos{} }
			return
		}
		if peer.Width != client.Width || peer.Height != client.Height {
			client.ToMe <- &events.Event{ events.ErrorEvent, "`" + client.WaitingFor + "` does not have the same terminal size as you", []events.Pos{ events.Pos{ peer.Width, peer.Height } } }
			return
		}

		NewGame(client, peer)
	}
}


func Start(port string) error {
	listener, err := net.Listen("tcp", port)
	if err != nil { return err }

	fmt.Printf("server: running\n")

	for {
		conn, err := listener.Accept()
		if err != nil { panic(err) }

		fmt.Printf("server: new connection\n")
		go handleConnection(conn);
	}

	return nil
}
