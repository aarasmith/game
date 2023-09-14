package main

import (
	"fmt"
	"math/rand"
	"net"
	// "strconv"
	"time"
	// "strings"
	"io"
	"encoding/json"
)

const (
	port = ":9999"
)

type Game struct {
    TargetNumber int `json:"target_number"`
    Guesses      map[net.Conn]int
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
	listener, err := net.Listen("tcp", port)
	if err != nil {
		fmt.Println("Error starting server:", err)
		return
	}
	defer listener.Close()

	fmt.Println("Game server started on port", port)

	game := NewGame()

	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("Error accepting connection:", err)
			continue
		}
		go handleClient(conn, game)
	}
}

func NewGame() *Game {
    return &Game{
        TargetNumber: rand.Intn(100),
        Guesses:      make(map[net.Conn]int),
    }
}

func handleClient(conn net.Conn, game *Game) {
    defer conn.Close()

    fmt.Println("New player connected:", conn.RemoteAddr())

    // Welcome message
    conn.Write([]byte("Welcome to the Guess the Number game!\n"))

    decoder := json.NewDecoder(conn)
    encoder := json.NewEncoder(conn)

    for {
        var clientMessage ClientMessage
        if err := decoder.Decode(&clientMessage); err == io.EOF {
            fmt.Println("Client disconnected:", conn.RemoteAddr())
            return
        } else if err != nil {
            fmt.Println("Error decoding message:", err)
            return
        }

        switch clientMessage.Command {
        case "guess":
            guess := clientMessage.Guess
            game.Guesses[conn] = guess

            if guess == game.TargetNumber {
                response := ServerMessage{Message: "Congratulations! You guessed the correct number."}
                encoder.Encode(response)
                return
            } else if guess < game.TargetNumber {
                response := ServerMessage{Message: "Try a higher number."}
                encoder.Encode(response)
            } else {
                response := ServerMessage{Message: "Try a lower number."}
                encoder.Encode(response)
            }
        case "quit":
            fmt.Println("Client disconnected:", conn.RemoteAddr())
            return
        default:
            response := ServerMessage{Message: "Invalid command."}
            encoder.Encode(response)
        }
    }
}