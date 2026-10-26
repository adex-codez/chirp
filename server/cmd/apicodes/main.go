// Command apicodes writes the mobile client's copy of the API's error codes.
//
// The server's codes are defined once, in internal/response/codes.go. This
// command reads that file and writes chirp-mobile/src/lib/api-codes.ts, so the
// client cannot fall behind the server.
package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"backend/internal/apicodegen"
)

func main() {
	start := flag.String("from", ".", "directory to search for the repository root")
	out := flag.String("out", "", "output path, relative to the repository root")
	flag.Parse()

	root, err := apicodegen.FindRepoRoot(*start)
	if err != nil {
		log.Fatal(err)
	}

	content, err := apicodegen.Generate(root)
	if err != nil {
		log.Fatal(err)
	}

	target := apicodegen.OutputPath
	if *out != "" {
		target = *out
	}
	path := filepath.Join(root, target)

	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		log.Fatal(err)
	}

	fmt.Printf("wrote %s\n", target)
}
