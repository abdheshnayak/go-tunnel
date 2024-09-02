package server

import (
	"fmt"
	"io"
	"net"

	"proxy.io/consts"
	"proxy.io/types"
)

func Run(ctx types.Context) error {
	send, receive, listener, err := Listen(&ctx.ServerAddr)
	if err != nil {
		return err
	}
	defer listener.Close()

	return ProxyListener(ctx, send, receive)
}

func ProxyListener(ctx types.Context, send chan types.Message, receive chan types.Message) error {
	l, err := net.Listen("tcp", ctx.ProxyAddr)
	if err != nil {
		return err
	}

	fmt.Println("Listening on: ", ctx.ProxyAddr)

	defer l.Close()

	conns := make(map[string]net.Conn, 100)

	for {
		conn, err := l.Accept()
		if err != nil {
			fmt.Println(err)
			continue
		}

		go func(conn net.Conn) {
			defer conn.Close()

			go func() {
				for {
					buf := make([]byte, consts.ProxyPayloadSize)
					n, err := conn.Read(buf)
					if err != nil {
						fmt.Println(err)
						if err == io.EOF {
							fmt.Println("eof found while reading proxy client")
							return
						}
						return
					}

					fmt.Println("<--[p]", conn.RemoteAddr().String(), string(buf[:n]))
					conns[conn.RemoteAddr().String()] = conn
					send <- types.Message{Id: conn.RemoteAddr().String(), Msg: buf[:n], Type: types.MessageTypeRequest}
				}
			}()

			for {
				msg := <-receive
				fmt.Println("-->[p]", msg.Id, msg.Type, string(msg.Msg))

				switch msg.Type {
				case types.MessageTypeResponse:
					conn, ok := conns[msg.Id]
					if !ok {
						fmt.Println("connection not found")
						return
					}

					_, err := conn.Write([]byte(msg.Msg))
					if err != nil {
						fmt.Println(err, "msg:", string(msg.Msg))
						return
					}
				case types.MessageTypeClose:
					fmt.Println("closing connection: ", msg.Id)
					if c, ok := conns[msg.Id]; ok {
						c.Close()
					}
					delete(conns, msg.Id)

				default:
					return
				}

			}
		}(conn)
	}
}
