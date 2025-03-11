package main

import (
	"fmt"
	"strings"
    "time"
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

    for i := range days {
        days[i] = MISSED_DAY
    }

    for i := 35; i < 50; i++ {
        days[i] = COMPLETED_DAY
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
