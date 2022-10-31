package apiresources

import (
	"context"
	"fmt"
	"strings"

	"golang.org/x/sync/errgroup"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/rest"
)

type resourceToMap struct {
	Key   schema.GroupVersionResource
	Value schema.GroupVersionResource
}

func GetAllApiResources(config *rest.Config) (map[schema.GroupVersionResource]schema.GroupVersionResource, error) {

	discoveryClient, err := discovery.NewDiscoveryClientForConfig(config)
	if err != nil {
		return nil, err
	}

	APIResourceListSlice, err := discoveryClient.ServerPreferredResources()
	if err != nil {
		return nil, err
	}
	errGroup, _ := errgroup.WithContext(context.Background())
	resourceChan := make(chan resourceToMap)
	for _, singleAPIResourceList := range APIResourceListSlice {
		apiResourcesList := singleAPIResourceList
		errGroup.Go(func() error {
			return getNamesFromResources(apiResourcesList, resourceChan)
		})
	}

	errorCh := make(chan error)
	go func() {
		defer close(resourceChan)
		errorCh <- errGroup.Wait()
	}()
	resourceKindMap := make(map[schema.GroupVersionResource]schema.GroupVersionResource)
	for {
		select {
		case err := <-errorCh:
			if err != nil {
				close(errorCh)
				return nil, err
			}
			return resourceKindMap, nil
		case resource := <-resourceChan:
			resourceKindMap[resource.Key] = resource.Value
		}
	}
}
func getNamesFromResources(apiResource *v1.APIResourceList, resourceChan chan<- resourceToMap) error {
	apiVersion, err := schema.ParseGroupVersion(apiResource.GroupVersion)
	if err != nil {
		return err
	}
	for _, resource := range apiResource.APIResources {
		groupVersionResource := schema.GroupVersionResource{
			Resource: resource.Name,
			Version:  apiVersion.Version,
			Group:    apiVersion.Group,
		}

		resourceKeyValue := resourceToMap{
			Key: schema.GroupVersionResource{
				Resource: resource.Name,
				Version:  apiVersion.Version,
				Group:    apiVersion.Group,
			},
			Value: groupVersionResource,
		}
		resourceChan <- resourceKeyValue

		resourceKeyValue.Key.Version = ""
		resourceChan <- resourceKeyValue
	}
	return nil
}

func CompareWithApiResources(config *rest.Config, flagList []string) ([]schema.GroupVersionResource, error) {
	apiResorces, err := GetAllApiResources(config)
	if err != nil {
		return nil, err
	}
	resultList := make([]schema.GroupVersionResource, len(flagList))
	for i, resource := range flagList {
		res, err := ParseResourceString(resource)
		if err != nil {
			return nil, err
		}
		if value, f := apiResorces[*res]; f {
			resultList[i] = value
			continue
		}
		return nil, fmt.Errorf("no such resource in kubernetes api: %v", res)
	}

	return resultList, nil
}

func ParseResourceString(arg string) (*schema.GroupVersionResource, error) {
	split := "/"
	resource := strings.Split(arg, split)
	if strings.Count(arg, split) == 1 {
		return &schema.GroupVersionResource{
			Resource: resource[0],
			Group:    resource[1],
		}, nil
	} else if strings.Count(arg, split) == 2 {
		return &schema.GroupVersionResource{
			Resource: resource[0],
			Version:  resource[1],
			Group:    resource[2],
		}, nil
	}
	return nil, fmt.Errorf("error parsing resource from flag: '%s'", arg)
}
