package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
)

type jsonResponse struct {
	Page  string   `json:"page"`
	Input string   `json:"input"`
	Words []string `json:"words"`
}

func main() {
	args := os.Args

	if len(args) < 2 {
		fmt.Println("Usage ./http-get <url>")
		os.Exit(1)
	}

	if _, err := url.ParseRequestURI(args[1]); err != nil {
		fmt.Printf("URL not valid format: %s\n", err)
	}

	response, err := http.Get(args[1])

	if err != nil {
		// Log is used for a system error
		// We want to be able to see the entire error
		log.Fatal(err)
	}

	defer response.Body.Close()
	/* Make sure that the response Body is closed
	when the function finishes running.
	It can be bigger than our memory alloc, it is streamed
	on demand, but we want to be sure we've closed it to
	free the memory. */

	body, err := io.ReadAll(response.Body)

	if err != nil {
		log.Fatal(err)
	}

	if response.StatusCode != 200 {
		fmt.Printf("Invalid request (HTTP Status Code: %d\n),Body: %s", response.StatusCode, body)
		os.Exit(1)
	}

	var words jsonResponse

	err = json.Unmarshal(body, &words)

	if err != nil {
		log.Fatal(err)
	}
	joined_words := strings.Join(words.Words, ", ")
	fmt.Printf("JSON Parsed\nPage: %s\nWords: %v\n", words.Page, joined_words)
}
