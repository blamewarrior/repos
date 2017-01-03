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

package main

import (
	// "encoding/json"

	"net/http"
	"net/http/httptest"

	"database/sql"
	"io/ioutil"
	"log"
	"os"
	"strings"

	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/blamewarrior/repos/blamewarrior"
	"github.com/blamewarrior/repos/blamewarrior/hooks"
)

func TestCreateRepositoryHandler(t *testing.T) {

	db, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/hooks/repositories", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusCreated)
	})

	log.SetOutput(ioutil.Discard)

	_, err := db.Exec("TRUNCATE repositories;")

	require.NoError(t, err)

	handlers := &Handlers{db}

	results := []struct {
		RequestBody  string
		ResponseCode int
		ResponseBody string
	}{
		{
			RequestBody:  `{"full_name":"blamewarrior/repos", "token":"test_token"}`,
			ResponseCode: http.StatusCreated,
			ResponseBody: "",
		},
		{
			RequestBody:  `{"full_name":"blamewarrior/repos"}`,
			ResponseCode: http.StatusUnprocessableEntity,
			ResponseBody: "",
		},
		{
			RequestBody:  `{"full_name":"blamewarrior&*()/repos", "token":"test_token"}`,
			ResponseCode: http.StatusInternalServerError,
			ResponseBody: "",
		},
	}

	for _, result := range results {
		req, err := http.NewRequest("POST", "/repositories", strings.NewReader(result.RequestBody))

		require.NoError(t, err)

		w := httptest.NewRecorder()

		handlers.CreateRepository(w, req)

		assert.Equal(t, result.ResponseCode, w.Code)
	}
}

func setup() (db *sql.DB, mux *http.ServeMux, teardownFn func()) {
	dbName := os.Getenv("DB_NAME")

	mux = http.NewServeMux()
	server := httptest.NewServer(mux)

	client := hooks.NewClient()
	client.BaseURL = server.URL

	if dbName == "" {
		log.Fatal("missing test database name (expected to be passed via ENV['DB_NAME'])")
	}

	opts := &blamewarrior.DatabaseOptions{
		Host:     os.Getenv("DB_HOST"),
		Port:     os.Getenv("DB_PORT"),
		User:     os.Getenv("DB_USER"),
		Password: os.Getenv("DB_PASSWORD"),
	}

	db, err := blamewarrior.ConnectDatabase(dbName, opts)
	if err != nil {
		log.Fatalf("failed to establish connection with test db %s using connection string %s: %s", dbName, opts.ConnectionString(), err)
	}

	return db, mux, func() {
		if err := db.Close(); err != nil {
			log.Printf("failed to close database connection: %s", err)
		}
	}
}
