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
	"encoding/json"
	"fmt"
	"github.com/SENERGY-Platform/process-io-api/pkg/model"
	"io"
	"log"
	"net/http"
	"runtime/debug"
	"time"
)

func (this *Client) List(userid string, query model.VariablesQueryOptions) (result []model.VariableWithUnixTimestamp, err error) {
	token, err := this.auth.ExchangeUserToken(userid)
	if err != nil {
		return result, err
	}
	if this.debug {
		log.Printf("DEBUG: list %v %#v\n", userid, query)
	}
	client := http.Client{
		Timeout: 5 * time.Second,
	}
	req, err := http.NewRequest(
		"GET",
		this.apiUrl+"/variables?"+query.Encode(),
		nil,
	)
	if err != nil {
		debug.PrintStack()
		return result, err
	}
	req.Header.Set("Authorization", token.Jwt())
	req.Header.Set("X-UserId", userid)
	resp, err := client.Do(req)
	if err != nil {
		debug.PrintStack()
		return result, err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		debug.PrintStack()
		temp, _ := io.ReadAll(resp.Body)
		return result, fmt.Errorf("unexpected response: %v, %v", resp.StatusCode, string(temp))
	}

	err = json.NewDecoder(resp.Body).Decode(&result)
	return result, err
}

func (this *Client) Count(userid string, query model.VariablesQueryOptions) (result model.Count, err error) {
	token, err := this.auth.ExchangeUserToken(userid)
	if err != nil {
		return result, err
	}
	if this.debug {
		log.Printf("DEBUG: list %v %#v\n", userid, query)
	}
	client := http.Client{
		Timeout: 5 * time.Second,
	}
	req, err := http.NewRequest(
		"GET",
		this.apiUrl+"/count/variables?"+query.Encode(),
		nil,
	)
	if err != nil {
		debug.PrintStack()
		return result, err
	}
	req.Header.Set("Authorization", token.Jwt())
	req.Header.Set("X-UserId", userid)
	resp, err := client.Do(req)
	if err != nil {
		debug.PrintStack()
		return result, err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		debug.PrintStack()
		temp, _ := io.ReadAll(resp.Body)
		return result, fmt.Errorf("unexpected response: %v, %v", resp.StatusCode, string(temp))
	}

	err = json.NewDecoder(resp.Body).Decode(&result)
	return result, err
}
