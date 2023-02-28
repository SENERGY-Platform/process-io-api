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

import (
	"net/url"
	"strconv"
)

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

func (this VariablesQueryOptions) Encode() string {
	values := url.Values{}
	if this.Limit > 0 {
		values["limit"] = []string{strconv.Itoa(this.Limit)}
	}
	if this.Offset > 0 {
		values["offset"] = []string{strconv.Itoa(this.Offset)}
	}
	if this.Sort != "" {
		values["sort"] = []string{this.Sort}
	}
	if this.KeyRegex != "" {
		values["key_regex"] = []string{this.KeyRegex}
	}
	if this.ProcessInstanceId != "" {
		values["process_instance_id"] = []string{this.ProcessInstanceId}
	}
	if this.ProcessDefinitionId != "" {
		values["process_definition_id"] = []string{this.ProcessDefinitionId}
	}
	return values.Encode()
}

type BulkRequest struct {
	Get []string   `json:"get"`
	Set []Variable `json:"set"`
}

type BulkResponse = []VariableWithUnixTimestamp
