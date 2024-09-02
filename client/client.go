package client

import (
	"fmt"
	"io"
	"net"
	"time"

	"proxy.io/consts"
	"proxy.io/types"
)

func Run(ctx types.Context) error {
	send, receive, conn, err := Dial(&ctx.ServerAddr)
	if err != nil {
		return err
	}
	defer conn.Close()

	reqMsg, respMsg, err := ProxyDialer(ctx, send, receive)
	if err != nil {
		return err
	}

	for {
		select {
		case response := <-respMsg:
			fmt.Println("send:", response.Id, response.Type, string(response.Msg))
			send <- response
		case msg := <-receive:
			// fmt.Println("rc message from: ", string(msg.Id))
			fmt.Println("received:", msg.Id, msg.Type, string(msg.Msg))
			reqMsg <- msg
		}
	}
}

func ProxyDialer(ctx types.Context, send chan types.Message, receive chan types.Message) (chan types.Message, chan types.Message, error) {

	readConn := make(chan types.Message)
	reqMsg := make(chan types.Message)
	respMsg := make(chan types.Message)

	dials := make(map[string]net.Conn)

	getDial := func(addr string) (net.Conn, error) {
		if conn, ok := dials[addr]; ok {
			return conn, nil
		}

		conn, err := net.Dial("tcp", ctx.ProxyAddr)
		if err != nil {
			return nil, err
		}
		fmt.Println("Dialing: ", ctx.ProxyAddr)

		dials[addr] = conn
		return conn, nil
	}

	readingList := make(map[string]net.Conn)

	go func() {
		for {
			select {
			case msg := <-readConn:
				c, err := getDial(msg.Id)
				if err != nil {
					fmt.Println(err)
					continue
				}
				if _, ok := readingList[msg.Id]; ok {
					continue
				}

				readingList[msg.Id] = c
				go func() {
					for {
						buf := make([]byte, consts.ProxyPayloadSize)
						n, err := c.Read(buf)
						if err != nil {
							if err == io.EOF {
								buf = make([]byte, consts.ProxyPayloadSize)
								fmt.Println("connection closed, eof")

								time.After(time.Second)
								respMsg <- types.Message{Id: msg.Id, Msg: []byte("error"), Type: types.MessageTypeClose}
								time.After(time.Second)
								return
							}

							continue
						}

						respMsg <- types.Message{Id: msg.Id, Msg: buf[:n], Type: types.MessageTypeResponse}
					}
				}()
			}
		}
	}()

	go func() {
		for {
			select {
			case req := <-reqMsg:
				// send request to proxy
				conn, err := getDial(req.Id)
				if err != nil {
					fmt.Println(err)
					continue
				}

				readConn <- req

				_, err = conn.Write([]byte(req.Msg))
				if err != nil {
					continue
				}

			}

		}
	}()

	return reqMsg, respMsg, nil
}
