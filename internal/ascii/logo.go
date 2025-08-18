package ascii

import (
	"fmt"
	"strings"
	"github.com/helviojunior/adbcat/internal/version"
)

// Logo returns the adbcat ascii logo
func Logo() string {
	txt := `{W}{G}                      
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
 @@{B}    _____   ________  __________ {O}_________           __   {G}    
 @@{B}   /  _  \  \______ \ \______   \{O}\_   ___ \ _____  _/  |_ {G} 
 @@{B}  /  /_\  \  |    |  \ |    |  _/{O}/    \  \/ \__  \ \   __\{G} 
 @@{B} /    |    \ |    |   \|    |   \{O}\     \____ / __ \_|  |  {G} 
 @@{B} \____|__  //_______  /|______  /{O} \________/(______/|__|  {G} 
 @@{B}         \/         \/        \/   {GR}v {version} {G}{W}
`

	v := fmt.Sprintf("%s-%s", version.Version, version.GitHash)
	txt = strings.Replace(txt, "{version}", v, -1)
	txt = strings.Replace(txt, "{G}", "\033[32m", -1)
	txt = strings.Replace(txt, "{B}", "\033[36m", -1)
	txt = strings.Replace(txt, "{GR}", "\033[0m\033[1;90m", -1)
	txt = strings.Replace(txt, "{R}", "\033[1;31m", -1)
	txt = strings.Replace(txt, "{O}", "\033[33m", -1)
	txt = strings.Replace(txt, "{W}", "\033[0m", -1)
	return fmt.Sprintln(txt)
}

// LogoHelp returns the logo, with help
func LogoHelp(s string) string {
	return Logo() + "\n\n" + s
}
