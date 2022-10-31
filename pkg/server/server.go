// Copyright 2022.
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package server

import (
	"context"
	"net/http"

	"log"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func StartMetricsServer(ctx context.Context, address string, errorCh chan error) {
	mux := http.NewServeMux()

	mux.Handle("/metrics", promhttp.Handler())

	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`<!DOCTYPE html>
			<title>Annotations Exporter</title>
			<h1>Annotations Exporter</h1>
			<p><a href=/metrics>Metrics</a></p>`))
	})

	log.Printf("start exporting metrics on %q", address)

	srv := &http.Server{
		Addr:    address,
		Handler: mux,
	}

	go func() {
		<-ctx.Done()
		log.Println("closing metrics server ...")
		_ = srv.Shutdown(ctx)
	}()
	errorCh <- srv.ListenAndServe()
}
