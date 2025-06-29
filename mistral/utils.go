package mistral

import (
	"log"
	"os"
)

var (
	logger = log.New(os.Stdout, "mistral: ", log.LstdFlags|log.Lshortfile)
)
