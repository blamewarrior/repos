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
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"

	"database/sql"
	"fmt"
	"log"
	"os"

	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/blamewarrior/repos/blamewarrior"
	"github.com/blamewarrior/repos/github"
)

type githubClientMock struct {
	mock.Mock
}

func (ghClientMock *githubClientMock) UserRepositories(ctx github.Context, username string) (repos []blamewarrior.Repository, err error) {
	args := ghClientMock.Called(ctx, username)
	return args.Get(0).([]blamewarrior.Repository), args.Error(1)

}

type hooksClientMock struct {
	mock.Mock
}

func (hooksClientMock *hooksClientMock) CreateHook(repositoryName string) error {
	args := hooksClientMock.Called(repositoryName)
	return args.Error(0)
}

func (hooksClientMock *hooksClientMock) DeleteHook(repositoryName string) error {
	args := hooksClientMock.Called(repositoryName)
	return args.Error(0)
}

func TestGetRepositoryByFullName(t *testing.T) {
	db, teardown := setup()
	defer teardown()

	_, err := db.Exec("TRUNCATE repositories;")

	require.NoError(t, err)

	repo := &blamewarrior.Repository{Owner: "blamewarrior", Name: "test", Private: true}
	err = blamewarrior.CreateRepository(db, repo)
	require.NoError(t, err)

	hooksClient := new(hooksClientMock)
	ghClient := new(githubClientMock)

	handlers := &Handlers{
		db:          db,
		hooksClient: hooksClient,
		ghClient:    ghClient,
	}

	results := []struct {
		Owner        string
		Name         string
		ResponseCode int
		ResponseBody string
	}{
		{
			Owner:        "",
			Name:         "",
			ResponseCode: http.StatusBadRequest,
			ResponseBody: "Incorrect full name\n",
		},
		{
			Owner:        "blamewarrior",
			Name:         "test",
			ResponseCode: http.StatusOK,
			ResponseBody: "{\"full_name\":\"blamewarrior/test\",\"owner\":\"blamewarrior\",\"name\":\"test\",\"private\":true}\n",
		},
	}

	for _, result := range results {
		req, err := http.NewRequest("POST", "/repositories?:owner="+result.Owner+"&:name="+result.Name, nil)

		require.NoError(t, err)

		w := httptest.NewRecorder()

		handlers.GetRepositoryByFullName(w, req)

		assert.Equal(t, result.ResponseCode, w.Code)
		assert.Equal(t, result.ResponseBody, fmt.Sprintf("%v", w.Body))
	}
}

func TestCreateRepositoryHandler(t *testing.T) {

	db, teardown := setup()

	_, err := db.Exec("TRUNCATE repositories;")

	require.NoError(t, err)

	defer teardown()

	hooksClient := new(hooksClientMock)
	hooksClient.On("CreateHook", "blamewarrior/test").Return(nil)

	ghClient := new(githubClientMock)

	handlers := &Handlers{
		db:          db,
		hooksClient: hooksClient,
		ghClient:    ghClient,
	}

	log.SetOutput(ioutil.Discard)

	results := []struct {
		RequestBody  string
		ResponseCode int
		ResponseBody string
	}{
		{
			RequestBody:  `{"owner":"blamewarrior", "name":"test"}`,
			ResponseCode: http.StatusCreated,
			ResponseBody: "",
		},
		{
			RequestBody:  `{"owner":"blamewarrior&*()", "name":"repos"}`,
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
		assert.Equal(t, result.ResponseBody, fmt.Sprintf("%v", w.Body))
	}

	hooksClient.AssertExpectations(t)
}

func TestDeleteRepositoryHandler(t *testing.T) {
	db, teardown := setup()

	_, err := db.Exec("TRUNCATE repositories;")
	require.NoError(t, err)

	defer teardown()

	repo := &blamewarrior.Repository{Owner: "blamewarrior", Name: "repos", Private: true}
	err = blamewarrior.CreateRepository(db, repo)

	require.NoError(t, err)

	hooksClient := new(hooksClientMock)
	hooksClient.On("DeleteHook", "blamewarrior/test_repo").Return(nil)
	ghClient := new(githubClientMock)

	handlers := &Handlers{
		db:          db,
		hooksClient: hooksClient,
		ghClient:    ghClient,
	}

	urlValues := make(url.Values)
	urlValues[":owner"] = []string{"blamewarrior"}
	urlValues[":name"] = []string{"test_repo"}

	req, err := http.NewRequest("DELETE", "/repositories?"+urlValues.Encode(), nil)

	require.NoError(t, err)

	w := httptest.NewRecorder()

	handlers.DeleteRepository(w, req)

	assert.Equal(t, http.StatusNoContent, w.Code)

	hooksClient.AssertExpectations(t)
}

func TestGetListRepositoryByOwner(t *testing.T) {
	db, teardown := setup()
	defer teardown()

	_, err := db.Exec("TRUNCATE repositories;")

	require.NoError(t, err)

	repo := &blamewarrior.Repository{Owner: "blamewarrior", Name: "test", Private: true}
	err = blamewarrior.CreateRepository(db, repo)
	require.NoError(t, err)

	hooksClient := new(hooksClientMock)
	ghClient := new(githubClientMock)

	ghClient.On("UserRepositories").Return([]blamewarrior.Repository{*repo})

	handlers := &Handlers{
		db:          db,
		hooksClient: hooksClient,
		ghClient:    ghClient,
	}

	results := []struct {
		Owner        string
		Name         string
		ResponseCode int
		ResponseBody string
	}{
		{
			Owner:        "",
			ResponseCode: http.StatusBadRequest,
			ResponseBody: "Incorrect owner\n",
		},
		{
			Owner:        "blamewarrior",
			ResponseCode: http.StatusOK,
			ResponseBody: "[{\"full_name\":\"blamewarrior/test\",\"owner\":\"blamewarrior\",\"name\":\"test\",\"private\":true}]\n",
		},
	}

	for _, result := range results {
		req, err := http.NewRequest("POST", "/repositories?:owner="+result.Owner, nil)

		require.NoError(t, err)

		w := httptest.NewRecorder()

		handlers.GetListRepositoryByOwner(w, req)

		assert.Equal(t, result.ResponseCode, w.Code)
		assert.Equal(t, result.ResponseBody, fmt.Sprintf("%v", w.Body))
	}
}

func setup() (db *sql.DB, teardownFn func()) {
	dbName := os.Getenv("DB_NAME")

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

	tx, err := db.Begin()

	if err != nil {
		log.Fatalf("failed to create transaction, %s", err)
	}

	return db, func() {
		tx.Rollback()

		if err := db.Close(); err != nil {
			log.Printf("failed to close database connection: %s", err)
		}
	}
}
