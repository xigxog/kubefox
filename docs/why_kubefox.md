<!-- markdownlint-disable MD033 -->
# Why KubeFox?

KubeFox is involved in the areas of the software lifecycle that inflict the greatest pain on engineering and DevOps teams.  That starts with CI/CD.  CI/CD pipelines are so named because they truly are pipelines, complete with a host of products and capabilities that spawn the creation of internal cottage industries with dedicated staff to manage them.  

At a high level, they’re all some variation of Figure 1 below:

<figure markdown>
  <img src="../diagrams/why_kubefox/high_level_deployment_workflow.svg" width=100% height=100%>
  <figcaption>Figure 1 - High Level Deployment Workflow</figcaption>
</figure>

KubeFox automates significant parts of this workflow.

In KubeFox, you build and deploy [**Systems**](concepts.md).  A **System** is simply a collection of [**Applications**](concepts.md), and applications are collections of [**Components**](concepts.md).  Components can be microservices or functions.  

The System construct exists for a few reasons:

- It enables KubeFox to help conveniently manage a group of applications that share components (and the sharing of components is good, as we’ve all been taught)
- It provides a simple control point for one or more applications.  In KubeFox, you deploy and release Systems.  When a System is deployed, KubeFox determines what components have changed, and distills the deployment to those unique components.

Building applications with KubeFox yields a host of benefits in the software lifecycle:

- Application lifecycle is focused on the Application component development and interaction – not on associated DevOps tasks and logistics.
- A primary goal of KubeFox is to enable engineering teams to interact with Kubernetes in an as frictionless a manner as possible.  Developers can rapidly prototype new concepts, enhance existing applications, and test their work.
- It is not incumbent on developers or QA / Release teams to manually ascertain whether a component has changed and manage component compatibility.  KubeFox automatically deploys the correct versions of components and distills deployments to only those components that are unique.  
- Different versions of the same application can coexist on the same cluster – without incurring the overhead and resource drain of needlessly duplicated unchanged components.
- Sophisticated deployments like Canary and Blue/Green are facilitated with KubeFox.  You can arrange for a subdomain for canary testing for early adopters, and run the advanced version of the application side-by-side with the current version.  
- KubeFox dynamically shapes traffic so shifting from one application version to another is lightswitch quick.  
- Environments are virtual constructs, enabling you to rapidly spool up and tear down isolated environments for teams or even individual developers.  Configuration is abstracted from environment, yielding great flexibility and fostering rapid engineering.
- Span-based federated telemetry (logs, metrics, traces) is immediately available, enabling you to view context-relevant application and component behavior.
- Your applications are zero trust from the beginning.

In the following sections, we’ll delve more deeply into some of KubeFox’s mkconcepts and capabilities.
