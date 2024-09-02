package server

import (
	"fmt"
	"io"
	"net"
	// "sync"
	"time"

	// json "encoding/json"
	// json "encoding/gob"
	json "github.com/vmihailenco/msgpack/v5"

	"proxy.io/types"
)

// var mu sync.Mutex

func Listen(serverAddr *string) (chan types.Message, chan types.Message, net.Listener, error) {
	send := make(chan types.Message, 100)
	receive := make(chan types.Message, 100)

	listener, err := net.Listen("tcp", *serverAddr)
	if err != nil {
		return nil, nil, nil, err
	}
	fmt.Println("Listening on: ", *serverAddr)

	type Context struct {
		conn    net.Conn
		send    chan types.Message
		receive chan types.Message
	}

	mctx := Context{
		send:    send,
		receive: receive,
	}

	go func(ctx *Context) {
		for {
			if ctx.conn != nil {
				time.Sleep(1 * time.Second)
				continue
			}

			fmt.Println("accept connection is nil")
			cn, err := listener.Accept()
			if err != nil {
				fmt.Println(err)
				continue
			}

			ctx.conn = cn
		}
	}(&mctx)

	go func(ctx *Context) {
		for {
			msg := <-send

			// data, err := msg.Bytes()
			// if err != nil {
			// 	fmt.Println(err)
			// 	continue
			// }

			if ctx.conn == nil {
				fmt.Println("write connection is nil")
				time.Sleep(1 * time.Second)
				continue
			}

			e := json.NewEncoder(ctx.conn)
			if err = e.Encode(&msg); err != nil {
				fmt.Println("error while encoding message", err)
				continue
			}

			fmt.Println("-->[c]", msg.Id, msg.Type, string(msg.Msg))
		}
	}(&mctx)

	go func(ctx *Context) {
		for {
			if ctx.conn == nil {
				fmt.Println("read connection is nil")
				time.Sleep(1 * time.Second)
				continue
			}

			d := json.NewDecoder(ctx.conn)
			var msg types.Message
			err := d.Decode(&msg)
			if err != nil {
				if err == io.EOF {
					fmt.Println("eof found while reading main client")
					ctx.conn.Close()
					ctx.conn = nil
					continue
				}
				fmt.Println("error while decoding message", err)
				continue
			}

			// mu.Lock()
			fmt.Println("<--[c]", msg.Id, msg.Type, string(msg.Msg))
			receive <- msg
		}
	}(&mctx)

	return send, receive, listener, nil
}
