package main

import (
    "database/sql"
    "fmt"
    "log"
    "net/http"

    _ "github.com/go-sql-driver/mysql"
)

var (
    db *sql.DB
)

func main() {
    /// BASE DE DONNÉES
    var err error
    db, err = sql.Open("mysql", "goteam:root@tcp(localhost:3306)/golang")
    if err != nil {
        log.Fatal(err)
    }
    defer db.Close()

    err = db.Ping()
    if err != nil {
        log.Fatal(err)
    }
    port := 8080
    fmt.Printf("Server is running on port %d...\n", port)
    log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", port), nil))
}

