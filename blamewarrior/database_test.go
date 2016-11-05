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
	"testing"

	"github.com/blamewarrior/hooks/blamewarrior"
	"github.com/stretchr/testify/assert"
)

func TestDatabaseOptions_ConnectionString_FullConfig(t *testing.T) {
	examples := map[string]struct {
		Opts             blamewarrior.DatabaseOptions
		Present, Missing []string
	}{
		"empty": {
			Opts:    blamewarrior.DatabaseOptions{},
			Missing: []string{`user=`, `password=`, `host=`, `port=`},
		},
		"only user": {
			Opts:    blamewarrior.DatabaseOptions{User: "dba"},
			Present: []string{` user=dba`},
			Missing: []string{`password=`, `host=`, `port=`},
		},
		"only password": {
			Opts:    blamewarrior.DatabaseOptions{Password: "h4X0r"},
			Present: []string{` password="h4X0r"`},
			Missing: []string{`user=`, `host=`, `port=`},
		},
		"only host": {
			Opts:    blamewarrior.DatabaseOptions{Host: "db-1"},
			Present: []string{` host=db-1`},
			Missing: []string{`user=`, `password=`, `port=`},
		},
		"only port": {
			Opts:    blamewarrior.DatabaseOptions{Port: "1234"},
			Present: []string{` port=1234`},
			Missing: []string{`user=`, `password=`, `host=`},
		},
		"user and password": {
			Opts:    blamewarrior.DatabaseOptions{User: "dba", Password: "h4X0r"},
			Present: []string{` user=dba`, ` password="h4X0r"`},
			Missing: []string{`host=`, `port=`},
		},
		"user and host": {
			Opts:    blamewarrior.DatabaseOptions{User: "dba", Host: "db-1"},
			Present: []string{` user=dba`, ` host=db-1`},
			Missing: []string{`password=`, `port=`},
		},
		"user and port": {
			Opts:    blamewarrior.DatabaseOptions{User: "dba", Port: "1234"},
			Present: []string{` user=dba`, ` port=1234`},
			Missing: []string{`password=`, `host=`},
		},
		"password and host": {
			Opts:    blamewarrior.DatabaseOptions{Password: "h4X0r", Host: "db-1"},
			Present: []string{` password="h4X0r"`, ` host=db-1`},
			Missing: []string{`user=`, `port=`},
		},
		"password and port": {
			Opts:    blamewarrior.DatabaseOptions{Password: "h4X0r", Port: "1234"},
			Present: []string{` password="h4X0r"`, ` port=1234`},
			Missing: []string{`user=`, `host=`},
		},
		"missing user": {
			Opts:    blamewarrior.DatabaseOptions{Password: "h4X0r", Host: "db-1", Port: "1234"},
			Present: []string{` password="h4X0r"`, ` host=db-1`, ` port=1234`},
			Missing: []string{`user=`},
		},
		"missing password": {
			Opts:    blamewarrior.DatabaseOptions{User: "dba", Host: "db-1", Port: "1234"},
			Present: []string{` user=dba`, ` host=db-1`, ` port=1234`},
			Missing: []string{`password=`},
		},
		"missing host": {
			Opts:    blamewarrior.DatabaseOptions{User: "dba", Password: "h4X0r", Port: "1234"},
			Present: []string{` user=dba`, ` password="h4X0r"`, ` port=1234`},
			Missing: []string{`host=`},
		},
		"missing port": {
			Opts:    blamewarrior.DatabaseOptions{User: "dba", Password: "h4X0r", Host: "db-1"},
			Present: []string{` user=dba`, ` password="h4X0r"`, ` host=db-1`},
			Missing: []string{`port=`},
		},
		"all fields": {
			Opts:    blamewarrior.DatabaseOptions{User: "dba", Password: "h4X0r", Host: "db-1", Port: "1234"},
			Present: []string{` user=dba`, ` password="h4X0r"`, ` host=db-1`, ` port=1234`},
		},
	}

	for name, example := range examples {
		t.Run(name, func(t *testing.T) {
			connStr := example.Opts.ConnectionString()

			for _, part := range example.Present {
				assert.Contains(t, connStr, part)
			}

			for _, part := range example.Missing {
				assert.NotContains(t, connStr, part)
			}
		})
	}
}
