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

package postgres

import (
	"database/sql"
	"encoding/json"
	"process-io-api/pkg/model"
	"strconv"
	"strings"
)

const createVariablesTableSql = `CREATE TABLE IF NOT EXISTS variables (
    user_id VARCHAR ( 50 ) NOT NULL,
    variable_key VARCHAR ( 255 ) NOT NULL,
    process_definition_id VARCHAR ( 64 ),
    process_instance_id VARCHAR ( 64 ),
    unix_timestamp_in_s INT,
    variable_value json,
    PRIMARY KEY (user_id, variable_key)
);`

const createVariablesIndexesSql = `
CREATE INDEX IF NOT EXISTS variable_process_definition ON variables (process_definition_id);
CREATE INDEX IF NOT EXISTS variable_process_instance ON variables (process_instance_id);
`

func init() {
	CreateTable = append(CreateTable, func(db *Pg) error {
		ctx, _ := getTimeoutContext()
		_, err := db.db.ExecContext(ctx, createVariablesTableSql)
		if err != nil {
			return err
		}
		_, err = db.db.ExecContext(ctx, createVariablesIndexesSql)
		if err != nil {
			return err
		}
		return nil
	})
}

const getVariableSql = `SELECT user_id, variable_key, process_definition_id, process_instance_id, unix_timestamp_in_s, variable_value  FROM variables WHERE user_id = $1 AND variable_key = $2`

func (this *Pg) GetVariable(userId string, key string) (result model.VariableWithUser, err error) {
	ctx, _ := getTimeoutContext()
	var jsonValue []byte
	err = this.db.QueryRowContext(ctx, getVariableSql, userId, key).Scan(
		&result.UserId,
		&result.Key,
		&result.ProcessDefinitionId,
		&result.ProcessInstanceId,
		&result.UnixTimestampInS,
		&jsonValue,
	)
	if err == sql.ErrNoRows {
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
		return result, err
	}
	err = json.Unmarshal(jsonValue, &result.Value)
	return result, err
}

const setVariableSql = `
INSERT INTO variables (user_id, variable_key, process_definition_id, process_instance_id, unix_timestamp_in_s, variable_value) 
VALUES ($1, $2, $3, $4, $5, $6)
ON CONFLICT (user_id, variable_key) DO UPDATE 
  SET process_definition_id = excluded.process_definition_id, 
      process_instance_id = excluded.process_instance_id,
      unix_timestamp_in_s = excluded.unix_timestamp_in_s,
      variable_value = excluded.variable_value;
`

func (this *Pg) SetVariable(variable model.VariableWithUser) error {
	jsonValue, err := json.Marshal(variable.Value)
	if err != nil {
		return err
	}
	ctx, _ := getTimeoutContext()
	_, err = this.db.ExecContext(ctx, setVariableSql,
		variable.UserId,
		variable.Key,
		variable.ProcessDefinitionId,
		variable.ProcessInstanceId,
		variable.UnixTimestampInS,
		jsonValue,
	)
	return err
}

const deleteVariableSql = `DELETE FROM variables WHERE user_id = $1 AND variable_key = $2;`

func (this *Pg) DeleteVariable(userId string, key string) error {
	ctx, _ := getTimeoutContext()
	_, err := this.db.ExecContext(ctx, deleteVariableSql,
		userId,
		key,
	)
	return err
}

func (this *Pg) ListVariables(userId string, query model.VariablesQueryOptions) (result []model.VariableWithUnixTimestamp, err error) {
	sqlQueryParts := []string{"SELECT variable_key, process_definition_id, process_instance_id, unix_timestamp_in_s, variable_value FROM variables"}
	args := []interface{}{
		userId,
	}

	sqlQueryParts = append(sqlQueryParts, "WHERE user_id = $1")
	if query.ProcessDefinitionId != "" {
		sqlQueryParts = append(sqlQueryParts, "process_definition_id = $"+(strconv.Itoa(len(args)+1)))
		args = append(args, query.ProcessDefinitionId)
	}
	if query.ProcessInstanceId != "" {
		sqlQueryParts = append(sqlQueryParts, "process_instance_id = $"+(strconv.Itoa(len(args)+1)))
		args = append(args, query.ProcessInstanceId)
	}

	if query.Sort != "" {
		sortField := ""
		sortDir := ""
		if strings.HasSuffix(query.Sort, ".asc") {
			sortField = strings.TrimSuffix(query.Sort, ".asc")
			sortDir = "ASC"
		} else if strings.HasSuffix(query.Sort, ".desc") {
			sortField = strings.TrimSuffix(query.Sort, ".desc")
			sortDir = "DESC"
		} else {
			sortField = query.Sort
		}
		switch sortField {
		case "key":
			sortField = "variable_key"
		}
		sqlQueryParts = append(sqlQueryParts, "ORDER BY "+sortField)
		if sortDir != "" {
			sqlQueryParts = append(sqlQueryParts, sortDir)
		}
	}

	if query.Limit > 0 {
		sqlQueryParts = append(sqlQueryParts, "Limit $"+(strconv.Itoa(len(args)+1))+" OFFSET $"+(strconv.Itoa(len(args)+2)))
		args = append(args, query.Limit, query.Offset)
	}

	sqlQuery := strings.Join(sqlQueryParts, " ")
	rows, err := this.db.Query(sqlQuery, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		element := model.VariableWithUnixTimestamp{}
		var jsonValue []byte
		err = rows.Scan(&element.Key,
			&element.ProcessDefinitionId,
			&element.ProcessInstanceId,
			&element.UnixTimestampInS,
			&jsonValue)
		if err != nil {
			return nil, err
		}
		err = json.Unmarshal(jsonValue, &element.Value)
		if err != nil {
			return nil, err
		}
		result = append(result, element)
	}
	return result, nil
}

const deleteProcessDefinitionSql = `DELETE FROM variables WHERE process_definition_id = $1;`

func (this *Pg) DeleteVariablesOfProcessDefinition(definitionId string) error {
	ctx, _ := getTimeoutContext()
	_, err := this.db.ExecContext(ctx, deleteProcessDefinitionSql, definitionId)
	return err
}

const deleteProcessInstanceSql = `DELETE FROM variables WHERE process_instance_id = $1;`

func (this *Pg) DeleteVariablesOfProcessInstance(instanceId string) error {
	ctx, _ := getTimeoutContext()
	_, err := this.db.ExecContext(ctx, deleteProcessInstanceSql, instanceId)
	return err
}
