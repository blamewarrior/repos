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
	"log"
	"net/http"

	"github.com/bmizerany/pat"
)

func TrackRepositoryHandler(w http.ResponseWriter, req *http.Request) {
	fmt.Fprint(w, "OK")
}

func main() {
	mux := pat.New()
	mux.Get("/repositories/track", http.HandlerFunc(TrackRepositoryHandler))

	http.Handle("/", mux)

	log.Printf("blamewarrior hooks is running on 8080 port")

	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Panic(err)
	}

}
