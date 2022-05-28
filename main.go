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

	ar "github.com/alex123012/annotations-exporter/src/apiresources"
	rc "github.com/alex123012/annotations-exporter/src/resourcecontroller"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/spf13/cobra"
	"golang.org/x/sync/errgroup"
	v1 "k8s.io/api/core/v1"
	_ "k8s.io/client-go/plugin/pkg/client/auth/oidc"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/klog/v2"
)

func main() {
	var (
		port        int
		local       bool
		stats       bool
		namespaces  []string
		annotations []string
		labels      []string
		resources   []string
	)

	run := func(ctx context.Context) error {
		clusterConfig := GenerateNewConfig(local)
		resources := rc.ResourcesConfig{
			Resources:  ar.GetResourceList(clusterConfig, resources),
			NameSpaces: namespaces,
		}
		grp, ctx := errgroup.WithContext(ctx)
		controller := rc.NewResourceController(ctx, resources, clusterConfig, annotations, labels)
		grp.Go(func() error {
			return controller.Run(ctx)
		})

		grp.Go(func() error {
			return RunMetricsServer(ctx, controller, port)
		})
		if stats {
			grp.Go(func() error {
				return SystemStats(ctx)

			})
		}
		return grp.Wait()
	}

	cmd := &cobra.Command{
		Use:     "annoexp",
		Short:   "Export annotations and labels from k8s resources to prometheus metrics",
		Version: "0.0.2",
		Run: func(cmd *cobra.Command, args []string) {
			ctx := cmd.Context()

			if err := run(ctx); err != nil {
				klog.Exitln(err)
			}
		},
	}

	flags := cmd.PersistentFlags()
	flags.BoolVar(&local, "local", false, "Use local (kubeconfig) or in cluster(serviceaccount) configuration for connecting to cluster")
	flags.BoolVar(&stats, "stats", false, "Log cpu and memory allocation")
	flags.IntVarP(&port, "port", "p", 8888, "Port to use for metrics server")
	flags.StringSliceVarP(&annotations, "annotations", "A", []string{}, "annotations names to use in prometheus metric labels")
	flags.StringSliceVarP(&labels, "labels", "L", []string{}, "labels names to use in prometheus metric labels")
	flags.StringSliceVarP(&resources, "resources", "R", []string{"deployments", "ingresses", "pods"}, "Resource types to export labels and annotations")
	flags.StringSliceVarP(&namespaces, "namespaces", "n", []string{v1.NamespaceAll}, "Specifies the namespace that the exporter will monitor resources in, defaults to all")
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	if err := cmd.ExecuteContext(ctx); err != nil {
		klog.Exitln(err)
	}
}

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
