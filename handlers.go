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

func (h *Handlers) CreateRepository(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")

	body, err := requestBody(req)

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	repository := &blamewarrior.Repository{}

	if err := json.Unmarshal(body, &repository); err != nil {
		w.WriteHeader(http.StatusUnprocessableEntity)
		fmt.Fprintf(w, "Error when unmarshalling json")
		return
	}

	if err := repository.Validate(); err != nil {
		w.WriteHeader(http.StatusUnprocessableEntity)
		fmt.Fprintf(w, "Error when creating repository: %s", err)
		return
	}

	tx, err := h.db.Begin()

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if err := blamewarrior.CreateRepository(tx, repository); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Printf("%s\t%s\t%v\t%s", "POST", req.RequestURI, http.StatusInternalServerError, err)
		return
	}

	if err := h.client.CreateHook(repository.FullName); err != nil {
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
