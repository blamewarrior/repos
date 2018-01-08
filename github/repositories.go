/*
   Copyright (C) 2017 The BlameWarrior Authors.
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

package github

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/blamewarrior/repos/blamewarrior/tokens"

	bw "github.com/blamewarrior/repos/blamewarrior"

	gh "github.com/google/go-github/github"
)

var (
	ErrRateLimitReached = errors.New("GitHub API request rate limit reached")
	ErrNoSuchUser       = errors.New("no such user")
)

type Client interface {
	UserRepositories(ctx Context, username string) ([]bw.Repository, error)
}

type GithubClient struct {
	tokenClient tokens.Client
}

func NewGithubClient(tokenClient tokens.Client) *GithubClient {
	return &GithubClient{tokenClient}
}

func (c *GithubClient) UserRepositories(ctx Context, username string) (repos []bw.Repository, err error) {

	api, err := initAPIClient(ctx, c.tokenClient, username)
	if err != nil {
		return nil, err
	}

	opt := &gh.RepositoryListOptions{
		ListOptions: gh.ListOptions{PerPage: 100},
	}
	for {
		ghRepositories, resp, err := api.Repositories.List(ctx, "", opt)
		if err != nil {
			switch err.(type) {
			case *gh.RateLimitError:
				return nil, ErrRateLimitReached
			case *gh.ErrorResponse:
				apiErr := err.(*gh.ErrorResponse)
				if apiErr.Response.StatusCode == http.StatusNotFound {
					return nil, ErrNoSuchUser
				}
			}

			return nil, fmt.Errorf("request failed: %s", err)
		}

		for _, repo := range ghRepositories {
			repos = append(repos, bw.Repository{
				Owner:   *repo.Owner.Login,
				Name:    *repo.Name,
				Private: *repo.Private,
			})
		}

		if resp.NextPage == 0 {
			break
		}
		opt.Page = resp.NextPage
	}

	return repos, nil
}
