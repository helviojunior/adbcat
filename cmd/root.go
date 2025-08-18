package cmd

import (
	"os"
	"fmt"
	"os/signal"
    "syscall"
    "time"
    //"runtime"

	"github.com/helviojunior/adbcat/internal/ascii"
	"github.com/helviojunior/adbcat/internal/tools"
	"github.com/helviojunior/adbcat/pkg/log"
	"github.com/helviojunior/adbcat/pkg/readers"
    "github.com/spf13/cobra"
)

var tempFolder string
var workspacePath string
var opts = &readers.Options{}
var rootCmd = &cobra.Command{
	Use:   "adbcat",
	Short: "Get colored and formatted Android logs",
	Long:  ascii.Logo(),
	Example: `
- adbcat logcat
- adbcat logcat -o logcat.txt
- adbcat logcat -p com.android.chrome
`,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		
	    if cmd.CalledAs() != "version" && !opts.Logging.Silence {
			fmt.Println(ascii.Logo())
		}

		if opts.Logging.Silence {
			log.EnableSilence()
		}

		if opts.Logging.Debug && !opts.Logging.Silence {
			log.EnableDebug()
			log.Debug("debug logging enabled")
		}

		return nil
	},
}

func Execute() {
	
	ascii.SetConsoleColors()

	c := make(chan os.Signal)
    signal.Notify(c, os.Interrupt, syscall.SIGTERM)
    go func() {
        <-c
        ascii.ClearLine()
        fmt.Fprintf(os.Stderr, "\r\n")
        ascii.ClearLine()
        ascii.ShowCursor()
        log.Warn("interrupted, shutting down...                            ")
        ascii.ClearLine()
        fmt.Printf("\n")
        tools.RemoveFolder(tempFolder)
        os.Exit(2)
    }()

	rootCmd.CompletionOptions.DisableDefaultCmd = true
	rootCmd.SilenceErrors = true
	err := rootCmd.Execute()
	if err != nil {
		var cmd string
		c, _, cerr := rootCmd.Find(os.Args[1:])
		if cerr == nil {
			cmd = c.Name()
		}

		v := "\n"

		if cmd != "" {
			v += fmt.Sprintf("An error occured running the `%s` command\n", cmd)
		} else {
			v += "An error has occured. "
		}

		v += "The error was:\n\n" + fmt.Sprintf("```%s```", err)
		fmt.Println(ascii.Markdown(v))

		os.Exit(1)
	}

	//Time to wait the logger flush
	time.Sleep(time.Second/4)
    tools.RemoveFolder(tempFolder)
    ascii.ShowCursor()
    fmt.Printf("\n")
}

func init() {
	
	rootCmd.PersistentFlags().BoolVarP(&opts.Logging.Debug, "debug-log", "D", false, "Enable debug logging")
	rootCmd.PersistentFlags().BoolVarP(&opts.Logging.Silence, "quiet", "q", false, "Silence (almost all) logging")
        
}
