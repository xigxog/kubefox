# Overview

KubeFox is a software platform that makes creating, deploying, releasing, and
maintaining software applications on Kubernetes easier. It consists of a core
set of services for recording telemetry and storing configuration and a
framework to develop, test, and run components of an application. The KubeFox
Platform is highly opinionated greatly reducing its complexity and endless
configuration for it to work.

KubeFox is built on the concept of deploying a System. Concretely, a KubeFox
System is a collection of Apps, which in turn are a collection of Components and
Routes to those Components. The definition of a System's Apps and Routes as well
as the source code of its Components is stored in a Git repository, referred to
as a System Repo.

A KubeFox System can be deployed to an instance of the Kubefox Platform running
on a Kubernetes Cluster. KubeFox will create Kubernetes Pods for each of the
System's Components. Even though the Components are running no requests will be
sent to the System until it is released. A System is released to an Environment
which defines configuration needed by the System's Components to process
requests. Once a System is released requests will be sent to it if they match
any of its Routes.
