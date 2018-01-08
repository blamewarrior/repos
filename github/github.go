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
	"context"
	"fmt"
	"net/url"
	"strings"

	"github.com/blamewarrior/repos/blamewarrior/tokens"
	"golang.org/x/oauth2"

	gh "github.com/google/go-github/github"
)

type Context struct {
	context.Context
	// BaseURL overrides GitHub API endpoint and is intended for use in tests.
	BaseURL *url.URL
}

// SplitRepositoryName splits full GitHub repository name into owner and name parts.
func SplitRepositoryName(fullName string) (owner, repo string) {
	sep := strings.IndexByte(fullName, '/')
	if sep <= 0 || sep == len(fullName)-1 {
		return "", ""
	}

	return fullName[0:sep], fullName[sep+1:]
}

func initAPIClient(ctx Context, tokenClient tokens.Client, owner string) (*gh.Client, error) {

	token, err := tokenClient.GetToken(owner)

	if err != nil {
		return nil, fmt.Errorf("unable to get token to init API client: %s", err)
	}

	tokenSource := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: token})
	oauthClient := oauth2.NewClient(ctx, tokenSource)

	api := gh.NewClient(oauthClient)
	if ctx.BaseURL != nil {
		api.BaseURL = ctx.BaseURL
	}

	return api, nil

}
