package client

import "net/http"

// Client interface is used for dependency injection pourposes
// and avoiding package dependencies
type Client interface {
	Get(url string) (*http.Response, error)
}

type client struct{}

// New creates a client interface
func New() Client {
	return &client{}
}

func (c *client) Get(u string) (*http.Response, error) {
	return http.Get(u)
}
