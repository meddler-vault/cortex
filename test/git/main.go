package main

import (
	"log"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/meddler-vault/cortex/bootstrap"
)

func main() {

	log.Println("Clone options")
	repository, err := bootstrap.Clone(
		"https://github.com/studiogangster/sensibull-realtime-options-api-ingestor.git",
		"/tmp/test-clone",
		"no_asuth",
		"",
		"",
		// "refs/heads/master",
		"",
		1,
	)

	log.Println("Error", err)

	if err != nil {
		return
	}

	commitID := "6cd7ffabf88c3ea295e35c3334781987bd651843"
	// return

	// Checkout the specific commit ID
	hash := plumbing.NewHash(commitID)

	// Checkout the commit
	worktree, err := repository.Worktree()
	if err != nil {
		log.Println("failed to get worktree:", err)
		return
	}

	err = worktree.Checkout(&git.CheckoutOptions{
		Hash: hash,
	})
	if err != nil {
		log.Println("failed to checkout commit ", commitID, err)
		return
	}

}
