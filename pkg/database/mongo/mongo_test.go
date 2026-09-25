/*
 * Copyright (c) 2026 InfAI (CC SES)
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
	"net"
	"reflect"
	"strings"
	"sync"
	"testing"

	"github.com/SENERGY-Platform/process-io-api/pkg/configuration"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const replicaSetURL = "mongodb://mongo-0.mongo:27017,mongo-1.mongo:27017/?replicaSet=rs0&readPreference=primary"

func TestValidateConfig(t *testing.T) {
	tests := []struct {
		name    string
		cfg     configuration.Config
		wantErr error
	}{
		{"no auth", configuration.Config{MongoDatabase: "process_io"}, nil},
		{"user and password", configuration.Config{MongoDatabase: "process_io", MongoUser: "u", MongoPassword: "p"}, nil},
		{"password without user", configuration.Config{MongoDatabase: "process_io", MongoPassword: "p"}, nil},
		{"user without password", configuration.Config{MongoDatabase: "process_io", MongoUser: "u"}, errMissingPassword},
		{"empty database", configuration.Config{}, errEmptyDatabase},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := validateConfig(tt.cfg); !errors.Is(err, tt.wantErr) {
				t.Errorf("err = %v, want %v", err, tt.wantErr)
			}
		})
	}
}

func TestClientOptions_AuthWhenUserGiven(t *testing.T) {
	opts := clientOptions(configuration.Config{
		MongoUrl:        replicaSetURL,
		MongoUser:       "process-io-api",
		MongoPassword:   "s3cr3t",
		MongoAuthSource: "admin",
		MongoDatabase:   "process_io",
	})
	if err := opts.Validate(); err != nil {
		t.Fatal(err)
	}
	want := &options.Credential{Username: "process-io-api", Password: "s3cr3t", AuthSource: "admin"}
	if !reflect.DeepEqual(opts.Auth, want) {
		t.Errorf("auth = %+v, want %+v", opts.Auth, want)
	}
}

func TestClientOptions_NoAuthWhenUserEmpty(t *testing.T) {
	// A password without a user must not switch auth on.
	opts := clientOptions(configuration.Config{
		MongoUrl:        "mongodb://localhost:27017",
		MongoPassword:   "s3cr3t",
		MongoAuthSource: "admin",
		MongoDatabase:   "process_io",
	})
	if err := opts.Validate(); err != nil {
		t.Fatal(err)
	}
	if opts.Auth != nil {
		t.Errorf("auth = %+v, want nil", opts.Auth)
	}
}

func TestClientOptions_ConfiguredCredentialsReplaceURICredentials(t *testing.T) {
	opts := clientOptions(configuration.Config{
		MongoUrl:        "mongodb://old:oldpw@localhost:27017/?authSource=other&authMechanism=SCRAM-SHA-1",
		MongoUser:       "process-io-api",
		MongoPassword:   "newpw",
		MongoAuthSource: "admin",
		MongoDatabase:   "process_io",
	})
	if err := opts.Validate(); err != nil {
		t.Fatal(err)
	}
	want := &options.Credential{Username: "process-io-api", Password: "newpw", AuthSource: "admin"}
	if !reflect.DeepEqual(opts.Auth, want) {
		t.Errorf("auth = %+v, want %+v", opts.Auth, want)
	}
}

// unreachableURL points at a port that was just free, so only the startup check can fail.
func unreachableURL(t *testing.T) string {
	t.Helper()
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	addr := l.Addr().String()
	if err = l.Close(); err != nil {
		t.Fatal(err)
	}
	return "mongodb://" + addr + "/?directConnection=true"
}

// The startup check would fail as well, so these check for the specific validation error and
// confirm New never dials out for a config that is already invalid.
func TestNew_RejectsBeforeConnecting(t *testing.T) {
	cases := map[string]configuration.Config{
		"empty database":        {MongoUser: "process-io-api", MongoPassword: "s3cr3t"},
		"user without password": {MongoUser: "process-io-api", MongoDatabase: "process_io"},
	}
	want := map[string]error{"empty database": errEmptyDatabase, "user without password": errMissingPassword}
	for name, cfg := range cases {
		t.Run(name, func(t *testing.T) {
			cfg.MongoUrl = unreachableURL(t)
			wg := &sync.WaitGroup{}
			db, err := New(context.Background(), wg, cfg)
			if !errors.Is(err, want[name]) {
				t.Fatalf("err = %v, want %v", err, want[name])
			}
			if db != nil {
				t.Error("expected no db on failure")
			}
			if strings.Contains(err.Error(), "s3cr3t") {
				t.Errorf("error leaks the password: %v", err)
			}
		})
	}
}

// TestNew_StartupCheckFailsOnUnreachableServer confirms a valid but unreachable config fails
// startup instead of returning a lazily-connected client.
func TestNew_StartupCheckFailsOnUnreachableServer(t *testing.T) {
	cfg := configuration.Config{MongoUrl: unreachableURL(t), MongoDatabase: "process_io"}
	wg := &sync.WaitGroup{}
	db, err := New(context.Background(), wg, cfg)
	if err == nil {
		t.Fatal("expected an error for an unreachable server")
	}
	if !strings.Contains(err.Error(), "mongo startup check failed") {
		t.Errorf("err = %v, want it wrapped as a startup check failure", err)
	}
	if db != nil {
		t.Error("expected no db on failure")
	}
}
