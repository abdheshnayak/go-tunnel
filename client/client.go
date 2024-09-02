package client

import (
	"encoding/base64"
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

	go func() {
		for {
			response := <-respMsg
			fmt.Println("-->[c]", response.Id, response.Type, string(response.Msg))
			send <- response
		}
	}()

	for {
		msg := <-receive
		fmt.Println("<--[c]", msg.Id, msg.Type, string(msg.Msg))
		reqMsg <- msg
	}
}

func ProxyDialer(ctx types.Context, send chan types.Message, receive chan types.Message) (chan types.Message, chan types.Message, error) {

	readConn := make(chan types.Message, 100)
	reqMsg := make(chan types.Message, 100)
	respMsg := make(chan types.Message, 100)

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
								fmt.Println("connection closed, eof")

								time.Sleep(time.Millisecond * 100)
								respMsg <- types.Message{Id: msg.Id, Msg: "error", Type: types.MessageTypeClose}
								time.Sleep(time.Millisecond * 100)
								return
							}

							continue
						}

						encoded := base64.StdEncoding.EncodeToString(buf[:n])
						respMsg <- types.Message{Id: msg.Id, Msg: encoded, Type: types.MessageTypeResponse}
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

				data, err := base64.StdEncoding.DecodeString(req.Msg)
				if err != nil {
					fmt.Println(err)
					continue
				}

				_, err = conn.Write(data)
				if err != nil {
					continue
				}

			}

		}
	}()

	return reqMsg, respMsg, nil
}
