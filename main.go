package main

import (
	"./server"
	"./client"
	"flag"
	"fmt"
)

func main() {
	addr := flag.String("addr", "", "addr to connect to or to bind server to")
	nick := flag.String("nick", "", "your nickname")
	peer := flag.String("peer", "", "nickname of the person you want to play with")
	isserver := flag.Bool("server", false, "be server")

	flag.Parse()

	if *isserver {
		if err := server.Start(*addr); err != nil {
			panic(err)
		}
	} else if *nick != "" && *peer != "" {
		if err := client.Start(*addr, *nick, *peer); err != nil {
			panic(err)
		}
	} else {
		fmt.Printf("usage: gosnake --addr <server-address> (--server | --nick <your-nick> --peer <your-friend>)\n")
	}
}
