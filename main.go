package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
)

type Response interface {
	GetResponse() string
}

/* STRUCTS TO LAYOUT THE GET RESPONSE */
type Page struct {
	Name string `json:"page"`
}

type Words struct {
	Input string   `json:"input"`
	Words []string `json:"words"`
}

func (w Words) GetResponse() string {
	return fmt.Sprintf("%v", strings.Join(w.Words, ", "))
}

type Occurrence struct {
	Words map[string]int `json:"words"`
}

func (o Occurrence) GetResponse() string {
	out := []string{}

	for word, occurence := range o.Words {
		out = append(out, fmt.Sprintf("%s {%d}", word, occurence))
	}
	return fmt.Sprintf("%v", strings.Join(out, ", "))
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

	if res == nil {
		fmt.Printf("No response: %s\n", res)
		os.Exit(1)
	}

	fmt.Printf("Response: %s\n", res.GetResponse())
}

func doRequests(requestURL string) (Response, error) {

	if _, err := url.ParseRequestURI(requestURL); err != nil {
		return nil, fmt.Errorf("URL is not valid error: %s", err)
	}
	/* GET REQUEST LOGIC */
	response, err := http.Get(requestURL)

	if err != nil {
		// Log is used for a system error
		// We want to be able to see the entire error
		return nil, fmt.Errorf("HTTP Get error: %s", err)
	}

	defer response.Body.Close()
	/* Make sure that the response Body is closed
	when the function finishes running.
	It can be bigger than our memory alloc, it is streamed
	on demand, but we want to be sure we've closed it to
	free the memory. */

	body, err := io.ReadAll(response.Body)

	if err != nil {
		return nil, fmt.Errorf("readAll Error: %s", err)
	}

	if response.StatusCode != 200 {
		return nil, fmt.Errorf("invalid output (HTTP Code %d): %s", response.StatusCode, string(body))
	}

	var page Page

	err = json.Unmarshal(body, &page)

	if err != nil {
		return nil, fmt.Errorf("unmarshal error: %s", err)
	}

	switch page.Name {
	case "words":
		var words Words

		err = json.Unmarshal(body, &words)

		if err != nil {
			return nil, fmt.Errorf("unmarshal error: %s", err)
		}

		return words, nil
	case "occurrence":
		var occurrence Occurrence

		err = json.Unmarshal(body, &occurrence)

		if err != nil {
			return nil, fmt.Errorf("unmarshal error: %s", err)
		}

		return occurrence, nil
	}
	return nil, nil
}
