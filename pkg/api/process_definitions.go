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
)

func init() {
	endpoints = append(endpoints, &ProcessDefinitions{})
}

type ProcessDefinitions struct{}

// SetWithInstance godoc
// @Summary      set the value associated with the given key
// @Description  set the value associated with the given key
// @Tags         values, process-definitions, process-instances
// @Accept       json
// @Param        key path string true "key of value"
// @Param        definitionId path string true "definitionId associated with value"
// @Param        instanceId path string true "instanceId associated with value"
// @Param        message body Anything true "Anything"
// @Success      204
// @Failure      400
// @Failure      500
// @Router       /process-definitions/{definitionId}/process-instances/{instanceId}/values/{key} [put]
func (this *ProcessDefinitions) SetWithInstance(config configuration.Config, router *httprouter.Router, ctrl Controller) {
	router.PUT("/process-definitions/:definitionId/process-instances/:instanceId/values/:key", func(writer http.ResponseWriter, request *http.Request, params httprouter.Params) {
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
		instanceId := params.ByName("instanceId")
		if instanceId == "" {
			http.Error(writer, "missing instanceId", http.StatusBadRequest)
			return
		}
		definitionId := params.ByName("definitionId")
		if definitionId == "" {
			http.Error(writer, "missing definitionId", http.StatusBadRequest)
			return
		}

		var value interface{}
		err = json.NewDecoder(request.Body).Decode(&value)
		if err != nil {
			http.Error(writer, err.Error(), http.StatusBadRequest)
			return
		}

		err = ctrl.Set(token, model.Variable{
			Key:                 key,
			Value:               value,
			ProcessDefinitionId: definitionId,
			ProcessInstanceId:   instanceId,
		})
		if err != nil {
			http.Error(writer, err.Error(), http.StatusInternalServerError)
			return
		}
		writer.WriteHeader(http.StatusNoContent)
	})
}

// Set			 godoc
// @Summary      set the value associated with the given key
// @Description  set the value associated with the given key
// @Tags         values, process-definitions
// @Accept       json
// @Param        key path string true "key of value"
// @Param        definitionId path string true "definitionId associated with value"
// @Param        message body Anything true "Anything"
// @Success      204
// @Failure      400
// @Failure      500
// @Router       /process-definitions/{definitionId}/values/{key} [put]
func (this *ProcessDefinitions) Set(config configuration.Config, router *httprouter.Router, ctrl Controller) {
	router.PUT("/process-definitions/:definitionId/values/:key", func(writer http.ResponseWriter, request *http.Request, params httprouter.Params) {
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
		definitionId := params.ByName("definitionId")
		if definitionId == "" {
			http.Error(writer, "missing definitionId", http.StatusBadRequest)
			return
		}
		var value interface{}
		err = json.NewDecoder(request.Body).Decode(&value)
		if err != nil {
			http.Error(writer, err.Error(), http.StatusBadRequest)
			return
		}

		err = ctrl.Set(token, model.Variable{
			Key:                 key,
			Value:               value,
			ProcessDefinitionId: definitionId,
			ProcessInstanceId:   "",
		})
		if err != nil {
			http.Error(writer, err.Error(), http.StatusInternalServerError)
			return
		}
		writer.WriteHeader(http.StatusNoContent)
	})
}
