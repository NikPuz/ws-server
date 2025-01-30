package main

import (
	"github.com/gorilla/websocket"
	"sync"
)

type Game struct {
	width   int
	height  int
	players map[string]Point
	state   GameState
	conns   []*websocket.Conn
	mu      sync.Mutex
}

type GameState struct {
	Players    []Point
	Towers     []Point
	Explosions []Point
}

func (g *Game) RespawnPlayer(id string) {
	x, y := 0, 0
	for g.isPointBusy(x, y) {
		x += 3
	}

	g.players[id] = Point{x, y}
}

func (g *Game) MovePlayer(id string, dx, dy int) {
	g.mu.Lock()
	defer g.mu.Unlock()

	player, exists := g.players[id]
	if !exists {
		return
	}

	newX := player.X + dx
	newY := player.Y + dy

	if g.isPointBusy(newX, newY) {
		ok, secondId := g.getPlayerIdByPoint(newX, newY)
		if ok {
			delete(g.players, id)
			delete(g.players, secondId)
			g.RespawnPlayer(id)
			g.RespawnPlayer(secondId)
		}

		g.state.Explosions = append(g.state.Explosions, Point{newX, newY})
		return
	}

	if newX >= 0 && newX < g.width {
		player.X = newX
	}
	if newY >= 0 && newY < g.height {
		player.Y = newY
	}

	g.players[id] = player
}

func (g *Game) SendStates() {
	state := g.GetState()
	for _, conn := range g.conns {
		conn.WriteJSON(state)
	}
}

func (g *Game) GetState() GameState {
	g.mu.Lock()
	defer g.mu.Unlock()

	g.state.Players = make([]Point, 0, len(g.players))
	for _, player := range g.players {
		g.state.Players = append(g.state.Players, player)
	}

	return g.state
}

func (g *Game) isPointBusy(x, y int) bool {

	for _, p := range g.state.Towers {
		if x == p.X && y == p.Y {
			return true
		}
	}

	isBusy, _ := g.getPlayerIdByPoint(x, y)

	return isBusy
}

func (g *Game) getPlayerIdByPoint(x, y int) (bool, string) {

	for id, p := range g.players {
		if p.X == x && p.Y == y {
			return true, id
		}
	}

	return false, ""
}
