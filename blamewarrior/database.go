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

package blamewarrior

import (
	"database/sql"
	"fmt"
	"strconv"

	_ "github.com/lib/pq"
)

// DatabaseOptions is a configuration object type to pass PostgreSQL connection options.
type DatabaseOptions struct {
	// Host is PostgreSQL database host
	Host string
	// Port is PostgreSQL database port
	Port string
	// User is connection user
	User string
	// Password is connection password to authenticate with
	Password string
}

// ConnectionString returns a connection string suitable to for sql.DB().
func (opts *DatabaseOptions) ConnectionString() string {
	if opts == nil {
		return ""
	}

	var connStr string
	if opts.User != "" {
		connStr += " user=" + opts.User
	}

	if opts.Password != "" {
		connStr += " password=" + strconv.Quote(opts.Password)
	}

	if opts.Host != "" {
		connStr += " host=" + opts.Host
	}

	if opts.Port != "" {
		connStr += " port=" + opts.Port
	}

	return connStr
}

func ConnectDatabase(dbName string, opts ...*DatabaseOptions) (*sql.DB, error) {
	connStr := "sslmode=disable dbname=" + dbName

	if len(opts) > 0 && opts[0] != nil {
		connStr += " " + opts[0].ConnectionString()
	}

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to db: %s", err)
	}

	return db, db.Ping()
}
