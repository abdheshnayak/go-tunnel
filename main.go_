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

	proxyAddr = flag.String("proxyAddr", ":4455", "Proxy address to tunnel")

	// Shared secret for handshake
	secret = flag.String("secret", "mySecretKey", "Secret key for client-server handshake")
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

		fmt.Println("checking creds")
		// Perform the handshake with the server
		if err := performHandshake(conn); err != nil {
			fmt.Printf("Handshake failed: %v\n", err)
			conn.Close()
			time.Sleep(1 * time.Second)
			continue
		}

		handleClientConnection(conn)
		fmt.Println("creds matched")
	}
}

func performHandshake(conn net.Conn) error {
	// Send the secret to the server
	if _, err := conn.Write([]byte(*secret + "\n")); err != nil {
		return fmt.Errorf("failed to send secret: %v", err)
	}

	// Wait for the server to confirm the handshake
	response := make([]byte, len("OK\n"))
	if _, err := io.ReadFull(conn, response); err != nil {
		return fmt.Errorf("failed to read handshake response: %v", err)
	}

	if string(response) != "OK\n" {
		return fmt.Errorf("invalid handshake response: %s", string(response))
	}

	fmt.Println("handshake success")
	return nil
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

		// Perform the handshake with the client
		if err := validateHandshake(conn); err != nil {
			log.Printf("Handshake failed: %v", err)
			conn.Close()
			continue
		}

		clientConnMutex.Lock()
		clientConn = conn
		clientConnMutex.Unlock()

		log.Println("Client connected")
	}
}

func validateHandshake(conn net.Conn) error {
	// Read the secret from the client
	receivedSecret := make([]byte, len(*secret)+1)
	if _, err := io.ReadFull(conn, receivedSecret); err != nil {
		return fmt.Errorf("failed to read secret: %v", err)
	}

	// Validate the secret
	if string(receivedSecret) != *secret+"\n" {
		return fmt.Errorf("invalid secret: %s", string(receivedSecret))
	}

	// Send confirmation back to the client
	if _, err := conn.Write([]byte("OK\n")); err != nil {
		return fmt.Errorf("failed to send handshake confirmation: %v", err)
	}

	return nil
}
