/*
 * Copyright 2024 InfAI (CC SES)
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

package metrics

import (
	"encoding/json"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"log"
	"net/http"
)

func New() *Metrics {
	reg := prometheus.NewRegistry()

	result := &Metrics{
		httphandler: promhttp.HandlerFor(
			reg,
			promhttp.HandlerOpts{
				Registry: reg,
			},
		),
		writeSizes: promauto.With(reg).NewCounterVec(prometheus.CounterOpts{
			Name: "process_io_api_writes_size_sum",
			Help: "write size sum in bytes",
		}, []string{"user_id"}),
		readSizes: promauto.With(reg).NewCounterVec(prometheus.CounterOpts{
			Name: "process_io_api_read_size_sum",
			Help: "read size sum in bytes",
		}, []string{"user_id"}),
	}

	return result
}

type Metrics struct {
	httphandler http.Handler

	writeSizes *prometheus.CounterVec
	readSizes  *prometheus.CounterVec
}

func (this *Metrics) LogWriteSize(userId string, writtenElement interface{}) {
	if this == nil {
		return
	}
	buf, err := json.Marshal(writtenElement)
	if err != nil {
		log.Printf("ERROR: in LogWriteSize(): %v\n", err.Error())
	} else {
		this.writeSizes.WithLabelValues(userId).Add(float64(len(buf)))
	}
}

func (this *Metrics) LogReadSize(userId string, readElement interface{}) {
	if this == nil {
		return
	}
	buf, err := json.Marshal(readElement)
	if err != nil {
		log.Printf("ERROR: in LogWriteSize(): %v\n", err.Error())
	} else {
		this.readSizes.WithLabelValues(userId).Add(float64(len(buf)))
	}
}
