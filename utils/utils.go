package utils

import (
	"unicode"
	"time"
)

type Error struct {
	Message string
}

func (e Error) Error() string {
	return e.Message
}

func Timestamp() int {
	return int(time.Now().Unix())
}

func IsNickValid(nick string) bool {
	if len(nick) < 3 || len(nick) > 20 {
		return false
	}

	for _, c := range nick {
		if !(unicode.IsDigit(c) || unicode.IsLetter(c)) {
			return false
		}
	}
	return true
}
