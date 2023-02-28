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
	"encoding/json"
	"github.com/julienschmidt/httprouter"
	"net/http"
	"process-io-api/pkg/auth"
	"process-io-api/pkg/configuration"
	"process-io-api/pkg/model"
	"strings"
)

func init() {
	endpoints = append(endpoints, &Values{})
}

type Values struct{}

type Anything = interface{}

// Get godoc
// @Summary      returns the value associated with the given key
// @Description  returns the value associated with the given key
// @Tags         values
// @Param        key path string true "key of value"
// @Produce      json
// @Success      200 {object} Anything
// @Failure      400
// @Failure      500
// @Router       /values/{key} [get]
func (this *Values) Get(config configuration.Config, router *httprouter.Router, ctrl Controller) {
	router.GET("/values/*key", func(writer http.ResponseWriter, request *http.Request, params httprouter.Params) {
		token, err := auth.GetParsedToken(request)
		if err != nil {
			http.Error(writer, err.Error(), http.StatusUnauthorized)
			return
		}
		key := strings.TrimPrefix(params.ByName("key"), "/")
		if key == "" {
			http.Error(writer, "missing id", http.StatusBadRequest)
			return
		}
		result, err := ctrl.Get(token.GetUserId(), key)
		if err != nil {
			http.Error(writer, err.Error(), http.StatusInternalServerError)
			return
		}

		writer.Header().Set("Content-Type", "application/json")
		json.NewEncoder(writer).Encode(result.Value)
	})
}

// Set godoc
// @Summary      set the value associated with the given key
// @Description  set the value associated with the given key
// @Tags         values
// @Accept       json
// @Param        key path string true "key of value"
// @Param        message body Anything true "Anything"
// @Success      204
// @Failure      400
// @Failure      500
// @Router       /values/{key} [put]
func (this *Values) Set(config configuration.Config, router *httprouter.Router, ctrl Controller) {
	router.PUT("/values/:key", func(writer http.ResponseWriter, request *http.Request, params httprouter.Params) {
		token, err := auth.GetParsedToken(request)
		if err != nil {
			http.Error(writer, err.Error(), http.StatusUnauthorized)
			return
		}
		key := params.ByName("key")
		if key == "" {
			http.Error(writer, "missing key", http.StatusBadRequest)
			return
		}

		var value interface{}
		err = json.NewDecoder(request.Body).Decode(&value)
		if err != nil {
			http.Error(writer, err.Error(), http.StatusBadRequest)
			return
		}

		err = ctrl.Set(token.GetUserId(), model.Variable{
			Key:                 key,
			Value:               value,
			ProcessDefinitionId: "",
			ProcessInstanceId:   "",
		})
		if err != nil {
			http.Error(writer, err.Error(), http.StatusInternalServerError)
			return
		}
		writer.WriteHeader(http.StatusNoContent)
	})
}

// Delete godoc
// @Summary      delete the value associated with the given key
// @Description  delete the value associated with the given key
// @Tags         values
// @Param        key path string true "key of value"
// @Success      204
// @Failure      400
// @Failure      500
// @Router       /values/{key} [delete]
func (this *Values) Delete(config configuration.Config, router *httprouter.Router, ctrl Controller) {
	router.DELETE("/values/:key", func(writer http.ResponseWriter, request *http.Request, params httprouter.Params) {
		token, err := auth.GetParsedToken(request)
		if err != nil {
			http.Error(writer, err.Error(), http.StatusUnauthorized)
			return
		}
		key := params.ByName("key")
		if key == "" {
			http.Error(writer, "missing key", http.StatusBadRequest)
			return
		}
		err = ctrl.Delete(token.GetUserId(), key)
		if err != nil {
			http.Error(writer, err.Error(), http.StatusInternalServerError)
			return
		}
		writer.WriteHeader(http.StatusNoContent)
	})
}
