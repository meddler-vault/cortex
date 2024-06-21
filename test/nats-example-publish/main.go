package main

import (
	"log"

	producernats "github.com/meddler-io/watchdog/producer-nats"
)


func main(){


	er :=producernats.Produce( "whitehat" ,  "4Jy6P)$Ep@c^SenL",  "rmq.meddler.io:443", "tasks_test", "jai-shree-ram")
	log.Println("Error", er)
}