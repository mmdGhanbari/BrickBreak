package main

import (
    "fmt"
    bb "brickBreak_server/source"
)

func main() {
    go bb.RunServer()

    var input string
    fmt.Scanln(&input)
}
