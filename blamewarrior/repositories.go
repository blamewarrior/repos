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
	"encoding/json"
	"fmt"
	"strings"

	_ "github.com/lib/pq"
)

type Repository struct {
	ID      int    `json:"-"`
	Owner   string `json:"owner"`
	Name    string `json:"name"`
	Private bool   `json:"private"`
}

func (repo *Repository) MarshalJSON() ([]byte, error) {
	type Alias Repository
	return json.Marshal(&struct {
		FullName string `json:"full_name"`
		*Alias
	}{
		FullName: repo.FullName(),
		Alias:    (*Alias)(repo),
	})
}

var IncorrectFullName = fmt.Errorf("incorrect full name for repository")

func (repo *Repository) FullName() string {
	return fmt.Sprintf("%s/%s", repo.Owner, repo.Name)
}

func (repo *Repository) Validate() error {
	if repo.Owner == "" {
		return fmt.Errorf("owner must not be empty")
	}

	if repo.Name == "" {
		return fmt.Errorf("name must not be empty")
	}

	return nil
}

func GetListRepositoryByOwner(runner SQLRunner, owner string) (repositories []Repository, err error) {
	rows, err := runner.Query(GetListRepositoryByOwnerQuery, owner)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var repo Repository

		if err := rows.Scan(&repo.Owner, &repo.Name, &repo.Private); err != nil {
			return nil, err
		}

		repositories = append(repositories, repo)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return repositories, nil
}

func GetRepositoryByFullName(runner SQLRunner, fullName string) (*Repository, error) {

	repo := &Repository{}

	owner, name, err := parseFullName(fullName)

	if err != nil {
		return nil, err
	}

	err = runner.QueryRow(GetRepositoryQuery, owner, name).Scan(&repo.Owner, &repo.Name, &repo.Private)

	if err != nil {
		return nil, fmt.Errorf("failed to fetch repositories: %s", err)
	}

	return repo, nil

}

func CreateRepository(runner SQLRunner, repo *Repository) (err error) {
	err = runner.QueryRow(CreateRepositoryQuery, repo.Owner, repo.Name, repo.Private).Scan(&repo.ID)

	if err != nil {
		return fmt.Errorf("failed to create repository: %s", err)
	}

	return err
}

func DeleteRepository(runner SQLRunner, fullName string) (err error) {
	owner, name, err := parseFullName(fullName)

	if err != nil {
		return err
	}

	_, err = runner.Exec(DeleteRepositoryQuery, owner, name)

	if err != nil {
		return fmt.Errorf("failed to delete repository: %s", err)
	}

	return err
}

func parseFullName(fullName string) (owner string, name string, err error) {
	parameters := strings.Split(fullName, "/")
	if len(parameters) != 2 {
		return "", "", IncorrectFullName
	}

	owner, name = parameters[0], parameters[1]

	if owner == "" || name == "" {
		return "", "", IncorrectFullName
	}

	return owner, name, nil

}

const (
	GetListRepositoryByOwnerQuery = `SELECT owner, name, private FROM repositories WHERE owner=$1`
	GetRepositoryQuery            = `SELECT owner, name, private FROM repositories WHERE owner=$1 AND name=$2`
	CreateRepositoryQuery         = `INSERT INTO repositories (owner, name, private) VALUES ($1, $2, $3) RETURNING id`
	DeleteRepositoryQuery         = `DELETE FROM repositories WHERE owner=$1 and name=$2`
)
