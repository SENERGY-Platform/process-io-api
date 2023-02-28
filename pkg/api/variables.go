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
	"strconv"
	"strings"
)

func init() {
	endpoints = append(endpoints, &Variables{})
}

type Variables struct{}

// List godoc
// @Summary      returns a list of variables
// @Description  returns a list of variables
// @Tags         variables
// @Param        limit query integer false "limits size of result; 0 means unlimited"
// @Param        offset query integer false "offset to be used in combination with limit"
// @Param        sort query string false "describes the sorting in the form of key.asc"
// @Param        process_instance_id query string false "filter by process instance id"
// @Param        process_definition_id query string false "filter by process definition id"
// @Produce      json
// @Success      200 {array} model.VariableWithUnixTimestamp
// @Failure      400
// @Failure      500
// @Router       /variables [get]
func (this *Variables) List(config configuration.Config, router *httprouter.Router, ctrl Controller) {
	router.GET("/variables", func(writer http.ResponseWriter, request *http.Request, params httprouter.Params) {
		token, err := auth.GetParsedToken(request)
		if err != nil {
			http.Error(writer, err.Error(), http.StatusUnauthorized)
			return
		}

		query := model.VariablesQueryOptions{}
		limit := request.URL.Query().Get("limit")
		if limit != "" {
			query.Limit, err = strconv.Atoi(limit)
			if err != nil {
				http.Error(writer, err.Error(), http.StatusBadRequest)
				return
			}
		}
		offset := request.URL.Query().Get("offset")
		if offset != "" {
			query.Offset, err = strconv.Atoi(offset)
			if err != nil {
				http.Error(writer, err.Error(), http.StatusBadRequest)
				return
			}
		}
		query.Sort = request.URL.Query().Get("sort")
		if query.Sort == "" {
			query.Sort = "key.asc"
		}

		query.ProcessInstanceId = request.URL.Query().Get("process_instance_id")
		query.ProcessDefinitionId = request.URL.Query().Get("process_definition_id")
		query.KeyRegex = request.URL.Query().Get("key_regex")

		result, err := ctrl.List(token.GetUserId(), query)
		if err != nil {
			http.Error(writer, err.Error(), http.StatusInternalServerError)
			return
		}

		writer.Header().Set("Content-Type", "application/json")
		json.NewEncoder(writer).Encode(result)

	})
}

// Count godoc
// @Summary      counts variables
// @Description  counts variables
// @Tags         variables, count
// @Param        process_instance_id query string false "filter by process instance id"
// @Param        process_definition_id query string false "filter by process definition id"
// @Produce      json
// @Success      200 {object} model.Count
// @Failure      400
// @Failure      500
// @Router       /count/variables [get]
func (this *Variables) Count(config configuration.Config, router *httprouter.Router, ctrl Controller) {
	router.GET("/count/variables", func(writer http.ResponseWriter, request *http.Request, params httprouter.Params) {
		token, err := auth.GetParsedToken(request)
		if err != nil {
			http.Error(writer, err.Error(), http.StatusUnauthorized)
			return
		}

		query := model.VariablesQueryOptions{}

		query.ProcessInstanceId = request.URL.Query().Get("process_instance_id")
		query.ProcessDefinitionId = request.URL.Query().Get("process_definition_id")
		query.KeyRegex = request.URL.Query().Get("key_regex")

		result, err := ctrl.Count(token.GetUserId(), query)
		if err != nil {
			http.Error(writer, err.Error(), http.StatusInternalServerError)
			return
		}

		writer.Header().Set("Content-Type", "application/json")
		json.NewEncoder(writer).Encode(result)

	})
}

// Get godoc
// @Summary      returns the variable associated with the given key
// @Description  returns the variable associated with the given key
// @Tags         variables
// @Param        key path string true "key of variable/value"
// @Produce      json
// @Success      200 {object} model.VariableWithUnixTimestamp
// @Failure      400
// @Failure      500
// @Router       /variables/{key} [get]
func (this *Variables) Get(config configuration.Config, router *httprouter.Router, ctrl Controller) {
	router.GET("/variables/*key", func(writer http.ResponseWriter, request *http.Request, params httprouter.Params) {
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
		json.NewEncoder(writer).Encode(result)
	})
}

// Set godoc
// @Summary      set the variable associated with the given key
// @Description  set the variable associated with the given key
// @Tags         variables
// @Accept       json
// @Param        key path string true "key of variable/value"
// @Param        message body model.Variable true "model.Variable"
// @Success      204
// @Failure      400
// @Failure      500
// @Router       /variables/{key} [put]
func (this *Variables) Set(config configuration.Config, router *httprouter.Router, ctrl Controller) {
	router.PUT("/variables/:key", func(writer http.ResponseWriter, request *http.Request, params httprouter.Params) {
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
		msg := model.Variable{}
		err = json.NewDecoder(request.Body).Decode(&msg)
		if err != nil {
			http.Error(writer, err.Error(), http.StatusBadRequest)
			return
		}
		if msg.Key != key {
			http.Error(writer, "path.key != body.key", http.StatusBadRequest)
			return
		}

		err = ctrl.Set(token.GetUserId(), msg)
		if err != nil {
			http.Error(writer, err.Error(), http.StatusInternalServerError)
			return
		}

		writer.WriteHeader(http.StatusNoContent)
	})
}

// Delete godoc
// @Summary      delete the variables associated with the given key
// @Description  delete the variables associated with the given key
// @Tags         variables
// @Param        key path string true "key of variable/value"
// @Success      204
// @Failure      400
// @Failure      500
// @Router       /variables/{key} [delete]
func (this *Variables) Delete(config configuration.Config, router *httprouter.Router, ctrl Controller) {
	router.DELETE("/variables/:key", func(writer http.ResponseWriter, request *http.Request, params httprouter.Params) {
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
