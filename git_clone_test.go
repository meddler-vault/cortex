package main

import (
	"log"
	"testing"

	"github.com/meddler-vault/cortex/bootstrap"
)

func TestClone(t *testing.T) {

	rep, err := bootstrap.Clone("https://github.com/tinygrad/tinygrad.git", "/tmp/watch-dog-test/", "basicauth", "studiogangster", "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", "", 1)
	log.Println(rep)
	log.Println(err)
	// t.Errorf("Add(2, 3) = %d; want %d")

}
