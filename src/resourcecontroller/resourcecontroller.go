// Inspired by https://blog.dsb.dev/posts/creating-dynamic-informers/
// and https://github.com/davidsbond/kollect/blob/master/internal/agent/agent.go
package resourcecontroller

import (
	"context"
	"fmt"
	"sync"
	"time"

	"golang.org/x/sync/errgroup"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/dynamic/dynamicinformer"
	_ "k8s.io/client-go/plugin/pkg/client/auth/oidc"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
	"k8s.io/klog/v2"
)

type ResourceController struct {
	sycnMutex     *sync.RWMutex
	handlerMux    *sync.Mutex
	cacheSynced   bool
	Resources     ResourcesConfig
	clusterClient dynamic.Interface
	metric        MetricInterface
}

type ResourcesConfig struct {
	Resources     []schema.GroupVersionResource
	NameSpace     string
	LogPodsStatus bool
}

func (c *ResourceController) Run(ctx context.Context) error {
	factory := dynamicinformer.NewFilteredDynamicSharedInformerFactory(c.clusterClient, time.Minute, c.Resources.NameSpace, nil)

	group, ctx := errgroup.WithContext(ctx)
	cacheSyncs := make([]cache.InformerSynced, len(c.Resources.Resources))
	for i, resource := range c.Resources.Resources {
		informer := factory.ForResource(resource).Informer()

		cacheSyncs[i] = informer.HasSynced
		handler := c.informerHandler(ctx, informer)
		group.Go(handler)
	}
	isSynced := cache.WaitForCacheSync(ctx.Done(), cacheSyncs...)
	c.sycnMutex.Lock()
	c.cacheSynced = isSynced
	c.sycnMutex.Unlock()

	if !c.CheckCacheSync() {
		klog.Fatal("failed to sync cache")
	}

	return group.Wait()
}

func (c *ResourceController) CheckCacheSync() bool {
	c.sycnMutex.RLock()
	defer c.sycnMutex.RUnlock()
	return c.cacheSynced
}
func (c *ResourceController) addHandler() func(obj interface{}) {
	return func(obj interface{}) {
		if !c.CheckCacheSync() {
			return
		}

		c.handlerMux.Lock()
		defer c.handlerMux.Unlock()

		res := obj.(*unstructured.Unstructured)

		klog.Infof("Add %s %s in namespace %s", res.GetKind(), res.GetName(), res.GetNamespace())
		generatedLabels := c.metric.ExportLabelsAndAnnotations(res)
		c.metric.ExportMetricForResource(generatedLabels, 1)
	}
}

func (c *ResourceController) updateHandler() func(oldObj interface{}, newObj interface{}) {
	return func(oldObj interface{}, newObj interface{}) {
		if !c.CheckCacheSync() {
			return
		}

		c.handlerMux.Lock()
		defer c.handlerMux.Unlock()

		resOld := oldObj.(*unstructured.Unstructured)
		resNew := newObj.(*unstructured.Unstructured)

		klog.Errorln(fmt.Sprintf("Update %s %s to -> %s %s in namespace %s(%s)", resOld.GetKind(), resOld.GetName(), resNew.GetKind(), resNew.GetName(), resNew.GetNamespace(), resOld.GetNamespace()))
		generatedOldLabels := c.metric.ExportLabelsAndAnnotations(resOld)
		generatedNewLabels := c.metric.ExportLabelsAndAnnotations(resNew)
		c.metric.DeletetMetricForResource(generatedOldLabels)
		c.metric.ExportMetricForResource(generatedNewLabels, 1)
	}
}

func (c *ResourceController) deleteHandler() func(obj interface{}) {
	return func(obj interface{}) {
		if !c.CheckCacheSync() {
			return
		}

		c.handlerMux.Lock()
		defer c.handlerMux.Unlock()

		res := obj.(*unstructured.Unstructured)
		klog.Infof("Delete %s %s in namespace %s", res.GetKind(), res.GetName(), res.GetNamespace())
		generatedLabels := c.metric.ExportLabelsAndAnnotations(res)
		c.metric.DeletetMetricForResource(generatedLabels)
	}
}

func (a *ResourceController) informerHandler(ctx context.Context, informer cache.SharedIndexInformer) func() error {
	return func() error {
		ctx, cancel := context.WithCancel(ctx)
		defer cancel()

		informer.AddEventHandler(cache.ResourceEventHandlerFuncs{
			AddFunc:    a.addHandler(),
			UpdateFunc: a.updateHandler(),
			DeleteFunc: a.deleteHandler(),
		})
		err := informer.SetWatchErrorHandler(func(_ *cache.Reflector, err error) {
			klog.Errorln(err)
			cancel()
		})
		if err != nil {
			return fmt.Errorf("failed to set watch error handler: %w", err)
		}

		go informer.Run(ctx.Done())
		<-ctx.Done()

		return nil
	}
}

func NewResourceController(resources ResourcesConfig, clusterConfig *rest.Config, annotations, labels []string) *ResourceController {
	clusterClient, err := dynamic.NewForConfig(clusterConfig)

	if err != nil {
		klog.Exitln(err)
	}

	metric := NewGaugeMetric(annotations, labels)

	var allResourcesNow []unstructured.Unstructured
	for _, res := range resources.Resources {
		list, err := clusterClient.Resource(res).Namespace(resources.NameSpace).List(context.Background(), metav1.ListOptions{})
		if err != nil {
			return nil
		}
		allResourcesNow = append(allResourcesNow, list.Items...)
	}
	for _, res := range allResourcesNow {
		generatedLabels := metric.ExportLabelsAndAnnotations(&res)
		metric.ExportMetricForResource(generatedLabels, 0)
	}
	klog.Infoln("Generated current resource list")
	return &ResourceController{
		sycnMutex:     &sync.RWMutex{},
		handlerMux:    &sync.Mutex{},
		Resources:     resources,
		clusterClient: clusterClient,
		metric:        metric,
	}
}
