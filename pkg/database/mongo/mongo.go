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

package mongo

import (
	"context"
	"process-io-api/pkg/configuration"
	"process-io-api/pkg/model"
)

func New(ctx context.Context, config configuration.Config) (*Mongo, error) {
	return &Mongo{config: config}, nil
}

type Mongo struct {
	config configuration.Config
}

func (this *Mongo) GetVariable(userId string, key string) (model.VariableWithUser, error) {
	//TODO implement me
	panic("implement me")
}

func (this *Mongo) SetVariable(variable model.VariableWithUser) error {
	//TODO implement me
	panic("implement me")
}

func (this *Mongo) DeleteVariable(userId string, key string) error {
	//TODO implement me
	panic("implement me")
}

func (this *Mongo) ListVariables(userId string, query model.VariablesQueryOptions) ([]model.VariableWithUnixTimestamp, error) {
	//TODO implement me
	panic("implement me")
}
