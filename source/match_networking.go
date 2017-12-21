package brick_break

import (
    "time"
    // "math/rand"
    "encoding/json"
)

func SendPlatformAngles(players []*Connection) {
    for players[0].match.State == "playing" {
        time.Sleep(80 * time.Millisecond)

        values := make([]int, len(players))
        for i, _ := range players {
            if !players[i].inGameData.Dead {
                values[i] = players[i].inGameData.Angle
            } else {
                values[i] = -1
            }
        }
        outMessage := struct {
            Msg string
            Vls []int
        } {"agl", values}
        outJson, _ := json.Marshal(outMessage)

        for _, c := range players {
            writeSocket(string(outJson), c.conn)
        }
    }
}

func UpdateBallInfo(match *Match, ball Ball) {
    colors := make([]int, match.LivePlayers)
    for _, p := range match.Players {
        if !p.inGameData.Dead {
            colors = append(colors, p.inGameData.Color)
        }
    }

    // s1 := rand.NewSource(time.Now().UnixNano())
    // r1 := rand.New(s1)
    // match.Ball.Clr = r1.Intn(match.LivePlayers)
    // if match.LivePlayers > 1 {
    //     for match.Ball.Clr == ball.Clr {
    //         match.Ball.Clr = r1.Intn(match.LivePlayers)
    //     }
    // }
    match.Ball.Clr = (ball.Clr + 1) % match.LivePlayers
    match.Ball.X = ball.X
    match.Ball.Y = ball.Y
    match.Ball.Dir = ball.Dir
    match.Ball.Spd = ball.Spd

    outMessage := struct {
        Msg string
        B Ball
    } {"bll", match.Ball}
    outJson, _ := json.Marshal(outMessage)

    for _, c := range match.Players {
        writeSocket(string(outJson), c.conn)
    }
}

func HandlePlayerLose(match *Match, index int) {
    if !match.Players[index].inGameData.Dead {
        match.Players[index].inGameData.Dead = true
        match.LivePlayers--
        outMessage := struct {
            Msg string
            Vls []int
        } {"ls", []int{index}}
        outJson, _ := json.Marshal(outMessage)

        for _, c := range match.Players {
            writeSocket(string(outJson), c.conn)
        }
    }
}
