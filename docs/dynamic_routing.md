# Dynamic Routing and Guaranteed Delivery

KubeFox can achieve the capabilities described in [**Virtual Environments**](virtual_environments.md) and [**Versioned Deployments**](versioned_deployments.md) because it dynamically routes requests at runtime.  

When requests originate - [**Genesis Events**](concepts.md) in KubeFox – metadata is associated with them that informs the KubeFox runtime how the requests should be routed.  This occurs not only at origination time, but throughout the System and for all components that are part of that System.  In so doing, KubeFox can guarantee delivery of requests to the correct version of each component, irrespective of the number of versions of that component that may be running.

