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
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"runtime/debug"
	"time"
)

func (this *Client) Delete(userid string, key string) error {
	token, err := this.auth.ExchangeUserToken(userid)
	if err != nil {
		return err
	}
	if this.debug {
		log.Println("DEBUG: delete", userid, key)
	}
	client := http.Client{
		Timeout: 5 * time.Second,
	}
	req, err := http.NewRequest(
		"DELETE",
		this.apiUrl+"/variables/"+url.PathEscape(key),
		nil,
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

	if resp.StatusCode >= 300 {
		debug.PrintStack()
		temp, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("unexpected response: %v, %v", resp.StatusCode, string(temp))
	}
	return nil
}

func (this *Client) DeleteProcessDefinition(userid string, definitionId string) error {
	token, err := this.auth.ExchangeUserToken(userid)
	if err != nil {
		return err
	}
	if this.debug {
		log.Println("DEBUG: delete process-definition", userid, definitionId)
	}
	client := http.Client{
		Timeout: 5 * time.Second,
	}
	req, err := http.NewRequest(
		"DELETE",
		this.apiUrl+"/process-definitions/"+url.PathEscape(definitionId),
		nil,
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

	if resp.StatusCode >= 300 {
		debug.PrintStack()
		temp, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("unexpected response: %v, %v", resp.StatusCode, string(temp))
	}
	return nil
}

func (this *Client) DeleteProcessInstance(userid string, instanceId string) error {
	token, err := this.auth.ExchangeUserToken(userid)
	if err != nil {
		return err
	}
	if this.debug {
		log.Println("DEBUG: delete process-instance", userid, instanceId)
	}
	client := http.Client{
		Timeout: 5 * time.Second,
	}
	req, err := http.NewRequest(
		"DELETE",
		this.apiUrl+"/process-instances/"+url.PathEscape(instanceId),
		nil,
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

	if resp.StatusCode >= 300 {
		debug.PrintStack()
		temp, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("unexpected response: %v, %v", resp.StatusCode, string(temp))
	}
	return nil
}
