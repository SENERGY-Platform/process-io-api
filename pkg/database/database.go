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

package database

import (
	"context"
	"errors"
	"process-io-api/pkg/configuration"
	"process-io-api/pkg/database/mongo"
	"process-io-api/pkg/model"
)

type Database interface {
	GetVariable(userId string, key string) (model.VariableWithUser, error)
	SetVariable(variable model.VariableWithUser) error
	DeleteVariable(userId string, key string) error
	ListVariables(userId string, query model.VariablesQueryOptions) ([]model.VariableWithUnixTimestamp, error)
}

func New(ctx context.Context, config configuration.Config) (Database, error) {
	switch config.DatabaseSelection {
	case "mongodb":
		return mongo.New(ctx, config)
	default:
		return nil, errors.New("unknown database: " + config.DatabaseSelection)
	}
}
