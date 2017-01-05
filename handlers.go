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
)

type Handlers struct {
	client *hooks.Client
	db     *sql.DB
}

func (h *Handlers) GetRepositories(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")

	token := req.URL.Query().Get("token")

	if token == "" {
		w.WriteHeader(http.StatusUnprocessableEntity)
		fmt.Fprintf(w, "Token must not be empty")
		return
	}

	results, err := blamewarrior.GetRepositories(h.db, token)

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Printf("%s\t%s\t%v\t%s", "GET", req.RequestURI, http.StatusInternalServerError, err)
		return
	}

	if err := json.NewEncoder(w).Encode(results); err != nil {
		w.WriteHeader(http.StatusUnprocessableEntity)
		fmt.Fprintf(w, "Error when unmarshalling json")
		return
	}

	w.WriteHeader(http.StatusOK)

	return
}

func (h *Handlers) CreateRepository(w http.ResponseWriter, req *http.Request) {

	var err error

	w.Header().Set("Content-Type", "application/json; charset=UTF-8")

	body, err := requestBody(req)

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	repository := &blamewarrior.Repository{}

	if err = json.Unmarshal(body, &repository); err != nil {
		w.WriteHeader(http.StatusUnprocessableEntity)
		fmt.Fprintf(w, "Error when unmarshalling json")
		return
	}

	if err = repository.Validate(); err != nil {
		w.WriteHeader(http.StatusUnprocessableEntity)
		fmt.Fprintf(w, "Error when creating repository: %s", err)
		return
	}

	tx, err := h.db.Begin()

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if err = blamewarrior.CreateRepository(tx, repository); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Printf("%s\t%s\t%v\t%s", "POST", req.RequestURI, http.StatusInternalServerError, err)
		return
	}

	if err = h.client.CreateHook(repository.FullName); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Printf("%s\t%s\t%v\t%s", "POST", req.RequestURI, http.StatusInternalServerError, err)
		return
	}

	defer func() {
		if err != nil {
			tx.Rollback()
			return
		}
		err = tx.Commit()
	}()

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

	if err = h.client.DeleteHook(repositoryName); err != nil {
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
