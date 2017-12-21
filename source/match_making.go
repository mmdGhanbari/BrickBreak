package brick_break

import (
    "time"
    "encoding/json"
)

const matchCapacity = 2
var matchPool = make([]*Match, 0, 10)

func MatchRequest(newReq *Connection, capacity int) {
    match := FindProperMatch(capacity)

    if match == nil {
        playersList := []*Connection{newReq}
        newMatch := Match{capacity, playersList, map[int]bool{newReq.player.Id : false}, "lobby", time.Now(), Ball{}, 0}
        matchPool = append(matchPool, &newMatch)
        newReq.match = &newMatch
        InformMatchPlayers(newMatch.Players, newReq)
        ResetReadyStates(&newMatch)
    } else {
        match.Players = append(match.Players, newReq)
        newReq.match = match
        InformMatchPlayers(match.Players, newReq)
        ResetReadyStates(match)
    }
}

func FindProperMatch(capacity int) *Match {
    for _, match := range matchPool {
        if match.State == "lobby" && match.Capacity == capacity && len(match.Players) < match.Capacity {
            return match
        }
    }
    return nil
}

func InformMatchPlayers(players []*Connection, newMember *Connection) {
    data := make([]PlayerSign, 0, len(players))
    for _, p := range players {
        data = append(data, GetPlayerSignByPlayer(p.player))
    }

    for _, conn := range players {
        msg := "matchFounded"
        if conn != newMember {
            msg = "updateMatchPlayers"
        }

        outMessage := struct {
            Message string
            Players []PlayerSign
        } {msg, data}

        outJson, _ := json.Marshal(outMessage)
        writeSocket(string(outJson), conn.conn)
    }
}

func SendReadyState(players []*Connection, username string) {
    for _, conn := range players {
        outMessage := struct {
            Message string
            Username string
        } {"updateReadyState", username}

        outJson, _ := json.Marshal(outMessage)
        writeSocket(string(outJson), conn.conn)
    }
}

func ResetReadyStates(match *Match) {
    for key, _ := range match.ReadyState {
        delete(match.ReadyState, key)
    }
}

func ReadyMatch(conn *Connection) {
    conn.match.ReadyState[conn.player.Id] = true
    SendReadyState(conn.match.Players, conn.player.Username)

    allready := true
    for _, p := range conn.match.Players {
        if !conn.match.ReadyState[p.player.Id] {
            allready = false
            break
        }
    }

    if allready {
        conn.match.LivePlayers = conn.match.Capacity
        SendInitialMatchData(conn.match.Players)
        go SendStartMatchAfterSeconds(conn.match.Players, 3)
    }
}

func SendInitialMatchData(players []*Connection) {
    for i, _ := range players {
        players[i].inGameData = PlayerInGameData{i * (360 / players[0].match.Capacity), i, false}
    }
    sendData := make([]PlayerInGameData, len(players))
    for i, _ := range players {
        sendData[i] = players[i].inGameData
    }

    players[0].match.Ball = Ball{0, 0, 90, 2, players[0].inGameData.Color}

    outMessage := struct {
        Message string
        Data []PlayerInGameData
        Ball Ball
    } {"initialData", sendData, players[0].match.Ball}
    outJson, _ := json.Marshal(outMessage)

    for _, c := range players {
        writeSocket(string(outJson), c.conn)
    }
}

func SendStartMatchAfterSeconds(players []*Connection, t time.Duration) {
    time.Sleep(t * time.Second)

    players[0].match.State = "playing"
    for _, conn := range players {
        sendJsonMessageToClient("startMatch", conn)
    }

    go SendPlatformAngles(players)
}

func LeaveMatch(conn *Connection) {
    for i, p := range conn.match.Players {
        if conn == p {
            conn.match.Players = append(conn.match.Players[:i], conn.match.Players[i + 1:]...)
            if len(conn.match.Players) == 0 {
                RemoveMatchFromPool(conn.match)
            }
            sendJsonMessageToClient("success", conn)
        }
    }
    InformMatchPlayers(conn.match.Players, nil)
    ResetReadyStates(conn.match)
}

func RemoveMatchFromPool(match *Match) {
    for i, m := range matchPool {
        if m == match {
            matchPool = append(matchPool[:i], matchPool[i + 1:]...)
        }
    }
}
