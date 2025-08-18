package adb

import (
    "context"
    "fmt"
    "os"
    "os/exec"
    "os/user"
    "path/filepath"
    "regexp"
    "strings"
    "time"
)

var (
    // Regex to get the full adb path from `adb version`
    reFullPath = regexp.MustCompile(`Installed as ([\S\/\\]+)`)
)

type Client struct {
    ADBPath       string   // Path to the ADB binary like /path/to/adb
    BaseCmd       []string // Base command like /path/to/adb -d
    BaseCmdLogcat []string // Base logcat command like /path/to/adb -d logcat
}

// Creates a new ADB client with the passed binPath like '/path/to/adb' and the connectionStr like '-d'
func NewClient(binPath string, connectionStr []string) (client *Client, err error) {
    client = &Client{}

    if err := client.setADBPath(binPath); err != nil {
        return nil, err
    }

    client.BaseCmd = []string{}
    client.BaseCmd = append(client.BaseCmd, client.ADBPath)
    client.BaseCmd = append(client.BaseCmd, connectionStr...)

    return client, nil
}

// Checks if ADB at binPath exists and gets the full path from the output of 'adb version'
func (client *Client) setADBPath(binPath string) (err error) {
    if binPath != "" {
        // User specified a custom path to adb

        // The tilde char will create an error, expand home
        if strings.HasPrefix(binPath, "~") {
            usr, _ := user.Current()
            homeDir := usr.HomeDir
            binPath = filepath.Join(homeDir, binPath[2:])
        }

        if _, err = os.Stat(binPath); err != nil {
            return fmt.Errorf("adb not found at %s", binPath)
        }
    } else {
        // User did not specify a path, look at $PATH
        binPath, err = exec.LookPath("adb")
        if err != nil {
            return fmt.Errorf("adb not found in $PATH")
        }
    }

    client.BaseCmd = append(client.BaseCmd, binPath) // Set this temporarily to make client.Run below work

    // Run adb version
    out, err := client.Run(5, "version")
    if err != nil {
        return err
    }

    // Parse out the full path displayed in 'adb version's output (makes sure it's actually adb)
    matches := reFullPath.FindStringSubmatch(out)
    if len(matches) != 2 {
        return fmt.Errorf("cloud not parse 'adb version' output: %s", out)
    }

    client.ADBPath = string(matches[1])
    if client.ADBPath == "" {
        return fmt.Errorf("cloud not parse 'adb version' output: %s", out)
    }

    return nil
}

// Runs an adb command with the passed timeout and arguments. Returns the output of the command
func (client *Client) Run(timeoutSeconds int, args ...string) (string, error) {
    ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeoutSeconds)*time.Second)
    defer cancel()

    adbCmd := []string{}
    adbCmd = append(adbCmd, client.BaseCmd...) // BaseCmd like /path/to/adb -d
    adbCmd = append(adbCmd, args...)

    cmd := exec.CommandContext(ctx, adbCmd[0], adbCmd[1:]...) // First string will be adb path
    out, err := cmd.CombinedOutput()

    if err != nil {
        if ctx.Err() == context.DeadlineExceeded {
            return "", fmt.Errorf("%s: timeout", strings.Join(adbCmd, " "))
        }

        return "", fmt.Errorf("%s\n%s\n%s", strings.Join(adbCmd, " "), out, err)
    }

    return string(out), nil
}

// Clears the logcat output via 'logcat -c'
func (client *Client) ClearLogcatOutput() (err error) {
    if _, err := client.Run(5, "logcat", "-c"); err != nil {
        return err
    }

    return nil
}