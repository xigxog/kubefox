# KubeFox Kit SDK for Go

## Kubernetes CRDs

This project includes Go models in `pkg/api/kubernetes` for KubeFox Kubernetes
CRDs. The [Helm Charts](https://github.com/xigxog/kubefox/helm-charts) project uses
these models to generate the CRD YAMLs. When changes are made to the models be
sure to update the CRDs. See the `README` for the [Helm
Charts](https://github.com/xigxog/kubefox/helm-charts) project. To add a new CRD simply
copy an existing CRD Go file and update as needed.
