# Virtual Environments

When we think about environments today, we think about them as physical constructs, because that is what they are.  Environments familiar to all of us - dev, QA, UAT, prod - exist for various reasons:

- In dev, we want to support the work of engineering teams to provide for rapid development of new capabilities;
- In QA, we’re collecting the works of various teams and testing them in various capacities (functional, performance, smoke, load, verification etc.);
- In UAT, we’re running pre-production code for the purpose of enabling user and customers to validate that the proposed release meets requirements;
- In prod, we’re running final code.

Each of these environments has a set of requirements, some of them unique.  For instance, in dev and QA, we don’t want customer data, UAT should have only limited or synthetic customer data, HA is unimportant in dev but can be mission critical in prod etc.  Environments enable us to build policy controls supporting these tenets.  

<img src="../diagrams/environments/dev_environment.png" width=60% height=60%>

To prevent things like adverse impacts to  production customers from destructive prototypes, we segregate the environments.  That segregation can take various forms.  Whether the segregation occurs through different namespaces, different environments or even different clusters, significant DevOps overhead is incurred - largely due to the need to determine the means by which segregation will occur and then building the constructs that will provide for that segregation - e.g., manually building and configuring a new namespace, providing for connectivity, determining and accounting for the blast radius of the new works etc.  This is especially difficult in dev because dev has competing agendas:

1. We want developers to be able to innovate, test out new ideas, collaborate on changes, morph approaches - and do all of this freely and rapidly;
2. We want to prevent developers from stepping on each other; from a developer productivity standpoint, dedicated systems would be ideal;
3. We want to keep costs as low as possible, so we want to distill the supporting infrastructure (compute/storage resources, developer, QA and DevOps personnel etc.) to the smallest possible subset necessary

In essence we have a dichotomy:  we want an independent, flexible dev workspace that gracefully supports the innovative efforts of our teams *but* we want it at the lowest cost possible.  And by the way, nothing can affect our production systems.

Cost-reduction efforts today focus on reducing overall required resources by actively analyzing different teams and developers within the environments - for instance, manually determining common factors across teams and developers, building namespaces around those factors and continually reallocating - resulting in a lot of overhead and expense.  These tasks are usually performed by DevOps teams.  KubeFox usurps responsibility for the difficult aspects of this process and drastically reduces DevOps overhead while simultaneously increasing the speed at which accommodations can be made for new projects.

How?  KubeFox turns the concept of environment on its head.  Environments become very lightweight, malleable, virtual constructs.  Custom configurations (consisting of environment variables and secrets) can be created and deployed easily  - either alongside a developer’s modifications or independently.  In effect, it appears to the developer that their code is running inside a private environment unique to them.  Behind the scenes, KubeFox takes care of the efficiencies by deploying only what is actually unique to that developer’s efforts and sharing those components that have not changed.  The developer can coexist with their colleagues - even if those colleagues are working on the same component(s) for different reasons.  And DevOps overhead is drastically reduced.  

One can think about KubeFox environments as retaining the aspects of environments that help accomplish useful things like segregating workloads, providing sandboxes for POCs, developer experimentation, all the varieties of QA testing and others.  However, the constraints one associates with environments are gone.  

KubeFox ‘environments’ are lightweight, agile, virtual constructs.  KubeFox’s dynamic routing empowers developers to easily and simply spool up sandboxes to support their efforts.  So developers are no longer constrained by environment or namespace related logistics.  

Polina can test her new code in Environment A, which is unique to her:

<img src="../diagrams/environments/polina_dev_environment_a.png" width=60% height=60%>

and then shift instantly to Environment B, perhaps to compare her changes with the prior version of software:

<img src="../diagrams/environments/polina_dev_environment_b.png" width=60% height=60%>

KubeFox takes care of routing the requests appropriately and injecting the appropriate credentials - not just for the datastores, but for all components composing Polina’s deployment.  And by the way, Polina can run her application in both environments simultaneously.

One way to think about it is that the developer environments are spooled up in minutes.  But what is really happening is that the environment is created on-the-fly, and overlaid onto a deployment.  Because KubeFox abstracts the configuration from the deployment, teams can easily and simply shift their workloads - even for deployments that have already occurred.

KubeFox segregates the developer sandboxes while leveraging its capabilities to distill the number of Pods running to only those that are unique and necessary (this is discussed in greater depth in Versioned Deployments below).

<img src="../diagrams/environments/dev_environment_sandboxes.png" width=60% height=60%>

Qiao, Jose, Polina, Pradeep, Fred and Sally could all be working on their own copies of Component X for Application 3, and KubeFox will ensure that traffic destined for Joe’s environment within Dev will use Joe’s version of Component X.





