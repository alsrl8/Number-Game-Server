package main

import (
	"fmt"
	"github.com/gorilla/websocket"
	"log"
	"net/http"
	"strconv"
)

var webSocketUpgrade = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}
var clients = make(map[*websocket.Conn]bool)
var scores = make(map[*websocket.Conn]int)

func run(ws *websocket.Conn) {
	defer func(ws *websocket.Conn) {
		err := ws.Close()
		if err != nil {
			log.Printf("Failed to close connection: %+v", ws)
		}
		log.Printf("Close the websocket")
		delete(clients, ws)
		delete(scores, ws)
		checkWinner() // Check for a winner when a player disconnects
	}(ws)

	for {
		_, msg, err := ws.ReadMessage()
		if err != nil {
			log.Println("Failed to read message: ", err)
			break
		}

		log.Println("Message from client:", string(msg))

		if string(msg) == "New Player" {
			if len(clients) > 2 {
				log.Printf("Too many clients")
				continue
			}
			for client := range clients {

				client.WriteMessage(websocket.TextMessage, []byte(fmt.Sprintf("Player Num: %d", len(clients))))
				log.Printf("Send message to client(%s): %s", client.LocalAddr(), fmt.Sprintf("Player Num: %d", len(clients)))
			}
		} else {
			// Convert the received message to an integer
			num, err := strconv.Atoi(string(msg))
			if err != nil {
				log.Println("Invalid number: ", err)
				continue
			}

			if scores[ws] > 0 {
				log.Printf("Number for this player(%s) is already assigned", ws.LocalAddr())
				continue
			}

			scores[ws] = num
			log.Printf("Set number(%d) for player(%s)", num, ws.LocalAddr())

			// Check if all players have sent their numbers
			if len(scores) == 2 {
				checkWinner()
			}
		}

	}
}

func checkWinner() {
	if len(scores) == 2 {
		var winner *websocket.Conn
		for client, score := range scores {
			if winner == nil || scores[winner] < score {
				winner = client
			} else if scores[winner] == score {
				winner = nil // It's a draw
			}
		}

		// Send the result to all clients
		for client := range clients {
			if client == winner {
				client.WriteMessage(websocket.TextMessage, []byte("You win!"))
			} else if winner == nil {
				client.WriteMessage(websocket.TextMessage, []byte("It's a draw!"))
			} else {
				client.WriteMessage(websocket.TextMessage, []byte("You lose!"))
			}
		}
	}
}

func main() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		ws, err := webSocketUpgrade.Upgrade(w, r, nil)
		if err != nil {
			log.Println("Failed to upgrade request: ", err)
			return
		}
		clients[ws] = true
		go run(ws)
	})

	log.Printf("Server is listening on :8080...")
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
