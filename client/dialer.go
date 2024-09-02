package client

import (
	// json "encoding/json"
	json "encoding/gob"
	"fmt"
	// json "github.com/vmihailenco/msgpack/v5"
	"io"
	"net"
	"sync"
	"time"

	"proxy.io/types"
)

var mu sync.Mutex

func Dial(serverAddr *string) (chan types.Message, chan types.Message, net.Conn, error) {
	send := make(chan types.Message, 100)
	receive := make(chan types.Message, 100)

	getConn := func() net.Conn {
		for {
			conn, err := net.Dial("tcp", *serverAddr)
			if err != nil {
				fmt.Println(err)
				time.Sleep(1 * time.Second)
				continue
			}

			fmt.Println("Dialing: ", *serverAddr)
			return conn
		}
	}

	conn := getConn()

	go func() {
		for {
			d := json.NewDecoder(conn)

			var msg types.Message
			err := d.Decode(&msg)
			if err != nil {
				if err == io.EOF {
					fmt.Println("connection closed")
					conn.Close()
					conn = nil

					conn = getConn()
					continue
				}
				fmt.Println("error while decoding message", err)
				continue
			}

			// buf := make([]byte, consts.PayloadSize)
			// conn.SetReadDeadline(time.Now().Add(time.Second * 2))
			// n, err := conn.Read(buf)
			// if err != nil {
			// 	if err == io.EOF {
			// 		fmt.Println("connection closed")
			// 		conn.Close()
			// 		conn = nil
			//
			// 		conn = getConn()
			// 		continue
			// 	}
			//
			// 	continue
			// }
			//
			// fmt.Println("received message from: ", len(buf), n)
			//
			// var msg types.Message
			// err = msg.FromBytes(buf[:n])
			// if err != nil {
			// 	fmt.Println(err)
			// 	continue
			// }

			fmt.Println("received message from: ", msg.Id)
			receive <- msg
		}
	}()

	// send message

	go func() {
		for {
			msg := <-send
			if conn == nil {
				fmt.Println("connection is nil")
				time.Sleep(1 * time.Second)
				continue
			}

			e := json.NewEncoder(conn)
			if err := e.Encode(&msg); err != nil {
				fmt.Println("error while encoding message", err)
				continue
			}

			// data, err := msg.Bytes()
			// if err != nil {
			// 	fmt.Println(err)
			// 	continue
			// }
			//
			// // make buffer of size payload size
			// buf := make([]byte, consts.PayloadSize)
			//
			// // copy data to buffer
			// copy(buf, data)
			//
			// mu.Lock()
			// _, err = conn.Write(buf)
			// mu.Unlock()
			//
			// if err != nil {
			// 	fmt.Println(err)
			// 	continue
			// }
		}
	}()

	return send, receive, conn, nil
}
