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

package tokens

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

type Client interface {
	GetToken(nickname string) (token string, err error)
}

type Response struct {
	Token string `json:"token"`
}

type TokenClient struct {
	BaseURL string
	c       *http.Client
}

func (client *TokenClient) GetToken(nickname string) (token string, err error) {

	resp, err := client.c.Get(client.BaseURL + "/users/" + nickname)

	if err != nil {
		return "", fmt.Errorf("impossible to get data for %s: %s", nickname, err)
	}

	b, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		return "", fmt.Errorf("cannot read response body when getting data for %s: %s", nickname, err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("got unsuccessful response for %s, status %d: %s", nickname, resp.StatusCode, string(b))
	}

	tokenResp := new(Response)

	err = json.Unmarshal(b, &tokenResp)

	if err != nil {
		return "", fmt.Errorf("cannot unmarshal responded json from users service: %s", err)
	}

	token = tokenResp.Token

	if token == "" {
		return "", fmt.Errorf("token for %s user cannot be empty", nickname)
	}

	return token, nil
}

func NewTokenClient() *TokenClient {
	client := &TokenClient{
		BaseURL: "https://blamewarrior.com",
		c:       http.DefaultClient,
	}

	return client
}
