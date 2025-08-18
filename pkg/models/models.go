package models

import (
    "fmt"
    "os"
    "strings"

    "github.com/helviojunior/adbcat/internal/ascii"
    "github.com/fatih/color"
    "github.com/nathan-fiscaletti/consolesize-go"
)

const (
    // The available log levels
    LevelVerbose = iota // 0
    LevelDebug          // 1
    LevelInfo           // 2
    LevelWarning        // 3
    LevelError          // 4
    LevelFatal          // 5

    MaxLenTag = 24 // The maximum length of a tag for the terminal UI
    MaxLenTime = 13 // The maximum length of a time for the terminal UI
    MaxLenPid = 5 // The maximum length of a pid/tid for the terminal UI
)

var (
    // The colors for the log levels
    colorLevel = []*color.Color{
        color.New(color.BgBlack, color.FgWhite),  // Verbose
        color.New(color.BgBlack, color.FgCyan),   // Debug
        color.New(color.BgBlack, color.FgGreen),  // Info
        color.New(color.BgBlack, color.FgYellow), // Warning
        color.New(color.BgBlack, color.FgRed),    // Error
        color.New(color.BgBlack, color.FgRed),    // Fatal
    }

    // A slice of colors for the tags that gets rotated when a new tag is encountered
    colorTags = []*color.Color{
        color.New(color.FgWhite).AddBgRGB(60,60,60),  // Verbose
        color.New(color.FgCyan).AddBgRGB(60,60,60),   // Debug
        color.New(color.FgGreen).AddBgRGB(60,60,60),  // Info
        color.New(color.FgYellow).AddBgRGB(60,60,60), // Warning
        color.New(color.FgRed).AddBgRGB(60,60,60),    // Error
        color.New(color.FgRed).AddBgRGB(60,60,60),    // Fatal
    }

    LevelMap = map[string]int{
        "V": LevelVerbose,
        "D": LevelDebug,
        "I": LevelInfo,
        "W": LevelWarning,
        "E": LevelError,
        "F": LevelFatal,
    }

)

type LogcatEntry struct {
    Date        string       `json:"date"`
    Time        string       `json:"time"`
    Level       string       `json:"level"`
    Tag         string       `json:"tag"`
    PID         string       `json:"pid"`
    TID         string       `json:"tid"`
    Message     string       `json:"message"`
}


func (entry LogcatEntry) ToAnsiString() string {
    return entry.FormatAnsiString(true, true)
}

func (entry LogcatEntry) FormatAnsiString(showTime bool, showPid bool) string {

    time := ""
    if showTime {
        time = formatTime(entry.Time)
    }
    pid := ""
    if showPid {
        pid = entry.GetFormattedPidTid()
    }
    name := formatTag(entry.Tag)

    // Color the level based on the log level
    c1 := colorLevel[LevelMap[entry.Level]]
    coloredLevel := colorTags[LevelMap[entry.Level]].Sprintf(" %s ", entry.Level)
    coloredName := c1.Sprintf("%s", name)

    prefix := "\033[0m\033[1;90m" + time + pid + "\033[0m\033[1;90m\033[0m" + coloredName
    prefixLen := len(ascii.ScapeAnsi(prefix))
    prefixLen2 := len(ascii.ScapeAnsi(prefix+coloredLevel))
    coloredMsg := ""
    for i, line := range strings.Split(entry.Message, "\n") {
        if i == 0 {
            coloredMsg += prefix + coloredLevel + c1.Sprint(formatMsg(line, prefixLen2))
        }else{
            coloredMsg += fmt.Sprintf("\n%*s%s%s", prefixLen, "", coloredLevel, c1.Sprint(formatMsg(line, prefixLen2)))
        }
    }

    return coloredMsg
}

func (entry LogcatEntry) ToString() string {
    
    time := formatTime(entry.Time)
    pid := entry.GetFormattedPidTid()
    name := fmt.Sprintf("%*s", MaxLenTag, entry.Tag) 

    level := fmt.Sprintf(" %s ", entry.Level)

    prefix := ascii.ScapeAnsi(time+pid+name)
    prefixLen := len(prefix)
    msg := ""
    for i, line := range strings.Split(entry.Message, "\n") {
        if i == 0 {
            msg += prefix + level + ascii.ScapeAnsi(line)
        }else{
            msg += fmt.Sprintf("\n%*s%s%s", prefixLen, "", level, ascii.ScapeAnsi(line))
        }
    }

    return msg
}

// Prints a logcat line with colors
func (entry LogcatEntry) Print() {
    fmt.Fprintln(color.Output, entry.ToAnsiString())
}

// Writes a logcat line to a file
func (entry LogcatEntry) ToFile(fh *os.File) (err error) {
    if fh == nil {
        return nil
    }

    _, err = fh.WriteString(fmt.Sprintf("%s\n", entry.ToString()))
    if err != nil {
        return err
    }

    return nil
}


func (entry LogcatEntry) GetFormattedPidTid() string {
    pid := entry.PID

    if len(pid) > MaxLenPid {
        return pid[:MaxLenPid]
    }

    for len(pid) < MaxLenPid {
        pid = " " + pid
    }

    tid := entry.TID

    if len(tid) > MaxLenPid {
        return tid[:MaxLenPid]
    }

    for len(tid) < MaxLenPid {
        tid += " "
    }

    return fmt.Sprintf("%s-%s ", pid, tid)
}

// Formats the tag to be colored and have a fixed length
func formatTime(time string) string {
    // Add a space if the tag is empty or does not end with a space
    if len(time) == 0 || time[len(time)-1] != ' ' {
        time = time + " "
    }

    // Trim the tag if it's too long
    if len(time) > MaxLenTime {
        return time[:MaxLenTime-1] + " "
    }

    // Add spaces to fill the rest of the line
    for len(time) < MaxLenTime {
        time += " "
    }

    return time
}

// Formats the tag to be colored and have a fixed length
func formatTag(tag string) string {
    // Add a space if the tag is empty or does not end with a space
    if len(tag) == 0 || tag[len(tag)-1] != ' ' {
        tag = tag + " "
    }

    // Trim the tag if it's too long
    if len(tag) > MaxLenTag {
        tag = tag[:MaxLenTag-3] + "..."
    }

    return fmt.Sprintf("%*s", MaxLenTag, tag)
}

// Formats the message to have a fixed length
func formatMsg(msg string, prefixSize int) string {
    // Get the console width (3rd party because the stdlib does not provide a working solution for Windows)
    width, _ := consolesize.GetConsoleSize()

    // Calculate the maximum width for the message
    //maxWidthMsg := width - MaxLenTime - MaxLenPid - MaxLenPid - MaxLenTag - 5 // 5 = 3 spaces, 1 char for level and 1 char for -
    maxWidthMsg := width - prefixSize - 1

    // Add a space if the message is empty or does not start with a space
    if len(msg) == 0 || msg[0] != ' ' {
        msg = " " + msg
    }

    // Trim the message if it's too long
    if len(msg) > maxWidthMsg {
        return msg[:maxWidthMsg-4] + "... "
    }

    // Add spaces to fill the rest of the line
    for len(msg) < maxWidthMsg {
        msg += " "
    }

    return msg
}


type NoDataError struct {
	Message string
}

func (e NoDataError) Error() string {
	return e.Message
}

