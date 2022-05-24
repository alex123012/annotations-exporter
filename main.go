package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"k8s.io/client-go/kubernetes"
	_ "k8s.io/client-go/plugin/pkg/client/auth/oidc"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

const (
	INGRESS     = "Ingress"
	STATEFULSET = "StatefulSet"
	POD         = "Pod"
	DEPLOYMENT  = "Deployment"
	SERVICE     = "Service"
)

func main() {
	var local bool
	var config *rest.Config
	var err error
	var annotations, labels ArrayFlags

	flag.BoolVar(&local, "local", false, "local or in cluster")
	flag.Var(&annotations, "annotation", "annotations to export")
	flag.Var(&labels, "label", "labels to export")
	flag.Parse()
	prometheus_labels := append(annotations, labels...)

	clear_prometheus_labels := GetClearLabels(prometheus_labels)

	fmt.Println(prometheus_labels)
	fmt.Println(clear_prometheus_labels)

	if local {
		fmt.Println("Using local configuration")
		config, err = clientcmd.BuildConfigFromFlags("", filepath.Join(os.Getenv("HOME"), ".kube", "config"))
	} else {
		fmt.Println("Using in cluster configuration")
		config, err = rest.InClusterConfig()
	}
	if err != nil {
		panic(err.Error())
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Setuping prometheus")
	gaugeVec := prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "custom_app_annotations_and_labels",
		Help: "annotation and labels from kube resources",
	}, append([]string{"kind", "name", "namespace"}, clear_prometheus_labels...),
	)
	prometheus.Register(gaugeVec)
	fmt.Println("Starting scrapper")
	go GenerateMetrics(clientset, gaugeVec, prometheus_labels)
	http.Handle("/metrics", promhttp.Handler())
	http.ListenAndServe(":2112", nil)
}

func GenerateMetrics(clientset *kubernetes.Clientset, metric *prometheus.GaugeVec, prometheus_labels ArrayFlags) {
	for {
		start := time.Now()
		namespaces, err := clientset.CoreV1().Namespaces().List(context.Background(), metav1.ListOptions{})
		if err != nil {
			log.Fatal(err)
		}
		const max_goroutines = 5
		semaphore := make(chan struct{}, max_goroutines)
		wg := sync.WaitGroup{}
		for i, ns := range namespaces.Items {
			ns_name := ns.GetName()
			if strings.HasPrefix(ns_name, "d8-") || strings.HasPrefix(ns_name, "kube-") {
				continue
			}
			semaphore <- struct{}{}
			wg.Add(1)
			go func(ns_name string, i int) {
				defer wg.Done()
				resourceList := []ResourcesInterface{
					NewPod(),
					NewDeployment(),
					NewStatefulSets(),
					NewIngress(),
				}
				for _, res := range resourceList {
					res.GetObjectItems(clientset, ns_name)
					err = res.CombineLabels(metric, prometheus_labels)
					if err != nil {
						log.Fatal(err)
					}
					fmt.Println(res.GetResourceType(), len(res.GetItems()), ns_name)
				}
				<-semaphore
			}(ns_name, i)
		}
		wg.Wait()
		elapsed := time.Since(start)
		log.Printf("\ntook %s", elapsed)
	}
}
