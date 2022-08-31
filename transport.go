package main

import "net/http"

type MyJWTTransport struct {
	transport http.RoundTripper // this makes sure we'll use the DefaultTransport
	token     string
}

func (m MyJWTTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	/* In order to find this function's signature I went to
	client.Transport -> right click -> Go to Definition
	-> RoundTripper interface. */

	/* It seems that our struct is implementing the RoundTrip method
	but what it does its actually using the DefaultTransport's
	RoundTrip method */

	/* In this step for it to actually do something
	we'll make it inject the Header Bearer to the GET request */
	if m.token != "" {
		req.Header.Add("Authorization", "Bearer "+m.token)
	}
	return m.transport.RoundTrip(req)
}
