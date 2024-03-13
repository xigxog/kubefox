# KubeFox
#### A XigXog Product

[![build](https://github.com/xigxog/kubefox/actions/workflows/build.yaml/badge.svg)](https://github.com/xigxog/kubefox/actions/workflows/build.yaml)
[![release](https://github.com/xigxog/kubefox/actions/workflows/release.yaml/badge.svg)](https://github.com/xigxog/kubefox/actions/workflows/release.yaml)
[![Go Report Card](https://goreportcard.com/badge/github.com/xigxog/kubefox)](https://goreportcard.com/report/github.com/xigxog/kubefox)

<!-- <figure markdown>
    <img src="./docs/images/XigXog Horizontal Layout (250 x 150) revised.svg" width=30% height=30%>
    <figcaption>A XigXog Product</figcaption>
</figure>
 -->

- Website: https://www.kubefox.io
- Documentation: https://docs.kubefox.io
- Quickstart: https://docs.kubefox.io/quickstart.html

<figure markdown>
<img src="./docs/images/KubeFox Horizontal Layout (400 x 100) revised.svg" width=100% height=100%>
</figure>

KubeFox is an SDK, platform and infrastructure to enable rapid construction and
deployment of secure and robust applications for Kubernetes, and which
drastically reduces bureaucracy and burdensome DevOps activities.

Teams and even individual developers can create and rapidly prototype
code on the same cluster in what appear to be individual sandboxes.  Behind the scenes, KubeFox
provides the following capabilities:

- **Deployment Distillation:**  

    With KubeFox, you deploy at an application level.  You don't need to worry
    about which components you have added or modified.  KubeFox tracks the
    repository and builds, containerizes and distills deployments to only those
    components that have changed.  You can read more about Deployment
    Distillation 
    [here](https://docs.kubefox.io/concepts/deployment_distillation.html).

- **Versioned Deployments**

    When you deploy an application with KubeFox, that application is
    automatically versioned.  KubeFox will ensure that traffic is restricted to
    the components that composed the application when it was deployed. Note that
    individual deployments may share one or more components.  This enables
    KubeFox to prevent over-provisioning; the deployments can run on the same
    cluster but it appears that each deployment is running in its own invidual
    sandbox.  That extends to deployment telemetry, which will reflect data from
    each version.  You can read more about Versioned Deployments [here](https://docs.kubefox.io/concepts/versioned_deployments.html).

- **Virtual Environments**

    KubeFox Virtual Environments (VEs for short) are lightweight, malleable
    constructs that enable engineering teams and developers to create virtual
    sandboxes in which they can run prototypes or POCs. Custom configurations (consisting of environment variables and
    secrets) can be created and deployed easily - either in concert with a
    deployment or independently.  And they're controllable by the developers
    themselves, eliminating the overhead associated with traditional
    environments.  You can read more about Virtual Environments [here](https://docs.kubefox.io/concepts/virtual_environments.html).

- **Federated metrics, tracing and logging**

    Multiple versions of components and applications can be deployed to a
    single cluster.  That would seem to complicate telemetry but KubeFox
    provides for telemetry at global, application and component levels.  You can
    read more about Telemetry [here](https://docs.kubefox.io/concepts/telemetry.html).

- **Dynamic Routing and Guaranteed Delivery**

    KubeFox is capable of managing multiple versions of applications and
    components because it is aware of the composition of those applications, and
    traffic is shaped and validated to a specific version of an application at
    runtime.  You can read more about Dynamic Routing [here](https://docs.kubefox.io/concepts/dynamic_routing.html).

- **Zero Trust**

    Zero Trust is an intrinsic property of a KubeFox application.  All requests
    are validated at origin (Genesis Events in KubeFox) and at all
    component-to-component boundaries.  

### Getting Started and Documentation

The best way to get started with KubeFox is to go through our
[Quickstart](https://docs.kubefox.io/quickstart.html) and read through our [docs](https://docs.kubefox.io).  The Quickstart
currently supports
[kind](https://kind.sigs.k8s.io/) and [Microsoft
Azure](https://learn.microsoft.com/en-us/azure/?product=popular). The Quickstart
can be navigated in less than a half hour and gives you a good overview of the
power of KubeFox.

### License

XigXog is committed to open source and our software is licensed under the MPL-2.0.

[MPL-2.0](https://github.com/xigxog/kubefox/blob/main/LICENSE)



