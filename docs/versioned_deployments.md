# Versioned Deployments

Deployments are fraught with a host of risks and their management absorbs excessive resources.  Factors like version compatibility, roll-back, availability, blue/green, A/B and canary complicate deployment.  KubeFox simplifies the software lifecycle by providing for elegant versioned deployments that serve to reduce risk and simplify engineering workflows.  But perhaps most importantly, versioning reduces the application and microservice sprawl that is endemic to many of today’s Kubernetes teams. 

Let’s start with a System “MySystem”, which is composed of 3 applications as follows:

<img src="../diagrams/deployments/mysystem_version1_composition.png" width=60% height=60%>

Note that the applications all share Components A, B and C.  If the MySystem is deployed by KubeFox, then the cluster will look like this:

<img src="../diagrams/deployments/mysystem_version1_deployment.png" width=70% height=70%>

KubeFox will handle the routing from component to component, and provide telemetry both on a discrete basis (by application), and on a global basis (across the cluster).  You can visualize the behavior of Shared Component A both in the context of Application 1, and in the context of the cluster.  Similarly, you can visualize the behavior of each of the 3 applications, as if they were running by themselves in the cluster.

At deployment time, KubeFox determines what components need to be deployed, and deploys only the unique components necessary to support that deployment.  That applies to both components, and unique versions of those components.  In the current example, if we change component H in Application 3 – let’s call it H’ - and deployed the new version of MySystem (any change within a System will yield a new version of that System), when we deploy the new version of MySystem, then the actual deployment to the cluster will be distilled to only deploy component H’.

Let’s look at a different example to explore the power of KubeFox further.  Suppose in the prior diagram, we enhance Application 3.  In our new deployment, Application 3 v2 consists of 5 components, 1 additional unshared component (J), and a new version of Shared component A – component A’.  When we deploy the new version of MySystem, the cluster will look like this:

<img src="../diagrams/deployments/mysystem_deployment_app3_enhancement.png" width=70% height=70%>

Note that both the original Component A – because it is still shared by Applications 1 and 2 – is still running, and the new Component A’ is deployed because it is needed by Application 3 v2.  If we later create new versions of Application 1 and 2 that use the updated Component A’, when we deploy the new version of MySystem, the cluster will look like this:

<img src="../diagrams/deployments/mysystem_deployment_app1and2_enhancement.png" width=70% height=70%>

## KubeFox Deploy

A KubeFox deployment results in the following:

1. The components unique to the new version of the System are deployed discretely and the Pods to support them are spooled up.  Think of this as a diff between the collection of components associated with version A of a System and those associated with version 2 of that System.  Any updated component or new component in version B will be deployed when version B of the System is deployed.

    See [**Deployment Management**](deployment_management.md) for a deeper discussion of the component distillation facilities of KubeFox
   
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

Let’s revisit one of the prior examples and describe what is actually happening.

<img src="../diagrams/deployments/mysystem_version1_deployment_2.png" width=70% height=70%>

In this example, MySystem Version 1 is constructed as before:

<img src="../diagrams/deployments/mysystem_version1_composition.png" width=60% height=60%>

When we release MySystem Version 1, all default traffic is routed to it.

Now we create a new version of MySystem – Version 2.  In our new version and as before, we’ve modified Application 3.  We’ve added a new component J, and enhanced component A – which we’ll call A’:

<img src="../diagrams/deployments/mysystem_version2_composition.png" width=60% height=60%>

When we deploy MySystem Version 2, KubeFox will only deploy updated component A’ and new component J.  And it will spool up the Pods necessary to support these components.  

<img src="../diagrams/deployments/mysystem_version2_deployment.png" width=70% height=70%>

Default traffic will still be routed to Version 1 of MySystem – including Application 3 – because we have not yet released MySystem Version 2.  Version 2 of MySystem – including Application 3 v2 - is accessible via explicit URLs.  If and when we release Version 2 of MySystem, default traffic will be routed to it instead of Version 1.








