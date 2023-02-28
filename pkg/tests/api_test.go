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
	"github.com/SENERGY-Platform/process-io-api/pkg/configuration"
	"github.com/SENERGY-Platform/process-io-api/pkg/model"
	"io"
	"net/http"
	"net/url"
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

	config, _, err := StartTestEnv(ctx, wg, "mongodb")
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

	config, _, err := StartTestEnv(ctx, wg, "postgres")
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

	t.Run("create value d1 i1 v2", testRequest(config, "PUT", "/process-definitions/d1/process-instances/i1/values/v2", "a", http.StatusNoContent, nil))
	t.Run("create value d2 i2 v3", testRequest(config, "PUT", "/process-definitions/d2/process-instances/i2/values/v3", "b", http.StatusNoContent, nil))
	t.Run("create value d3 i3 v4", testRequest(config, "PUT", "/process-definitions/d3/process-instances/i3/values/v4", "c", http.StatusNoContent, nil))
	t.Run("create value d4 v5", testRequest(config, "PUT", "/process-definitions/d4/values/v5", "d", http.StatusNoContent, nil))
	t.Run("create value d5 v6", testRequest(config, "PUT", "/process-definitions/d5/values/v6", "e", http.StatusNoContent, nil))

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
		{
			Variable: model.Variable{
				Key:                 "v2",
				Value:               "a",
				ProcessDefinitionId: "d1",
				ProcessInstanceId:   "i1",
			},
			UnixTimestampInS: configuration.TimeNow().Unix(),
		},
		{
			Variable: model.Variable{
				Key:                 "v3",
				Value:               "b",
				ProcessDefinitionId: "d2",
				ProcessInstanceId:   "i2",
			},
			UnixTimestampInS: configuration.TimeNow().Unix(),
		},
		{
			Variable: model.Variable{
				Key:                 "v4",
				Value:               "c",
				ProcessDefinitionId: "d3",
				ProcessInstanceId:   "i3",
			},
			UnixTimestampInS: configuration.TimeNow().Unix(),
		},
		{
			Variable: model.Variable{
				Key:                 "v5",
				Value:               "d",
				ProcessDefinitionId: "d4",
				ProcessInstanceId:   "",
			},
			UnixTimestampInS: configuration.TimeNow().Unix(),
		},
		{
			Variable: model.Variable{
				Key:                 "v6",
				Value:               "e",
				ProcessDefinitionId: "d5",
				ProcessInstanceId:   "",
			},
			UnixTimestampInS: configuration.TimeNow().Unix(),
		},
	}))

	t.Run("delete unknown", testRequest(config, "DELETE", "/values/unknown", nil, http.StatusNoContent, nil))

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
		{
			Variable: model.Variable{
				Key:                 "v2",
				Value:               "a",
				ProcessDefinitionId: "d1",
				ProcessInstanceId:   "i1",
			},
			UnixTimestampInS: configuration.TimeNow().Unix(),
		},
		{
			Variable: model.Variable{
				Key:                 "v3",
				Value:               "b",
				ProcessDefinitionId: "d2",
				ProcessInstanceId:   "i2",
			},
			UnixTimestampInS: configuration.TimeNow().Unix(),
		},
		{
			Variable: model.Variable{
				Key:                 "v4",
				Value:               "c",
				ProcessDefinitionId: "d3",
				ProcessInstanceId:   "i3",
			},
			UnixTimestampInS: configuration.TimeNow().Unix(),
		},
		{
			Variable: model.Variable{
				Key:                 "v5",
				Value:               "d",
				ProcessDefinitionId: "d4",
				ProcessInstanceId:   "",
			},
			UnixTimestampInS: configuration.TimeNow().Unix(),
		},
		{
			Variable: model.Variable{
				Key:                 "v6",
				Value:               "e",
				ProcessDefinitionId: "d5",
				ProcessInstanceId:   "",
			},
			UnixTimestampInS: configuration.TimeNow().Unix(),
		},
	}))

	t.Run("delete v1", testRequest(config, "DELETE", "/values/v1", nil, http.StatusNoContent, nil))

	t.Run("get variables", testRequest(config, "GET", "/variables", nil, http.StatusOK, []model.VariableWithUnixTimestamp{
		{
			Variable: model.Variable{
				Key:                 "v2",
				Value:               "a",
				ProcessDefinitionId: "d1",
				ProcessInstanceId:   "i1",
			},
			UnixTimestampInS: configuration.TimeNow().Unix(),
		},
		{
			Variable: model.Variable{
				Key:                 "v3",
				Value:               "b",
				ProcessDefinitionId: "d2",
				ProcessInstanceId:   "i2",
			},
			UnixTimestampInS: configuration.TimeNow().Unix(),
		},
		{
			Variable: model.Variable{
				Key:                 "v4",
				Value:               "c",
				ProcessDefinitionId: "d3",
				ProcessInstanceId:   "i3",
			},
			UnixTimestampInS: configuration.TimeNow().Unix(),
		},
		{
			Variable: model.Variable{
				Key:                 "v5",
				Value:               "d",
				ProcessDefinitionId: "d4",
				ProcessInstanceId:   "",
			},
			UnixTimestampInS: configuration.TimeNow().Unix(),
		},
		{
			Variable: model.Variable{
				Key:                 "v6",
				Value:               "e",
				ProcessDefinitionId: "d5",
				ProcessInstanceId:   "",
			},
			UnixTimestampInS: configuration.TimeNow().Unix(),
		},
	}))

	t.Run("delete d1", testRequestWithToken(config, admintoken, "DELETE", "/process-definitions/d1", nil, http.StatusNoContent, nil))

	t.Run("get variables", testRequest(config, "GET", "/variables", nil, http.StatusOK, []model.VariableWithUnixTimestamp{
		{
			Variable: model.Variable{
				Key:                 "v3",
				Value:               "b",
				ProcessDefinitionId: "d2",
				ProcessInstanceId:   "i2",
			},
			UnixTimestampInS: configuration.TimeNow().Unix(),
		},
		{
			Variable: model.Variable{
				Key:                 "v4",
				Value:               "c",
				ProcessDefinitionId: "d3",
				ProcessInstanceId:   "i3",
			},
			UnixTimestampInS: configuration.TimeNow().Unix(),
		},
		{
			Variable: model.Variable{
				Key:                 "v5",
				Value:               "d",
				ProcessDefinitionId: "d4",
				ProcessInstanceId:   "",
			},
			UnixTimestampInS: configuration.TimeNow().Unix(),
		},
		{
			Variable: model.Variable{
				Key:                 "v6",
				Value:               "e",
				ProcessDefinitionId: "d5",
				ProcessInstanceId:   "",
			},
			UnixTimestampInS: configuration.TimeNow().Unix(),
		},
	}))

	t.Run("delete i2", testRequestWithToken(config, admintoken, "DELETE", "/process-instances/i2", nil, http.StatusNoContent, nil))

	t.Run("get variables", testRequest(config, "GET", "/variables", nil, http.StatusOK, []model.VariableWithUnixTimestamp{
		{
			Variable: model.Variable{
				Key:                 "v4",
				Value:               "c",
				ProcessDefinitionId: "d3",
				ProcessInstanceId:   "i3",
			},
			UnixTimestampInS: configuration.TimeNow().Unix(),
		},
		{
			Variable: model.Variable{
				Key:                 "v5",
				Value:               "d",
				ProcessDefinitionId: "d4",
				ProcessInstanceId:   "",
			},
			UnixTimestampInS: configuration.TimeNow().Unix(),
		},
		{
			Variable: model.Variable{
				Key:                 "v6",
				Value:               "e",
				ProcessDefinitionId: "d5",
				ProcessInstanceId:   "",
			},
			UnixTimestampInS: configuration.TimeNow().Unix(),
		},
	}))

	t.Run("delete d4", testRequestWithToken(config, admintoken, "DELETE", "/process-definitions/d4", nil, http.StatusNoContent, nil))

	t.Run("get variables", testRequest(config, "GET", "/variables", nil, http.StatusOK, []model.VariableWithUnixTimestamp{
		{
			Variable: model.Variable{
				Key:                 "v4",
				Value:               "c",
				ProcessDefinitionId: "d3",
				ProcessInstanceId:   "i3",
			},
			UnixTimestampInS: configuration.TimeNow().Unix(),
		},
		{
			Variable: model.Variable{
				Key:                 "v6",
				Value:               "e",
				ProcessDefinitionId: "d5",
				ProcessInstanceId:   "",
			},
			UnixTimestampInS: configuration.TimeNow().Unix(),
		},
	}))

	t.Run("delete d-unknown ", testRequestWithToken(config, admintoken, "DELETE", "/process-definitions/d-unknown", nil, http.StatusNoContent, nil))
	t.Run("delete i-unknown ", testRequestWithToken(config, admintoken, "DELETE", "/process-instances/i-unknown", nil, http.StatusNoContent, nil))

	t.Run("get variables", testRequest(config, "GET", "/variables", nil, http.StatusOK, []model.VariableWithUnixTimestamp{
		{
			Variable: model.Variable{
				Key:                 "v4",
				Value:               "c",
				ProcessDefinitionId: "d3",
				ProcessInstanceId:   "i3",
			},
			UnixTimestampInS: configuration.TimeNow().Unix(),
		},
		{
			Variable: model.Variable{
				Key:                 "v6",
				Value:               "e",
				ProcessDefinitionId: "d5",
				ProcessInstanceId:   "",
			},
			UnixTimestampInS: configuration.TimeNow().Unix(),
		},
	}))

	t.Run("list variables instance i3", testRequest(config, "GET", "/variables?process_instance_id=i3", nil, http.StatusOK, []model.VariableWithUnixTimestamp{
		{
			Variable: model.Variable{
				Key:                 "v4",
				Value:               "c",
				ProcessDefinitionId: "d3",
				ProcessInstanceId:   "i3",
			},
			UnixTimestampInS: configuration.TimeNow().Unix(),
		},
	}))
	t.Run("list variables definition d5", testRequest(config, "GET", "/variables?process_definition_id=d5", nil, http.StatusOK, []model.VariableWithUnixTimestamp{
		{
			Variable: model.Variable{
				Key:                 "v6",
				Value:               "e",
				ProcessDefinitionId: "d5",
				ProcessInstanceId:   "",
			},
			UnixTimestampInS: configuration.TimeNow().Unix(),
		},
	}))

	t.Run("count variables", testRequest(config, "GET", "/count/variables", nil, http.StatusOK, model.Count{Count: 2}))
	t.Run("count variables instance i3", testRequest(config, "GET", "/count/variables?process_instance_id=i3", nil, http.StatusOK, model.Count{Count: 1}))
	t.Run("count variables definition d5", testRequest(config, "GET", "/count/variables?process_definition_id=d5", nil, http.StatusOK, model.Count{Count: 1}))

	t.Run("create value foo", testRequest(config, "PUT", "/values/foo", 13, http.StatusNoContent, nil))
	t.Run("create value bar", testRequest(config, "PUT", "/values/bar", 13, http.StatusNoContent, nil))
	t.Run("create value foobar", testRequest(config, "PUT", "/values/foobar", 13, http.StatusNoContent, nil))

	t.Run("search foo", testRequest(config, "GET", "/variables?key_regex="+url.QueryEscape("foo"), nil, http.StatusOK, []model.VariableWithUnixTimestamp{
		{
			Variable: model.Variable{
				Key:                 "foo",
				Value:               13,
				ProcessDefinitionId: "",
				ProcessInstanceId:   "",
			},
			UnixTimestampInS: configuration.TimeNow().Unix(),
		},
		{
			Variable: model.Variable{
				Key:                 "foobar",
				Value:               13,
				ProcessDefinitionId: "",
				ProcessInstanceId:   "",
			},
			UnixTimestampInS: configuration.TimeNow().Unix(),
		},
	}))
	t.Run("count foo variables", testRequest(config, "GET", "/count/variables?key_regex="+url.QueryEscape("foo"), nil, http.StatusOK, model.Count{Count: 2}))

	t.Run("search bar", testRequest(config, "GET", "/variables?key_regex="+url.QueryEscape("bar"), nil, http.StatusOK, []model.VariableWithUnixTimestamp{
		{
			Variable: model.Variable{
				Key:                 "bar",
				Value:               13,
				ProcessDefinitionId: "",
				ProcessInstanceId:   "",
			},
			UnixTimestampInS: configuration.TimeNow().Unix(),
		},
		{
			Variable: model.Variable{
				Key:                 "foobar",
				Value:               13,
				ProcessDefinitionId: "",
				ProcessInstanceId:   "",
			},
			UnixTimestampInS: configuration.TimeNow().Unix(),
		},
	}))
	t.Run("count bar variables", testRequest(config, "GET", "/count/variables?key_regex="+url.QueryEscape("bar"), nil, http.StatusOK, model.Count{Count: 2}))

	t.Run("search b.r", testRequest(config, "GET", "/variables?key_regex="+url.QueryEscape("b.r"), nil, http.StatusOK, []model.VariableWithUnixTimestamp{
		{
			Variable: model.Variable{
				Key:                 "bar",
				Value:               13,
				ProcessDefinitionId: "",
				ProcessInstanceId:   "",
			},
			UnixTimestampInS: configuration.TimeNow().Unix(),
		},
		{
			Variable: model.Variable{
				Key:                 "foobar",
				Value:               13,
				ProcessDefinitionId: "",
				ProcessInstanceId:   "",
			},
			UnixTimestampInS: configuration.TimeNow().Unix(),
		},
	}))
	t.Run("count b.r variables", testRequest(config, "GET", "/count/variables?key_regex="+url.QueryEscape("b.r"), nil, http.StatusOK, model.Count{Count: 2}))
	t.Run("test calculated utcOffset Asia/Hong_Kong", testRequest(config, "GET", "/values/calculate_UtcOffset_Asia/Hong_Kong", nil, http.StatusOK, 8*60))
	t.Run("test calculated utcOffset US/Hawaii", testRequest(config, "GET", "/values/calculate_UtcOffset_US/Hawaii", nil, http.StatusOK, -10*60))

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
		defer resp.Body.Close()
		defer io.ReadAll(resp.Body) // ensure reuse of connection
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
