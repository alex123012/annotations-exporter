package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/alex123012/annotations-exporter/pkg/apiresources"
	"github.com/alex123012/annotations-exporter/pkg/collector"
	"github.com/alex123012/annotations-exporter/pkg/kube"
	"github.com/alex123012/annotations-exporter/pkg/server"
	"github.com/spf13/cobra"
	v1 "k8s.io/api/core/v1"
	_ "k8s.io/client-go/plugin/pkg/client/auth/oidc"
)

var (
	exporterAddress string   = ":8000"
	namespaces      []string = []string{v1.NamespaceAll}
	annotations     []string
	labels          []string
	resources       []string = []string{
		"deployments/apps",
		"ingresses/v1/networking.k8s.io",
		"statefulsets/apps",
		"daemonsets/apps",
	}
	maxRevisions int = 3
	logLevel     string
	kubeconfig   string

	onlyLabelsAndAnnotations bool
	referenceAnnotations     []string
	referenceLabels          []string
)

func main() {

	cmd := &cobra.Command{
		Use:     "annotations-exporter",
		Short:   "Export annotations and labels from k8s resources to prometheus metrics",
		Version: "0.1.0",
		Run: func(cmd *cobra.Command, args []string) {
			if err := Run(cmd.Context()); err != nil {
				log.Fatal(err)
			}
		},
	}

	flags := cmd.PersistentFlags()
	flags.StringVar(&exporterAddress, "server.exporter-address", exporterAddress, "Address to export prometheus metrics")
	flags.StringVar(&logLevel, "server.log-level", logLevel, "Log level")
	flags.StringSliceVar(&annotations, "kube.annotations", annotations, "Annotations names to use in prometheus metric labels")
	flags.StringSliceVar(&labels, "kube.labels", labels, "Labels names to use in prometheus metric labels")
	flags.StringSliceVar(&resources, "kube.resources", resources, "Resources (<resource>/<version>/<api> or <resource>/<api>) to export labels and annotations")
	flags.StringSliceVar(&namespaces, "kube.namespaces", namespaces, "Specifies the namespace that the exporter will monitor resources in (default 'all namespaces')")
	flags.IntVar(&maxRevisions, "kube.max-revisions", maxRevisions, "Max revisions of resource labels to store")
	flags.StringVar(&kubeconfig, "kube.config", kubeconfig, "Path to kubeconfig (optional)")
	flags.StringSliceVar(&referenceAnnotations, "kube.reference-annotations", referenceAnnotations, "Annotations names to use in prometheus metric labels and for count revisions (reference names)")
	flags.StringSliceVar(&referenceLabels, "kube.reference-labels", referenceLabels, "Labels names to use in prometheus metric labels and for count revisions (reference names)")
	flags.BoolVar(&onlyLabelsAndAnnotations, "kube.only-labels-and-annotations", onlyLabelsAndAnnotations, "Export only labels and annotations defined by flags (default false)")

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()
	if err := cmd.ExecuteContext(ctx); err != nil {
		log.Fatal(err)
	}
}

func Run(ctx context.Context) error {
	namespaces, err := validateNamespaces(namespaces)
	if err != nil {
		return err
	}

	clusterConfig, err := GenerateNewConfig(kubeconfig)
	if err != nil {
		return err
	}
	apiResources, err := apiresources.CompareWithApiResources(clusterConfig, resources)
	if err != nil {
		return err
	}
	for _, res := range apiResources {
		log.Printf("Starting watching for resource: %s", res.String())
	}

	metricVault := collector.NewVault()
	if err := metricVault.RegisterMappings([]collector.Mapping{
		kube.ResourceMapping(labels, annotations, maxRevisions,
			onlyLabelsAndAnnotations, referenceLabels, referenceAnnotations),
	}); err != nil {
		log.Fatal(err)
	}

	errorCh := make(chan error)

	informerController, err := kube.NewResourcesInformer(clusterConfig, namespaces, apiResources, metricVault)
	if err != nil {
		log.Fatalf("kubernetes informer: %v", err)
	}

	go server.StartMetricsServer(ctx, exporterAddress, errorCh)

	go informerController.Run(ctx, errorCh)

	for {
		select {
		case s := <-ctx.Done():
			log.Printf("signal received: %v, exiting...", s)
			return nil
		case err := <-errorCh:
			log.Fatalf("error received: %v", err)
			return err
		}
	}
}
