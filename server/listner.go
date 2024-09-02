package server

import (
	"fmt"
	"io"
	"net"
	"sync"
	"time"

	"proxy.io/consts"
	"proxy.io/types"
)

var mu sync.Mutex

func Listen(serverAddr *string) (chan types.Message, chan types.Message, net.Listener, error) {
	send := make(chan types.Message)
	receive := make(chan types.Message)

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
			data, err := msg.Bytes()
			if err != nil {
				fmt.Println(err)
				continue
			}

			if ctx.conn == nil {
				fmt.Println("write connection is nil")
				time.Sleep(1 * time.Second)
				continue
			}

			// make buffer of size payload size
			buf := make([]byte, consts.PayloadSize)

			// copy data to buffer
			copy(buf, data)

			_, err = ctx.conn.Write(buf)
			if err != nil {
				fmt.Println(err)
				return
			}

			fmt.Println("sent message for: ", msg.Id)
		}
	}(&mctx)

	go func(ctx *Context) {
		for {
			if ctx.conn == nil {
				fmt.Println("read connection is nil")
				time.Sleep(1 * time.Second)
				continue
			}

			buf := make([]byte, consts.PayloadSize)
			n, err := ctx.conn.Read(buf)
			if err != nil {
				if err == io.EOF {
					fmt.Println("connection closed")
					ctx.conn = nil
				}

				fmt.Println(err, "read error")
				continue
			}

			var msg types.Message
			err = msg.FromBytes(buf[:n])
			if err != nil {
				fmt.Println(err)
				continue
			}

			mu.Lock()
			fmt.Println("msg1:", msg.Id, msg.Type, string(msg.Msg))

			receive <- msg
		}
	}(&mctx)

	return send, receive, listener, nil
}
