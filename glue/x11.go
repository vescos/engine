// +build linux,!android

package glue

import (
	"log"
	"os"
	"runtime"
	"strings"
)

func init() {
	log.Print(">>>>> Status: Initializing...")
}

func (a *App) InitPlatform() {
	log.Printf(">>>>> Platform: %v/%v", runtime.GOARCH, runtime.GOOS)
	
	// Parse flags of type -flag=string 
	a.Flags = make(map[string]string)
	for _, v := range os.Args[1:] {
		sp := strings.SplitN(v, "=", 2)
		if len(sp) < 2 {
			log.Printf("Can't parse flag: %v. Use -flag=string", v)
			continue
		}
		// remove leading -
		if sp[0][0] != []byte("-")[0] {
			log.Printf("Missing '-' in flag definition: %v. Use -flag=string", v)
			continue
		}
		key := sp[0][1:]
		a.Flags[key] = sp[1]
	}
}
