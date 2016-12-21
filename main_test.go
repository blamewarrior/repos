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
	// "encoding/json"

	"strings"

	"net/http"
	"net/http/httptest"

	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreateRepositoryHandler(t *testing.T) {

	req, err := http.NewRequest("POST", "/repositories", strings.NewReader(`{"full_name":"blamewarrior/repos", "token":"test_token"}`))

	require.NoError(t, err)

	w := httptest.NewRecorder()

	CreateRepositoryHandler(w, req)

	assert.Equal(t, w.Code, http.StatusOK)
}
