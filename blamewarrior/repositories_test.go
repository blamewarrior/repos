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

package blamewarrior_test

import (
	"database/sql"
	"errors"
	"log"
	"os"
	"testing"

	"github.com/blamewarrior/repos/blamewarrior"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRepositoryValidation(t *testing.T) {
	db, teardown := setup()
	defer teardown()

	_, err := db.Exec("TRUNCATE repositories;")

	require.NoError(t, err)

	results := []struct {
		Repo *blamewarrior.Repository
		Err  error
	}{
		{
			Repo: &blamewarrior.Repository{Owner: "blamewarrior", Name: "repos"},
			Err:  nil,
		},
		{
			Repo: &blamewarrior.Repository{Owner: "blamewarrior"},
			Err:  errors.New(`name must not be empty`),
		},
	}

	for _, result := range results {
		repo := result.Repo
		err = repo.Validate()
		assert.Equal(t, err, result.Err)
	}
}

func TestGetRepositoryByFullName(t *testing.T) {
	db, teardown := setup()
	defer teardown()

	_, err := db.Exec("TRUNCATE repositories;")

	_, err = db.Exec(blamewarrior.CreateRepositoryQuery, "blamewarrior", "repos", true)

	require.NoError(t, err)

	results, err := blamewarrior.GetRepositoryByFullName(db, "blamewarrior/repos")

	require.NoError(t, err)
	require.NotEmpty(t, results)
}

func TestGetListRepositoryByOwner(t *testing.T) {
	db, teardown := setup()
	defer teardown()

	_, err := db.Exec("TRUNCATE repositories;")

	_, err = db.Exec(blamewarrior.CreateRepositoryQuery, "blamewarrior", "repos", true)

	require.NoError(t, err)

	results, err := blamewarrior.GetListRepositoryByOwner(db, "blamewarrior")

	require.NoError(t, err)
	require.NotEmpty(t, results)
	assert.Equal(t, 1, len(results))
}

func TestCreateRepository(t *testing.T) {

	results := []struct {
		Repo *blamewarrior.Repository
		Err  error
	}{
		{
			Repo: &blamewarrior.Repository{Owner: "blamewarrior", Name: "repos", Private: true},
			Err:  nil,
		},
		{
			Repo: &blamewarrior.Repository{Owner: "blamewarrior&*()", Name: "repos", Private: true},
			Err:  errors.New(`failed to create repository: pq: new row for relation "repositories" violates check constraint "proper_owner"`),
		},
		{
			Repo: &blamewarrior.Repository{Owner: "blamewarrior", Name: "repos&*(", Private: true},
			Err:  errors.New(`failed to create repository: pq: new row for relation "repositories" violates check constraint "proper_name"`),
		},
	}

	for _, result := range results {
		db, teardown := setup()

		_, err := db.Exec("TRUNCATE repositories;")

		require.NoError(t, err)

		repo := result.Repo
		err = blamewarrior.CreateRepository(db, repo)
		assert.Equal(t, result.Err, err)

		teardown()
	}
}

func TestDeleteRepository(t *testing.T) {
	db, teardown := setup()
	defer teardown()

	_, err := db.Exec("TRUNCATE repositories;")

	require.NoError(t, err)

	repo := &blamewarrior.Repository{Owner: "blamewarrior", Name: "repos", Private: true}
	err = blamewarrior.CreateRepository(db, repo)
	require.NoError(t, err)

	err = blamewarrior.DeleteRepository(db, repo.FullName())

	require.NoError(t, err)
}

func setup() (tx *sql.Tx, teardownFn func()) {
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

	tx, err = db.Begin()

	if err != nil {
		log.Fatal("failed to create transaction, %s", err)
	}

	return tx, func() {
		tx.Rollback()
		if err := db.Close(); err != nil {
			log.Printf("failed to close database connection: %s", err)
		}
	}
}
