/*
 * Copyright (c) 2023 InfAI (CC SES)
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *    http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package client

import (
	"github.com/SENERGY-Platform/process-io-api/pkg/api/client/auth"
)

func New(apiUrl string, authEndpoint string, authClientId string, authClientSecret string, debug bool) (*Client, error) {
	a, err := auth.New(authEndpoint, authClientId, authClientSecret, nil)
	if err != nil {
		return nil, err
	}
	return NewWithAuth(apiUrl, a, debug), nil
}

func NewWithAuth(apiUrl string, a Auth, debug bool) *Client {
	return &Client{
		apiUrl: apiUrl,
		auth:   a,
		debug:  debug,
	}
}

type Client struct {
	apiUrl string
	debug  bool
	auth   Auth
}

type Auth interface {
	ExchangeUserToken(userid string) (token string, err error)
}

type Token interface {
	Jwt() string
	IsAdmin() bool
	GetUserId() string
}
