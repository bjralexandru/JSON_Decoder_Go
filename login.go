package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type LoginRequest struct {
	Password string `json:"password"`
}

type LoginResponse struct {
	Token string `json:"token"`
}

func doLoginRequest(requestURL, password string) (string, error) {
	loginRequest := LoginRequest{
		Password: password,
	}

	body, err := json.Marshal(loginRequest)
	// We must marshall the password in json format
	// Which will be sent to the server
	if err != nil {
		return "", fmt.Errorf("marhsal error: %s", err)
	}

	// Make a post request with the json body
	// Body is of type []bytes and must be converted
	// for the io.Reader's Read() function
	// Because bytes.Buffer implements the Read().
	response, err := http.Post(requestURL, "app/json", bytes.NewBuffer(body))

	// The following is copied from the get-request logic
	// and modified to suit the POST request

	if err != nil {
		// Log is used for a system error
		// We want to be able to see the entire error
		return "", fmt.Errorf("HTTP POST error: %s", err)
	}

	defer response.Body.Close()
	/* Make sure that the response Body is closed
	when the function finishes running.
	It can be bigger than our memory alloc, it is streamed
	on demand, but we want to be sure we've closed it to
	free the memory. */

	resBody, err := io.ReadAll(response.Body)

	if err != nil {
		return "", fmt.Errorf("readAll Error: %s", err)
	}

	if response.StatusCode != 200 {
		return "", fmt.Errorf("invalid output (HTTP Code %d): %s", response.StatusCode, string(resBody))
	}

	/* Did this beacuse the errors werent very nice for the
	   end user */
	if !json.Valid(resBody) {
		return "", RequestError{
			HTTPCode: response.StatusCode,
			Body:     string(resBody),
			Err:      "no valid JSON returned",
		}
	}

	var loginResponse LoginResponse

	err = json.Unmarshal(resBody, &loginResponse)

	if err != nil {
		return "", RequestError{
			HTTPCode: response.StatusCode,
			Body:     string(resBody),
			Err:      fmt.Sprintf("page unmarshall error: %v", err),
		}
	}

	if loginResponse.Token == "" {
		return "", RequestError{
			HTTPCode: response.StatusCode,
			Body:     string(resBody),
			Err:      "Empty token replied",
		}
	}
	// We need to pass this Token to our doRequest func()
	return loginResponse.Token, nil
}
