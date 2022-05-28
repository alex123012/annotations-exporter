package apiresources

import (
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/rest"
	"k8s.io/klog/v2"
)

type ApiResource struct {
	Resource schema.GroupVersionResource
	AllNames []string
}

type AllResources struct {
	allResources []ApiResource
	mapResource  map[string]schema.GroupVersionResource
}

func (a *AllResources) getAllApiResources(config *rest.Config) {
	discoveryClient, err := discovery.NewDiscoveryClientForConfig(config)

	if err != nil {
		panic(err.Error())
	}
	_, APIResourceListSlice, err := discoveryClient.ServerGroupsAndResources()
	if err != nil {
		panic(err.Error())
	}
	a.allResources = make([]ApiResource, 0)
	for _, singleAPIResourceList := range APIResourceListSlice {

		gv, err := schema.ParseGroupVersion(singleAPIResourceList.GroupVersion)
		if err != nil {
			panic(err.Error())
		}

		for _, resource := range singleAPIResourceList.APIResources {

			a.allResources = append(a.allResources, ApiResource{
				Resource: schema.GroupVersionResource{
					Resource: resource.Name,
					Group:    gv.Group,
					Version:  gv.Version,
				},
				AllNames: append(resource.ShortNames, resource.Name),
			})
		}
	}
}

func (a *AllResources) generateMap() {
	a.mapResource = make(map[string]schema.GroupVersionResource)
	for _, resource := range a.allResources {
		for _, name := range resource.AllNames {
			a.mapResource[name] = resource.Resource
		}
	}
}
func GetResourceList(config *rest.Config, resources []string) []schema.GroupVersionResource {
	a := AllResources{}
	a.getAllApiResources(config)
	a.generateMap()

	checkUniq := make(map[schema.GroupVersionResource]int)
	resultList := make([]schema.GroupVersionResource, 0)
	for _, res := range resources {
		switch value, f := a.mapResource[res]; f {
		case true:
			switch _, notUniq := checkUniq[value]; notUniq {
			case true:
				klog.Warningln("Using different names for one resource type")
			case false:
				resultList = append(resultList, value)
				checkUniq[value] = 1
			}
		case false:
			klog.Exitf("Resource %s can't be found in any api groups", res)
		}
	}
	klog.Infoln("Resources to watch for:", resultList)
	return resultList
}
