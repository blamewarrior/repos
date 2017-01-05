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
			Repo: &blamewarrior.Repository{FullName: "blamewarrior/repos", Token: "test_token"},
			Err:  nil,
		},
		{
			Repo: &blamewarrior.Repository{FullName: "blamewarrior/repos"},
			Err:  errors.New(`token must not be empty`),
		},
		{
			Repo: &blamewarrior.Repository{Token: "test_token"},
			Err:  errors.New(`full name must not be empty`),
		},
	}

	for _, result := range results {
		repo := result.Repo
		err = repo.Validate()
		assert.Equal(t, err, result.Err)
	}
}

func TestGetRepositories(t *testing.T) {
	db, teardown := setup()
	defer teardown()

	_, err := db.Exec("TRUNCATE repositories;")

	_, err = db.Exec(blamewarrior.CreateRepositoryQuery, "blamewarrior/hooks", "test_token", true)

	require.NoError(t, err)

	results, err := blamewarrior.GetRepositories(db)

	require.NoError(t, err)
	require.NotEmpty(t, results)
}

func TestCreateRepository(t *testing.T) {
	db, teardown := setup()
	defer teardown()

	_, err := db.Exec("TRUNCATE repositories;")

	require.NoError(t, err)

	results := []struct {
		Repo *blamewarrior.Repository
		Err  error
	}{
		{
			Repo: &blamewarrior.Repository{FullName: "blamewarrior/repos", Token: "test_token", Private: true},
			Err:  nil,
		},
		{
			Repo: &blamewarrior.Repository{FullName: "blamewarrior&*()/repos", Token: "test_token", Private: true},
			Err:  errors.New(`failed to create repository: pq: new row for relation "repositories" violates check constraint "proper_full_name"`),
		},
	}

	for _, result := range results {
		repo := result.Repo
		err = blamewarrior.CreateRepository(db, repo)
		assert.Equal(t, err, result.Err)
	}
}

func TestDeleteRepository(t *testing.T) {
	db, teardown := setup()
	defer teardown()

	_, err := db.Exec("TRUNCATE repositories;")

	require.NoError(t, err)

	repo := &blamewarrior.Repository{FullName: "blamewarrior/repos", Token: "test_token", Private: true}
	err = blamewarrior.CreateRepository(db, repo)
	require.NoError(t, err)

	err = blamewarrior.DeleteRepository(db, repo.ID)

	require.NoError(t, err)
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

	return db, func() {
		if err := db.Close(); err != nil {
			log.Printf("failed to close database connection: %s", err)
		}
	}
}

func checkRepositoryConsistentWithDB(t *testing.T, repo *blamewarrior.Repository, db *sql.DB) error {
	var (
		token   string
		private bool
	)

	require.NoError(t, db.QueryRow(
		`SELECT token, private FROM repositories WHERE id = $1 LIMIT 1;`,
		repo.ID,
	).Scan(
		&token,
		&private,
	))

	assert.Equal(t, token, repo.Token)

	if private {
		assert.True(t, repo.Private)
	} else {
		assert.False(t, repo.Private)
	}

	return nil
}
