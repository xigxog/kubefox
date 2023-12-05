<!-- markdownlint-disable MD033 -->
# Deployment Distillation

KubeFox automatically distills System Deployments to only those components that
are new or which have changed.  You can think of this process as a diff between
what components are currently running in the cluster, and what components have
changed in the new deployment.  In so doing KubeFox helps you control
provisioning.

For purposes of illustration, consider an Order [System](../concepts.md).

Note: There are additional things happening during these deployment cycles which
are discussed toward the bottom of the page. For this part of the discussion,
we're contemplating multiple deployments where only the most-recently deployed
version of the System is running.

In our scenario, there is an Order System that comprises 2 applications:

1. Fulfillment
2. Web-UI

The Fulfillment app is composed of 4 components, 2 of which are adapters:

1. CRON adapter
2. State Store adapter

[Adapters](../concepts.md) are Brokers for External Components, in this case
serving to proxy events, requests and responses to CRON and a State Store (like
a database).

The user-written components are:

1. Worker
2. API server (api-srv)

The web-ui app also composes 4 components, 2 of which being adapters:

1. HTTP adapter
2. State Store adapter

and two of which user-written:

1. Order user interface (order-ui)
2. API server (api-srv)

Note that the fulfillment app and web-ui app both employ the api-srv components,
i.e., these components are shared.

## Deployment a

When the Order System is initially deployed (we'll call this the 'a'
deployment), it will look like this in KubeFox (Figure 1):

<figure markdown>
  <img src="../../diagrams/deployments/distillation/light/deployment_a(light).svg#only-light" width=100% height=100%>
  <img src="../../diagrams/deployments/distillation/dark/deployment_a(dark).svg#only-dark" width=100% height=100%>
  <figcaption>Figure 1 - Initial Deployment [a]</figcaption>
</figure>

KubeFox will spool up 3 Pods, the Worker[a], api-srv[a] and the order-ui[a].
Because the api-srv component is shared by the fulfillment and web-ui apps,
KubeFox will deploy it only once.

## Deployment b

Now things get interesting!

Let's say that the user needs to make a change to the order-ui component. When
System b is deployed, it will look as shown in Figure 2:

<figure markdown>
  <img src="../../diagrams/deployments/distillation/light/deployment_b(light).svg#only-light" width=100% height=100%>
  <img src="../../diagrams/deployments/distillation/dark/deployment_b(dark).svg#only-dark" width=100% height=100%>
  <figcaption>Figure 2 - Deployment [b] to change the order-ui component</figcaption>
</figure>

Note that in our deployment table, only the order-ui component was deployed.
KubeFox checks the state of the Order System and deploys only the components
that have changed.

## Deployment c

For our next deployment [c], the user decides to make a change to the api-srv
component. Again, KubeFox checks the state of the System and deploys only
api-srv[c]:

<figure markdown>
  <img src="../../diagrams/deployments/distillation/light/deployment_c(light).svg#only-light" width=100% height=100%>
  <img src="../../diagrams/deployments/distillation/dark/deployment_c(dark).svg#only-dark" width=100% height=100%>
  <figcaption>Figure 3 - Deployment [c] to change the api-srv component</figcaption>
</figure>

Note that the highest component version is tracked in the System (now [c]), and
in the applications (now [c] as well).

## Deployment d

In our final deployment [d], the user does a couple of things:

- Creates a new component (reviews)
- Modifies the order-ui component

When the System is deployed, it looks as shown in Figure 4:

<figure markdown>
  <img src="../../diagrams/deployments/distillation/light/deployment_d(light).svg#only-light" width=100% height=100%>
  <img src="../../diagrams/deployments/distillation/dark/deployment_d(dark).svg#only-dark" width=100% height=100%>
  <figcaption>Figure 4 - Deployment [d] to add a component and modify order-ui</figcaption>
</figure>

Only the new (reviews) and modified component (order-ui) are deployed. Only the
web-ui application was modified - so now it's at version [d], while the
fulfillment app was not modified - so it's still at version [c].

## Summary Notes

There are a number things of note here:

- To put a fine point on it, each of the deployments is a version in KubeFox.
- All of the deployments (a - d) are actually available via explicit URLs (unless they're deprecated).
- Systems are released monolithically - which greatly simplifies deployent for users. Under the covers KubeFox is doing a few things:
  - Distilling the component set to the minimum required to run the new deployment, thereby preserving resources and preventing over-provisioning
  - Shaping traffic dynamically to enable the use of shared, common components
    both within a release (for instance,same component inside different
    applications) and across versions (for instance, same component version used
    for different deployments)
- Only one copy of api-srv will be running at any given time _unless_ there is a different version of api-srv deployed for one of the apps. In that case, KubeFox will run the appropriate versions of api-srv to service requests for each individual application.
- Default traffic will be routed to the most recently released version. So if [b] is released, default traffic will be running through version b even after version d is deployed. That provides development, QA and release teams with a great deal of power and flexibility.
