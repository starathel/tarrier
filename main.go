package main

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

    _ "github.com/mattn/go-sqlite3"
)

type DayType int

const (
    EMPTY_DAY DayType = iota
    MISSED_DAY
    COMPLETED_DAY
)

// TODO: Maybe better chars?
const (
	FILLED_BLOCK = "██"
    SHADED_BLOCK = "░░"
	EMPTY_BLOCK  = "  "
)

func main() {
    days := make([]DayType, 366)
    db := getDbConnection()
    defer db.Close()

    fmt.Println(getMarkedDays(db, time.Now().Year()))

    for i := range days {
        days[i] = MISSED_DAY
    }

    for _, i := range getMarkedDays(db, time.Now().Year()) {
        days[i-1] = COMPLETED_DAY
    }

    for i := time.Now().YearDay() - 1; i < len(days); i++ {
        days[i] = MISSED_DAY
    }

    adjustedTable := make([]DayType, int(firstWeekdayOfYear(2025)), 400)
    adjustedTable = append(adjustedTable, days...)
    printTable(adjustedTable)
}

func printTable(days []DayType) {
    rows := make([]string, 7)
    for i := range 7 {
        daysInRow := make([]string, 0, 60)
        daysInRow = append(daysInRow, dayToDayOfWeek(i) + " ")
        daysInRow = append(daysInRow, " ")
        for j := i; j < len(days); j += 7 {
            daysInRow = append(daysInRow, dayTypeToBlock(days[j]))
        }
        rows[i] = strings.Join(daysInRow, "")
    }
    for _, row := range rows {
        fmt.Println(row)
    }
}

func dayTypeToBlock(dayType DayType) string {
    switch dayType {
    case EMPTY_DAY:
        return EMPTY_BLOCK
    case MISSED_DAY:
        return SHADED_BLOCK
    case COMPLETED_DAY:
        return FILLED_BLOCK
    default:
        panic("Should never ever happen")
    }
}

func dayToDayOfWeek(day int) string {
    switch day {
    case 0:
        return "M"
    case 1:
        return "T"
    case 2:
        return "W"
    case 3:
        return "T"
    case 4:
        return "F"
    case 5:
        return "S"
    case 6:
        return "S"
    default:
        panic("Should never see this")
    }
}

func firstWeekdayOfYear(year int) time.Weekday {
    return time.Date(year, time.January, 0, 0, 0, 0, 0, time.UTC).Weekday()    
}

func getDbConnection() *sql.DB {
    var shouldFillDb bool
    if _, err := os.Stat("./aboba.db"); errors.Is(err, os.ErrNotExist) {
        shouldFillDb = true
    } else if err != nil {
        log.Fatal(err)
    }

    db, err := sql.Open("sqlite3", "./aboba.db")
    if err != nil {
        log.Fatal(err)
    }

    if shouldFillDb {
        fillDb(db)
    }

    return db
}

func fillDb(db *sql.DB) {
    var err error
    _, err = db.Exec(`
        CREATE TABLE IF NOT EXISTS hobbies (
            id INTEGER PRIMARY KEY AUTOINCREMENT,
            name VARCHAR(64) NOT NULL UNIQUE
        );
`)
    if err != nil {
        log.Fatal(err)
    }

    _, err = db.Exec(`
        CREATE TABLE IF NOT EXISTS tracks (
            hobby_id INTEGER NOT NULL REFERENCES hobbies (id) ON DELETE CASCADE,
            mark DATE NOT NULL,

            CONSTRAINT unique_hobby_check PRIMARY KEY (hobby_id, mark)
                ON CONFLICT IGNORE
        );
`)
    if err != nil {
        log.Fatal(err)
    }
}

func getMarkedDays(db *sql.DB, year int) []int {
    markedDays := make([]int, 0, 366)

    rows, err := db.Query(`
        SELECT strftime('%j', mark) from tracks
        WHERE strftime('%Y', mark) = $1
`, strconv.Itoa(year))
    if err != nil {
        log.Fatal(err)
    }
    for rows.Next() {
        var day int
        if err := rows.Scan(&day); err != nil {
            log.Fatal(err)
        }
        markedDays = append(markedDays, day)
    }
    return markedDays
}
