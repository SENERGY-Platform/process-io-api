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
	"errors"
	"fmt"
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

var (
	errEmptyDatabase   = errors.New("mongo database name must not be empty")
	errMissingPassword = errors.New("mongo password must not be empty when a mongo user is set")
)

const startupCheckTimeout = 10 * time.Second

func New(ctx context.Context, wg *sync.WaitGroup, config configuration.Config) (*Mongo, error) {
	if err := validateConfig(config); err != nil {
		return nil, err
	}
	reg := bson.NewRegistryBuilder().RegisterTypeMapEntry(bsontype.EmbeddedDocument, reflect.TypeOf(bson.M{})).Build() //ensure map marshalling to interface
	client, err := connect(ctx, clientOptions(config).SetRegistry(reg), config.MongoDatabase, startupCheckTimeout)
	if err != nil {
		return nil, err
	}

	db := &Mongo{config: config, client: client}
	for _, creators := range CreateCollections {
		err = creators(db)
		if err != nil {
			db.disconnect()
			return nil, err
		}
	}

	wg.Add(1)
	go func() {
		<-ctx.Done()
		db.disconnect()
		wg.Done()
	}()

	return db, nil
}

// connect runs listCollections on database because Connect is lazy and ping needs no
// authentication; unreachable servers and wrong or missing credentials then fail at startup.
func connect(ctx context.Context, opts *options.ClientOptions, database string, timeout time.Duration) (*mongo.Client, error) {
	connectCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	client, err := mongo.Connect(connectCtx, opts)
	if err != nil {
		return nil, err
	}
	checkCtx, checkCancel := context.WithTimeout(ctx, timeout)
	defer checkCancel()
	listOpts := options.ListCollections().SetNameOnly(true).SetAuthorizedCollections(true)
	if _, err = client.Database(database).ListCollectionNames(checkCtx, bson.D{}, listOpts); err != nil {
		disconnectCtx, disconnectCancel := context.WithTimeout(context.Background(), timeout)
		defer disconnectCancel()
		_ = client.Disconnect(disconnectCtx)
		return nil, fmt.Errorf("mongo startup check failed: %w", err)
	}
	return client, nil
}

func validateConfig(config configuration.Config) error {
	if config.MongoDatabase == "" {
		return errEmptyDatabase
	}
	if config.MongoUser != "" && config.MongoPassword == "" {
		return errMissingPassword
	}
	return nil
}

// clientOptions applies the credentials after the URI so they replace any given in MONGO_URL.
func clientOptions(config configuration.Config) *options.ClientOptions {
	opts := options.Client().ApplyURI(config.MongoUrl)
	if config.MongoUser != "" {
		opts.SetAuth(options.Credential{
			Username:   config.MongoUser,
			Password:   config.MongoPassword,
			AuthSource: config.MongoAuthSource,
		})
	}
	return opts
}

func getTimeoutContext() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), 10*time.Second)
}

type Mongo struct {
	config configuration.Config
	client *mongo.Client
}

func (db *Mongo) disconnect() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	_ = db.client.Disconnect(ctx)
}
