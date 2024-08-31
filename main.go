package main

import (
	"flag"
	"log"

	"proxy.io/client"
	"proxy.io/server"
	"proxy.io/types"
)

var (
	mode = flag.String("mode", "server", "Mode to run: server or client")

	serverAddr = flag.String("serverAddr", ":3000", "Address to listen on")
	proxyAddr  = flag.String("proxyAddr", ":4455", "Address to listen on")

	secret = flag.String("secret", "secret", "Secret to use for encryption")
)

func main() {
	flag.Parse()

	ctx := types.Context{
		ServerAddr: *serverAddr,
		ProxyAddr:  *proxyAddr,
	}

	switch *mode {
	case "server":
		if err := server.Run(ctx); err != nil {
			log.Fatal(err)
		}
	case "client":
		if err := client.Run(ctx); err != nil {
			log.Fatal(err)
		}
	default:
		log.Fatalf("Invalid mode: %s. Use 'server' or 'client'.", *mode)
	}
}
