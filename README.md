# Annotations Exporter

## Overview

Prometheus-exporter, which converts any Kubernetes resources annotations and labels to Prometheus samples.

## Usage
```
Usage of annotations-exporter:
  -h, --help
          help for annotations-exporter
  --kube.annotations strings
          Annotations names to use in prometheus metric labels
  --kube.labels strings
          Labels names to use in prometheus metric labels
  --kube.max-revisions int
          Max revisions of resource labels to store (default 3)
  --kube.namespaces strings
          Specifies the namespace that the exporter will monitor resources in (default 'all namespaces')
  --kube.resources strings
          Resources (<resource>/<version>/<api> or <resource>/<api>) to export labels and annotations (default [deployments/apps,ingresses/v1/networking.k8s.io,statefulsets/apps,daemonsets/apps])
  -v, --version
          version for annotations-exporter
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

## Dashboards

TBA
