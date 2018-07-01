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
	"log"
	"net/http"
	"os"

	"github.com/blamewarrior/repos/github"
	"github.com/bmizerany/pat"

	"github.com/blamewarrior/repos/blamewarrior"
	"github.com/blamewarrior/repos/blamewarrior/hooks"
	"github.com/blamewarrior/repos/blamewarrior/tokens"
)

func main() {

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

	hooksBaseURL := os.Getenv("BW_HOOKS_BASE_URL")
	if hooksBaseURL == "" {
		log.Fatal("missing hooks base url (expected to be passed via ENV['BW_HOOKS_BASE_URL'])")
	}

	tokensBaseURL := os.Getenv("BW_TOKENS_BASE_URL")
	if tokensBaseURL == "" {
		log.Fatal("missing tokens base url (expected to be passed via ENV['BW_HOOKS_BASE_URL'])")
	}

	db, err := blamewarrior.ConnectDatabase(dbName, opts)
	if err != nil {
		log.Fatalf("failed to establish connection with test db %s using connection string %s: %s", dbName, opts.ConnectionString(), err)
	}

	tokenClient := tokens.NewTokenClient(tokensBaseURL)
	ghClient := github.NewGithubClient(tokenClient)

	hooksclient := hooks.NewHooksClient(hooksBaseURL)

	handlers := &Handlers{
		db:          db,
		hooksClient: hooksclient,
		ghClient:    ghClient,
	}

	mux := pat.New()

	mux.Get("/repositories/:owner/:name", http.HandlerFunc(handlers.GetRepositoryByFullName))
	mux.Get("/repositories/:owner", http.HandlerFunc(handlers.GetListRepositoryByOwner))
	mux.Get("/repositories/:owner/github", http.HandlerFunc(handlers.GetListGithubRepositories))
	mux.Post("/repositories", http.HandlerFunc(handlers.CreateRepository))
	mux.Del("/repositories/:owner/:name", http.HandlerFunc(handlers.DeleteRepository))

	http.Handle("/", mux)

	log.Printf("blamewarrior repositories is running on 8080 port")

	err = http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Panic(err)
	}

}
