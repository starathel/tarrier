package main

import (
	"database/sql"
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

const usage = `
    Usage: tarrier [-m] [-year <year>] <hobby>

    -m              Mark today as completed

    --year <year>   Select year to get progress of previous years. Ignored if used with -m

`

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

var dayToBlock = map[DayType]string{
    EMPTY_DAY: EMPTY_BLOCK,
    MISSED_DAY: SHADED_BLOCK,
    COMPLETED_DAY: FILLED_BLOCK,
}
var dayToDayOfWeek = map[int]string{
    0: "M",
    1: "T",
    2: "W",
    3: "T",
    4: "F",
    5: "S",
    6: "S",
}

var (
	current_year  int
	selected_year int
	mark_today    bool
	print_help    bool
)

func init() {
	current_year = time.Now().Year()
	flag.IntVar(&selected_year, "year", current_year, "Select year to get progress of previous years. Ignored if used with -m")
	flag.BoolVar(&mark_today, "m", false, "Mark today as completed")
	flag.BoolVar(&print_help, "help", false, "Get help")
	flag.BoolVar(&print_help, "h", false, "Get help")
}

func main() {
	flag.Parse()
	if print_help {
		fmt.Print(usage)
		os.Exit(0)
	}
	args := flag.Args()
	if len(args) != 1 {
		fmt.Fprint(os.Stderr, usage)
		os.Exit(1)
	}
	selected_hobby := args[0]
    if selected_year != current_year {
        mark_today = false
    }

	var days []DayType
	if isLeapYear(selected_year) {
		days = make([]DayType, 366)
	} else {
		days = make([]DayType, 365)
	}

	if current_year == selected_year {
		for i := range time.Now().YearDay() {
			days[i] = MISSED_DAY
		}
	} else {
		for i := range days {
			days[i] = MISSED_DAY
		}
	}

	db := getDbConnection()
	defer db.Close()

	if mark_today {
		err := markToday(db, selected_hobby)
		if err != nil {
			log.Fatal(err)
		}
	}

	for _, i := range getMarkedDays(db, selected_year, selected_hobby) {
		days[i-1] = COMPLETED_DAY
	}

	adjustedTable := make([]DayType, int(firstWeekdayOfYear(selected_year)), 400)
	adjustedTable = append(adjustedTable, days...)
	printTable(adjustedTable)
}

func printTable(days []DayType) {
	rows := make([]string, 7)
	for i := range 7 {
		daysInRow := make([]string, 0, 60)
		daysInRow = append(daysInRow, dayToDayOfWeek[i]+" ")
		daysInRow = append(daysInRow, " ")
		for j := i; j < len(days); j += 7 {
			daysInRow = append(daysInRow, dayToBlock[days[j]])
		}
		rows[i] = strings.Join(daysInRow, "")
	}
	for _, row := range rows {
		fmt.Println(row)
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
		err = fillDb(db)
		if err != nil {
			db.Close()
			log.Fatal(err)
		}
	}

	return db
}

func fillDb(db *sql.DB) error {
	var err error
	_, err = db.Exec(`
        CREATE TABLE IF NOT EXISTS hobbies (
            id INTEGER PRIMARY KEY AUTOINCREMENT,
            name VARCHAR(64) NOT NULL UNIQUE
        );
`)
	if err != nil {
		return err
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
		return err
	}
	return nil
}

func getMarkedDays(db *sql.DB, year int, hobby string) []int {
	markedDays := make([]int, 0, 366)

	rows, err := db.Query(`
        SELECT strftime('%j', mark) from tracks
        JOIN hobbies on hobbies.id = tracks.hobby_id
        WHERE strftime('%Y', mark) = $1
        AND LOWER(hobbies.name) = LOWER($2)
`, strconv.Itoa(year), hobby)
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

func markToday(db *sql.DB, hobby string) error {
	_, err := db.Exec(`
        INSERT INTO hobbies (name) VALUES (LOWER($1))
        ON CONFLICT DO NOTHING
    `, hobby)
	if err != nil {
		return err
	}
	_, err = db.Exec(`
        INSERT INTO tracks (hobby_id, mark) VALUES
        (
            (SELECT id FROM hobbies WHERE LOWER(name) = LOWER($1)),
            date()
        )
    `, hobby)
	if err != nil {
		return err
	}
	return nil
}

func isLeapYear(year int) bool {
	if year%4 == 0 {
		if year%100 == 0 {
			if year%400 == 0 {
				return true
			}
			return false
		}
		return true
	}
	return false
}
