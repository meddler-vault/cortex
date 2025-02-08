package main

import (
	"log"

	"github.com/meddler-vault/cortex/bootstrap"
)

func TestClone() {

	rep, err := bootstrap.Clone("https://github.com/meddler-io/cyclops-ui.git", "/tmp/watch-dog-test/", "basicauth", "x-token-auth", "{{secret}}", "", 1)
	log.Println(rep)
	log.Println(err)
	// t.Errorf("Add(2, 3) = %d; want %d")

}

func main() {
	TestClone()
}
