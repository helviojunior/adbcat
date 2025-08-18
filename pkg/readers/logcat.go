package readers

import (
    "context"
    "strings"
    "fmt"

    "bufio"
    "os"
    "os/exec"
    "os/signal"
    "sync"
    "syscall"
    "time"
    "slices"

    "github.com/fatih/color"
    "github.com/helviojunior/adbcat/pkg/models"
    "github.com/helviojunior/adbcat/pkg/log"
    "github.com/helviojunior/adbcat/pkg/adb"
)

type LogcatRunner struct {
    
    ADBClient *adb.Client
    Logcat    *adb.LogcatOptions

    options Options

    //Context
    ctx    context.Context
    cancel context.CancelFunc

    logFile *os.File

    running bool

}

func NewRunner(opts Options) (*LogcatRunner, error) {
    var err error
    ctx, cancel := context.WithCancel(context.Background())

    // Select the connection option
    connectionStr := []string{}
    if opts.DeviceSerial != "" {
        connectionStr = append(connectionStr, []string{"-s", opts.DeviceSerial}...)
    } else if opts.UseDevice {
        connectionStr = append(connectionStr, "-d")
    } else if opts.UseEmulator {
        connectionStr = append(connectionStr, "-e")
    } else {
        //return nil, fmt.Errorf("mission adb option, chooose '-s/--serial', '-d/--device' or '-e/--emulator'")
    }

    runner := LogcatRunner{
        ctx:        ctx,
        cancel:     cancel,
        options:    opts,
        Logcat: &adb.LogcatOptions{},
        logFile: nil,
        running: true,
    }

    runner.ADBClient, err = adb.NewClient(opts.AdbBinPath, connectionStr)
    if err != nil {
        return nil, err
    }

    if opts.PackageName == "" {
        // Users wants all packages, do not filter

    } else {
        runner.Logcat.Packages = append(runner.Logcat.Packages, opts.PackageName)
    }

    runner.ADBClient.BaseCmdLogcat = append(runner.ADBClient.BaseCmd, "logcat")

    if opts.ClearOutput {
        if err := runner.ADBClient.ClearLogcatOutput(); err != nil {
            return nil, err
        }
    }

    minLevel := strings.ToUpper(opts.MinLevel)
    if _, ok := models.LevelMap[minLevel]; !ok {
        return nil, fmt.Errorf("invalid level '%s'", minLevel)
    }
    runner.Logcat.MinLevel = minLevel


    if opts.LogFile != "" {
        runner.logFile, err = os.OpenFile(opts.LogFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0600)
        if err != nil {
            return nil, err
        }
    }

    return &runner, nil
}

func (run LogcatRunner) Run() {
    defer run.cancel()

    pids := []string{}

    cmd := exec.CommandContext(run.ctx, run.ADBClient.BaseCmdLogcat[0], run.ADBClient.BaseCmdLogcat[1:]...)

    // Capture the output of the logcat command
    stdout, err := cmd.StdoutPipe()
    if err != nil {
        log.Errorf("%s", err)
        return
    }

    // Start logcat
    if err := cmd.Start(); err != nil {
        log.Errorf("%s", err)
        os.Exit(1)
    }

    if len(run.Logcat.Packages) > 0 {
        pids = append(pids, "invalid")  // Create an pid to ignore all other packeges logs
    }

    // Create a go function that every two seconds checks for the PIDs of the wanted packages
    stopChanPidWatchDog := make(chan bool)
    wgPidWatchDog := new(sync.WaitGroup)
    wgPidWatchDog.Add(1)
    go func() {
        defer wgPidWatchDog.Done()

        for {
            select {
            case <-stopChanPidWatchDog:
                // We got a stop signal, return
                return
            default:
                for _, slug := range run.Logcat.Packages {
                    pid, err := run.ADBClient.GetPID(slug)
                    if err != nil {
                        log.Errorf("%s", err)
                        os.Exit(1)
                    }

                    // Add the pid to the slice if it's not already there
                    if !slices.Contains(pids, pid) {
                        pids = append(pids, pid)
                    }
                }
            }

            time.Sleep(time.Second * 2)
        }
    }()

    // Channel were the logcat lines are sent to
    chanLogcatLines := make(chan string)
    wgOutputWriter := new(sync.WaitGroup)

    // Start a go function that reads the logcat lines and prints them to the terminal after filtering and formatting
    wgOutputWriter.Add(1)
    go func() {
        defer wgOutputWriter.Done()

        var lastLine *adb.AdbLineEntry
        var currentEntry *models.LogcatEntry
        for line := range chanLogcatLines {
            entry, err := adb.ParseLogcatLine(line)
            if err != nil {
                continue // Ignore parse errors
            }

            if !entry.EqualTimePidLevel(lastLine) && currentEntry != nil {
                run.DispatchEntry(currentEntry)
                currentEntry = nil
            }

            // Check if the PID of the entry is not in the wanted PIDs
            if len(pids) > 0 && !slices.Contains(pids, entry.PID) {
                continue
            }

            // Check if the level is in scope to be processed
            if !adb.IsLevelInScope(entry.Level, run.Logcat.MinLevel) {
                continue
            }

            if run.CheckIgnore(entry) {
                continue
            }

            // Print the logcat line
            //Check if is the same time/pid/level
            if entry.EqualTimePidLevel(lastLine) && currentEntry != nil {
                currentEntry.Message += "\n" + entry.Message
            }else{
                currentEntry = &models.LogcatEntry{
                    Date:       entry.Date,
                    Time:       entry.Time,
                    Level:      entry.Level,
                    Tag:        entry.Tag,
                    PID:        entry.PID,
                    TID:        entry.TID,
                    Message:    entry.Message,
                }
                
            }

            lastLine = &entry
            
        }
    }()

    // Start a go function that reads the logcat lines and sends them to the channel
    wgLogcatReader := new(sync.WaitGroup)
    wgLogcatReader.Add(1)
    go func() {
        defer wgLogcatReader.Done()

        // Read the logcat lines
        scanner := bufio.NewScanner(stdout)
        for scanner.Scan() {
            chanLogcatLines <- scanner.Text()
        }

        if err := scanner.Err(); err != nil {
            log.Errorf("%s", err)
            os.Exit(1)
        }

        close(chanLogcatLines)
    }()

    // Wait for the user to press CTRL+C
    c := make(chan os.Signal, 1)
    signal.Notify(c, os.Interrupt, syscall.SIGTERM)
    go func() {
        <-c
        run.cancel()
        run.running = false
    }()

    // Wait for the logcat process to finish
    cmd.Wait()

    wgLogcatReader.Wait()
    wgOutputWriter.Wait()
}

func (run LogcatRunner) CheckIgnore(logEntry adb.AdbLineEntry) bool {

    txt := fmt.Sprintf("%s %s %s", logEntry.PID, logEntry.Tag, logEntry.Message)

    // Check if the tag is to be included
    if len(run.options.IncludeFilterList) > 0 {
        for _, slug := range run.options.IncludeFilterList {
            if strings.Contains(txt, slug) {
                return false
            }
        }
        return true
    }

    // Check if the tag is to be ignored
    if len(run.options.ExcludeFilterList) > 0 {
        for _, slug := range run.options.ExcludeFilterList {
            if strings.Contains(txt, slug) {
                return true
            }
        }
    }


    return false
}

func (run LogcatRunner) DispatchEntry(logEntry *models.LogcatEntry) {
    if !run.running {
        return
    }

    fmt.Fprintln(color.Output, logEntry.FormatAnsiString(run.options.ShowTime, run.options.ShowPid))

    logEntry.ToFile(run.logFile)
}