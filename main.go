package main

import (
	"flag"
	"fmt"
	"os"
)

func main() {
	serverAddr := flag.String("server", "", "run relay server on address (e.g. :9000)")
	connectAddr := flag.String("connect", "", "connect client to relay server (e.g. localhost:9000)")
	name := flag.String("name", "", "display name for chat client")
	flag.Parse()

	switch {
	case *serverAddr != "":
		if err := runServer(*serverAddr); err != nil {
			fmt.Fprintf(os.Stderr, "server: %v\n", err)
			os.Exit(1)
		}
	case *connectAddr != "":
		n := *name
		if n == "" {
			n = defaultName()
		}
		if err := runClient(*connectAddr, n); err != nil {
			fmt.Fprintf(os.Stderr, "client: %v\n", err)
			os.Exit(1)
		}
	default:
		fmt.Fprintf(os.Stderr, "usage:\n")
		fmt.Fprintf(os.Stderr, "  %s -server :9000\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s -connect localhost:9000 [-name alice]\n", os.Args[0])
		os.Exit(2)
	}
}
