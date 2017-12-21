package brick_break

import (
    "time"
)

type Player struct {
    Id int
    Username string
    Password string
    Level int
    XP int
}

type PlayerSign struct {
    Username string
    Level int
}

type Match struct {
    Capacity int
    Players []*Connection
    ReadyState map[int]bool
    State string   // lobby, playing, finished
    CreationTime time.Time
    Ball Ball
    LivePlayers int
}

type PlayerInGameData struct {
    Angle int
    Color int
    Dead bool
}

type Ball struct {
    X float64
    Y float64
    Dir float64
    Spd float64
    Clr int
}
