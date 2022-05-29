package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"net/http/pprof"
	"os"
	"os/signal"
	"path/filepath"
	"runtime"
	"syscall"
	"time"

	rc "github.com/alex123012/annotations-exporter/pkg/resourcecontroller"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	_ "k8s.io/client-go/plugin/pkg/client/auth/oidc"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/klog/v2"
)

func GenerateNewConfig(local bool) *rest.Config {
	var clusterConfig *rest.Config
	var err error

	switch local {
	case true:
		clusterConfig, err = clientcmd.BuildConfigFromFlags("", filepath.Join(os.Getenv("HOME"), ".kube", "config"))
	case false:
		clusterConfig, err = rest.InClusterConfig()
	}
	if err != nil {
		klog.Exitln(err)
	}

	return clusterConfig
}

func RunMetricsServer(ctx context.Context, controller *rc.ResourceController, port int) error {
	mux := http.NewServeMux()
	svr := &http.Server{Addr: fmt.Sprintf(":%d", port), Handler: mux}

	mux.Handle("/metrics", promhttp.Handler())
	mux.HandleFunc("/__/pprof/profile", pprof.Profile)
	mux.HandleFunc("/__/pprof/trace", pprof.Trace)
	mux.HandleFunc("/__/pprof/cmdline", pprof.Cmdline)
	mux.HandleFunc("/__/pprof/symbol", pprof.Symbol)
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		switch controller.CheckCacheSync() {
		case true:
			w.WriteHeader(http.StatusOK)
		case false:
			w.WriteHeader(http.StatusPreconditionFailed)
		}
	})

	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		defer svr.Shutdown(ctx)
		<-c
	}()
	return svr.ListenAndServe()
}
func SystemStats(ctx context.Context) error {
	mem := &runtime.MemStats{}
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	go func() {
		for {
			cpu := runtime.NumCPU()
			log.Println("CPU:", cpu)

			rot := runtime.NumGoroutine()
			log.Println("Goroutine:", rot)

			// Byte
			runtime.ReadMemStats(mem)
			log.Println("Memory:", mem.Alloc/1024)

			time.Sleep(2 * time.Second)
			log.Println("-------")
		}
	}()
	<-ctx.Done()
	return nil
}
