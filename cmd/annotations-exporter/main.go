package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	ar "github.com/alex123012/annotations-exporter/pkg/apiresources"
	rc "github.com/alex123012/annotations-exporter/pkg/resourcecontroller"
	"github.com/spf13/cobra"
	"golang.org/x/sync/errgroup"
	v1 "k8s.io/api/core/v1"
	_ "k8s.io/client-go/plugin/pkg/client/auth/oidc"
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
		klog.Infoln("***Namespaces to watch:")
		for _, ns := range namespaces {
			klog.Infof(ns)
		}
		klog.Infoln()
		klog.Infoln("***Objects to watch***")
		for _, res := range resources.Resources {
			klog.Infof("apiVersion: %s/%s, kind: %s", res.Group, res.Version, res.Resource)
		}
		klog.Infoln()
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
