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

package hooks

import (
	"bytes"
	"fmt"
	"net/http"
)

type Client interface {
	CreateHook(repositoryName string) error
	DeleteHook(repositoryName string) error
}

type HooksClient struct {
	BaseURL string
	c       *http.Client
}

func (client *HooksClient) CreateHook(repositoryName string) error {

	payload := []byte(fmt.Sprintf(`{"full_name":"%s"}`, repositoryName))

	response, err := client.c.Post(client.BaseURL+"/repositories", "application/json", bytes.NewBuffer(payload))

	if err != nil {
		return err
	}

	if response.StatusCode != http.StatusCreated {
		return fmt.Errorf("Impossible to create hook for %s", repositoryName)
	}

	return nil

}

func (client *HooksClient) DeleteHook(repositoryName string) error {
	url := client.BaseURL + "/repositories/" + repositoryName

	req, err := http.NewRequest("DELETE", url, nil)
	if err != nil {
		return err
	}

	response, err := client.c.Do(req)

	if err != nil {
		return err
	}

	if response.StatusCode != http.StatusNoContent {
		return fmt.Errorf("Impossible to delete hook for %s", repositoryName)
	}
	return nil
}

func NewHooksClient() *HooksClient {
	client := &HooksClient{
		BaseURL: "https://blamewarrior.com/hooks",
		c:       http.DefaultClient,
	}

	return client
}
