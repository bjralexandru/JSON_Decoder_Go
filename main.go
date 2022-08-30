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

/* STRUCTS TO LAYOUT THE GET RESPONSE */
type Page struct {
	Name string `json:"page"`
}

type Words struct {
	Input string   `json:"input"`
	Words []string `json:"words"`
}

type Occurrence struct {
	Words map[string]int `json:"words"`
}

func main() {
	/* CHECK URL VALIDITY */
	args := os.Args

	if len(args) < 2 {
		fmt.Println("Usage ./http-get <url>")
		os.Exit(1)
	}
	res, err := doRequests(args[1])

	if err != nil {
		fmt.Printf("Error: %s\n", err)
		os.Exit(1)
	}
}

func doRequests(requestURL string) (Response, error) {

	if _, err := url.ParseRequestURI(requestURL); err != nil {
		fmt.Printf("URL not valid format: %s\n", err)

		/* GET REQUEST LOGIC */
		response, err := http.Get(requestURL)

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

		var page Page

		err = json.Unmarshal(body, &page)

		if err != nil {
			log.Fatal(err)
		}

		switch page.Name {
		case "words":
			var words Words

			err = json.Unmarshal(body, &words)

			if err != nil {
				log.Fatal(err)
			}

			joined_words := strings.Join(words.Words, ", ")
			fmt.Printf("JSON Parsed\nPage: %s\nWords: %v\n", page.Name, joined_words)
		case "occurrence":
			var occurrence Occurrence

			err = json.Unmarshal(body, &occurrence)

			if err != nil {
				log.Fatal(err)
			}
			for key, value := range occurrence.Words {
				fmt.Printf("%s: %d\n", key, value)
			}
		default:
			fmt.Println("Page not found!")
		}

	}
}
