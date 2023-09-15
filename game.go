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

// Define a Player struct
type Player struct {
    ID   int
    Conn *websocket.Conn
}

// Create a map to store players by their IDs
var players = make(map[int]*Player)
var currentPlayerID = 1 // Initialize the player ID counter

// Define a list to store connected clients
var connectedClients = make([]*websocket.Conn, 0)

// Create a variable to track the game state
var gameActive = true


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

    // Generate a unique player ID for this client
    playerID := currentPlayerID
    currentPlayerID++

    // Create a new Player object and store it in the map
    player := &Player{
        ID:   playerID,
        Conn: conn,
    }
    players[playerID] = player

    // Log client connection
    fmt.Printf("Player %d connected: %s\n", playerID, conn.RemoteAddr())

    // Add the client's connection to the list of connected clients
    connectedClients = append(connectedClients, conn)

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

            // Log client disconnection
            fmt.Printf("Player %d disconnected: %s\n", playerID, conn.RemoteAddr())

            // Remove the client's connection from the list of connected clients
            for i, client := range connectedClients {
                if client == conn {
                    connectedClients = append(connectedClients[:i], connectedClients[i+1:]...)
                    break
                }
            }

            // Delete the player from the map
            delete(players, playerID)
            break
        }

        switch message.Command {
        case "guess":
            // Check if the game is still active
            if gameActive {
                // Use playerID to identify the player making the guess
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

                    // Notify all other connected clients except the current player
                    notification := struct {
                        Message string `json:"message"`
                    }{
                        Message: fmt.Sprintf("Player %d guessed the correct number!", playerID),
                    }

                    for _, client := range connectedClients {
                        if client != conn {
                            // Send the notification to other clients
                            err := client.WriteJSON(notification)
                            if err != nil {
                                fmt.Println("Error sending notification:", err)
                            }
                        }
                    }

                    // Update the game state to indicate that it's over
                    gameActive = false
                } else if message.Guess < game.TargetNumber {
                    conn.WriteJSON(struct{ Message string `json:"message"` }{Message: "Try a higher number."})
                } else {
                    conn.WriteJSON(struct{ Message string `json:"message"` }{Message: "Try a lower number."})
                }
            } else {
                // If the game is not active, inform the player that the game is over
                conn.WriteJSON(struct{ Message string `json:"message"` }{Message: "The game is over. You can't guess anymore."})
            }

        case "quit":
            // Log client disconnection
            fmt.Printf("Player %d disconnected: %s\n", playerID, conn.RemoteAddr())

            // Remove the client's connection from the list of connected clients
            for i, client := range connectedClients {
                if client == conn {
                    connectedClients = append(connectedClients[:i], connectedClients[i+1:]...)
                    break
                }
            }

            // Delete the player from the map
            delete(players, playerID)
            delete(game.Guesses, conn)
            break

        // ...
        }
    }
}