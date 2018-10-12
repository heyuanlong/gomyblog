package main

import (
	"fmt"

	"github.com/microcosm-cc/bluemonday"
	"gopkg.in/russross/blackfriday.v2"
)

func main() {
	input := []byte("Hello.\n\n* This is markdown.\n* It is fun\n* Love it or leave it.")
	unsafe := blackfriday.Run(input)
	html := bluemonday.UGCPolicy().SanitizeBytes(unsafe)
	fmt.Println(string(unsafe))
	fmt.Println(string(html))
}
