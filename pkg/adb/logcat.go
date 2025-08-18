package adb

import (
    "fmt"
    "regexp"
    "strings"

    "github.com/helviojunior/adbcat/pkg/models"
)

type LogcatOptions struct {
    Packages   []string // The packages to filter for
    MinLevel   string   // The minimum log level to show
    Tags       []string // The tags to filter for
    IgnoreTags []string // The tags to ignore
}

// The struct to represent a logcat line
type AdbLineEntry struct {
    Date  string
    Time  string
    Level string
    Tag   string
    PID   string
    TID   string
    MSG   string
}

var (
    // The regex to parse a logcat line
    reLine = regexp.MustCompile(`(?i)(\d{2}-\d{2})\s(\d{2}:\d{2}:\d{2}\.\d{3})\s+(\d+)\s+(\d+)\s+(\S){1}\s+([^\(:]*):\s+[\s]*([^\n]*)`)
)

// Parses a logcat line into a LogcatEntry struct
func ParseLogcatLine(line string) (entry AdbLineEntry, err error) {
    matches := reLine.FindStringSubmatch(line)
    if len(matches) < 8 {
        return entry, fmt.Errorf("could not parse logcat line")
    }

    entry = AdbLineEntry{
        Date:  strings.TrimSpace(matches[1]),
        Time:  strings.TrimSpace(matches[2]),
        PID:   strings.TrimSpace(matches[3]),
        TID:   strings.TrimSpace(matches[4]),
        Level: strings.TrimSpace(matches[5]),
        Tag:   strings.TrimSpace(matches[6]),
        MSG:   strings.TrimRight(strings.Replace(strings.TrimSpace(matches[7]), "\r", "", -1), "\n"),
    }

    return entry, err
}

func (entry AdbLineEntry) EqualTimePidLevel(e2 *AdbLineEntry) bool {

    if e2 == nil {
        return false
    }

    tm1 := entry.getTimePidLevel()
    tm2 := e2.getTimePidLevel()

    return (tm1 == tm2)
}

func (entry AdbLineEntry) getTimePidLevel() string {
    return fmt.Sprintf("%s|%s|%s|%s|%s", entry.Date, entry.Time, entry.PID, entry.TID, entry.Level)
}

// Checks if the level is higher or the same as the wanted one
func IsLevelInScope(entryLevel string, wantedLevel string) bool {
    return models.LevelMap[entryLevel] >= models.LevelMap[wantedLevel]
}
