/*
   Copyright (C) 2016 The BlameWarrior Authors.

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

		client := hooks.NewHooksClient()
		client.BaseURL = testAPIEndpoint

		err := client.CreateHook("blamewarrior/test_repo")

		assert.Equal(t, result.ResponseError, err)

		teardown()
	}
}

func TestDeleteHook(t *testing.T) {

	results := []struct {
		ResponseStatus int
		ResponseError  error
	}{
		{ResponseStatus: http.StatusNoContent, ResponseError: nil},
		{ResponseStatus: http.StatusNotFound, ResponseError: errors.New("Impossible to delete hook for blamewarrior/test_repo")},
	}

	for _, result := range results {
		testAPIEndpoint, mux, teardown := setup()

		mux.HandleFunc("/repositories/blamewarrior/test_repo", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(result.ResponseStatus)
		})

		client := hooks.NewHooksClient()
		client.BaseURL = testAPIEndpoint

		err := client.DeleteHook("blamewarrior/test_repo")

		assert.Equal(t, result.ResponseError, err)

		teardown()
	}
}

func setup() (baseURL string, mux *http.ServeMux, teardown func()) {
	mux = http.NewServeMux()
	server := httptest.NewServer(mux)

	return server.URL, mux, server.Close
}
