package adb

import (
    "fmt"
    "regexp"
    "slices"
    "strings"

    "github.com/helviojunior/adbcat/pkg/log"
)

var (
    // A regex to parse the output of 'adb shell ps'
    rePSOutput = regexp.MustCompile(`(\S+)\s+(\d+)\s+(\d+)\s+(\d+)\s+(\d+)\s+(\S+)\s+(\S+)\s+(\S)\s+(\S+)`)
    // Regex to parse out the slug of the 'adb dumsys' output
    reForegroundApp = regexp.MustCompile(`.*Recent #0: \S+{\S+ \S+ \S+ \S+:([\S.]+)}.*`)
    // Regex to parse out the slug of the 'adb devices' output
    reDevicesApp = regexp.MustCompile(`(.*)\s+(\S+)`)
)

// One line of the output from 'adb shell ps'
type Process struct {
    USER  string
    PID   string
    PPID  string
    VSZ   string
    RSS   string
    WCHAN string
    ADDR  string
    LEVEL string
    NAME  string
}

// Returns the PID of the passed app identified by its slug (com.example.app)
//
// If no process is matched, an error is returned
func (client *Client) GetPID(slug string) (pid string, err error) {
    processes, err := client.GetProcesses()
    if err != nil {
        return "", err
    }

    for _, process := range processes {
        if strings.EqualFold(process.NAME, slug) { // Trim should not be necessary
            return process.PID, nil
        }
    }

    return "", fmt.Errorf("no PID for '%s' found", slug)
}

// Runs 'adb shell ps' to check conectivity with device
func (client *Client) CheckConn() (err error) {
    processes, err := client.GetProcesses()
    if err != nil {
        return err
    }

    if len(processes) == 0 {
        return fmt.Errorf("error checking device connection, do you have an connected device?")
    }

    return err
}

// Runs 'adb shell ps' and parses the output into a custom struct
func (client *Client) GetProcesses() (processes []*Process, err error) {
    out, err := client.Run(10, "shell", "ps")
    if err != nil {
        return nil, err
    }

    processes = []*Process{}

    for _, line := range strings.Split(out, "\n") {
        matches := rePSOutput.FindStringSubmatch(line)
        if len(matches) != 10 {
            continue // Ignore lines that don't match (header for example)
        } else {
            processes = append(processes, &Process{
                USER:  strings.TrimSpace(matches[1]),
                PID:   strings.TrimSpace(matches[2]),
                PPID:  strings.TrimSpace(matches[3]),
                VSZ:   strings.TrimSpace(matches[4]),
                RSS:   strings.TrimSpace(matches[5]),
                WCHAN: strings.TrimSpace(matches[6]),
                ADDR:  strings.TrimSpace(matches[7]),
                LEVEL: strings.TrimSpace(matches[8]),
                NAME:  strings.TrimSpace(matches[9]),
            })
        }
    }

    return processes, err
}

// Returns a list of all packages installed on the device via 'adb shell pm list packages'
func (client *Client) ListDevices() (devices []string, err error) {
    out, err := client.Run(5, "devices")
    if err != nil {
        return nil, err
    }

    emulatorOnly := false
    if slices.Contains(client.BaseCmd, "-e") {
        emulatorOnly = true
    }

    devices = []string{}
    for _, line := range strings.Split(out, "\n") {
        if !strings.Contains(strings.ToLower(line), "list of devices attached") {
            matches := reDevicesApp.FindStringSubmatch(line)
            if len(matches) == 3 {
                d1 := strings.TrimSpace(matches[1])
                if d1 != "" && strings.Contains(matches[2], "device") {
                    if emulatorOnly && !strings.Contains(strings.ToLower(d1), "emulator-") {
                        log.Warnf("Ignoring non emulator device: %s", d1)
                        continue
                    }
                    devices = append(devices, d1) 
                }
            }
        }
    }

    if len(devices) == 0 {
        return devices, fmt.Errorf("error parsing devices output, do you have an connected device?")
    }

    return devices, nil
}

// Returns a list of all packages installed on the device via 'adb shell pm list packages'
func (client *Client) ListAllPackages() (packages []string, err error) {
    out, err := client.Run(5, "shell", "pm", "list", "packages")
    if err != nil {
        return nil, err
    }

    packages = []string{}
    for _, line := range strings.Split(out, "\n") {
        if strings.HasPrefix(line, "package:") {
            packages = append(packages, strings.TrimSpace(strings.TrimPrefix(line, "package:")))
        }
    }

    slices.Sort(packages)

    return packages, nil
}

// Returns a list of all third party packages installed on the device via 'adb shell pm list packages -3'
func (client *Client) ListThirdPartyPackages() (packages []string, err error) {
    out, err := client.Run(5, "shell", "pm", "list", "packages", "-3")
    if err != nil {
        return nil, err
    }

    packages = []string{}
    for _, line := range strings.Split(out, "\n") {
        if strings.HasPrefix(line, "package:") {
            packages = append(packages, strings.TrimSpace(strings.TrimPrefix(line, "package:")))
        }
    }

    slices.Sort(packages)

    return packages, nil
}

// Returns the slug (com.example.app) of the app in the foreground via 'adb shell dumsys'
func (client *Client) GetCurrentApp() (slug string, err error) {
    out, err := client.Run(5, "shell", "dumpsys", "activity", "recents")
    if err != nil {
        return "", err
    }

    matches := reForegroundApp.FindStringSubmatch(out)
    if len(matches) != 2 {
        return "", fmt.Errorf("error parsing dumpsys output, do you have an app open?")
    }

    slug = matches[1]
    if slug == "" {
        return "", fmt.Errorf("error parsing dumpsys output, do you have an app open?")
    }

    return slug, nil
}
