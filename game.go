package main

import (
	"fmt"
	"math/rand"
	// "net"
	// "strconv"
	"time"
	// "strings"
	// "io"
	"net/http"
	// "encoding/json"
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
)

const (
	port = ":9999"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

type Game struct {
    TargetNumber int `json:"target_number"`
    Guesses      map[*websocket.Conn]int
}

type Client struct {
    Conn  *websocket.Conn
    Guess int
}

type ClientMessage struct {
    Command string `json:"command"`
    Guess   int    `json:"guess"`
}

type ServerMessage struct {
    Message string `json:"message"`
}

func main() {
    rand.Seed(time.Now().UnixNano())

    game := NewGame()

    r := mux.NewRouter()
    r.HandleFunc("/", ServeHome)
    r.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
        ServeWebSocket(w, r, game)
    })

    http.Handle("/", r)

    fmt.Println("Game server started on port", port)
    err := http.ListenAndServe(port, nil)
    if err != nil {
        fmt.Println("Error starting server:", err)
    }
}

func NewGame() *Game {
    return &Game{
        TargetNumber: rand.Intn(100),
        Guesses:      make(map[*websocket.Conn]int),
    }
}

func ServeHome(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "index.html")
}

func ServeWebSocket(w http.ResponseWriter, r *http.Request, game *Game) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		fmt.Println("Error upgrading to WebSocket:", err)
		return
	}
	defer conn.Close()

	game.Guesses[conn] = -1 // Initialize the guess

	for {
		var message struct {
			Command string `json:"command"`
			Guess   int    `json:"guess"`
		}

		err := conn.ReadJSON(&message)
		if err != nil {
			fmt.Println("Error reading message:", err)
			delete(game.Guesses, conn)
			break
		}

		switch message.Command {
		case "guess":
			game.Guesses[conn] = message.Guess

			if message.Guess == game.TargetNumber {
				response := struct {
					Message string `json:"message"`
					Win     bool   `json:"win"`
				}{
					Message: "Congratulations! You guessed the correct number.",
					Win:     true,
				}
				conn.WriteJSON(response)
			} else if message.Guess < game.TargetNumber {
				conn.WriteJSON(struct{ Message string `json:"message"` }{Message: "Try a higher number."})
			} else {
				conn.WriteJSON(struct{ Message string `json:"message"` }{Message: "Try a lower number."})
			}
		case "quit":
			delete(game.Guesses, conn)
			break
		default:
			conn.WriteJSON(struct{ Message string `json:"message"` }{Message: "Invalid command."})
		}
	}
}