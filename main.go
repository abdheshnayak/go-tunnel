package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"sync"
	"time"
)

var (
	mode = flag.String("mode", "server", "Mode to run: server or client")

	serverAddr = flag.String("serverAddr", ":3000", "Server address to listen on or connect to")

	proxyAddr = flag.String("proxyAddr", ":4455", "proxy address to tunnel")
)

var (
	clientConn      net.Conn
	clientConnMutex sync.Mutex
)

func main() {
	flag.Parse()

	switch *mode {
	case "server":
		runServer()
	case "client":
		runClient()
	default:
		log.Fatalf("Invalid mode: %s. Use 'server' or 'client'.", *mode)
	}
}

func runClient() error {
	for {
		// Connect to the server on Kubernetes
		conn, err := net.Dial("tcp", *serverAddr)
		if err != nil {
			fmt.Printf("Error connecting to server: %v\n", err)
			time.Sleep(1 * time.Second)
			continue
		}

		defer conn.Close()
		handleClientConnection(conn)
	}
}

func handleClientConnection(conn net.Conn) {
	defer conn.Close()

	var localConn net.Conn

	// Connect to the local WebSocket server
	for {
		var err error
		localConn, err = net.Dial("tcp", *proxyAddr)
		if err != nil {
			log.Printf("Error connecting to local WebSocket server: %v", err)
			time.Sleep(1 * time.Second)
			continue
		}
		defer localConn.Close()
		break
	}

	// Forward data between the WebSocket server and the server connection
	done := make(chan struct{})

	go func() {
		io.Copy(localConn, conn)
		done <- struct{}{}
	}()

	go func() {
		io.Copy(conn, localConn)
		done <- struct{}{}
	}()

	<-done // Wait for either direction to complete
}

func runServer() {
	// Listen for connections from external clients
	listener, err := net.Listen("tcp", *proxyAddr)
	if err != nil {
		log.Fatalf("Error starting server: %v", err)
	}
	defer listener.Close()

	go listenForClient()

	for {
		conn, err := listener.Accept() // Accept incoming connections
		if err != nil {
			log.Printf("Error accepting connection: %v", err)
			continue
		}

		go handleConnection(conn)
	}
}

func handleConnection(conn net.Conn) {
	defer conn.Close()

	clientConnMutex.Lock()
	localClientConn := clientConn
	clientConnMutex.Unlock()

	if localClientConn == nil {
		log.Println("No client connected")
		return
	}

	// Handle each request independently
	done := make(chan struct{})

	go func() {
		io.Copy(localClientConn, conn)
		done <- struct{}{}
	}()

	go func() {
		io.Copy(conn, localClientConn)
		done <- struct{}{}
	}()

	<-done // Wait for either direction to complete
}

func listenForClient() {
	// Listen for connection from the local client
	listener, err := net.Listen("tcp", *serverAddr)
	if err != nil {
		log.Fatalf("Error starting client listener: %v", err)
	}
	defer listener.Close()

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Printf("Error accepting client connection: %v", err)
			continue
		}

		clientConnMutex.Lock()
		clientConn = conn
		clientConnMutex.Unlock()

		log.Println("Client connected")
	}
}
