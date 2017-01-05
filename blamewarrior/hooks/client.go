package hooks

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
)

type API interface {
	Do(req *http.Request) (resp *http.Response, err error)
	Get(url string) (resp *http.Response, err error)
	Post(url string, bodyType string, body io.Reader) (resp *http.Response, err error)
}

type Client struct {
	BaseURL string
	c       API
}

func (client *Client) CreateHook(repositoryName string) error {

	payload := []byte(fmt.Sprintf(`{"full_name":"%s"}`, repositoryName))

	response, err := client.c.Post(client.BaseURL+"/repositories", "application/json", bytes.NewBuffer(payload))

	if err != nil {
		return err
	}

	if response.StatusCode != http.StatusCreated {
		return fmt.Errorf("Impossible to create hook for %s", repositoryName)
	}

	return nil

}

func NewClient() *Client {
	client := &Client{
		BaseURL: "https://blamewarrior.com/hooks",
		c:       http.DefaultClient,
	}

	return client
}
