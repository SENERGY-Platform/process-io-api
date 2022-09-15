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
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"process-io-api/pkg/model"
	"runtime/debug"
)

var VariableBson = getBsonFieldObject[model.VariableWithUser]()

func init() {
	CreateCollections = append(CreateCollections, func(db *Mongo) error {
		var err error
		collection := db.client.Database(db.config.MongoTable).Collection(db.config.MongoVariablesCollection)
		err = db.ensureCompoundIndex(collection, "variables_user_key_index", true, true, VariableBson.UserId, VariableBson.Key)
		if err != nil {
			debug.PrintStack()
			return err
		}
		err = db.ensureIndex(collection, "variables_user_index", VariableBson.UserId, true, false)
		if err != nil {
			debug.PrintStack()
			return err
		}
		err = db.ensureIndex(collection, "variables_p_instance_index", VariableBson.ProcessInstanceId, true, false)
		if err != nil {
			debug.PrintStack()
			return err
		}
		err = db.ensureIndex(collection, "variables_p_definition_index", VariableBson.ProcessDefinitionId, true, false)
		if err != nil {
			debug.PrintStack()
			return err
		}
		return nil
	})
}

func (this *Mongo) variablesCollection() *mongo.Collection {
	return this.client.Database(this.config.MongoTable).Collection(this.config.MongoVariablesCollection)
}

func (this *Mongo) GetVariable(userId string, key string) (result model.VariableWithUser, err error) {
	ctx, _ := getTimeoutContext()
	filter := bson.M{VariableBson.UserId: userId, VariableBson.Key: key}
	temp := this.variablesCollection().FindOne(ctx, filter)
	err = temp.Err()
	if err == mongo.ErrNoDocuments {
		return model.VariableWithUser{
			VariableWithUnixTimestamp: model.VariableWithUnixTimestamp{
				Variable: model.Variable{
					Key:                 key,
					Value:               nil,
					ProcessDefinitionId: "",
					ProcessInstanceId:   "",
				},
				UnixTimestampInS: 0,
			},
			UserId: userId,
		}, nil
	}
	if err != nil {
		return
	}
	err = temp.Decode(&result)
	if err == mongo.ErrNoDocuments {
		return model.VariableWithUser{
			VariableWithUnixTimestamp: model.VariableWithUnixTimestamp{
				Variable: model.Variable{
					Key:                 key,
					Value:               nil,
					ProcessDefinitionId: "",
					ProcessInstanceId:   "",
				},
				UnixTimestampInS: 0,
			},
			UserId: userId,
		}, nil
	}
	return result, nil
}

func (this *Mongo) SetVariable(variable model.VariableWithUser) error {
	ctx, _ := getTimeoutContext()
	_, err := this.variablesCollection().ReplaceOne(
		ctx,
		bson.M{
			VariableBson.UserId: variable.UserId,
			VariableBson.Key:    variable.Key,
		},
		variable,
		options.Replace().SetUpsert(true))
	if err != nil {
		return err
	}
	return nil
}

func (this *Mongo) DeleteVariable(userId string, key string) error {
	ctx, _ := getTimeoutContext()
	_, err := this.variablesCollection().DeleteMany(ctx, bson.M{
		VariableBson.UserId: userId,
		VariableBson.Key:    key,
	})
	return err
}

func (this *Mongo) ListVariables(userId string, query model.VariablesQueryOptions) (result []model.VariableWithUnixTimestamp, err error) {
	opt := createFindOptions(query)
	filter := bson.M{VariableBson.UserId: userId}
	if query.ProcessDefinitionId != "" {
		filter[VariableBson.ProcessDefinitionId] = query.ProcessDefinitionId
	}
	if query.ProcessInstanceId != "" {
		filter[VariableBson.ProcessInstanceId] = query.ProcessInstanceId
	}
	ctx, _ := getTimeoutContext()
	cursor, err := this.variablesCollection().Find(ctx, filter, opt)
	if err != nil {
		return result, err
	}
	temp, err := readCursorResult[model.VariableWithUser](ctx, cursor)
	if err != nil {
		return result, err
	}
	for _, e := range temp {
		result = append(result, e.VariableWithUnixTimestamp)
	}
	return result, err
}

func (this *Mongo) CountVariables(userId string, query model.VariablesQueryOptions) (result model.Count, err error) {
	filter := bson.M{VariableBson.UserId: userId}
	if query.ProcessDefinitionId != "" {
		filter[VariableBson.ProcessDefinitionId] = query.ProcessDefinitionId
	}
	if query.ProcessInstanceId != "" {
		filter[VariableBson.ProcessInstanceId] = query.ProcessInstanceId
	}
	ctx, _ := getTimeoutContext()
	result.Count, err = this.variablesCollection().CountDocuments(ctx, filter)
	return
}

func (this *Mongo) DeleteVariablesOfProcessDefinition(definitionId string) error {
	ctx, _ := getTimeoutContext()
	_, err := this.variablesCollection().DeleteMany(ctx, bson.M{
		VariableBson.ProcessDefinitionId: definitionId,
	})
	return err
}

func (this *Mongo) DeleteVariablesOfProcessInstance(instanceId string) error {
	ctx, _ := getTimeoutContext()
	_, err := this.variablesCollection().DeleteMany(ctx, bson.M{
		VariableBson.ProcessInstanceId: instanceId,
	})
	return err
}
