package cmd

import (
	"regexp"
    "strings"
    "errors"
    "fmt"
    "os"

    "github.com/helviojunior/adbcat/internal/ascii"
    "github.com/helviojunior/adbcat/internal/tools"
    "github.com/helviojunior/adbcat/pkg/readers"
    //"github.com/helviojunior/adbcat/pkg/models"
    "github.com/helviojunior/adbcat/pkg/log"
    resolver "github.com/helviojunior/gopathresolver"
    "github.com/spf13/cobra"
)

var runner *readers.LogcatRunner
var tmpExcludeFilter = []string{}
var tmpIncludeFilter = []string{}

var logcatCmd = &cobra.Command{
    Use:   "logcat",
    Short: "Get colored and formatted Android logs",
    Long: ascii.LogoHelp(ascii.Markdown(`
# logcat

Get colored and formatted Android logs
`)),
    Example: `
- adbcat logcat
- adbcat logcat -o logcat.txt
- adbcat logcat -p com.android.chrome
- adbcat logcat --show-time --show-pid
`,
    PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
        var err error

        // Annoying quirk, but because I'm overriding PersistentPreRun
        // here which overrides the parent it seems.
        // So we need to explicitly call the parent's one now.
        if err = rootCmd.PersistentPreRunE(cmd, args); err != nil {
            return err
        }

        if opts.LogFile != "" {
            fp1, err := resolver.ResolveFullPath(opts.LogFile)
            if err != nil {
                return err
            }

            opts.LogFile = fp1
        }

        re := regexp.MustCompile("[^a-zA-Z0-9@-_.]")
        for _, s1 := range tmpIncludeFilter {
            incLines := []string{}

            s1 = strings.Trim(s1, " ")
            if len(s1) > 1 {
                if s1[0:1] == "@" {

                    f1, err := resolver.ResolveFullPath(s1[1:])
                    if err != nil {
                        return errors.New(fmt.Sprintf("Invalid file path (%s): %s", s1[1:], err.Error()))
                    }
                    if !tools.FileExists(f1) {
                        return errors.New(fmt.Sprintf("Invalid file path (%s): %s", s1[1:], "File not found"))
                    }

                    readers.ReadAllLines(f1, &incLines)

                }else{
                    incLines = append(incLines, s1)
                }
                for _, s2 := range incLines {
                
                    s3 := strings.ToLower(strings.Trim(s2, " "))
                    s3 = re.ReplaceAllString(s2, "")
                    if s3 != "" {
                        opts.IncludeFilterList = append(opts.IncludeFilterList, s3)
                    }
                }
            }
        }

        for _, s1 := range tmpExcludeFilter {
            incLines := []string{}

            s1 = strings.Trim(s1, " ")
            if len(s1) > 1 {
                if s1[0:1] == "@" {

                    f1, err := resolver.ResolveFullPath(s1[1:])
                    if err != nil {
                        return errors.New(fmt.Sprintf("Invalid file path (%s): %s", s1[1:], err.Error()))
                    }
                    if !tools.FileExists(f1) {
                        return errors.New(fmt.Sprintf("Invalid file path (%s): %s", s1[1:], "File not found"))
                    }

                    readers.ReadAllLines(f1, &incLines)

                }else{
                    incLines = append(incLines, s1)
                }
                for _, s2 := range incLines {
                    s3 := strings.ToLower(strings.Trim(s2, " "))
                    s3 = re.ReplaceAllString(s2, "")
                    if s3 != "" {
                        opts.ExcludeFilterList = append(opts.ExcludeFilterList, s3)
                    }
                }
            }
        }

        return nil
    },
    PreRunE: func(cmd *cobra.Command, args []string) error {
        var err error

        runner, err = readers.NewRunner(*opts)
        if err != nil {
            return err
        }

        if opts.LogFile != "" && opts.ClearOutput {
            err := os.Truncate(opts.LogFile, 0)
            if err != nil {
                return err
            }
        }

        return nil
    },
    Run: func(cmd *cobra.Command, args []string) {
        //var ft string
        //var err error

        log.Info("Starting process...")

        runner.Run()

    },
}


func init() {
    rootCmd.AddCommand(logcatCmd)

    logcatCmd.PersistentFlags().StringSliceVar(&tmpExcludeFilter, "exclude", []string{}, "Exclude all messages with specified strings. You can specify multiple values by comma-separated terms or by repeating the flag. Use @filename to load from text file.")
    logcatCmd.PersistentFlags().StringSliceVar(&tmpIncludeFilter, "include", []string{}, "Include only messages with specified strings. You can specify multiple values by comma-separated terms or by repeating the flag. Use @filename to load from text file.")    
    logcatCmd.PersistentFlags().StringVarP(&opts.LogFile, "log-file", "o", "", "Write logcat output to file.")
    logcatCmd.PersistentFlags().BoolVar(&opts.UseAnsiLog, "log-file-ansi", false, "Use ANSI colors at log file.")
    logcatCmd.PersistentFlags().StringVarP(&opts.MinLevel, "min-level", "l", "V", "Minimum log level to be displayed (V,D,I,W,E,F) (default 'V').")

    logcatCmd.PersistentFlags().BoolVarP(&opts.ClearOutput, "clear", "c", false, "Clear the log before running")
    logcatCmd.PersistentFlags().BoolVarP(&opts.UseDevice, "device", "d", false, "Use the first device (adb -d)")
    logcatCmd.PersistentFlags().BoolVarP(&opts.UseEmulator, "emulator", "e", false, "use the first emulator (adb -e)")
    logcatCmd.PersistentFlags().StringVarP(&opts.DeviceSerial, "serial", "s", "", "Sevice serial number (adb -s)")

    logcatCmd.Flags().StringVarP(&opts.PackageName, "package", "p", "", "Application package name.")

    logcatCmd.Flags().StringVar(&opts.AdbBinPath, "adb-path", "", "Path to the ADB binary")

    logcatCmd.PersistentFlags().BoolVar(&opts.ShowTime, "show-time", false, "Display time")
    logcatCmd.PersistentFlags().BoolVar(&opts.ShowPid, "show-pid", false, "Displey PID/TID")
}
