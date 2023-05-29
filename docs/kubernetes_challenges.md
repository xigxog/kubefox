# Challenges in Kubernetes

It’s one thing but it’s a really big thing:  the lifecycle of software in the
Kubernetes world is fraught with complexity and requires continual analysis and
realignment.  And as a company and its teams grow, those difficulties increase
disproportionately.  

That analysis is performed by the scarcest and one of the most expensive teams
in any company:  DevOps.  DevOps needs to address a variety of overlapping goals
and deal with sometimes competing objectives:

- Providing a frictionless workflow to engineering, to enable rapid prototyping and iterations
- Application security, with the clear objective being zero trust
- Application monitoring, testing and debugging
- What tools will be chosen, how they’ll be implemented, provisioned, monitored and supported
- Microservice and Pod proliferation - and resultant and continuous increases in costly resources (compute, memory, storage)
- Providing for n - 1, n and n + 1 versions in multiple environments - a challenge magnified in dev, where the need to provide isolation results in drastic increases in resource demands
- Dealing with multiple vendors on a discrete basis for licenses, support contracts, maintenance and upgrades

… and this is just the tip of the iceberg
