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

	conns := make(map[string]net.Conn)

	for {
		conn, err := l.Accept()
		if err != nil {
			fmt.Println(err)
			continue
		}

		go func() {
			go func(conn net.Conn) {
				defer conn.Close()

				go func() {
					buf := make([]byte, consts.ProxyPayloadSize)
					n, err := conn.Read(buf)
					if err != nil {
						fmt.Println(err)
						if err == io.EOF {
							return
						}
						return
					}

					conns[conn.RemoteAddr().String()] = conn
					send <- types.Message{Id: conn.RemoteAddr().String(), Msg: buf[:n], Type: types.MessageTypeRequest}
				}()

				for {
					msg := <-receive

					// fmt.Println("msg:", msg.String())
					switch msg.Type {
					case types.MessageTypeResponse:
						conn, ok := conns[msg.Id]
						if !ok {
							fmt.Println("connection not found")
							continue
						}

						_, err := conn.Write([]byte(msg.Msg))
						if err != nil {
							fmt.Println(err)
							continue
						}
					case types.MessageTypeClose:
						fmt.Println("closing connection: ", msg.Id)
						if c, ok := conns[msg.Id]; ok {
							c.Close()
						}
						// delete(conns, msg.Id)

					default:
						continue
					}
					if msg.Type != types.MessageTypeResponse {
						continue
					}

				}
			}(conn)
		}()
	}
}
