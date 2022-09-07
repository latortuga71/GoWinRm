package main

import (
	"log"

	"github.com/latortuga71/GoWinRm"
)

func main() {
	result, err := GoWinRm.WinRmExecuteCommand("HACKERLAB", "turtleadmin", "dawoof7123!!", "192.168.56.109", "5985", "hostname && whoami")
	if err != nil {
		log.Println(result, err)
		return
	}
	log.Println("-> ", result)
	// SSL untested but it should work
	// does not work with self signed certs because i was not able to get the session options to work
	// :(
	//GoWinRm.WinRmExecuteCommandSSL("HACKERLAB", "admin", "adminapss1123", "dc01", "5986", "hostname")
}
