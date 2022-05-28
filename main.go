package main

import (
	"context"
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
		namespace   string
		local       bool
		stats       bool
		annotations []string
		labels      []string
		resources   []string
	)

	run := func(ctx context.Context) error {
		clusterConfig := GenerateNewConfig(local)
		resources := rc.ResourcesConfig{
			Resources: ar.GetResourceList(clusterConfig, resources),
			NameSpace: namespace,
		}
		controller := rc.NewResourceController(resources, clusterConfig, annotations, labels)
		grp, ctx := errgroup.WithContext(ctx)
		grp.Go(func() error {
			return controller.Run(ctx)
		})

		grp.Go(func() error {
			mux := http.NewServeMux()
			svr := &http.Server{Addr: ":2112", Handler: mux}

			mux.Handle("/metrics", promhttp.Handler())
			mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {})
			mux.HandleFunc("/__/pprof/profile", pprof.Profile)
			mux.HandleFunc("/__/pprof/trace", pprof.Trace)
			mux.HandleFunc("/__/pprof/cmdline", pprof.Cmdline)
			mux.HandleFunc("/__/pprof/symbol", pprof.Symbol)
			mux.HandleFunc("/__/shutdown", func(w http.ResponseWriter, r *http.Request) {
				svr.Shutdown(ctx)
			})
			mux.HandleFunc("/ready", func(w http.ResponseWriter, r *http.Request) {
				if !controller.CheckCacheSync() {
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
		})
		if stats {
			grp.Go(func() error {
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
	flags.StringVar(&namespace, "namespace", v1.NamespaceAll, "Specifies the namespace that the exporter will monitor resources in, defaults to all")
	flags.BoolVar(&local, "local", false, "local or in cluster configuration")
	flags.BoolVar(&stats, "stats", false, "Show cpu and memory allocation")
	flags.StringSliceVarP(&annotations, "annotations", "A", []string{}, "annotations names to use in metric labels")
	flags.StringSliceVarP(&labels, "labels", "L", []string{}, "labels names to use in metric labels")
	flags.StringSliceVarP(&resources, "resources", "R", []string{"deployments", "ingresses"}, "Resource types to export labels and annotations")
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
