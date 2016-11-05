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

	_ "github.com/lib/pq"
)

type Repository struct {
	Id       int
	FullName string
	Private  bool
}

func GetRepositories(db *sql.DB) (repos []Repository, err error) {

	rows, err := db.Query(GetRepositoriesQuery)

	if err != nil {
		return nil, fmt.Errorf("failed to fetch repositories: %s", err)
	}

	defer rows.Close()

	for rows.Next() {

		var repo Repository

		if err = rows.Scan(&repo.FullName, &repo.Private); err != nil {
			return nil, fmt.Errorf("failed to fetch repository: %s", err)
		}

		repos = append(repos, repo)
	}

	return repos, rows.Err()

}

func CreateRepository(db *sql.DB, repo *Repository) (err error) {
	err = db.QueryRow(CreateRepositoryQuery, repo.FullName, repo.Private).Scan(&repo.Id)

	if err != nil {
		return fmt.Errorf("failed to create repository: %s", err)
	}

	return err
}

func DeleteRepository(db *sql.DB, repositoryId int) (err error) {
	_, err = db.Exec(DeleteRepositoryQuery, repositoryId)

	if err != nil {
		return fmt.Errorf("failed to delete repository: %s", err)
	}

	return err
}

const (
	GetRepositoriesQuery  = `SELECT full_name, private FROM repositories`
	CreateRepositoryQuery = `INSERT INTO repositories (full_name, private) VALUES ($1, $2) RETURNING id`
	DeleteRepositoryQuery = `DELETE FROM repositories WHERE id = $1`
)
