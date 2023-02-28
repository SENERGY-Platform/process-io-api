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
	"github.com/SENERGY-Platform/process-io-api/pkg/api/client/auth"
	"github.com/SENERGY-Platform/process-io-api/pkg/configuration"
	"github.com/SENERGY-Platform/process-io-api/pkg/model"
	"github.com/julienschmidt/httprouter"
	"net/http"
)

func init() {
	endpoints = append(endpoints, &Bulk{})
}

type Bulk struct{}

// Bulk godoc
// @Summary      bulk write of variables and read of values
// @Description  bulk write of variables and read of values
// @Tags         bulk
// @Accept       json
// @Produce      json
// @Param        message body model.BulkRequest true "model.BulkRequest; 'get' contains a list of value keys; 'set' contains a list of model.Variable"
// @Success      200 {object} model.BulkResponse
// @Failure      500
// @Router       /bulk [post]
func (this *Bulk) Bulk(config configuration.Config, router *httprouter.Router, ctrl Controller) {
	router.POST("/bulk", func(writer http.ResponseWriter, request *http.Request, params httprouter.Params) {
		token, err := auth.GetParsedToken(request)
		if err != nil {
			http.Error(writer, err.Error(), http.StatusUnauthorized)
			return
		}
		msg := model.BulkRequest{}
		err = json.NewDecoder(request.Body).Decode(&msg)
		if err != nil {
			http.Error(writer, err.Error(), http.StatusBadRequest)
			return
		}

		result, err := ctrl.Bulk(token.GetUserId(), msg)
		if err != nil {
			http.Error(writer, err.Error(), http.StatusInternalServerError)
			return
		}

		writer.Header().Set("Content-Type", "application/json")
		json.NewEncoder(writer).Encode(result)
	})
}
