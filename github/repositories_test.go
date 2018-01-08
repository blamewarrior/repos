package github_test

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/blamewarrior/repos/github"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	bw "github.com/blamewarrior/repos/blamewarrior"
)

type tokenServiceMock struct {
	mock.Mock
}

func (tsMock *tokenServiceMock) GetToken(username string) (string, error) {
	args := tsMock.Called(username)
	return args.String(0), args.Error(1)

}

func TestGithubService_UserRepositories(t *testing.T) {
	baseURL, mux, teardown := setup()
	defer teardown()

	ts := new(tokenServiceMock)

	ts.On("GetToken", "user1").Return("test-token", nil)

	c := github.NewGithubClient(ts)

	mux.HandleFunc("/user/repos", func(w http.ResponseWriter, req *http.Request) {
		url := baseURL.String() + "/" + req.URL.Path
		w.Header().Set("Link", `<`+url+`?page=2>; rel="last"`)

		assert.Equal(t, "Bearer test-token", req.Header.Get("Authorization"))

		if req.FormValue("page") != "2" {
			w.Header().Set("Link", `<`+url+`?page=2>; rel="next", `+w.Header().Get("Link"))
			w.Write([]byte(`[{"name":"repo1","private":false,"owner":{"login":"user1"}},{"name":"repo2","private":true,"owner":{"login":"user1"}}]`))
		} else {
			w.Write([]byte(`[{"name":"repo3","private":true,"owner":{"login":"user1"}}]`))
		}
	})

	ctx := github.Context{context.Background(), baseURL}

	repositories, err := c.UserRepositories(ctx, "user1")
	require.NoError(t, err)
	assert.Len(t, repositories, 3)

	assert.Contains(t, repositories, bw.Repository{Name: "repo1", Private: false, Owner: "user1"})
	assert.Contains(t, repositories, bw.Repository{Name: "repo2", Private: true, Owner: "user1"})
	assert.Contains(t, repositories, bw.Repository{Name: "repo3", Private: true, Owner: "user1"})
}

func setup() (baseURL *url.URL, mux *http.ServeMux, teardownFn func()) {
	mux = http.NewServeMux()
	srv := httptest.NewServer(mux)
	baseURL, _ = url.Parse(fmt.Sprintf("%s/", srv.URL))

	return baseURL, mux, srv.Close
}
