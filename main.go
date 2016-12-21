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
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/bmizerany/pat"

	"github.com/blamewarrior/repos/blamewarrior"
)

func CreateRepositoryHandler(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")

	body, err := requestBody(req)

	if err != nil {
		log.Printf("%s\t%s\t%s\t%s", "POST", req.RequestURI, http.StatusInternalServerError, err)
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(w, "500: Internal server error")
	}

	repository := &blamewarrior.Repository{}

	if err := json.Unmarshal(body, &repository); err != nil {

		fmt.Fprintf(w, "Error: %s", err)
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

func main() {
	mux := pat.New()
	mux.Post("/repositories", http.HandlerFunc(CreateRepositoryHandler))

	http.Handle("/", mux)

	log.Printf("blamewarrior repositories is running on 8080 port")

	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Panic(err)
	}

}
