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

package auth

import (
	"encoding/json"
	"errors"
	"github.com/SENERGY-Platform/service-commons/pkg/cache"
	"io"
	"log"
	"net/http"
	"net/url"
	"time"
)

func (this *Auth) ExchangeUserToken(userid string) (token string, err error) {
	return cache.UseWithExpInGet(this.cache, "user-token."+userid, func() (string, time.Duration, error) {
		return this.exchangeUserToken(userid)
	}, func(s string) error {
		if s == "" {
			return errors.New("invalid token received from cache")
		}
		return nil
	}, time.Hour)
}

func (this *Auth) exchangeUserToken(userid string) (token string, expiration time.Duration, err error) {
	resp, err := http.PostForm(this.endpoint+"/auth/realms/master/protocol/openid-connect/token", url.Values{
		"client_id":         {this.clientId},
		"client_secret":     {this.clientSecret},
		"grant_type":        {"urn:ietf:params:oauth:grant-type:token-exchange"},
		"requested_subject": {userid},
	})
	if err != nil {
		return
	}
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		log.Println("ERROR: GetUserToken()", resp.StatusCode, string(body))
		err = errors.New("access denied")
		resp.Body.Close()
		return
	}
	var openIdToken OpenidToken
	err = json.NewDecoder(resp.Body).Decode(&openIdToken)
	if err != nil {
		return
	}
	return "Bearer " + openIdToken.AccessToken, time.Duration(openIdToken.ExpiresIn-5) * time.Second, err // subtract 5 seconds from expiration as a buffer
}

type OpenidToken struct {
	AccessToken      string    `json:"access_token"`
	ExpiresIn        float64   `json:"expires_in"`
	RefreshExpiresIn float64   `json:"refresh_expires_in"`
	RefreshToken     string    `json:"refresh_token"`
	TokenType        string    `json:"token_type"`
	RequestTime      time.Time `json:"-"`
}
