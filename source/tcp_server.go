package brick_break

import (
    "fmt"
    "bufio"
    "net"
    "database/sql"
    "encoding/json"
)

type Connection struct {
    conn net.Conn
    player Player
    match *Match
    inGameData PlayerInGameData
}

func (c *Connection) Close() {
    if c.match != nil && c.match.State == "lobby" {
        LeaveMatch(c)
    }
    for i, conn := range connections {
        if conn == c {
            connections = append(connections[:i], connections[i + 1:]...)
        }
    }
    c.conn.Close()
    c = nil
}

var connections []*Connection

func RunServer() {
    ln, err := net.Listen("tcp", ":8888")
    fmt.Println("listening ...")

    if err != nil {
        fmt.Println(err)
        return
    }
    var db *sql.DB = InitDB("database.db")
    CreateTable(db)

    connections = make([]*Connection, 0, 10)
    loginHashs = make(map[string]int)

    for {
        c, err := ln.Accept()
        if err != nil {
            fmt.Println(err)
            continue
        }
        fmt.Println("new connection ...")

        newConnection := Connection{c, Player{}, nil, PlayerInGameData{}}
        connections = append(connections, &newConnection)
        go handleServerConnection(&newConnection, db)
    }
}

func handleServerConnection(conn *Connection, db *sql.DB) {
    for {
        message, err := bufio.NewReader(conn.conn).ReadString('\n')

        if err != nil {
            fmt.Println("Error reading:", err.Error())
            break
        } else {
            message = message[:len(message) - 1]
            // fmt.Println("message : " + message)
            onNewMessage(message, conn, db)
        }
    }
    conn.Close()
}

func onNewMessage(message string, conn *Connection, db *sql.DB) {
    jsonMsg := []byte(message)
    var receivedData map[string]interface{}

    if err := json.Unmarshal(jsonMsg, &receivedData); err != nil {
        handleError(err, conn)
        return
    }
    command := receivedData["Message"].(string)
    if command != "UA" {
        fmt.Println("message : " + message)
    }

    if command == "register" {
        username := receivedData["Username"].(string)
        pass := receivedData["Password"].(string)
        exists, err := UsernameExists(username, db)
        if err != nil {
            handleError(err, conn)
            return
        }

        if exists {
            sendJsonMessageToClient("exists", conn)
        } else {
            id, err := RegisterPlayer(Player{0, username, pass, 1, 0}, db)
            if err != nil {
                handleError(err, conn)
                return
            }
            loginHash := GetRandomHash()
            for loginHashs[loginHash] != 0 {
                loginHash = GetRandomHash()
            }
            loginHashs[loginHash] = id

            conn.player, _ = GetPlayerById(id, db)
            outMessage := struct {
                Message string
                LoginHash string
                Id int
            } {"success", loginHash, id}

            outJson, _ := json.Marshal(outMessage)
            writeSocket(string(outJson), conn.conn)
        }

    } else if command == "set_pass" {
        // username := receivedData["Username"].(string)
        // pass := receivedData["Password"].(string)
        //
        // exists, err := UsernameExists(username, db)
        // if err != nil {
        //     handleError(err, conn)
        //     return
        // }
        //
        // if !exists {
        //     sendJsonMessageToClient("username_not_exists", conn)
        // } else {
        //     err := SetPlayerPassword(Player{0, username, pass}, db)
        //     if err != nil {
        //         handleError(err, conn)
        //         return
        //     }
        //     sendJsonMessageToClient("success", conn)
        // }

    } else if command == "login" {
        username := receivedData["Username"].(string)
        pass := receivedData["Password"].(string)

        player, err := CheckPlayerInfo(username, pass, db)
        if err != nil {
            handleError(err, conn)
            return
        }

        if player.Id != 0 {
            conn.player = player

            loginHash := GetRandomHash()
            deleteFromHashsById(player.Id)
            loginHashs[loginHash] = player.Id

            outMessage := struct {
                Message string
                LoginHash string
                Id int
            } {"success", loginHash, player.Id}

            outJson, _ := json.Marshal(outMessage)
            writeSocket(string(outJson), conn.conn)
        } else {
            sendJsonMessageToClient("notfound", conn)
        }

    } else if command == "loginByHash" {
        hash := receivedData["LoginHash"].(string)
        playerId, exists := loginHashs[hash]
        if !exists {
            sendJsonMessageToClient("notfound", conn)
        } else {
            player, err := GetPlayerById(playerId, db)
            if err != nil {
                handleError(err, conn)
                return
            }
            conn.player = player
            outMessage := struct {
                Message string
                Id int
            } {"success", playerId}

            outJson, _ := json.Marshal(outMessage)
            writeSocket(string(outJson), conn.conn)
        }

    } else if command == "logout" {
        deleteFromHashsById(conn.player.Id)

    } else if command == "getPlayerData" {
        outJson, _ := json.Marshal(conn.player)
        writeSocket(string(outJson), conn.conn)

    } else if command == "findMatch" {
        player, err := GetPlayerById(conn.player.Id, db)
        if err != nil {
            handleError(err, conn)
            return
        }
        conn.player = player
        MatchRequest(conn, matchCapacity)

    } else if command == "readyMatch" {
        ReadyMatch(conn)

    } else if command == "leaveMatch" {
        LeaveMatch(conn)

    } else if command == "UA" {
        angle := int(receivedData["A"].(float64))
        conn.inGameData.Angle = angle

    } else if command == "UB" {
        x := receivedData["X"].(float64)
        y := receivedData["Y"].(float64)
        direction := receivedData["Dir"].(float64)
        speed := receivedData["Spd"].(float64)
        color := int(receivedData["Clr"].(float64))
        UpdateBallInfo(conn.match, Ball{x, y, direction, speed, color})

    } else if command == "LS" {
        index := int(receivedData["ID"].(float64))
        HandlePlayerLose(conn.match, index)
    }
}

func deleteFromHashsById(id int) {
    for hash, value := range loginHashs {
        if value == id {
            delete(loginHashs, hash)
        }
    }
}

func sendJsonMessageToClient(msg string, c *Connection) {
    outMessage := struct {
        Message string
    } {msg}

    outJson, _ := json.Marshal(outMessage)
    writeSocket(string(outJson), c.conn)
}

func writeSocket(message string, conn net.Conn) {
    var serverMsg string = message + "\n"
    conn.Write([]byte(serverMsg))
}

func handleError(err error, conn *Connection) {
    fmt.Println(err)
    sendJsonMessageToClient("error", conn)
}
