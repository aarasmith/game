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
    "math"
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
    X    int // x-coordinate
    Y    int // y-coordinate
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
        X: 1,
        Y: 1,
    }
    players[playerID] = player

    // Log client connection
    fmt.Printf("Player %d connected: %s\n", playerID, conn.RemoteAddr())

    // Add the client's connection to the list of connected clients
    connectedClients = append(connectedClients, conn)

    // Assign initial (x, y) positions to all players
    assignInitialPositions()

    for {
        var message struct {
            Command       string `json:"command"`
            TargetPlayerID int    `json:"target_player_id"`
            GuessX        int    `json:"guess_x"`
            GuessY        int    `json:"guess_y"`
        }

        err := conn.ReadJSON(&message)
        if err != nil {
            fmt.Println("Error reading message:", err)
            // Handle client disconnection here if needed
            break
        }

        conn.WriteJSON(struct {
            PlayerX       int    `json:"player_x"`
            PlayerY       int    `json:"player_y"`
        }{
            PlayerX: player.X,
            PlayerY: player.Y,
        })

        switch message.Command {
        case "guess":
            // Check if the game is still active (if needed)
            
            // Extract the target player ID and guessed (x, y) position
            targetPlayerID := message.TargetPlayerID
            guessedX := message.GuessX
            guessedY := message.GuessY

            // Check if the target player exists
            _, targetExists := players[targetPlayerID]

            if targetExists {
                // Calculate the distance between the guessed position and the actual position
                distance := calculateDistance(player.X, player.Y, guessedX, guessedY)

                // Define a threshold for a correct guess
                threshold := 0.0 // You can adjust this value based on your game's rules

                // Check if the guess is correct based on the threshold
                if distance <= threshold {
                    // Provide feedback to the player
                    conn.WriteJSON(struct {
                        Message string `json:"message"`
                    }{
                        Message: fmt.Sprintf("Congratulations! You guessed the correct position of Player %d.", targetPlayerID),
                    })
                } else {
                    // Provide feedback that the guess is incorrect
                    conn.WriteJSON(struct {
                        Message string `json:"message"`
                    }{
                        Message: fmt.Sprintf("Your guess for Player %d is not correct. Try again!", targetPlayerID),
                    })
                }
            } else {
                // Inform the player that the target player does not exist
                conn.WriteJSON(struct {
                    Message string `json:"message"`
                }{
                    Message: fmt.Sprintf("Player %d does not exist.", targetPlayerID),
                })
            }

        case "quit":
            // Handle client disconnection here (if needed)
            break

        // Handle other commands as needed

        default:
            // Handle invalid command
            conn.WriteJSON(struct {
                Message string `json:"message"`
            }{
                Message: "Invalid command.",
            })
        }
    }
}

// Function to calculate the distance between two (x, y) points
func calculateDistance(x1, y1, x2, y2 int) float64 {
    dx := x2 - x1
    dy := y2 - y1
    return math.Sqrt(float64(dx*dx + dy*dy))
}

// Function to assign initial (x, y) positions to all players
func assignInitialPositions() {
    for _, player := range players {
        player.X = rand.Intn(10) // MaxX is the maximum x-coordinate
        player.Y = rand.Intn(10) // MaxY is the maximum y-coordinate
    }
}
