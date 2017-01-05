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
	ID       int
	FullName string `json:"full_name"`
	Token    string `json:"token"`
	Private  bool   `json:"private"`
}

func (repo *Repository) Validate() error {
	if repo.FullName == "" {
		return fmt.Errorf("full name must not be empty")
	}

	if repo.Token == "" {
		return fmt.Errorf("token must not be empty")
	}
	return nil
}

func GetRepositories(db *sql.DB) (repos []Repository, err error) {

	rows, err := db.Query(GetRepositoriesQuery)

	if err != nil {
		return nil, fmt.Errorf("failed to fetch repositories: %s", err)
	}

	defer rows.Close()

	for rows.Next() {

		var repo Repository

		if err = rows.Scan(&repo.FullName, &repo.Token, &repo.Private); err != nil {
			return nil, fmt.Errorf("failed to fetch repository: %s", err)
		}

		repos = append(repos, repo)
	}

	return repos, rows.Err()

}

func CreateRepository(db Queryer, repo *Repository) (err error) {
	err = db.QueryRow(CreateRepositoryQuery, repo.FullName, repo.Token, repo.Private).Scan(&repo.ID)

	if err != nil {
		return fmt.Errorf("failed to create repository: %s", err)
	}

	return err
}

func DeleteRepository(db Queryer, repositoryID int) (err error) {
	_, err = db.Exec(DeleteRepositoryQuery, repositoryID)

	if err != nil {
		return fmt.Errorf("failed to delete repository: %s", err)
	}

	return err
}

const (
	GetRepositoriesQuery  = `SELECT full_name, token, private FROM repositories`
	CreateRepositoryQuery = `INSERT INTO repositories (full_name, token, private) VALUES ($1, $2, $3) RETURNING id`
	UpdateRepositoryQuery = `UPDATE repositories SET token = $2, private = $3 WHERE id = $1 RETURNING id;`
	DeleteRepositoryQuery = `DELETE FROM repositories WHERE id = $1`
)
