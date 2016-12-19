package util

import "log"
import "os"

func Check(e error) {
	if e != nil {
		log.Println(e)
		os.Exit(1)
	}
}
