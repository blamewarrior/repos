/*
   Copyright (C) 2017 The BlameWarrior Authors.
   This file is a part of BlameWarrior service.
   This program is free software: you can redistribute it and/or modify
   it under the terms of the GNU General Public License as published by
   the Free Software Foundation, either version 3 of the License, or
   (at your option) any later version.
   This program is distributed in the hope that it will be useful,
   but WITHOUT ANY WARRANTY; without even the implied warranty of
   MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
   GNU General Public License for more details.
   You should have received a copy of the GNU General Public License
   along with this program.  If not, see <http://www.gnu.org/licenses/>.
*/

package tokens_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/blamewarrior/repos/blamewarrior/tokens"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetToken(t *testing.T) {
	testAPIEndpoint, mux, teardown := setup()

	defer teardown()

	mux.HandleFunc("/users/blamewarrior", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(userResponse))
	})

	client := tokens.NewTokenClient()
	client.BaseURL = testAPIEndpoint

	token, err := client.GetToken("blamewarrior")

	require.NoError(t, err)

	assert.Equal(t, "test_token", token)

}

func setup() (baseURL string, mux *http.ServeMux, teardown func()) {
	mux = http.NewServeMux()
	server := httptest.NewServer(mux)

	return server.URL, mux, server.Close
}

const userResponse = `{"token": "test_token"}`
