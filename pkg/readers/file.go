package readers

import (
    "bufio"
    "os"
    "strings"
)

// Read from a file.
func ReadAllLines(fileName string, outList *[]string) error {

    var file *os.File
    var err error

    file, err = os.Open(fileName)
    if err != nil {
        return err
    }
    defer file.Close()

    reader := bufio.NewReader(file)
    for {
        lastLine := false
        line, err := reader.ReadString('\n')

        if err != nil {
            // Print last line even without newline
            lastLine = true
        }

        line = strings.Trim(strings.Replace(strings.Replace(strings.ToLower(line), "\n", "", -1), "\r", "", -1), " ")
        if line != "" {
            *outList = append(*outList, line)
        }

        if lastLine {
            break
        }
    }

    return nil
}
