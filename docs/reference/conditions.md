# Conditions

Conditions provide a standard mechanism for higher-level status reporting from
the KubeFox Operator. Conditions provide summary information about resources
without needing to understand resource-specific status details. They complement
more detailed information about the observed status of an object.

## Platform

| Type      | Status | Reason                      | Description                                          |
| --------- | ------ | --------------------------- | ---------------------------------------------------- |
| Available | True   | PlatformComponentsAvailable | KubeFox Broker, HTTP Server, and NATS are available. |
|           | False  | BrokerUnavailable           | KubeFox Broker is unavailable.                       |
|           |        | HTTPSrvUnavailable          | KubeFox HTTP Server is unavailable.                  |
|           |        | NATSUnavailable             | NATS is unavailable.                                 |

## AppDeployment

| Type        | Status | Reason                         | Description                                                                      |
| ----------- | ------ | ------------------------------ | -------------------------------------------------------------------------------- |
| Available   | True   | ComponentsAvailable            | Components have minimum required Pods available.                                 |
|             | False  | ComponentUnavailable           | One or more Components do not have minimum required Pods available.              |
|             |        | DependencyNotFound             | One or more Component dependencies are missing.                                  |
| Progressing | False  | ComponentsDeployed             | Component Deployments are complete.                                              |
|             |        | ComponentDeploymentFailed      | One or more Component Deployments failed.                                        |
|             | True   | ComponentDeploymentProgressing | One or more Component Deployments are starting, scaling, or rolling out updates. |

## VirtualEnv

| Type             | Status | Reason                   | Description                                                                                                                                            |
| ---------------- | ------ | ------------------------ | ------------------------------------------------------------------------------------------------------------------------------------------------------ |
| ReleaseAvailable | True   | AppDeploymentAvailable   | Release AppDeployment is available, Routes and Adapters are valid and compatible with the VirtualEnv.                                                  |
|                  | False  | ReleaseEmpty             | Release for the VirtualEnv is not set.                                                                                                                 |
|                  |        | ReleasePending           | Release for the VirtualEnv is pending activation.                                                                                                      |
|                  |        | AppDeploymentUnavailable | AppDeployment has one or more Components that do not have minimum required Pods available.                                                             |
|                  |        | AppDeploymentFailed      | AppDeployment does not exist, is incompatible with the VirtualEnv, or is unavailable and failed to update or to show any progress within its deadline. |
|                  |        | VirtualEnvSnapshotFailed | VirtualEnvSnapshot does not exist or snapshot was not of VirtualEnv.                                                                                   |
|                  |        | RouteProcessingFailed    | One or more Component Routes could not be parsed or are incompatible with VirtualEnv.                                                                  |
|                  |        | AdapterProcessingFailed  | One or more Adapters could not be parsed or are incompatible with VirtualEnv.                                                                          |
|                  |        | PolicyViolation          | Release violates VirtualEnv release policy.                                                                                                            |
| ReleasePending   | False  | ReleaseActive            | Release is active.                                                                                                                                     |
|                  |        | ReleaseEmpty             | Release for the VirtualEnv is not set.                                                                                                                 |
|                  | True   | ReleaseUpdated           | Release updated but is pending activation.                                                                                                             |
|                  |        | AppDeploymentUnavailable | AppDeployment has one or more Components that do not have minimum required Pods available.                                                             |
|                  |        | AppDeploymentFailed      | AppDeployment does not exist, is incompatible with the VirtualEnv, or is unavailable and failed to update or to show any progress within its deadline. |
|                  |        | VirtualEnvSnapshotFailed | VirtualEnvSnapshot does not exist or snapshot was not of VirtualEnv.                                                                                   |
|                  |        | RouteProcessingFailed    | One or more Component Routes could not be parsed or are incompatible with VirtualEnv.                                                                  |
|                  |        | AdapterProcessingFailed  | One or more Adapters could not be parsed or are incompatible with VirtualEnv.                                                                          |
|                  |        | PolicyViolation          | Release violates VirtualEnv release policy.                                                                                                            |
|                  |        | PendingDeadlineExceeded  | Release was not activated within pending deadline.                                                                                                     |
