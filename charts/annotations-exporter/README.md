# annotations-exporter

![Version: 0.5.0](https://img.shields.io/badge/Version-0.5.0-informational?style=flat-square) ![Type: application](https://img.shields.io/badge/Type-application-informational?style=flat-square)

Prometheus-exporter, which converts any Kubernetes resources annotations and labels to Prometheus samples.

## Requirements

Kubernetes: `>=1.10.0-0`

## Values

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| replicaCount | int | `1` | Number of replicas (pods) to launch. |
| image.repository | string | `"ghcr.io/alex123012/annotations-exporter"` | Name of the image repository to pull the container image from. |
| image.pullPolicy | string | `"IfNotPresent"` | [Image pull policy](https://kubernetes.io/docs/concepts/containers/images/#updating-images) for updating already existing images on a node. |
| image.tag | string | `"v0.5.0"` | Image tag override for the default value (chart appVersion). |
| cmdArgs."kube.resources" | list | `["deployments/apps","ingresses/v1/networking.k8s.io","statefulsets/apps","daemonsets/apps"]` | Kubernetes resources for annotations and labels to export. |
| cmdArgs."kube.annotations" | list | `[]` | Kubernetes resources annotations to export. |
| cmdArgs."kube.labels" | list | `[]` | Kubernetes resources labels to export. |
| cmdArgs."kube.namespaces" | list | `[""]` | Kubernetes namespaces to watch. |
| cmdArgs."kube.max-revisions" | int | `3` | Max revisions of resource labels to store. |
| cmdArgs."server.log-level" | string | `"debug"` | Log level. |
| cmdArgs."kube.only-labels-and-annotations" | bool | `false` | Export only labels and annotations defined by flags (usefull for exposing ) |
| cmdArgs."kube.reference-labels" | list | `[]` | Labels names to use in prometheus metric labels and for count revisions (reference name) |
| cmdArgs."kube.reference-annotations" | list | `[]` | Annotations names to use in prometheus metric labels and for count revisions (reference name) |
| imagePullSecrets | list | `[]` | Reference to one or more secrets to be used when [pulling images](https://kubernetes.io/docs/tasks/configure-pod-container/pull-image-private-registry/#create-a-pod-that-uses-your-secret) (from private registries). |
| nameOverride | string | `""` | A name in place of the chart name for `app:` labels. |
| fullnameOverride | string | `""` | A name to substitute for the full names of resources. |
| hostAliases | list | `[]` | A list of hosts and IPs that will be injected into the pod's hosts file if specified. See the [API reference](https://kubernetes.io/docs/reference/kubernetes-api/workload-resources/pod-v1/#hostname-and-name-resolution) |
| envFrom | list | `[]` | Additional environment variables mounted from [secrets](https://kubernetes.io/docs/concepts/configuration/secret/#using-secrets-as-environment-variables) or [config maps](https://kubernetes.io/docs/tasks/configure-pod-container/configure-pod-configmap/#configure-all-key-value-pairs-in-a-configmap-as-container-environment-variables). See the [API reference](https://kubernetes.io/docs/reference/kubernetes-api/workload-resources/pod-v1/#environment-variables) for details. |
| env | object | `{}` | Additional environment variables passed directly to containers. See the [API reference](https://kubernetes.io/docs/reference/kubernetes-api/workload-resources/pod-v1/#environment-variables) for details. |
| envVars | list | `[]` | Similar to env but with support for all possible configurations. See the [API reference](https://kubernetes.io/docs/reference/kubernetes-api/workload-resources/pod-v1/#environment-variables) for details. |
| podAnnotations | object | `{}` | Annotations to be added to pods. |
| podDisruptionBudget.enabled | bool | `false` | Enable a [pod distruption budget](https://kubernetes.io/docs/tasks/run-application/configure-pdb/) to help dealing with [disruptions](https://kubernetes.io/docs/concepts/workloads/pods/disruptions/). It is **highly recommended** for webhooks as disruptions can prevent launching new pods. |
| podDisruptionBudget.minAvailable | int/percentage | `nil` | Number or percentage of pods that must remain available. |
| podDisruptionBudget.maxUnavailable | int/percentage | `nil` | Number or percentage of pods that can be unavailable. |
| priorityClassName | string | `""` | Specify a priority class name to set [pod priority](https://kubernetes.io/docs/concepts/scheduling-eviction/pod-priority-preemption/#pod-priority). |
| podSecurityContext | object | `{}` | Pod [security context](https://kubernetes.io/docs/tasks/configure-pod-container/security-context/#set-the-security-context-for-a-pod). See the [API reference](https://kubernetes.io/docs/reference/kubernetes-api/workload-resources/pod-v1/#security-context) for details. |
| securityContext | object | `{}` | Container [security context](https://kubernetes.io/docs/tasks/configure-pod-container/security-context/#set-the-security-context-for-a-container). See the [API reference](https://kubernetes.io/docs/reference/kubernetes-api/workload-resources/pod-v1/#security-context-1) for details. |
| resources | object | No requests or limits. | Container resource [requests and limits](https://kubernetes.io/docs/concepts/configuration/manage-resources-containers/). See the [API reference](https://kubernetes.io/docs/reference/kubernetes-api/workload-resources/pod-v1/#resources) for details. |
| nodeSelector | object | `{}` | [Node selector](https://kubernetes.io/docs/concepts/scheduling-eviction/assign-pod-node/#nodeselector) configuration. |
| tolerations | list | `[]` | [Tolerations](https://kubernetes.io/docs/concepts/scheduling-eviction/taint-and-toleration/) for node taints. See the [API reference](https://kubernetes.io/docs/reference/kubernetes-api/workload-resources/pod-v1/#scheduling) for details. |
| affinity | object | `{}` | [Affinity](https://kubernetes.io/docs/concepts/scheduling-eviction/assign-pod-node/#affinity-and-anti-affinity) configuration. See the [API reference](https://kubernetes.io/docs/reference/kubernetes-api/workload-resources/pod-v1/#scheduling) for details. |
| strategy | object | `{}` | Deployment [strategy](https://kubernetes.io/docs/concepts/workloads/controllers/deployment/#strategy) configuration. |
| service.annotations | object | `{}` | Annotations to be added to the service. |
| service.labels | object | `{}` | Labels to be added to the service. |
| service.type | string | `"ClusterIP"` | Kubernetes [service type](https://kubernetes.io/docs/concepts/services-networking/service/#publishing-services-service-types). |
| service.clusterIP | string | `""` | Internal cluster service IP (when applicable) |
| service.ports.port | int | `8000` | HTTP service port |
| service.ports.nodePort | int | `nil` | HTTP node port (when applicable) |

----------------------------------------------
Autogenerated from chart metadata using [helm-docs v1.5.0](https://github.com/norwoodj/helm-docs/releases/v1.5.0)