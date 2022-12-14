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
	"process-io-api/pkg/controller/calculate"
	"process-io-api/pkg/model"
	"strings"
	"time"
)

type Database interface {
	GetVariable(userId string, key string) (model.VariableWithUser, error)
	SetVariable(variable model.VariableWithUser) error
	DeleteVariable(userId string, key string) error
	ListVariables(userId string, query model.VariablesQueryOptions) ([]model.VariableWithUnixTimestamp, error)
	DeleteVariablesOfProcessDefinition(definitionId string) error
	DeleteVariablesOfProcessInstance(instanceId string) error
	CountVariables(userId string, query model.VariablesQueryOptions) (model.Count, error)
}

func New(config configuration.Config, db Database) *Controller {
	return &Controller{config: config, db: db, calc: calculate.New()}
}

type Controller struct {
	config configuration.Config
	db     Database
	calc   *calculate.Calculate
}

func (this *Controller) List(token auth.Token, query model.VariablesQueryOptions) (result []model.VariableWithUnixTimestamp, err error) {
	result, err = this.db.ListVariables(token.GetUserId(), query)
	if result == nil {
		result = []model.VariableWithUnixTimestamp{}
	}
	return
}

func (this *Controller) Count(token auth.Token, query model.VariablesQueryOptions) (model.Count, error) {
	return this.db.CountVariables(token.GetUserId(), query)
}

func (this *Controller) Get(token auth.Token, key string) (res model.VariableWithUnixTimestamp, err error) {
	if strings.HasPrefix(key, calculate.Prefix) {
		val, err := this.calc.Get(key)
		if err != nil {
			return res, err
		}
		res = model.VariableWithUnixTimestamp{
			Variable: model.Variable{
				Key:   key,
				Value: val,
			},
			UnixTimestampInS: time.Now().Unix(),
		}
	} else {
		variable, err := this.db.GetVariable(token.GetUserId(), key)
		if err != nil {
			return res, err
		}
		res = variable.VariableWithUnixTimestamp
	}
	return
}

func (this *Controller) Set(token auth.Token, variable model.Variable) error {
	return this.db.SetVariable(model.VariableWithUser{
		VariableWithUnixTimestamp: model.VariableWithUnixTimestamp{
			Variable:         variable,
			UnixTimestampInS: configuration.TimeNow().Unix(),
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
