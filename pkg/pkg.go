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

package pkg

import (
	"context"
	"github.com/SENERGY-Platform/process-io-api/pkg/api"
	"github.com/SENERGY-Platform/process-io-api/pkg/configuration"
	"github.com/SENERGY-Platform/process-io-api/pkg/controller"
	"github.com/SENERGY-Platform/process-io-api/pkg/database"
	"sync"
)

func Start(ctx context.Context, wg *sync.WaitGroup, config configuration.Config) (cmd *controller.Controller, err error) {
	db, err := database.New(ctx, wg, config)
	if err != nil {
		return nil, err
	}
	cmd = controller.New(config, db)
	return cmd, api.Start(ctx, config, cmd)
}
