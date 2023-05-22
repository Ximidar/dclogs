package main

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/google/uuid"
)

func print_messages() {
	id := uuid.New()
	t := time.Now()
	fmt.Printf("[%s] UUID: %s\n", t.Format("2006-01-02 15:04:05"), id.String())
}

func main() {
	min := 1
	max := 5
	offset := 10 + rand.Intn(max-min+1)
	duration := time.Duration(offset) * time.Second
	fmt.Println("Sleeping for", offset, "seconds")
	// Main loop
	for {
		print_messages()
		time.Sleep(duration)
	}
}
