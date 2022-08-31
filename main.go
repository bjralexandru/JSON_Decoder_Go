package main

import (
	"encoding/json"
	"flag"
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
	var (
		requestURL string
		password   string
		parsedURL  *url.URL // Used in doRequest() as a String()
		err        error
	)

	/* Instead of using CLI arguments we're using flags
	Which are implemented through the flag package */

	// Declaration
	flag.StringVar(&requestURL, "url", "", "url to access")
	flag.StringVar(&password, "password", "", "password for the api endpoint")

	// Parse flags
	flag.Parse()

	// Check URL Validity
	if parsedURL, err = url.ParseRequestURI(requestURL); err != nil {
		fmt.Printf("URL is not valid error: %s\n", err)
		flag.Usage()
		os.Exit(1)
	}

	// Check password not empty and  try to make
	// POST request.

	if password != "" {
		token, err := doLoginRequest(parsedURL.Scheme+"://"+parsedURL.Host+"/login", password)
		if err != nil {
			if requestErr, ok := err.(RequestError); ok {
				fmt.Printf("Error: %s (HTTP Code %d, Body: %s\n", requestErr.Err, requestErr.HTTPCode, requestErr.Body)
				os.Exit(1)
			}
			fmt.Printf("error: %s\n", err)
			os.Exit(1)
		}
		fmt.Printf("token: %s", token)
		os.Exit(0)
	}

	res, err := doRequests(parsedURL.String())

	if err != nil {
		if requestErr, ok := err.(RequestError); ok {
			fmt.Printf("Error occurred: %s (HTTP Error: %d, Body: %s)\n", requestErr.Error(), requestErr.HTTPCode, requestErr.Body)
			os.Exit(1)
		}
		fmt.Printf("Error occurred: %s\n", err)
		os.Exit(1)
	}
	if res == nil {
		fmt.Printf("No response\n")
		os.Exit(1)
	}
	fmt.Printf("Response: %s\n", res.GetResponse())
}

func doRequests(requestURL string) (Response, error) {

	/* GET REQUEST LOGIC */
	response, err := http.Get(requestURL)
	/* To this http.Get we must provide a header
	an Authorization header.
	In order to achieve this we will create an
	individual client for each session (token generated)
	through which the connection is done. */

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

	/* Did this beacuse the errors werent very nice for the
	   end user */
	if !json.Valid(body) {
		return nil, RequestError{
			HTTPCode: response.StatusCode,
			Body:     string(body),
			Err:      "no valid JSON returned",
		}
	}

	var page Page

	err = json.Unmarshal(body, &page)

	if err != nil {
		return nil, RequestError{
			HTTPCode: response.StatusCode,
			Body:     string(body),
			Err:      fmt.Sprintf("page unmarshall error: %v", err),
		}
	}

	switch page.Name {
	case "words":
		var words Words

		err = json.Unmarshal(body, &words)

		if err != nil {
			return nil, RequestError{
				HTTPCode: response.StatusCode,
				Body:     string(body),
				Err:      fmt.Sprintf(" words unmarshall error: %v", err),
			}
		}

		return words, nil

	case "occurrence":
		var occurrence Occurrence

		err = json.Unmarshal(body, &occurrence)

		if err != nil {
			return nil, RequestError{
				HTTPCode: response.StatusCode,
				Body:     string(body),
				Err:      fmt.Sprintf(" occurances unmarshall error: %v", err),
			}
		}

		return occurrence, nil
	}
	return nil, nil
}
