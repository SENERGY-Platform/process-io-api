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

package model

type Count struct {
	Count int64 `json:"count"`
}

type Variable struct {
	Key                 string      `json:"key"`
	Value               interface{} `json:"value"`
	ProcessDefinitionId string      `json:"process_definition_id,omitempty"`
	ProcessInstanceId   string      `json:"process_instance_id,omitempty"`
}

type VariableWithUnixTimestamp struct {
	Variable
	UnixTimestampInS int64 `json:"unix_timestamp_in_s"`
}

type VariableWithUser struct {
	VariableWithUnixTimestamp
	UserId string `json:"user_id"`
}

type VariablesQueryOptions struct {
	Limit               int
	Offset              int
	Sort                string
	KeyRegex            string
	ProcessDefinitionId string
	ProcessInstanceId   string
}

func (this VariablesQueryOptions) GetLimit() int64 {
	return int64(this.Limit)
}

func (this VariablesQueryOptions) GetOffset() int64 {
	return int64(this.Offset)
}

func (this VariablesQueryOptions) GetSort() string {
	if this.Sort == "" {
		return "key.asc"
	}
	return this.Sort
}

type BulkRequest struct {
	Get []string   `json:"get"`
	Set []Variable `json:"set"`
}

type BulkResponse = []VariableWithUnixTimestamp
