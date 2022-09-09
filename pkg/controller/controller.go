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

package controller

import (
	"process-io-api/pkg/auth"
	"process-io-api/pkg/configuration"
	"process-io-api/pkg/model"
	"time"
)

type Database interface {
	GetVariable(userId string, key string) (model.VariableWithUser, error)
	SetVariable(variable model.VariableWithUser) error
	DeleteVariable(userId string, key string) error
	ListVariables(userId string, query model.VariablesQueryOptions) ([]model.VariableWithUnixTimestamp, error)
	DeleteVariablesOfProcessDefinition(definitionId string) error
	DeleteVariablesOfProcessInstance(instanceId string) error
}

func New(config configuration.Config, db Database) *Controller {
	return &Controller{config: config, db: db}
}

type Controller struct {
	config configuration.Config
	db     Database
}

func (this *Controller) List(token auth.Token, query model.VariablesQueryOptions) ([]model.VariableWithUnixTimestamp, error) {
	return this.db.ListVariables(token.GetUserId(), query)
}

func (this *Controller) Get(token auth.Token, key string) (model.VariableWithUnixTimestamp, error) {
	variable, err := this.db.GetVariable(token.GetUserId(), key)
	return variable.VariableWithUnixTimestamp, err
}

func (this *Controller) Set(token auth.Token, variable model.Variable) error {
	return this.db.SetVariable(model.VariableWithUser{
		VariableWithUnixTimestamp: model.VariableWithUnixTimestamp{
			Variable:         variable,
			UnixTimestampInS: time.Now().Unix(),
		},
		UserId: token.GetUserId(),
	})
}

func (this *Controller) Delete(token auth.Token, key string) error {
	return this.db.DeleteVariable(token.GetUserId(), key)
}

func (this *Controller) Bulk(token auth.Token, bulk model.BulkRequest) (result model.BulkResponse, err error) {
	for _, variable := range bulk.Set {
		err = this.Set(token, variable)
		if err != nil {
			return result, err
		}
	}
	for _, key := range bulk.Get {
		var variable model.VariableWithUnixTimestamp
		variable, err = this.Get(token, key)
		if err != nil {
			return result, err
		}
		result = append(result, variable)
	}
	return result, nil
}

func (this *Controller) DeleteProcessDefinition(definitionId string) error {
	return this.db.DeleteVariablesOfProcessDefinition(definitionId)
}

func (this *Controller) DeleteProcessInstance(instanceId string) error {
	return this.db.DeleteVariablesOfProcessInstance(instanceId)
}
