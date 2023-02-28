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
	"github.com/SENERGY-Platform/process-io-api/pkg/configuration"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/bsontype"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"reflect"
	"sync"
	"time"
)

var CreateCollections = []func(db *Mongo) error{}

func New(ctx context.Context, wg *sync.WaitGroup, config configuration.Config) (*Mongo, error) {
	connectCtx, _ := context.WithTimeout(ctx, 10*time.Second)
	reg := bson.NewRegistryBuilder().RegisterTypeMapEntry(bsontype.EmbeddedDocument, reflect.TypeOf(bson.M{})).Build() //ensure map marshalling to interface
	client, err := mongo.Connect(connectCtx, options.Client().ApplyURI(config.MongoUrl), options.Client().SetRegistry(reg))
	if err != nil {
		return nil, err
	}

	db := &Mongo{config: config, client: client}
	for _, creators := range CreateCollections {
		err = creators(db)
		if err != nil {
			client.Disconnect(context.Background())
			return nil, err
		}
	}

	wg.Add(1)
	go func() {
		<-ctx.Done()
		client.Disconnect(nil)
		wg.Done()
	}()

	return db, nil
}

func getTimeoutContext() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), 10*time.Second)
}

type Mongo struct {
	config configuration.Config
	client *mongo.Client
}
