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
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/SENERGY-Platform/process-io-api/pkg/model"
	"io"
	"log"
	"net/http"
	"net/url"
	"runtime/debug"
	"time"
)

func (this *Client[TokenType]) Set(userid string, variable model.Variable) (err error) {
	token, err := this.auth.ExchangeUserToken(userid)
	if err != nil {
		return err
	}
	body, err := json.Marshal(variable)
	if err != nil {
		return err
	}
	if this.debug {
		log.Println("DEBUG: store", userid, variable.Key, string(body))
	}
	client := http.Client{
		Timeout: 5 * time.Second,
	}
	req, err := http.NewRequest(
		"PUT",
		this.apiUrl+"/variables/"+url.PathEscape(variable.Key),
		bytes.NewBuffer(body),
	)
	if err != nil {
		debug.PrintStack()
		return err
	}
	req.Header.Set("Authorization", token.Jwt())
	req.Header.Set("X-UserId", userid)
	resp, err := client.Do(req)
	if err != nil {
		debug.PrintStack()
		return err
	}
	defer resp.Body.Close()

	temp, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 300 {
		debug.PrintStack()
		return fmt.Errorf("unexpected response: %v, %v", resp.StatusCode, string(temp))
	}
	return err
}

func (this *Client[TokenType]) Get(userid string, key string) (value model.VariableWithUnixTimestamp, err error) {
	token, err := this.auth.ExchangeUserToken(userid)
	if err != nil {
		return value, err
	}
	if this.debug {
		log.Println("DEBUG: read", userid, key)
	}
	client := http.Client{
		Timeout: 5 * time.Second,
	}
	req, err := http.NewRequest(
		"GET",
		this.apiUrl+"/variables/"+url.PathEscape(key),
		nil,
	)
	if err != nil {
		debug.PrintStack()
		return value, err
	}
	req.Header.Set("Authorization", token.Jwt())
	req.Header.Set("X-UserId", userid)
	resp, err := client.Do(req)
	if err != nil {
		debug.PrintStack()
		return value, err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		debug.PrintStack()
		temp, _ := io.ReadAll(resp.Body)
		return value, fmt.Errorf("unexpected response: %v, %v", resp.StatusCode, string(temp))
	}

	err = json.NewDecoder(resp.Body).Decode(&value)
	return value, err
}
