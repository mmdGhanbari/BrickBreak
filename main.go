package main

import (
    "fmt"
    BB "brickBreak_server/source"
)

func main() {
    go BB.RunServer()

    var input string
    fmt.Scanln(&input)
}
