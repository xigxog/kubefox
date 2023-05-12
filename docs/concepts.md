#**Concepts**

We’ll start with brief descriptions of KubeFox constituents and gradually drop
deeper into detail.

####**`Applications`**

   : Applications are just what you think they are: collections of components which together 
   provide useful capabilities to end users of the software.

####**`Components`**

   : Components are user-written microservices or functions which provide capabilities to 
   fulfill requirements of an application. The same component can (of course) be used by one 
   or more applications. Providing for the sharing of components is at the core of why KubeFox 
   employs the concept of System.

   : Components are primitives of the KubeFox platform. Entrypoints are registered
   with KubeFox using the Kit SDK and invoked via events known to that
   component. The source code is stored in a registered Git repository that
   follows KubeFox convention. Building, packing, testing, and execution of the
   component is handled by KubeFox.

####**`System`**

   : A System is a collection of Applications. KubeFox employs the concept of
   System to help developers, QA and release teams with the reuse of components
   across applications.

   : When a release is deployed, it is deployed monolithically at the System
   level, greatly simplifying the deployment activity. Teams don't need to spend
   time determining what changed for a particular version - KubeFox handles that
   automatically.

####**`Repository`**

   : A repository is a Git repository. Repositories map 1:1 with Systems.

####**`Brokers`**

   : A Broker is a side-car service attached to a component that serves as a proxy
   for events, requests, and responses between that component and other
   components in an application. Brokers provide a number of advanced services
   and capabilities - we’ll dig more into them later.

**`External Components`**

   : External Components are third party software e.g., databases, K/V stores,
   CRON etc. Adapters provide connectivity to External Components.

####**`Adapters`**

   : Adapters are simply Brokers for External Components, serving to proxy events,
   requests and responses to and from External Components.

   : External Components are registered with KubeFox, enabling access control,
   secret injection, and telemetry recording via Adapters.

####**`System Object`**

   : When a System is built, a System Object is created in JSON format that
   includes everything a System needs to be deployed. A System Object comprises
   Applications, Components and Routes for a System.

####**`SysRef`**

   : A SysRef is a shorthand reference to a System Object that enables interaction
   with a System Object using its id, branch name, or tag name. A SysRef
   represents a snapshot of a System at a moment in time and once created, a
   SysRef is immutable. A SysRef can be created from a branch or tag.

####**`Deployment`**

   : A deployment is a SysRef that has been loaded into a cluster. The result of a
   deployment is that all constituents necessary to run a SysRef are available
   on the cluster.

   : When a SysRef is deployed, it is available via explicit URLs.

####**`Release`**

   : A Release results in the direction of default traffic to the specified
   SysRef. Any traffic not qualified with (for instance) explicit URLs would be
   routed to the currently Released SysRef. The current version of software
   running in Production would always be the released SysRef.

####**`Events`**

   : KubeFox is an event driven architecture. All requests and responses, even
   synchronous calls, are modeled as events and are exchanged as messages via
   brokers.

####**`Genesis Events`**
    
   : A Genesis Event is an originating event.  It is often created from an external request - 
   for instance an external HTTP call.  But an internal process like CRON or a batch job could 
   create a Genesis Event as well.

   : KubeFox determines how a Genesis Event will be routed dynamically.

####**`Environments`**

   : Environments are virtual overlays in KubeFox. Because KubeFox dynamically
   shapes and routes traffic, you are not bound to environments in the same way
   as you are today. For instance, in KubeFox, you can layer in an environment
   that shifts traffic in a running System from one database to another.

   : KubeFox environments are quite powerful and greatly simplify provisioning.
   We'll dive much more deeply into environments in a separate section.

####**`Routes`**

   : Routes are used to send events from external sources to a component of a
   SysRef. A route includes a set of criteria and a target component. A target
   can be an environment, SysRef, or component. Route criteria are expressed in
   adapter specific language and are processed when an event is received by the
   adapter.

   : Routes are matched once at event genesis. The target component as well as its
   application, system version, and environment are attached to the event to
   provide context for future routing of events generated by downstream
   components.

   : This context is needed as a single component instance might be shared between
   multiple deployed SysRefs. When a downstream component makes a call the of
   the context of the genesis event is used to ensure the next component being
   called is the correct version.

   : Routes optionally can contain logic to extract parts of the event into well
   name arguments which are passed to the component as key/value pairs.
