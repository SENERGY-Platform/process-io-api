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
	"io"
	"log"
	"net/http"
	"process-io-api/pkg/model"
	"runtime/debug"
	"time"
)

func (this *Client) Bulk(userid string, bulk model.BulkRequest) (outputs model.BulkResponse, err error) {
	token, err := this.auth.ExchangeUserToken(userid)
	if err != nil {
		return outputs, err
	}
	body, err := json.Marshal(bulk)
	if err != nil {
		return outputs, err
	}
	if this.debug {
		log.Println("DEBUG: bulk", userid, string(body))
	}
	client := http.Client{
		Timeout: 5 * time.Second,
	}
	req, err := http.NewRequest(
		"POST",
		this.apiUrl+"/bulk",
		bytes.NewBuffer(body),
	)
	if err != nil {
		debug.PrintStack()
		return outputs, err
	}
	req.Header.Set("Authorization", token.Jwt())
	req.Header.Set("X-UserId", userid)
	resp, err := client.Do(req)
	if err != nil {
		debug.PrintStack()
		return outputs, err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		debug.PrintStack()
		temp, _ := io.ReadAll(resp.Body)
		return outputs, fmt.Errorf("unexpected response: %v, %v", resp.StatusCode, string(temp))
	}

	err = json.NewDecoder(resp.Body).Decode(&outputs)
	return outputs, err
}
