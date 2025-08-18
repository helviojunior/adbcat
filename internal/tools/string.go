package tools

import (
	"strconv"
    
    "strings"
)

// LeftTrucate a string if its more than max
func LeftTrucate(s string, max int) string {
	if len(s) <= max {
		return s
	}

	return s[max:]
}

func FormatInt(n int) string {
	return FormatInt64(int64(n))
}

func FormatInt64(n int64) string {
    in := strconv.FormatInt(n, 10)
    numOfDigits := len(in)
    if n < 0 {
        numOfDigits-- // First character is the - sign (not a digit)
    }
    numOfCommas := (numOfDigits - 1) / 3

    out := make([]byte, len(in)+numOfCommas)
    if n < 0 {
        in, out[0] = in[1:], '-'
    }

    for i, j, k := len(in)-1, len(out)-1, 0; ; i, j = i-1, j-1 {
        out[j] = in[i]
        if i == 0 {
            return string(out)
        }
        if k++; k == 3 {
            j, k = j-1, 0
            out[j] = '.'
        }
    }
}

func FormatCN(cn string) string {
    if len(cn) <= 3 {
        return cn
    }
    txt := cn
    if strings.ToLower(txt[0:3]) == "cn=" {
        p := strings.Split(txt, ",")
        if len(p) >= 1 {
            txt = strings.Replace(strings.Replace(p[0], "CN=", "", -1), "cn=", "", -1)
        }
    }
    if txt == "" {
        txt = cn
    }
    txt = strings.Replace(txt, "\"", "", -1)
    txt = strings.Replace(txt, "'", "", -1)
    return txt
} 