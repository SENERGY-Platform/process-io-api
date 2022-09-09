/*
 * Copyright (c) 2022 InfAI (CC SES)
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

package tests

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"process-io-api/pkg/configuration"
	"process-io-api/pkg/model"
	"reflect"
	"sync"
	"testing"
	"time"
)

func TestApiMongo(t *testing.T) {
	wg := &sync.WaitGroup{}
	defer wg.Wait()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	config, err := StartTestEnv(ctx, wg, "mongodb")
	if err != nil {
		t.Error(err)
		return
	}

	runApiTests(t, config)
}

func TestApiPostgres(t *testing.T) {
	wg := &sync.WaitGroup{}
	defer wg.Wait()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	config, err := StartTestEnv(ctx, wg, "postgres")
	if err != nil {
		t.Error(err)
		return
	}

	runApiTests(t, config)
}

func runApiTests(t *testing.T, config configuration.Config) {
	now := time.Now()
	backup := configuration.TimeNow
	defer func() { configuration.TimeNow = backup }()
	configuration.TimeNow = func() time.Time {
		return now
	}

	t.Run("create value v1", testRequest(config, "PUT", "/values/v1", 13, http.StatusNoContent, nil))
	t.Run("get value v1", testRequest(config, "GET", "/values/v1", nil, http.StatusOK, 13))
	t.Run("get value unknown", testRequest(config, "GET", "/values/unknown", nil, http.StatusOK, nil))
	t.Run("update value v1", testRequest(config, "PUT", "/values/v1", 42, http.StatusNoContent, nil))
	t.Run("get updated value v1", testRequest(config, "GET", "/values/v1", nil, http.StatusOK, 42))

	t.Run("get variables", testRequest(config, "GET", "/variables", nil, http.StatusOK, []model.VariableWithUnixTimestamp{
		{
			Variable: model.Variable{
				Key:                 "v1",
				Value:               42,
				ProcessDefinitionId: "",
				ProcessInstanceId:   "",
			},
			UnixTimestampInS: configuration.TimeNow().Unix(),
		},
	}))
}

func testRequest(config configuration.Config, method string, path string, body interface{}, expectedStatusCode int, expected interface{}) func(t *testing.T) {
	return testRequestWithToken(config, testtoken, method, path, body, expectedStatusCode, expected)
}

func testRequestWithToken(config configuration.Config, token string, method string, path string, body interface{}, expectedStatusCode int, expected interface{}) func(t *testing.T) {
	return func(t *testing.T) {
		var requestBody io.Reader
		if body != nil {
			temp := new(bytes.Buffer)
			err := json.NewEncoder(temp).Encode(body)
			if err != nil {
				t.Error(err)
				return
			}
			requestBody = temp
		}

		req, err := http.NewRequest(method, "http://localhost:"+config.ServerPort+path, requestBody)
		if err != nil {
			t.Error(err)
			return
		}
		req.Header.Set("Authorization", token)
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Error(err)
			return
		}
		if resp.StatusCode != expectedStatusCode {
			temp, _ := io.ReadAll(resp.Body)
			t.Error(resp.StatusCode, string(temp))
			return
		}

		if expected != nil {
			temp, err := json.Marshal(expected)
			if err != nil {
				t.Error(err)
				return
			}
			var normalizedExpected interface{}
			err = json.Unmarshal(temp, &normalizedExpected)
			if err != nil {
				t.Error(err)
				return
			}

			var actual interface{}
			err = json.NewDecoder(resp.Body).Decode(&actual)
			if err != nil {
				t.Error(err)
				return
			}

			if !reflect.DeepEqual(actual, normalizedExpected) {
				a, _ := json.Marshal(actual)
				e, _ := json.Marshal(normalizedExpected)
				t.Error("\n", string(a), "\n", string(e))
				return
			}
		}
	}
}
