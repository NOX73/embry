package main

import (
	// I know than it's not good, but it so comfartably for development )
	"./client"
	"log"
)

func main() {
	var _ = client.NewClient()
	log.Println("Started")
}
