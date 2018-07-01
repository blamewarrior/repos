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
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/blamewarrior/repos/blamewarrior"
	"github.com/blamewarrior/repos/blamewarrior/hooks"

	"github.com/blamewarrior/repos/github"
)

type Handlers struct {
	ghClient    github.Client
	hooksClient hooks.Client
	db          *sql.DB
}

func (h *Handlers) GetRepositoryByFullName(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")

	owner := req.URL.Query().Get(":owner")

	name := req.URL.Query().Get(":name")

	fullName := fmt.Sprintf("%s/%s", owner, name)

	results, err := blamewarrior.GetRepositoryByFullName(h.db, fullName)

	if err != nil {

		if err == blamewarrior.IncorrectFullName {
			http.Error(w, "Incorrect full name", http.StatusBadRequest)
		} else {
			w.WriteHeader(http.StatusInternalServerError)
			log.Printf("%s\t%s\t%v\t%s", "GET", req.RequestURI, http.StatusInternalServerError, err)
		}
		return
	}

	if err := json.NewEncoder(w).Encode(results); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "Error when unmarshalling json")
		return
	}

	w.WriteHeader(http.StatusOK)

	return
}

func (h *Handlers) GetListRepositoryByOwner(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")

	owner := req.URL.Query().Get(":owner")

	if owner == "" {
		http.Error(w, "Incorrect owner", http.StatusBadRequest)
		return
	}

	results, err := blamewarrior.GetListRepositoryByOwner(h.db, owner)

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Printf("%s\t%s\t%v\t%s", "GET", req.RequestURI, http.StatusInternalServerError, err)
		return

	}

	if err := json.NewEncoder(w).Encode(results); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "Error when unmarshalling json")
		return
	}

	w.WriteHeader(http.StatusOK)

	return
}

func (h *Handlers) CreateRepository(w http.ResponseWriter, req *http.Request) {
	var err error
	var body []byte

	repository := &blamewarrior.Repository{}

	w.Header().Set("Content-Type", "application/json; charset=UTF-8")

	if body, err = requestBody(req); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if err = json.Unmarshal(body, &repository); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "Error when unmarshalling json")
		return
	}

	if err = repository.Validate(); err != nil {
		http.Error(w, fmt.Sprintf("Error when creating repository: %s", err), http.StatusUnprocessableEntity)
		return
	}

	tx, err := h.db.Begin()

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	defer tx.Rollback()

	if err = blamewarrior.CreateRepository(tx, repository); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Printf("%s\t%s\t%v\t%s", "POST", req.RequestURI, http.StatusInternalServerError, err)
		return
	}

	if err = h.hooksClient.CreateHook(repository.FullName()); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Printf("%s\t%s\t%v\t%s", "POST", req.RequestURI, http.StatusInternalServerError, err)
		return
	}

	if err = tx.Commit(); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Printf("%s\t%s\t%v\t%s", "POST", req.RequestURI, http.StatusInternalServerError, err)
		return
	}

	w.WriteHeader(http.StatusCreated)
	return

}

func (h Handlers) DeleteRepository(w http.ResponseWriter, req *http.Request) {
	var err error

	owner := req.URL.Query().Get(":owner")
	name := req.URL.Query().Get(":name")

	repositoryName := owner + "/" + name

	tx, err := h.db.Begin()

	if err = blamewarrior.DeleteRepository(tx, repositoryName); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Printf("%s\t%s\t%v\t%s", "DELETE", req.RequestURI, http.StatusInternalServerError, err)
		return
	}

	if err = h.hooksClient.DeleteHook(repositoryName); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Printf("%s\t%s\t%v\t%s", "DELETE", req.RequestURI, http.StatusInternalServerError, err)
		return
	}

	defer func() {
		if err != nil {
			tx.Rollback()
			return
		}
		err = tx.Commit()
	}()

	w.WriteHeader(http.StatusNoContent)
	return

}

func (h *Handlers) GetListGithubRepositories(w http.ResponseWriter, req *http.Request) {
	ctx := github.Context{}

	owner := req.URL.Query().Get(":owner")
	repositories, err := h.ghClient.UserRepositories(ctx, owner)

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Printf("%s\t%s\t%v\t%s", "GET", req.RequestURI, http.StatusInternalServerError, err)
		return
	}

	if err := json.NewEncoder(w).Encode(repositories); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Printf("%s\t%s\t%v\t%s", "GET", req.RequestURI, http.StatusInternalServerError, err)
		return
	}
}

func requestBody(r *http.Request) ([]byte, error) {
	body, err := ioutil.ReadAll(io.LimitReader(r.Body, 1048576))

	if err != nil {
		return nil, err
	}
	if err := r.Body.Close(); err != nil {
		return nil, err
	}
	return body, err
}
