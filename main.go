package main

import (
	"flag"
	"log"
	"rcon-sample/rcon"
)

func main() {
	var (
		host = flag.String("h", "host", "host")
		password = flag.String("p", "password", "password")
		command = flag.String("c", "command", "command")
	)

	flag.Parse()

	conn, err := rcon.New(*host, *password)
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	err = conn.Exec(*command)
	if err != nil {
		log.Fatal(err)
	}
}
