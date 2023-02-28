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

import "github.com/SENERGY-Platform/process-io-api/pkg/api/client/auth"

func New(apiUrl string, authEndpoint string, authClientId string, authClientSecret string, debug bool) *Client[auth.Token] {
	return NewWithAuth(apiUrl, auth.New(authEndpoint, authClientId, authClientSecret, nil), debug)
}

func NewWithAuth(apiUrl string, a Auth[auth.Token], debug bool) *Client[auth.Token] {
	return NewWithGenericAuth[auth.Token](apiUrl, a, debug)
}

func NewWithGenericAuth[TokenType Token](apiUrl string, auth Auth[TokenType], debug bool) *Client[TokenType] {
	return &Client[TokenType]{
		apiUrl: apiUrl,
		auth:   auth,
		debug:  debug,
	}
}

type Client[TokenType Token] struct {
	apiUrl string
	debug  bool
	auth   Auth[TokenType]
}

type Auth[TokenType Token] interface {
	ExchangeUserToken(userid string) (token TokenType, err error)
}

type Token interface {
	Jwt() string
	IsAdmin() bool
	GetUserId() string
}
