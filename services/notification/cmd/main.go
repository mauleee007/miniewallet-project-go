package main

import (
	"fmt"
	"time"
)

func main() {
	fmt.Println("Notification Service starting...")
	
	// Keep the process running
	for {
		fmt.Println("Notification Service background worker running...")
		time.Sleep(30 * time.Second)
	}
}
