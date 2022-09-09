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
	"context"
	"database/sql"
	_ "github.com/lib/pq"
	"process-io-api/pkg/configuration"
	"sync"
	"time"
)

var CreateTable = []func(db *Pg) error{}

func New(ctx context.Context, wg *sync.WaitGroup, config configuration.Config) (pg *Pg, err error) {
	pg = &Pg{}
	pg.db, err = sql.Open("postgres", config.PostgresConnString)
	if err != nil {
		return pg, err
	}

	for _, creators := range CreateTable {
		err = creators(pg)
		if err != nil {
			pg.db.Close()
			return nil, err
		}
	}

	wg.Add(1)
	go func() {
		<-ctx.Done()
		pg.db.Close()
		wg.Done()
	}()
	return pg, nil
}

type Pg struct {
	config configuration.Config
	db     *sql.DB
}

func getTimeoutContext() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), 10*time.Second)
}
