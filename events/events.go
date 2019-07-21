package events

import (
	"../utils"
	"strings"
	"bytes"
	"strconv"
)

type Pos struct {
	X, Y int
}

type EventType int8
const (
	// client -> server: Event{ "<my-nick>,<peer-nick>", [ Pos{ width, height } ] }
	WelcomeEvent EventType = iota

	// server -> client: Event{ "<reason>" }
	ErrorEvent

	// server -> client: Event{ "", [ Pos{ yourid, peerid }, Pos{ mystartx, ,mystarty }, Pos{ peerstartx, ,peerstarty } ] }
	StartGameEvent

	// client -> server: Event{ "", [ Pos{ yourid, timestamp }, Pos{ bx, by } ] } // client has eaten a berry
	// server -> client: Event{ "", [ Pos{ idofeater, timestamp }, Pos{ oldbx, oldby }, Pos{ newbx, newby } ] } // remove old and create new berry
	BerryEvent

	// client -> server: Event{ "", [ Pos{ yourid, timestamp }, Pos{ headx, heady }, Pos{ dx, dy } ] } // client wants to move
	// server -> client: Event{ "", [ Pos{ id, timestamp }, Pos{ headx, heady }, Pos{ dx, dy } ] }     // server tells snakes to move
	MoveEvent
)

type Event struct {
	Type EventType
	Message string // should not contain '\n' or ';''
	Data []Pos
}

func (e *Event) Serialize() string {
	var buf bytes.Buffer
	buf.WriteString(strconv.FormatInt(int64(e.Type), 32))
	buf.WriteString(";")
	buf.WriteString(e.Message)
	for i := 0; i < len(e.Data); i++ {
		buf.WriteString(";")
		buf.WriteString(strconv.FormatInt(int64(e.Data[i].X), 32))
		buf.WriteString(";")
		buf.WriteString(strconv.FormatInt(int64(e.Data[i].Y), 32))
	}
	buf.WriteString("\n")
	return buf.String()
}

func Parse(raw string) (*Event, error) {
	if len(raw) < 3 || raw[len(raw) - 1] != '\n' { return nil, utils.Error{ "Event Parsing Syntax Error" } }
	raw = raw[0:len(raw) - 1]

	sections := strings.Split(raw, ";")
	if len(sections) < 2 || len(sections) % 2 != 0 { return nil, utils.Error{ "Event Parsing Syntax Error" } }

	eventType, err := strconv.ParseInt(sections[0], 32, 8)
	if err != nil { return nil, err }

	e := new(Event)
	e.Type = EventType(eventType)
	e.Message = sections[1]

	for i := 2; i < len(sections); i += 2 {
		x, err := strconv.ParseInt(sections[i], 32, 0)
		if err != nil { return nil, err }

		y, err := strconv.ParseInt(sections[i + 1], 32, 0)
		if err != nil { return nil, err }

		e.Data = append(e.Data, Pos{ X: int(x), Y: int(y) })
	}

	return e, nil
}
