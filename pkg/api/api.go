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

package api

import (
	"context"
	"errors"
	"fmt"
	"github.com/julienschmidt/httprouter"
	"log"
	"net/http"
	"process-io-api/pkg/api/util"
	"process-io-api/pkg/auth"
	"process-io-api/pkg/configuration"
	"process-io-api/pkg/model"
	"reflect"
	"runtime/debug"
)

type Controller interface {
	List(token auth.Token, query model.VariablesQueryOptions) ([]model.VariableWithUnixTimestamp, error)
	Get(token auth.Token, key string) (model.VariableWithUnixTimestamp, error)
	Set(token auth.Token, variable model.Variable) error
	Delete(token auth.Token, key string) error
	Bulk(token auth.Token, bulk model.BulkRequest) (model.BulkResponse, error)
	DeleteProcessDefinition(definitionId string) error
	DeleteProcessInstance(instanceId string) error
	Count(token auth.Token, query model.VariablesQueryOptions) (model.Count, error)
}

type EndpointMethod = func(config configuration.Config, router *httprouter.Router, ctrl Controller)

var endpoints = []interface{}{} //list of objects with EndpointMethod

func Start(ctx context.Context, config configuration.Config, ctrl Controller) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = errors.New(fmt.Sprint(r))
		}
	}()
	router := GetRouter(config, ctrl)

	server := &http.Server{Addr: ":" + config.ServerPort, Handler: router}
	go func() {
		log.Println("listening on ", server.Addr)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			debug.PrintStack()
			log.Fatal("FATAL:", err)
		}
	}()
	go func() {
		<-ctx.Done()
		log.Println("api shutdown", server.Shutdown(context.Background()))
	}()
	return
}

// GetRouter
// @title         Smart-Service-Repository API
// @version       0.1
// @license.name  Apache 2.0
// @license.url   http://www.apache.org/licenses/LICENSE-2.0.html
// @host      localhost:8080
// @BasePath  /
// @securityDefinitions.apikey Bearer
// @in header
// @name Authorization
func GetRouter(config configuration.Config, command Controller) http.Handler {
	router := httprouter.New()
	for _, e := range endpoints {
		for name, call := range getEndpointMethods(e) {
			log.Println("add endpoint " + name)
			call(config, router, command)
		}
	}

	var handler http.Handler
	if config.DisableHttpLogger {
		handler = util.NewCors(router)
	} else {
		handler = util.NewLogger(util.NewCors(router))
	}

	return handler
}

func getEndpointMethods(e interface{}) map[string]func(config configuration.Config, router *httprouter.Router, ctrl Controller) {
	result := map[string]EndpointMethod{}
	objRef := reflect.ValueOf(e)
	methodCount := objRef.NumMethod()
	for i := 0; i < methodCount; i++ {
		m := objRef.Method(i)
		f, ok := m.Interface().(EndpointMethod)
		if ok {
			name := getTypeName(objRef.Type()) + "::" + objRef.Type().Method(i).Name
			result[name] = f
		}
	}
	return result
}

func getTypeName(t reflect.Type) (res string) {
	for t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	return t.Name()
}
