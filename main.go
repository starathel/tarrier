package main

import (
	"database/sql"
	"flag"
	"fmt"
	"log"
	"os"
	"path"
	"strconv"
	"strings"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

const usage = `
    Usage: tarrier [-m] [-year <year>] <habit>

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
	EMPTY_DAY:     EMPTY_BLOCK,
	MISSED_DAY:    SHADED_BLOCK,
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
	currentYear     int
	selectedYear    int
	shouldMarkToday bool
	printHelp       bool
    dbDirPath string
	dbPath          string
)

func init() {
	currentYear = time.Now().Year()
	flag.IntVar(&selectedYear, "year", currentYear, "Select year to get progress of previous years. Ignored if used with -m")
	flag.BoolVar(&shouldMarkToday, "m", false, "Mark today as completed")
	flag.BoolVar(&printHelp, "help", false, "Get help")
	flag.BoolVar(&printHelp, "h", false, "Get help")

	dbDirPath = path.Join(os.Getenv("HOME"), ".local/share/tarrier")
    dbPath  = path.Join(dbDirPath, "db.db")
}

func main() {
	flag.Parse()
	if printHelp {
		fmt.Print(usage)
		os.Exit(0)
	}
	args := flag.Args()
	if len(args) != 1 {
		fmt.Fprint(os.Stderr, usage)
		os.Exit(1)
	}
	selectedHabbit := args[0]
	if selectedYear != currentYear {
		shouldMarkToday = false
	}

	var days []DayType
	if isLeapYear(selectedYear) {
		days = make([]DayType, 366)
	} else {
		days = make([]DayType, 365)
	}

	if currentYear == selectedYear {
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

	if shouldMarkToday {
		err := markToday(db, selectedHabbit)
		if err != nil {
			log.Fatal(err)
		}
	}

	for _, i := range getMarkedDays(db, selectedYear, selectedHabbit) {
		days[i-1] = COMPLETED_DAY
	}

	adjustedTable := make([]DayType, int(firstWeekdayOfYear(selectedYear)), 400)
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
    os.MkdirAll(dbDirPath, 0750)
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		log.Fatal(err)
	}

	err = fillDb(db)
	if err != nil {
		db.Close()
		log.Fatal(err)
	}

	return db
}

func fillDb(db *sql.DB) error {
	var err error
	_, err = db.Exec(`
        CREATE TABLE IF NOT EXISTS habits (
            id INTEGER PRIMARY KEY AUTOINCREMENT,
            name VARCHAR(64) NOT NULL UNIQUE
        );
`)
	if err != nil {
		return err
	}

	_, err = db.Exec(`
        CREATE TABLE IF NOT EXISTS marks (
            habit_id INTEGER NOT NULL REFERENCES habits (id) ON DELETE CASCADE,
            mark DATE NOT NULL,

            CONSTRAINT unique_habit_mark PRIMARY KEY (habit_id, mark)
                ON CONFLICT IGNORE
        );
`)
	if err != nil {
		return err
	}
	return nil
}

func getMarkedDays(db *sql.DB, year int, habit string) []int {
	markedDays := make([]int, 0, 366)

	rows, err := db.Query(`
        SELECT strftime('%j', mark) from marks
        JOIN habits on habit.id = tracks.habit_id
        WHERE strftime('%Y', mark) = $1
        AND LOWER(habits.name) = LOWER($2)
`, strconv.Itoa(year), habit)
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

func markToday(db *sql.DB, habit string) error {
	_, err := db.Exec(`
        INSERT INTO habits (name) VALUES (LOWER($1))
        ON CONFLICT DO NOTHING
    `, habit)
	if err != nil {
		return err
	}
	_, err = db.Exec(`
        INSERT INTO marks (habit_id, mark) VALUES
        (
            (SELECT id FROM habits WHERE LOWER(name) = LOWER($1)),
            date()
        )
    `, habit)
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
