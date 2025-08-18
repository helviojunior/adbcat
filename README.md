# ADB Cat

Makes `adb logcat` colored and adds the feature of being able to filter by app or keywords.

![adbcat](screenshots/adbcat1.png)

## Get last release

Check how to get last release by your Operational Systems procedures here [INSTALL.md](https://github.com/helviojunior/adbcat/blob/main/INSTALL.md)


# Utilization

```
$ adbcat logcat -h

        @                        @
         @.                     @
          @@     *@@@@@@#.    %@
           @@@@@@@@@@@@@@@@@@@@
        @@@@@@@@@@@@@@@@@@@@@@@@@@
      @@@@@@@@@@@@@@@@@@@@@@@@@@@@@@
    @@@@@@#  @@@@@@@@@@@@@@@@  -@@@@@@
   @@@@@@@    @@@@@@@@@@@@@@.   @@@@@@@
  @@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@
 :@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@
 @@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@
 @@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@

 @@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@
 @@
 @@    _____   ________  __________ _________           __
 @@   /  _  \  \______ \ \______   \\_   ___ \ _____  _/  |_
 @@  /  /_\  \  |    |  \ |    |  _//    \  \/ \__  \ \   __\
 @@ /    |    \ |    |   \|    |   \\     \____ / __ \_|  |
 @@ \____|__  //_______  /|______  / \________/(______/|__|
 @@         \/         \/        \/   v dev-dev




   logcat

  Get colored and formatted Android logs

Usage:
  adbcat logcat [flags]

Examples:

- adbcat logcat
- adbcat logcat -o logcat.txt
- adbcat logcat -p com.android.chrome


Flags:
      --adb-path string    Path to the ADB binary
  -c, --clear              Clear the log before running
  -d, --device             Use the first device (adb -d)
  -e, --emulator           use the first emulator (adb -e)
      --exclude strings    Exclude all messages with specified strings. You can specify multiple values by comma-separated terms or by repeating the flag. Use @filename to load from text file.
  -h, --help               help for logcat
      --include strings    Include only messages with specified strings. You can specify multiple values by comma-separated terms or by repeating the flag. Use @filename to load from text file.
  -o, --log-file string    Write logcat output to file.
  -l, --min-level string   Minimum log level to be displayed (V,D,I,W,E,F) (default 'V'). (default "V")
  -p, --package string     Application package name.
  -s, --serial string      Sevice serial number (adb -s)

Global Flags:
  -D, --debug-log   Enable debug logging
  -q, --quiet       Silence (almost all) logging

```

## Info

This is a Golang port~ of [github.com/JakeWharton/pidcat](https://github.com/JakeWharton/pidcat).

### How does this work?

Run `adb logcat` and parse his output. If `--package` is provided, `adb shell ps` is used to get the PIDs of the wanted packages and filter by lines/entries have a PID assigned to them.

