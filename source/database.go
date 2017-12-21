package brick_break

import (
    "fmt"
    "database/sql"
    _ "github.com/mattn/go-sqlite3"
)

var loginHashs map[string]int

func InitDB(filePath string) *sql.DB {
    db, err := sql.Open("sqlite3", filePath)
    if err != nil { panic(err) }
    if db == nil { panic("db nil") }
    return db
}

func CreateTable(db *sql.DB) {
    sqlTable := `CREATE TABLE IF NOT EXISTS Players (
        Id integer NOT NULL PRIMARY KEY,
        Username text,
        Password text,
        Level integer,
        XP integer
    );`

    _, err := db.Exec(sqlTable)
    if err != nil { panic(err) }
}

func RegisterPlayer(player Player, db *sql.DB) (int, error) {
    sqlInsert := `INSERT INTO Players (Username, Password, Level, XP) values (?, ?, ?, ?)`
    stmt, err := db.Prepare(sqlInsert)
    if err != nil {
         return 0, err
    }
    defer stmt.Close()

    result, err2 := stmt.Exec(player.Username, player.Password, player.Level, player.XP)
    if err2 != nil {
         return 0, err2
    }

    playerId, err3 := result.LastInsertId()
    if err3 != nil {
        return 0, err3
    }

    return int(playerId), nil
}

func GetPlayerById(id int, db *sql.DB) (Player, error) {
    sqlGetPlayer := `SELECT * FROM Players WHERE Id=?`
    stmt, err := db.Prepare(sqlGetPlayer)
    if err != nil {
         return Player{}, err
    }
    defer stmt.Close()

    rows, err3 := stmt.Query(id)
    if err3 != nil {
        return Player{}, err3
    }
    defer rows.Close()

    player := Player{}
    if rows.Next() {
        err2 := rows.Scan(&player.Id, &player.Username, &player.Password, &player.Level, &player.XP)
        if err2 != nil {
             return Player{}, err2
        }
    }
    return player, nil
}

func GetPlayerSignByPlayer(player Player) PlayerSign {
    return PlayerSign{player.Username, player.Level}
}

func SetPlayerPassword(player Player, db *sql.DB) error {
    sqlUpdate := `UPDATE Players
    SET Password=?
    WHERE Username = ?
    `
    stmt, err := db.Prepare(sqlUpdate)
    defer stmt.Close()
    if err != nil {
         return err
    }
    defer stmt.Close()

    _, err2 := stmt.Exec(player.Password, player.Username)
    if err2 != nil {
         return err2
    }
    return nil
}

func UsernameExists(username string, db *sql.DB) (bool, error) {
    sqlSelect := `SELECT Id FROM Players WHERE Username=?`
    stmt, err := db.Prepare(sqlSelect)
    if err != nil {
         return false, err
    }
    defer stmt.Close()

    rows, err2 := stmt.Query(username)
    if err2 != nil {
         return false, err2
    }
    defer rows.Close()

    exists := false
    for rows.Next() {
        exists = true
    }
    return exists, nil
}

func CheckPlayerInfo(username, pass string, db *sql.DB) (Player, error) {
    sqlSelect := `SELECT Id FROM Players WHERE Username=? AND Password=?`
    stmt, err := db.Prepare(sqlSelect)
    if err != nil {
         return Player{}, err
    }
    defer stmt.Close()

    rows, err2 := stmt.Query(username, pass)
    if err2 != nil {
         return Player{}, err2
    }
    defer rows.Close()

    p := Player{}
    for rows.Next() {
        id := 0
        err3 := rows.Scan(&id)
        if err3 != nil {
            return Player{} ,err3
        }
        pTemp, err4 := GetPlayerById(id, db)
        p = pTemp
        if err4 != nil {
            return Player{} ,err4
        }
    }
    return p, nil
}

func PrintTable(db *sql.DB) {
    sqlSelect := `SELECT Id, Username, Password FROM Players`
    rows, err := db.Query(sqlSelect)
    if err != nil {
         return
    }
    defer rows.Close()

    for rows.Next() {
        player := Player{}
        err2 := rows.Scan(&player.Id, &player.Username, &player.Password)
        if err2 != nil {
             return
        }
        fmt.Println(player.Id, player.Username, player.Password)
    }
}
