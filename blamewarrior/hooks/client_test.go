package hooks_test

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/blamewarrior/repos/blamewarrior/hooks"
	"github.com/stretchr/testify/assert"
)

func TestCreateHook(t *testing.T) {

	results := []struct {
		ResponseStatus int
		ResponseError  error
	}{
		{ResponseStatus: http.StatusCreated, ResponseError: nil},
		{ResponseStatus: http.StatusNotFound, ResponseError: errors.New("Impossible to create hook for blamewarrior/test_repo")},
	}

	for _, result := range results {
		testAPIEndpoint, mux, teardown := setup()

		mux.HandleFunc("/repositories", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(result.ResponseStatus)
		})

		client := hooks.NewClient()
		client.BaseURL = testAPIEndpoint

		err := client.CreateHook("blamewarrior/test_repo")

		assert.Equal(t, result.ResponseError, err)

		teardown()
	}
}

func setup() (baseURL string, mux *http.ServeMux, teardown func()) {
	mux = http.NewServeMux()
	server := httptest.NewServer(mux)

	return server.URL, mux, server.Close
}
