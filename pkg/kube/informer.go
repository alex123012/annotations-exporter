package kube

import (
	"context"
	"fmt"
	"time"

	"log"

	"github.com/alex123012/annotations-exporter/pkg/collector"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/dynamic/dynamicinformer"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
)

var (
	ExporterMetricName = "kube_annotations_exporter"
)

// InformerController handles Kubernetes events for resourcess. The is the shim between metrics storage and Kubernetes cluster.
type InformerController struct {
	client     dynamic.Interface
	resources  []schema.GroupVersionResource
	namespaces []string

	metricCollector *collector.MetricsVault
}

// NewResourcesInformer creates cached informer to track resources from a Kubernetes cluster.
func NewResourcesInformer(config *rest.Config, namespaces []string, resources []schema.GroupVersionResource,
	metricCollector *collector.MetricsVault) (*InformerController, error) {
	client, err := dynamic.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	return &InformerController{
		client:          client,
		metricCollector: metricCollector,
		resources:       resources,
		namespaces:      namespaces,
	}, nil
}

func (i *InformerController) storeMetric(obj interface{}) {
	resource := obj.(*unstructured.Unstructured)
	i.metricCollector.Store(ExporterMetricName, ResourceToSample(resource))
}

func (i *InformerController) addHandler() func(obj interface{}) {
	return i.storeMetric
}

func (i *InformerController) updateHandler() func(old, new interface{}) {
	return func(old, new interface{}) {
		i.storeMetric(new)
	}
}

func (i *InformerController) deleteHandler() func(obj interface{}) {
	return func(obj interface{}) {
		resource := obj.(*unstructured.Unstructured)
		i.metricCollector.Clear(ExporterMetricName, ResourceToSample(resource))
	}
}

// Run starts the informers for different resources with various handlers and waits for the first cache synchronization.
func (c *InformerController) Run(ctx context.Context, errorCh chan<- error) {
	for _, namespace := range c.namespaces {
		go c.runInformerForNamespace(ctx, namespace, errorCh)
	}
	log.Println("started")
}

func (c *InformerController) runInformerForNamespace(ctx context.Context, namespace string, errorCh chan<- error) {
	factory := dynamicinformer.NewFilteredDynamicSharedInformerFactory(c.client, time.Minute, namespace, nil)
	cacheSyncs := make([]cache.InformerSynced, len(c.resources))
	for i, resource := range c.resources {
		informer, err := c.newInformer(factory, resource, errorCh)
		if err != nil {
			errorCh <- err
		}
		cacheSyncs[i] = informer.HasSynced
	}

	factory.Start(ctx.Done())
	log.Printf("started factory for namespace '%s'", namespace)
	if ok := cache.WaitForCacheSync(ctx.Done(), cacheSyncs...); !ok {
		log.Fatal(fmt.Errorf("informer cache is not synced"))
		errorCh <- fmt.Errorf("informer cache is not synced")
	}
}

func (i *InformerController) newInformer(factory dynamicinformer.DynamicSharedInformerFactory, resource schema.GroupVersionResource, errorCh chan<- error) (cache.SharedIndexInformer, error) {
	informer := factory.ForResource(resource).Informer()
	informer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    i.addHandler(),
		UpdateFunc: i.updateHandler(),
		DeleteFunc: i.deleteHandler(),
	})
	if err := informer.SetWatchErrorHandler(func(_ *cache.Reflector, err error) {
		errorCh <- fmt.Errorf("error for resource '%v': %v", resource, err)
	}); err != nil {
		return nil, fmt.Errorf("failed to set watch error handler: %w", err)
	}

	return informer, nil
}
