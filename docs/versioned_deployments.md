<!-- markdownlint-disable MD033 -->
# Versioned Deployments

Deployments are fraught with a host of risks and their management absorbs excessive resources.  Factors like version compatibility, roll-back, availability, blue/green, A/B and canary complicate deployment.  KubeFox simplifies the software lifecycle by providing for elegant versioned deployments that serve to reduce risk and simplify engineering workflows.  But perhaps most importantly, versioning reduces the application and microservice sprawl that is endemic to many of today’s Kubernetes teams.

Let’s start with a Retail System composed of 3 applications, a Web UI app, a Catalog app and a fulfillment app, as shown in Figure 1:

<figure markdown>
  <img src="../diagrams/deployments/versioned/light/retail_system_composition_v1(light).svg#only-light" width=90% height=90%>
  <img src="../diagrams/deployments/versioned/dark/retail_system_composition_v1(dark).svg#only-dark" width=90% height=90%>
  <figcaption>Figure 1 - Composition of Version 1 of the Retail System</figcaption>
</figure>

As the diagram notes, this is the first version of our Retail System.  Note that
the applications share some Components - specifically, the API Server, Vendor
Service and Member Service components.  When the Retail System is deployed by
KubeFox, the cluster will look like the diagram in Figure 2.  

<figure markdown>
  <img src="../diagrams/deployments/versioned/retail_system_v1.svg" width=100% height=100%>
  <figcaption>Figure 2 - First Deployment of the Retail System</figcaption>
</figure>

Each of the components runs in its own Pod.  KubeFox handles the routing from component to component, and provides telemetry both on a discrete basis - by component and application - and on a global basis (across the cluster).  You can visualize the behavior of the shared API Server both in the context of the Catalog app, and in the context of the cluster.  Similarly, you can visualize the behavior of each of the 3 applications, as if they were running by themselves in the cluster.

At deployment time, KubeFox determines what components need to be deployed, and
deploys only the unique components necessary to support that deployment.  That
applies to both components, and unique versions of those components.  In the
current example, if we update the Catalog Query component in the Catalog app,
our Retail System will be composed of the components shown in Figure 3.

<figure markdown>
  <img src="../diagrams/deployments/versioned/light/retail_system_composition_v2(light).svg#only-light" width=90% height=90%>
  <img src="../diagrams/deployments/versioned/dark/retail_system_composition_v2(dark).svg#only-dark" width=90% height=90%>
  <figcaption>Figure 3 - Composition of Version 2 of the Retail System</figcaption>
</figure>

When we redeploy, KubeFox will create a new  version of the Retail System; any
change within a System - whether an updated to a component, or addition of a
component - will yield a new version of that System.

<figure markdown>
  <img src="../diagrams/deployments/versioned/retail_system_v2.svg" width=100% height=100%>
  <figcaption>Figure 4 - Version 2 of the Retail System with the modified Catalog Query component</figcaption>
</figure>

Note that KubeFox is taking care of this, unburdening end users from the drudgery of determining what changed.  The end user simply redeploys the Retail System, and when KubeFox builds the deployment, it determines what changed and distills the deployment to only those components that have changed; in our example, only the Catalog Query component changed and it will be the only component redeployed (Figure 4).

Once the deployment occurs, the Retail System will be capable of supporting both versions of the Catalog application.  KubeFox dynamically shapes traffic and will route version 1 requests to the v1 version of the Catalog Query component, and version 2 requests to the v2 component.  Version 2 is accessible via explicit URLs, enabling teams to validate functionality and stability with minimal risk.  Default traffic will be routed to version 1 until version 2 is released.  The end user can optionally leave version 1 in place or choose to deprecate it.

Let’s look at a different example to explore the power of KubeFox further.  Suppose we enhance the Catalog app further by adding a new component “Catalog Vendor” and to support it, we need to update the shared component Vendor Services.  To keep things a little tidier, let’s say that we’ve deprecated version 1 of the Retail System, so we only need version 2 of the Catalog Query component.  The component building blocks will be those shown in Figure 5.

<figure markdown>
  <img src="../diagrams/deployments/versioned/light/retail_system_composition_v3(light).svg#only-light" width=90% height=90%>
  <img src="../diagrams/deployments/versioned/dark/retail_system_composition_v3(dark).svg#only-dark" width=90% height=90%>
  <figcaption>Figure 5 - Composition of Version 3 of the Retail System</figcaption>
</figure>

<figure markdown>
  <img src="../diagrams/deployments/versioned/retail_system_v3.svg" width=100% height=100%>
  <figcaption>Figure 6 - Version 3 of the Retail System</figcaption>
</figure>

Note that both the original v1 Vendor Service component – because it is still
shared by the Web UI and Fulfillment Applications – is still required, and the
new v2 Vendor Service component - because it is needed by the new Catalog Vendor
component - are present in the cluster.  If we later create new versions of the
Web UI and Fulfillment applications that use the updated v2 Vendor Service
component (thereby creating Version 4 of the Retail System), we can choose to
deprecate the v1 version of Vendor Services (Figure 7).

<figure markdown>
  <img src="../diagrams/deployments/versioned/retail_system_v4.svg" width=100% height=100%>
  <figcaption>Figure 7 - Version 4 of the Retail System</figcaption>
</figure>

Our examples illustrate some of the power of KubeFox.  We can choose to run all 4 versions of the Retail System in parallel if we wish.  Doing so would not mean that there would be 4 monolithic deployments of the Retail System, each demanding and consuming their own resources.  Instead, KubeFox would only deploy unique versions of needed components.  For instance, only one Pod for the Member Service would be necessary, as the Member Service did not change for versions 2 through 4.  KubeFox would handle traffic shaping for each version, and telemetry for version 2 would be specific to version 2, even for the Member Service.

Now is a good time to review KubeFox deployments and releases.

<!-- <object
  id="color-change-svg"
  data="../diagrams/deployments/dark/test.svg"
  type="image/svg+xml"
  >
 </object> -->

<!-- <p>Click the following button to see the function in action</p>  
<input type = "button" onclick = "changeSVGColor('#FFCD28')" value = "Display">   -->
## KubeFox Deploy

A KubeFox deployment results in the following:

1. The components unique to the new version of the System are deployed discretely and the Pods to support them are spooled up.  Think of this as a diff between the collection of components associated with version A of a System and those associated with version 2 of that System.  Any updated component or new component in version B will be deployed when version B of the System is deployed.

    See [**Deployment Distillation**](deployment_distillation.md) for a deeper discussion of the component distillation facilities of KubeFox

2. The new version of the System (version B) is available only via explicit calls, for instance, explicit URLs.  Public traffic continues to be routed to the currently released version of the System (version A).

## KubeFox Release

A KubeFox release results in the following:

1. Default traffic is routed to the newly released System (version B).

With KubeFox, you are always releasing a version of a System, even if a single component changed.  KubeFox shapes traffic delivery to each component of the system dynamically (see [**Dynamic Routing and Guaranteed Delivery**](dynamic_routing.md)).  Because release is such a lightweight operation, it is almost instantaneous.  And it yields a great deal of power and flexibility when dealing with new versions of Systems.
  
These are just some scenarios:

- Instead of spooling up a new cluster with dedicated nodes to enable a new customer to validate enhancements to their applications in UAT, you create a subdomain for them and route traffic from them to the new version of their System.  Meanwhile, your sales team is using the same system for demos.
- You want to canary a new release intelligently – by directing traffic of an early adopter to a new version of the software.  You have telemetry specific to that version (an innate capability of KubeFox) and can compare and contrast its behavior to that of the currently released version, running side-by-side in the same cluster.
- You have concerns with a new release of your applications, and require the ability to rollback extremely quickly should anything go awry.

KubeFox makes all of this possible.
