# Deployments and Releases

### KubeFox Deployments

A KubeFox deployment results in the following:

1. The components unique to the new version of an App are built and deployed
   discretely, and the Pods to support them are spooled up.  Think of this as a
   diff between the collection of components associated with version A of a
   App and those associated with version B of that App.  Any updated
   component or new component in version B will be deployed when version B of
   the App is deployed.

    See [Deployment Distillation](deployment_distillation.md) for a deeper
    discussion of the component distillation facilities of KubeFox

2. The new version of the App (version B) is available via explicit
   calls, for instance, explicit URLs.  Default public traffic continues to be routed to
   the currently released version of the App (version A).

### KubeFox Releases

A KubeFox release results in the following:

1. Default traffic is routed to the newly released System (version B).

With KubeFox, you are always releasing a version of a System, even if a single
component changed.  KubeFox shapes traffic delivery to each component of the
system dynamically (see [Dynamic Routing and Guaranteed
Delivery](dynamic_routing.md)).  Because release is such a lightweight
operation, it is almost instantaneous.  And it yields a great deal of power and
flexibility when dealing with new versions of Systems.
  
These are just some scenarios:

- Instead of spooling up a new cluster with dedicated nodes to enable a new
  customer to validate enhancements to their applications in UAT, you create a
  subdomain for them and route traffic from them to the new version of their
  System.  Meanwhile, your sales team is using the same system for demos.
- You want to canary a new release intelligently – by directing traffic of an
  early adopter to a new version of the software.  You have telemetry specific
  to that version (an innate capability of KubeFox) and can compare and contrast
  its behavior to that of the currently released version, running side-by-side
  in the same cluster.
- You have concerns with a new release of your applications, and require the
  ability to rollback extremely quickly should anything go awry.

KubeFox makes all of this possible.
