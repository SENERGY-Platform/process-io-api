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
)

func New(config configuration.Config) *Controller {
	return &Controller{config: config}
}

type Controller struct {
	config configuration.Config
}

func (this *Controller) List(token auth.Token, query model.VariablesQueryOptions) ([]model.VariableWithUnixTimestamp, error) {
	//TODO implement me
	panic("implement me")
}

func (this *Controller) Get(token auth.Token, key string) (model.VariableWithUnixTimestamp, error) {
	//TODO implement me
	panic("implement me")
}

func (this *Controller) Set(token auth.Token, variable model.Variable) error {
	//TODO implement me
	panic("implement me")
}

func (this *Controller) Delete(token auth.Token, key string) error {
	//TODO implement me
	panic("implement me")
}

func (this *Controller) Bulk(token auth.Token, bulk model.BulkRequest) (model.BulkResponse, error) {
	//TODO implement me
	panic("implement me")
}
