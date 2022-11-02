# Annotations Exporter

## Overview

Prometheus-exporter, which converts any Kubernetes resources annotations and labels to Prometheus samples.

## Usage
```text
Usage of annotations-exporter:
  -h, --help                                 help for annotations-exporter

      --kube.annotations strings             Annotations names to use in prometheus metric labels

      --kube.config string                   Path to kubeconfig (optional)

      --kube.labels strings                  Labels names to use in prometheus metric labels

      --kube.max-revisions int               Max revisions of resource labels to store (default 3)

      --kube.namespaces strings              Specifies the namespace that the exporter will monitor resources in (default 'all namespaces')

      --kube.only-labels-and-annotations     Export only labels and annotations defined by flags (default false)

      --kube.reference-annotations strings   Annotations names to use in prometheus metric labels and for count revisions (reference names)

      --kube.reference-labels strings        Labels names to use in prometheus metric labels and for count revisions (reference names)

      --kube.resources strings               Resources (<resource>/<version>/<api> or <resource>/<api>) to export labels and annotations (default [deployments/apps,ingresses/v1/networking.k8s.io,statefulsets/apps,daemonsets/apps])

      --server.exporter-address string       Address to export prometheus metrics (default ":8000")

      --server.log-level string              Log level

  -v, --version                              version for annotations-exporter
```

## Install

### Docker Container

Ready-to-use Docker images are [available on GitHub](https://github.com/alex123012/annotations-exporter/pkgs/container/annotations-exporter).

```bash
docker pull ghcr.io/alex123012/annotations-exporter:latest
```

### Helm Chart

The first version of helm chart is available.
1. Follow the instruction from [artifacthub](https://artifacthub.io/packages/helm/annotations-exporter/annotations-exporter) to install the chart
2. After the installation, metrics will be available on address `http://annotations-exporter.annotations-exporter:8000/metrics`

## How it works

Exporter will create one metric `kube_annotations_exporter` with constant labels:
* annotations_exporter_api_version - kubernetes resource apiVersion
* annotations_exporter_kind - kubernetes resource Kind
* annotations_exporter_namespace -  - kubernetes resource namespace
* annotations_exporter_name - kubernetes resource name

Other labels would be specified resource labels and annotations. For all specifeied kubernetes annotations and labels corresponding prometheus label would be lowercased and have replaced `/`, `-`, `.` symbols for `_`.

All kubernetes labels would have prefix `annotations_exporter_label_` in corresponding prometheus label and all kubernetes annotations would have prefix `annotations_exporter_annotation_`.
For example, if we run annotations exporter with flags:
`--kube.annotations=ci.werf.io/commit,gitlab.ci.werf.io/pipeline-url` and `--kube.labels=app`, it will create, for example for deployment, metrics like this:
```text
kube_annotations_exporter{annotations_exporter_api_version="apps/v1",annotations_exporter_kind="Deployment",annotations_exporter_annotation_ci_werf_io_commit="<deployment-annotation-value>",annotations_exporter_annotation_gitlab_ci_werf_io_pipeline_url="<deployment-annotation-value>",annotations_exporter_label_app="<deployment-label-value>",annotations_exporter_name="<deployment-name>",annotations_exporter_namespace="<deployment-namespace>",annotations_exporter_revision="0"}
```

Also, there is additional label `revision` - this label is for storing older values of annotation values.

If we update some resource and provided label or annotation changes - older metric would change label `revision` from `0` to `1` and new combination of annotations and labels would be stored with `revision` `0`. How many revisions are stored is controlled by `--kube.max-revisions` flag (defaults to 3). for example we have in cluster deployment `nginx` in namespace `default` with label `app=nginx` and ran annotations-exporter like this:
```bash
./annotations-exporter --kube.labels=app --kube.max-revisions=2 --kube.resources=deployments/apps
```

now metric for `nginx` deployment would be:

```text
kube_annotations_exporter{annotations_exporter_api_version="apps/v1",annotations_exporter_kind="Deployment",annotations_exporter_label_app="nginx",annotations_exporter_name="nginx",annotations_exporter_namespace="default",revision="0"}
```

if we update label `app` to `nginx-external` in `nginx` deployment, there would be one more metric for this deployment:
```text
kube_annotations_exporter{annotations_exporter_api_version="apps/v1",annotations_exporter_kind="Deployment",annotations_exporter_label_app="nginx",annotations_exporter_name="nginx",annotations_exporter_namespace="default",revision="1"}

kube_annotations_exporter{annotations_exporter_api_version="apps/v1",annotations_exporter_kind="Deployment",annotations_exporter_label_app="nginx-external",annotations_exporter_name="nginx",annotations_exporter_namespace="default",annotations_exporter_revision="0"}
```

Now, if we another time update label `app`, for example, to `nginx-app`, metrics would be:
```text
kube_annotations_exporter{annotations_exporter_api_version="apps/v1",annotations_exporter_kind="Deployment",annotations_exporter_label_app="nginx-app",annotations_exporter_name="nginx",annotations_exporter_namespace="default",revision="0"}

kube_annotations_exporter{annotations_exporter_api_version="apps/v1",annotations_exporter_kind="Deployment",annotations_exporter_label_app="nginx-external",annotations_exporter_name="nginx",namespace="default",annotations_exporter_revision="1"}
```

etc.

### Only Labels And Annotations
if `--kube.only-labels-and-annotations` flag provided - exporter won't collect resource meta for metrics (`apiVersion`, `kind`, `name`, `namespace`) and will only expose collected annotations and labels. This is useful when combined with `--kube.reference-annotations` and `--kube.reference-labels` to expose, for example, helm release name and namespace:

```bash
./annotations-exporter --kube.annotations=gitlab.ci.werf.io/pipeline-url --kube.only-labels-and-annotations=true --kube.reference-annotations=meta.helm.sh/release-name,meta.helm.sh/release-namespace
```
This will track revisions for `meta.helm.sh/release-name` and `meta.helm.sh/release-namespace` kubernetes resource labels. For example you deployed release from `https://gitlab.com/project/project/pipelines/1` and then from `https://gitlab.com/project/project/pipelines/2`, so the metrics would be:
```text
kube_annotations_exporter{annotations_exporter_annotation_gitlab_ci_werf_io_pipeline_url="https://gitlab.com/project/project/pipelines/1", annotations_exporter_annotation_meta_helm_sh_release_name="project-dev", annotations_exporter_annotation_meta_helm_sh_release_namespace="dev", annotations_exporter_revision="1"}

kube_annotations_exporter{annotations_exporter_annotation_gitlab_ci_werf_io_pipeline_url="https://gitlab.com/project/project/pipelines/2", annotations_exporter_annotation_meta_helm_sh_release_name="project-dev", annotations_exporter_annotation_meta_helm_sh_release_namespace="dev", annotations_exporter_revision="0"}
```
## Dashboards

Now there is only one [summary dashboard](charts/annotations-exporter/templates/dashboard.yaml), that will be autogenerated for helm-chart to ConfigMap. It is simple table that summarises all information about exported annotations and labels

<!--
helm upgrade --install annotations-exporter charts/annotations-exporter/ -n default -f test_values.yaml
-->
