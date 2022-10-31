package main

import (
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"

	v1 "k8s.io/api/core/v1"
	_ "k8s.io/client-go/plugin/pkg/client/auth/oidc"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

func GenerateNewConfig(kubeconfigPath string) (*rest.Config, error) {
	var (
		cfg *rest.Config
		err error
	)
	if kubeconfigPath != "" {
		cfg, err = clientcmd.BuildConfigFromFlags("", kubeconfigPath)
		if err != nil {
			return nil, fmt.Errorf("new kubernetes client from config %s: %w", kubeconfigPath, err)
		}
		log.Printf("using kubeconfig from file: %q", kubeconfigPath)
	} else {
		cfg, err = rest.InClusterConfig()
		switch {
		case err == nil:
			log.Printf("using in-cluster kubeconfig")
		case !errors.Is(err, rest.ErrNotInCluster):
			return nil, fmt.Errorf("new kubernetes from cluster: %w", err)
		default:
			home, _ := os.UserHomeDir()
			userKubeconfigPath := filepath.Join(home, ".kube", "config")

			cfg, err = clientcmd.BuildConfigFromFlags("", userKubeconfigPath)
			if err != nil {
				return nil, fmt.Errorf("new kubernetes client from homedir %s: %w", userKubeconfigPath, err)
			}
			log.Printf("using kubeconfig from homedir: %q", userKubeconfigPath)
		}
	}

	return cfg, nil
}

func validateNamespaces(namespaces []string) ([]string, error) {
	if len(namespaces) == 0 {
		return []string{v1.NamespaceAll}, nil
	}
	for _, namespace := range namespaces {
		if namespace == "" && len(namespaces) > 1 {
			return nil, fmt.Errorf("can't use several namespaces with all ('') namespaces specified")
		}
	}
	return namespaces, nil
}

// func RunMetricsServer(ctx context.Context, controller *rc.ResourceController, port int) error {
// 	mux := http.NewServeMux()
// 	svr := &http.Server{Addr: fmt.Sprintf(":%d", port), Handler: mux}

// 	mux.Handle("/metrics", promhttp.Handler())
// 	mux.HandleFunc("/__/pprof/profile", pprof.Profile)
// 	mux.HandleFunc("/__/pprof/trace", pprof.Trace)
// 	mux.HandleFunc("/__/pprof/cmdline", pprof.Cmdline)
// 	mux.HandleFunc("/__/pprof/symbol", pprof.Symbol)
// 	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
// 		switch controller.CheckCacheSync() {
// 		case true:
// 			w.WriteHeader(http.StatusOK)
// 		case false:
// 			w.WriteHeader(http.StatusPreconditionFailed)
// 		}
// 	})

// 	c := make(chan os.Signal)
// 	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
// 	go func() {
// 		defer svr.Shutdown(ctx)
// 		<-c
// 	}()
// 	return svr.ListenAndServe()
// }
// func SystemStats(ctx context.Context) error {

// 	for {
// 		select {
// 		default:
// 			cpu := runtime.NumCPU()
// 			log.Println("CPU:", cpu)

// 			rot := runtime.NumGoroutine()
// 			log.Println("Goroutine:", rot)

// 			mem := &runtime.MemStats{}
// 			// Byte
// 			runtime.ReadMemStats(mem)
// 			log.Println("Memory:", mem.Alloc/1024)

// 			time.Sleep(2 * time.Second)
// 			log.Println("-------")
// 		case <-ctx.Done():
// 			return nil
// 		}
// 	}
// }
