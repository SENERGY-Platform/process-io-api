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
	"strconv"
	"sync"
	"testing"
	"time"
)

func BenchmarkMongo(b *testing.B) {
	wg := &sync.WaitGroup{}
	defer wg.Wait()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	config, _, err := StartTestEnv(ctx, wg, "mongodb")
	if err != nil {
		b.Error(err)
		return
	}

	runBenchmark(b, config)
}

func BenchmarkPostgres(b *testing.B) {
	wg := &sync.WaitGroup{}
	defer wg.Wait()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	config, _, err := StartTestEnv(ctx, wg, "postgres")
	if err != nil {
		b.Error(err)
		return
	}

	b.ResetTimer()
	runBenchmark(b, config)
}

func runBenchmark(b *testing.B, config configuration.Config) {
	now := time.Now()
	backup := configuration.TimeNow
	defer func() { configuration.TimeNow = backup }()
	configuration.TimeNow = func() time.Time {
		return now
	}

	count := 1000

	b.Run("set", setBenchmark(config, count))
	b.Run("get", getBenchmark(config, count))

	b.Run("bulk", bulkBenchmark(config, count))
}

func setBenchmark(config configuration.Config, count int) func(b *testing.B) {
	return func(b *testing.B) {
		//reset db state
		variablesCleanup(config, b, count*b.N)
		b.ResetTimer()

		for j := 0; j < count*b.N; j++ {
			temp := new(bytes.Buffer)
			err := json.NewEncoder(temp).Encode(j)
			if err != nil {
				b.Error(err)
				return
			}
			req, err := http.NewRequest("PUT", "http://localhost:"+config.ServerPort+"/variables/v"+strconv.Itoa(j), temp)
			if err != nil {
				b.Error(err)
				return
			}
			req.Header.Set("Authorization", testtoken)
			resp, err := http.DefaultClient.Do(req)
			if err != nil {
				b.Error(err)
				return
			}
			io.ReadAll(resp.Body)
			resp.Body.Close()
		}
	}
}

func getBenchmark(config configuration.Config, count int) func(b *testing.B) {
	return func(b *testing.B) {
		for j := 0; j < count*b.N; j++ {
			req, err := http.NewRequest("GET", "http://localhost:"+config.ServerPort+"/variables/v"+strconv.Itoa(j), nil)
			if err != nil {
				b.Error(err)
				return
			}
			req.Header.Set("Authorization", testtoken)
			resp, err := http.DefaultClient.Do(req)
			if err != nil {
				b.Error(err)
				return
			}
			io.ReadAll(resp.Body)
			resp.Body.Close()
		}
	}

}

func bulkBenchmark(config configuration.Config, count int) func(b *testing.B) {
	return func(b *testing.B) {
		variablesCleanup(config, b, count*b.N)
		b.ResetTimer()

		bulkRequest := model.BulkRequest{
			Get: []string{},
			Set: []model.Variable{},
		}
		for j := 0; j < count*b.N; j++ {
			bulkRequest.Get = append(bulkRequest.Get, "v"+strconv.Itoa(j))
			bulkRequest.Set = append(bulkRequest.Set, model.Variable{
				Key:                 "v" + strconv.Itoa(j),
				Value:               j,
				ProcessDefinitionId: strconv.Itoa(j),
				ProcessInstanceId:   strconv.Itoa(j),
			})
		}
		temp := new(bytes.Buffer)
		err := json.NewEncoder(temp).Encode(bulkRequest)
		if err != nil {
			b.Error(err)
			return
		}
		req, err := http.NewRequest("POST", "http://localhost:"+config.ServerPort+"/bulk", temp)
		if err != nil {
			b.Error(err)
			return
		}
		req.Header.Set("Authorization", testtoken)
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			b.Error(err)
			return
		}
		io.ReadAll(resp.Body)
		resp.Body.Close()
	}
}

type errorHandler interface {
	Error(args ...any)
}

func variablesCleanup(config configuration.Config, e errorHandler, count int) {
	req, err := http.NewRequest("GET", "http://localhost:"+config.ServerPort+"/variables?limit="+strconv.Itoa(count), nil)
	if err != nil {
		e.Error(err)
		return
	}
	req.Header.Set("Authorization", testtoken)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		e.Error(err)
		return
	}

	variables := []model.Variable{}
	err = json.NewDecoder(resp.Body).Decode(&variables)
	if err != nil {
		e.Error(err)
		return
	}
	resp.Body.Close()

	for _, variable := range variables {
		req, err := http.NewRequest("DELETE", "http://localhost:"+config.ServerPort+"/variables/"+url.PathEscape(variable.Key), nil)
		if err != nil {
			e.Error(err)
			return
		}
		req.Header.Set("Authorization", testtoken)
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			e.Error(err)
			return
		}
		io.ReadAll(resp.Body)
		resp.Body.Close()
	}
}
