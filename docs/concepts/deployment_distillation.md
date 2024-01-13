# Deployment Distillation


KubeFox automatically distills application deployments to only those components that
are new or which have changed.  You can think of this process as a diff between
what components are currently running in the cluster, and what components have
changed in the new deployment.  In so doing, KubeFox helps you control
provisioning.

<h3>Example</h3>

For purposes of illustration, consider an simple Order application [App](./index.md#app).

In our scenario, our Order App comprises 2 modules:

1. Fulfillment
2. Web-UI

The Fulfillment module is composed of 4 components, 2 of which are adapters:

1. CRON adapter
2. State Store adapter

[Adapters](./index.md#adapter) are Brokers for External Components, in this case
serving to proxy events, requests and responses to CRON and a State Store (like
a database).

The user-written components are:

1. Worker
2. API server (api-srv)

The web-ui module also comprises 4 components, 2 of which being adapters:

1. HTTP adapter
2. State Store adapter

and two of which user-written:

1. Order user interface (order-ui)
2. API server (api-srv)

Note that the fulfillment and web-ui modules both employ the api-srv component,
i.e., this component is shared.

<h3>Version 1 Deployment</h3>

When our App is initially deployed (we'll call this the Version 1
deployment), it will look like this in KubeFox (Figure 1):

<figure markdown>
  <img src="../diagrams/deployments/distillation/app_deployment_v1.svg" width=100% height=100%>
  <figcaption>Figure 1 - Initial Deployment - Version 1</figcaption>
</figure>

KubeFox will spool up 3 Pods, the Worker [v1], api-srv [v1] and the order-ui [v1].
The api-srv component is shared by the fulfillment and web-ui modules,
KubeFox will deploy it only once.

<h3>Version 2 Deployment</h3>

Now things get interesting!

Suppose we decide to make a change to the api-srv component. To do so, we'll
deploy a new version of our App - Version 2.  KubeFox checks the state of our
deployment and deploys only those components that are new or which have changed.
Because only api-srv [v2] changed, only one additional Pod is created:

<figure markdown>
  <img src="../diagrams/deployments/distillation/app_deployment_v2.svg" width=100% height=100%>
  <figcaption>Figure 2 - Deploying Version 2 to change api-srv</figcaption>
</figure>

Remember when we said that now things get interesting?  Our cluster is now
capable of supporting both Version 1 and Version 2 traffic.  KubeFox looks at
the query parameters on the HTTP request and creates a [Genesis
Event](./index.md#genesis-event) that tells the system which version is to be employed to service that event.  The exception is if we release one of the
versions.  If we release Version 1, then traffic (sans query parameters) will be
defaulted to Version 1:

<figure markdown>
  <img src="../diagrams/deployments/distillation/app_deployment_v2_(v1_traffic).svg" width=100% height=100%>
  <figcaption>Figure 3 - Version 1 traffic (with Version 2 present)</figcaption>
</figure>

For all intents and purposes, it appears as though Version 1 is the only version
of our App running.  But we can easily access Version 2 via query parameters:

<figure markdown>
  <img src="../diagrams/deployments/distillation/app_deployment_v2_(v2_traffic).svg" width=100% height=100%>
  <figcaption>Figure 4 - Version 2 traffic (with Version 1 present)</figcaption>
</figure>

Let's take things a little further.  Suppose we want to create a new version of
our App - Version 3 - to enhance the order-ui component and to
add a new component ("review") to process order reviews.

When we deploy Version 3, our App will look like as shown in Figure 5:

<figure markdown>
  <img src="../diagrams/deployments/distillation/app_deployment_v3.svg" width=100% height=100%>
  <figcaption>Figure 5 - Deployment of Version 3</figcaption>
</figure>

As before, KubeFox deploys only the new and changed components.  And as before,
we now have access to 3 versions of our App.  If Version 1 is still the released
version, then default traffic will be routed to Version 1:

<figure markdown>
  <img src="../diagrams/deployments/distillation/app_deployment_v3_(v1_traffic).svg" width=100% height=100%>
  <figcaption>Figure 6 - Version 1 traffic (with Versions 2 and 3 present)</figcaption>
</figure>

Version 2 is accessible via query parameters:

<figure markdown>
  <img src="../diagrams/deployments/distillation/app_deployment_v3_(v2_traffic).svg" width=100% height=100%>
  <figcaption>Figure 7 - Version 2 traffic (with Versions 1 and 3 present)</figcaption>
</figure>

and similarly for Version 3:

<figure markdown>
  <img src="../diagrams/deployments/distillation/app_deployment_v3_(v3_traffic).svg" width=100% height=100%>
  <figcaption>Figure 8 - Version 3 traffic (with Versions 1 and 2 present)</figcaption>
</figure>

## Summary

There are a number things of note here:

- To put a fine point on it, each of the deployments is a version in KubeFox.
- All of the deployments (Versions 1, 2 and 3) are actually available via explicit URLs
  (unless they're deprecated).
- Apps are released monolithically - which greatly simplifies deployment for
  users. Under the covers KubeFox is doing a few things:
  - Distilling the component set to the minimum required to run the new
    deployment, thereby preserving resources and preventing over-provisioning
  - Shaping traffic dynamically at runtime to enable the use of shared, common components
    both within a release (for instance, the same component across versions)
- Default traffic will be routed to the most recently released version. So if
  Version 1 is released, default traffic will be running through Version 1 even after
  Versions 2 and 3 are deployed. That provides development, QA and release teams with a
  great deal of power and flexibility.
