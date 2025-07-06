package pkg

import (
	"log"
	"os"
)

var Logger = log.New(os.Stdout, "gollman: ", log.LstdFlags|log.Lshortfile)
