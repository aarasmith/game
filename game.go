package main

import (
	"fmt"
	"math/rand"
	"net"
	"strconv"
	"time"
	"strings"
)

const (
	port = ":9999"
)

type Game struct {
	targetNumber int
	guesses      map[net.Conn]int
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
		targetNumber: rand.Intn(100),
		guesses:      make(map[net.Conn]int),
	}
}

func handleClient(conn net.Conn, game *Game) {
	defer conn.Close()

	fmt.Println("New player connected:", conn.RemoteAddr())

	// Welcome message
	conn.Write([]byte("Welcome to the Guess the Number game!\n"))

	for {
		conn.Write([]byte("Enter your guess: "))
		buffer := make([]byte, 1024)
		n, err := conn.Read(buffer)
		if err != nil {
			fmt.Println("Client disconnected:", conn.RemoteAddr())
			return
		}

		input := strings.TrimSpace(string(buffer[:n-1]))

        if input == "quit" {
            fmt.Println("Client disconnected:", conn.RemoteAddr())
            return
        }

		guess, err := strconv.Atoi(strings.TrimSpace(string(buffer[:n-1])))
		if err != nil {
			conn.Write([]byte("Invalid input. Please enter a valid number.\n"))
			continue
		}

		// game.guesses[conn] = guess

		if guess == game.targetNumber {
			conn.Write([]byte("Congratulations! You guessed the correct number.\n"))
			return
		} else if guess < game.targetNumber {
			conn.Write([]byte("Try a higher number.\n"))
		} else {
			conn.Write([]byte("Try a lower number.\n"))
		}
	}
}
