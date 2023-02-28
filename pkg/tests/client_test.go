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
	"context"
	"github.com/SENERGY-Platform/process-io-api/pkg/api"
	"github.com/SENERGY-Platform/process-io-api/pkg/api/client"
	"github.com/SENERGY-Platform/process-io-api/pkg/api/client/auth"
	"github.com/SENERGY-Platform/process-io-api/pkg/configuration"
	"github.com/SENERGY-Platform/process-io-api/pkg/model"
	"reflect"
	"sync"
	"testing"
	"time"
)

type MockAuth map[string]string

func (this MockAuth) ExchangeUserToken(userid string) (token auth.Token, err error) {
	return auth.Parse(this[userid])
}

func TestClientApi(t *testing.T) {
	wg := &sync.WaitGroup{}
	defer wg.Wait()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	config, _, err := StartTestEnv(ctx, wg, "mongodb")
	if err != nil {
		t.Error(err)
		return
	}

	c := client.NewWithAuth("http://localhost:"+config.ServerPort, MockAuth(map[string]string{testTokenUser: testtoken, adminTokenUser: admintoken}), true)

	runClientTests(t, c)
}

func TestClientController(t *testing.T) {
	wg := &sync.WaitGroup{}
	defer wg.Wait()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	_, ctrl, err := StartTestEnv(ctx, wg, "postgres")
	if err != nil {
		t.Error(err)
		return
	}

	runClientTests(t, ctrl)
}

func runClientTests(t *testing.T, client api.Controller) {
	now := time.Now()
	backup := configuration.TimeNow
	defer func() { configuration.TimeNow = backup }()
	configuration.TimeNow = func() time.Time {
		return now
	}

	userid := testTokenUser
	adminid := adminTokenUser

	t.Run("create value v1", func(t *testing.T) {
		err := client.Set(userid, model.Variable{Key: "v1", Value: float64(13)})
		if err != nil {
			t.Error(err)
		}
	})
	t.Run("get value v1", func(t *testing.T) {
		actual, err := client.Get(userid, "v1")
		if err != nil {
			t.Error(err)
			return
		}
		expected := model.VariableWithUnixTimestamp{
			Variable:         model.Variable{Key: "v1", Value: float64(13)},
			UnixTimestampInS: configuration.TimeNow().Unix(),
		}
		if !reflect.DeepEqual(actual, expected) {
			t.Errorf("\n%#v\n%#v", actual, expected)
			return
		}
	})
	t.Run("get value unknown", func(t *testing.T) {
		actual, err := client.Get(userid, "unknown")
		if err != nil {
			t.Error(err)
			return
		}
		expected := model.VariableWithUnixTimestamp{
			Variable:         model.Variable{Key: "unknown"},
			UnixTimestampInS: 0,
		}
		if !reflect.DeepEqual(actual, expected) {
			t.Errorf("\n%#v\n%#v", actual, expected)
			return
		}
	})
	t.Run("update value v1", func(t *testing.T) {
		err := client.Set(userid, model.Variable{Key: "v1", Value: float64(42)})
		if err != nil {
			t.Error(err)
		}
	})
	t.Run("get updated value v1", func(t *testing.T) {
		actual, err := client.Get(userid, "v1")
		if err != nil {
			t.Error(err)
			return
		}
		if !reflect.DeepEqual(actual, model.VariableWithUnixTimestamp{
			Variable:         model.Variable{Key: "v1", Value: float64(42)},
			UnixTimestampInS: configuration.TimeNow().Unix(),
		}) {
			t.Errorf("%#v", actual)
			return
		}
	})

	t.Run("get variables", func(t *testing.T) {
		actual, err := client.List(userid, model.VariablesQueryOptions{})
		if err != nil {
			t.Error(err)
			return
		}
		if !reflect.DeepEqual(actual, []model.VariableWithUnixTimestamp{
			{
				Variable:         model.Variable{Key: "v1", Value: float64(42)},
				UnixTimestampInS: configuration.TimeNow().Unix(),
			},
		}) {
			t.Errorf("%#v", actual)
			return
		}
	})

	t.Run("create value d1 i1 v2", func(t *testing.T) {
		err := client.Set(userid, model.Variable{Key: "v2", Value: "a", ProcessDefinitionId: "d1", ProcessInstanceId: "i1"})
		if err != nil {
			t.Error(err)
		}
	})

	t.Run("create value d2 i2 v3", func(t *testing.T) {
		err := client.Set(userid, model.Variable{Key: "v3", Value: "b", ProcessDefinitionId: "d2", ProcessInstanceId: "i2"})
		if err != nil {
			t.Error(err)
		}
	})

	t.Run("create value d3 i3 v4", func(t *testing.T) {
		err := client.Set(userid, model.Variable{Key: "v4", Value: "c", ProcessDefinitionId: "d3", ProcessInstanceId: "i3"})
		if err != nil {
			t.Error(err)
		}
	})

	t.Run("create value d4 v5", func(t *testing.T) {
		err := client.Set(userid, model.Variable{Key: "v5", Value: "d", ProcessDefinitionId: "d4"})
		if err != nil {
			t.Error(err)
		}
	})

	t.Run("create value d5 v6", func(t *testing.T) {
		err := client.Set(userid, model.Variable{Key: "v6", Value: "e", ProcessDefinitionId: "d5"})
		if err != nil {
			t.Error(err)
		}
	})

	t.Run("get variables", func(t *testing.T) {
		actual, err := client.List(userid, model.VariablesQueryOptions{})
		if err != nil {
			t.Error(err)
			return
		}
		if !reflect.DeepEqual(actual, []model.VariableWithUnixTimestamp{
			{
				Variable: model.Variable{
					Key:                 "v1",
					Value:               float64(42),
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
		}) {
			t.Errorf("%#v", actual)
			return
		}
	})

	t.Run("delete unknown", func(t *testing.T) {
		err := client.Delete(userid, "unknown")
		if err != nil {
			t.Error(err)
		}
	})

	t.Run("get variables", func(t *testing.T) {
		actual, err := client.List(userid, model.VariablesQueryOptions{})
		if err != nil {
			t.Error(err)
			return
		}
		if !reflect.DeepEqual(actual, []model.VariableWithUnixTimestamp{
			{
				Variable: model.Variable{
					Key:                 "v1",
					Value:               float64(42),
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
		}) {
			t.Errorf("%#v", actual)
			return
		}
	})

	t.Run("delete v1", func(t *testing.T) {
		err := client.Delete(userid, "v1")
		if err != nil {
			t.Error(err)
		}
	})

	t.Run("get variables", func(t *testing.T) {
		actual, err := client.List(userid, model.VariablesQueryOptions{})
		if err != nil {
			t.Error(err)
			return
		}
		if !reflect.DeepEqual(actual, []model.VariableWithUnixTimestamp{
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
		}) {
			t.Errorf("%#v", actual)
			return
		}
	})

	t.Run("delete d1", func(t *testing.T) {
		err := client.DeleteProcessDefinition(adminid, "d1")
		if err != nil {
			t.Error(err)
		}
	})

	t.Run("get variables", func(t *testing.T) {
		actual, err := client.List(userid, model.VariablesQueryOptions{})
		if err != nil {
			t.Error(err)
			return
		}
		if !reflect.DeepEqual(actual, []model.VariableWithUnixTimestamp{
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
		}) {
			t.Errorf("%#v", actual)
			return
		}
	})

	t.Run("delete i2", func(t *testing.T) {
		err := client.DeleteProcessInstance(adminid, "i2")
		if err != nil {
			t.Error(err)
		}
	})

	t.Run("get variables", func(t *testing.T) {
		actual, err := client.List(userid, model.VariablesQueryOptions{})
		if err != nil {
			t.Error(err)
			return
		}
		if !reflect.DeepEqual(actual, []model.VariableWithUnixTimestamp{
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
		}) {
			t.Errorf("%#v", actual)
			return
		}
	})

	t.Run("delete d4", func(t *testing.T) {
		err := client.DeleteProcessDefinition(adminid, "d4")
		if err != nil {
			t.Error(err)
		}
	})

	t.Run("get variables", func(t *testing.T) {
		actual, err := client.List(userid, model.VariablesQueryOptions{})
		if err != nil {
			t.Error(err)
			return
		}
		if !reflect.DeepEqual(actual, []model.VariableWithUnixTimestamp{
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
		}) {
			t.Errorf("%#v", actual)
			return
		}
	})

	t.Run("delete d-unknown", func(t *testing.T) {
		err := client.DeleteProcessDefinition(adminid, "d-unknown")
		if err != nil {
			t.Error(err)
		}
	})
	t.Run("delete i-unknown", func(t *testing.T) {
		err := client.DeleteProcessInstance(adminid, "i-unknown")
		if err != nil {
			t.Error(err)
		}
	})

	t.Run("get variables", func(t *testing.T) {
		actual, err := client.List(userid, model.VariablesQueryOptions{})
		if err != nil {
			t.Error(err)
			return
		}
		if !reflect.DeepEqual(actual, []model.VariableWithUnixTimestamp{
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
		}) {
			t.Errorf("%#v", actual)
			return
		}
	})

	t.Run("list variables instance i3", func(t *testing.T) {
		actual, err := client.List(userid, model.VariablesQueryOptions{
			ProcessInstanceId: "i3",
		})
		if err != nil {
			t.Error(err)
			return
		}
		if !reflect.DeepEqual(actual, []model.VariableWithUnixTimestamp{
			{
				Variable: model.Variable{
					Key:                 "v4",
					Value:               "c",
					ProcessDefinitionId: "d3",
					ProcessInstanceId:   "i3",
				},
				UnixTimestampInS: configuration.TimeNow().Unix(),
			},
		}) {
			t.Errorf("%#v", actual)
			return
		}
	})

	t.Run("list variables instance d5", func(t *testing.T) {
		actual, err := client.List(userid, model.VariablesQueryOptions{
			ProcessDefinitionId: "d5",
		})
		if err != nil {
			t.Error(err)
			return
		}
		if !reflect.DeepEqual(actual, []model.VariableWithUnixTimestamp{
			{
				Variable: model.Variable{
					Key:                 "v6",
					Value:               "e",
					ProcessDefinitionId: "d5",
					ProcessInstanceId:   "",
				},
				UnixTimestampInS: configuration.TimeNow().Unix(),
			},
		}) {
			t.Errorf("%#v", actual)
			return
		}
	})

	t.Run("count variables", func(t *testing.T) {
		actual, err := client.Count(userid, model.VariablesQueryOptions{})
		if err != nil {
			t.Error(err)
			return
		}
		if !reflect.DeepEqual(actual, model.Count{Count: 2}) {
			t.Errorf("%#v", actual)
			return
		}
	})

	t.Run("count variables instance i3", func(t *testing.T) {
		actual, err := client.Count(userid, model.VariablesQueryOptions{ProcessInstanceId: "i3"})
		if err != nil {
			t.Error(err)
			return
		}
		if !reflect.DeepEqual(actual, model.Count{Count: 1}) {
			t.Errorf("%#v", actual)
			return
		}
	})

	t.Run("count variables definition d5", func(t *testing.T) {
		actual, err := client.Count(userid, model.VariablesQueryOptions{ProcessDefinitionId: "d5"})
		if err != nil {
			t.Error(err)
			return
		}
		if !reflect.DeepEqual(actual, model.Count{Count: 1}) {
			t.Errorf("%#v", actual)
			return
		}
	})

	t.Run("create value foo", func(t *testing.T) {
		err := client.Set(userid, model.Variable{Key: "foo", Value: float64(13)})
		if err != nil {
			t.Error(err)
		}
	})

	t.Run("create value bar", func(t *testing.T) {
		err := client.Set(userid, model.Variable{Key: "bar", Value: float64(13)})
		if err != nil {
			t.Error(err)
		}
	})

	t.Run("create value foobar", func(t *testing.T) {
		err := client.Set(userid, model.Variable{Key: "foobar", Value: float64(13)})
		if err != nil {
			t.Error(err)
		}
	})

	t.Run("search foo", func(t *testing.T) {
		actual, err := client.List(userid, model.VariablesQueryOptions{
			KeyRegex: "foo",
		})
		if err != nil {
			t.Error(err)
			return
		}
		if !reflect.DeepEqual(actual, []model.VariableWithUnixTimestamp{
			{
				Variable: model.Variable{
					Key:                 "foo",
					Value:               float64(13),
					ProcessDefinitionId: "",
					ProcessInstanceId:   "",
				},
				UnixTimestampInS: configuration.TimeNow().Unix(),
			},
			{
				Variable: model.Variable{
					Key:                 "foobar",
					Value:               float64(13),
					ProcessDefinitionId: "",
					ProcessInstanceId:   "",
				},
				UnixTimestampInS: configuration.TimeNow().Unix(),
			},
		}) {
			t.Errorf("%#v", actual)
			return
		}
	})

	t.Run("count foo variables", func(t *testing.T) {
		actual, err := client.Count(userid, model.VariablesQueryOptions{KeyRegex: "foo"})
		if err != nil {
			t.Error(err)
			return
		}
		if !reflect.DeepEqual(actual, model.Count{Count: 2}) {
			t.Errorf("%#v", actual)
			return
		}
	})

	t.Run("search bar", func(t *testing.T) {
		actual, err := client.List(userid, model.VariablesQueryOptions{
			KeyRegex: "bar",
		})
		if err != nil {
			t.Error(err)
			return
		}
		if !reflect.DeepEqual(actual, []model.VariableWithUnixTimestamp{
			{
				Variable: model.Variable{
					Key:                 "bar",
					Value:               float64(13),
					ProcessDefinitionId: "",
					ProcessInstanceId:   "",
				},
				UnixTimestampInS: configuration.TimeNow().Unix(),
			},
			{
				Variable: model.Variable{
					Key:                 "foobar",
					Value:               float64(13),
					ProcessDefinitionId: "",
					ProcessInstanceId:   "",
				},
				UnixTimestampInS: configuration.TimeNow().Unix(),
			},
		}) {
			t.Errorf("%#v", actual)
			return
		}
	})

	t.Run("count bar variables", func(t *testing.T) {
		actual, err := client.Count(userid, model.VariablesQueryOptions{KeyRegex: "bar"})
		if err != nil {
			t.Error(err)
			return
		}
		if !reflect.DeepEqual(actual, model.Count{Count: 2}) {
			t.Errorf("%#v", actual)
			return
		}
	})

	t.Run("search b.r", func(t *testing.T) {
		actual, err := client.List(userid, model.VariablesQueryOptions{
			KeyRegex: "b.r",
		})
		if err != nil {
			t.Error(err)
			return
		}
		if !reflect.DeepEqual(actual, []model.VariableWithUnixTimestamp{
			{
				Variable: model.Variable{
					Key:                 "bar",
					Value:               float64(13),
					ProcessDefinitionId: "",
					ProcessInstanceId:   "",
				},
				UnixTimestampInS: configuration.TimeNow().Unix(),
			},
			{
				Variable: model.Variable{
					Key:                 "foobar",
					Value:               float64(13),
					ProcessDefinitionId: "",
					ProcessInstanceId:   "",
				},
				UnixTimestampInS: configuration.TimeNow().Unix(),
			},
		}) {
			t.Errorf("%#v", actual)
			return
		}
	})

	t.Run("count b.r variables", func(t *testing.T) {
		actual, err := client.Count(userid, model.VariablesQueryOptions{KeyRegex: "b.r"})
		if err != nil {
			t.Error(err)
			return
		}
		if !reflect.DeepEqual(actual, model.Count{Count: 2}) {
			t.Errorf("%#v", actual)
			return
		}
	})

	t.Run("get calculated utcOffset Asia/Hong_Kong", func(t *testing.T) {
		actual, err := client.Get(userid, "calculate_UtcOffset_Asia/Hong_Kong")
		if err != nil {
			t.Error(err)
			return
		}
		expected := model.VariableWithUnixTimestamp{
			Variable:         model.Variable{Key: "calculate_UtcOffset_Asia/Hong_Kong", Value: float64(8 * 60)},
			UnixTimestampInS: configuration.TimeNow().Unix(),
		}
		expectedInt := model.VariableWithUnixTimestamp{
			Variable:         model.Variable{Key: "calculate_UtcOffset_Asia/Hong_Kong", Value: int64(8 * 60)},
			UnixTimestampInS: configuration.TimeNow().Unix(),
		}
		if !reflect.DeepEqual(actual, expected) && !reflect.DeepEqual(actual, expectedInt) {
			t.Errorf("\n%#v\n%#v", actual, expected)
			return
		}
	})

	t.Run("get calculated utcOffset US/Hawaii", func(t *testing.T) {
		actual, err := client.Get(userid, "calculate_UtcOffset_US/Hawaii")
		if err != nil {
			t.Error(err)
			return
		}
		expected := model.VariableWithUnixTimestamp{
			Variable:         model.Variable{Key: "calculate_UtcOffset_US/Hawaii", Value: float64(-10 * 60)},
			UnixTimestampInS: configuration.TimeNow().Unix(),
		}
		expectedInt := model.VariableWithUnixTimestamp{
			Variable:         model.Variable{Key: "calculate_UtcOffset_US/Hawaii", Value: int64(-10 * 60)},
			UnixTimestampInS: configuration.TimeNow().Unix(),
		}
		if !reflect.DeepEqual(actual, expected) && !reflect.DeepEqual(actual, expectedInt) {
			t.Errorf("\n%#v\n%#v", actual, expected)
			return
		}
	})

}
