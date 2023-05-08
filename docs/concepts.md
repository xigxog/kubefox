# Concepts

We’ll start with brief descriptions of KubeFox constituents and gradually drop
deeper into detail.

1. Cluster

   A cluster is simply a Kubernetes cluster with the KubeFox platform installed.

2. Repository

   A repository is a Git repository. Repositories map 1:1 with Systems.

3. System

   A System is a collection of Applications. KubeFox employs the concept of
   System to help developers, QA and, release teams with the reuse of components
   across applications.

   When a release is deployed, it is deployed monolithically at the System
   level, greatly simplifying the deployment activity. Teams don't need to
   spend time determining what changed for a particular version - KubeFox does
   that (and a lot more).

4. Application

   Applications are just what you think they are: collections of components
   which together provide useful capabilities to end users of the software.

5. Components

   Components are user-written microservices or functions which provide
   capabilities to fulfill requirements of an application. The same component
   can (of course) be used by one or more applications. This is really the core
   reason why we have the concept of System.

6. Broker

   A side-car service attached to a component that serves as a proxy for events,
   requests, and responses between that component and other components in the
   System. Brokers provide a number of advanced services and capabilities -
   we’ll dig more into them later.

7. Adapter

   Adapters are third party software Brokers; connectors to things like HTTP,
   state stores, gRPC-connected services, CRON etc. Adapters are to third party
   software what Brokers are to Components, serving to proxy requests to those
   third-party services.

8. Deployment

   A deployment is a SysRef that has been loaded into a cluster. The result of
   a deployment is that all constituents necessary to run a SysRef are
   available on the cluster.

9. Soft Release

   A Soft Release launches all constituents of a SysRef; however, the SysRef is
   available only via explicit URLs. By default, the Soft Release occurs at
   Deployment time, but it can be rendered separate and explicit.

10. Release

    A Release is the activation of a SysRef - general external traffic will be
    routed to the released SysRef.

## Events

KubeFox is an event driven architecture. All requests and responses, even
synchronous calls, are modeled as events and are exchanged as messages via
brokers.

## Cluster

Components are primitives of the KubeFox platform. They are services that
provide capabilities to fulfil requirements of an application. Entrypoints are
registered with KubeFox using the Kit SDK which are invoke by incoming events to
the component. The source code is stored in a registered Git repository that
follows the KubeFox conventions. Building, packing, testing, and execution of
the component is handled by KubeFox.

## Providers

[TODO]

Providers are services that exist outside of the control of KubeFox. A common
example of a provider is a state store such as a database but things like 3rd
party HTTP services are also applicable. Even though KubeFox does not completely
control Providers, they still must be registered with KubeFox. This allows
requests proxied through brokers to perform access control, secret injection,
and telemetry recording. Events from Providers are translated to KubeFox
compliant events by an adapter.

Based on the type of Provider certain variables must be specified in the
Environment.

## Adapter

Adapters bridge providers and KubeFox.

## Application

An Application is a logical group of one or more components that satisfy a set
of features. The same component can be used by one or more applications.
Components not assigned to an application will not be deployed.

Components within an application are able to communicate with each other without
explicit permission. Communication between applications requires explicit
authorization and is allowed on a component to component level.

## System

A group of one or more applications. Systems are represented by Git
repositories. All managed components in the Git repository are part of the
system.

## System Ref

An immutable snapshot of a System at a particular moment in time created from a
Git ref. The System Ref includes the container image hash of the components
built at that Git ref. A System Ref can be created from a commit to a branch or
a tag.

## Environment

A lightweight set of configuration that provide values for application
variables. Multiple systems and multiple tags of the same system can be assigned
to a environment. When multiple tags of the same system are assigned one of the
tags must be designated as the default.

## Variable

Variables are tuples that provide a key, value, and default value. They are
defined as part of the application specification. Their values are provide by
Environments. These variables are available to components at runtime and can be
used in configuration and policies.

## Broker

A side-car service attached to a component that proxies events, requests, and
responses between the KubeFox message transport and the component. When an event
is generated by the attached component the broker is responsible for applying
applicable policies, attaching relevant metadata, and to ensure proper routing.
Additionally the broker is responsible for gathering context, such as variables
and secrets, and providing it to the component.

## Route

Routes are used to send events from external sources to a component of a
particular SysRef. A route includes a set of criteria and a target component. A
target can be an environment, SysRef, or component. Route criteria are expressed
in adapter specific language and are processed when an event is received by the
adapter.

Routes are matched once at event genesis. The target component as well as its
application, system version, and environment are attached to the event to
provide context for future routing of events generated by downstream components.

This context is needed as a single component instance might be shared between
multiple deployed SysRefs. When a downstream component makes a call the of the
context of the genesis event is used to ensure the next component being called
is the correct version.

Routes optionally can contain logic to extract parts of the event into well name
arguments which are passed to the component as key/value pairs.

### Environment Route

Environment routes are used to match an event to an environment and optionally
to a specific SysRef in the environment. They are processed first, before
application routes. If no environment routes match then the default environment
of the cluster and the default SysRef of the environment is used before applying
application routes.

### Application Route

Application routes are used to match events and send them to a specified
component. They are processed after a SysRef is selected by environment routes.
If no application routes match then a `Not Found` error is returned to the
adapter.

## Subscription

Subscriptions are used to match messages and then send an event to a component.
Like routes, subscriptions contain a set of criteria to determine if an event is
relevant.

## Deployment

A deployment is a SysRef that has been applied to a cluster. The result of a
deployment is that all the components of the SysRef are running on the cluster.

## Release

A release is when a deployment is activated which allows traffic to flow to the
components of the SysRef.
