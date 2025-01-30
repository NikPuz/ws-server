package main

import (
	"encoding/json"
	"github.com/gorilla/websocket"
	"log"
	"net/http"
)

type Point struct {
	X int `json:"x"`
	Y int `json:"y"`
}

func NewGame(width, height int) *Game {
	return &Game{
		width:   width,
		height:  height,
		players: make(map[string]Point),
		state:   GameState{Players: make([]Point, 0), Towers: DefaultTowers(), Explosions: make([]Point, 0)},
	}
}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func handleWebSocket(w http.ResponseWriter, r *http.Request, game *Game) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("Ошибка при обновлении соединения:", err)
		return
	}
	defer conn.Close()
	game.conns = append(game.conns, conn)

	// Генерация уникального ID для игрока
	playerID := r.RemoteAddr
	game.RespawnPlayer(playerID)
	log.Println("Новый игрок:", playerID)

	// Отправка начального состояния
	game.SendStates()

	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			log.Println("Игрок отключился:", playerID)
			delete(game.players, playerID)
			break
		}

		var command map[string]int
		if err := json.Unmarshal(message, &command); err != nil {
			log.Println("Ошибка при разборе команды:", err)
			continue
		}

		// Обработка команды перемещения
		dx := command["dx"]
		dy := command["dy"]
		game.MovePlayer(playerID, dx, dy)

		// Отправка обновленного состояния всем игрокам
		game.SendStates()
	}
}

func main() {
	game := NewGame(30, 15) // Игровое поле 10x10

	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		handleWebSocket(w, r, game)
	})

	log.Println("Сервер запущен на :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
