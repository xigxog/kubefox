# Conditions

Conditions provide a standard mechanism for higher-level status reporting from the KubeFox Operator. Conditions provide
summary information about resources without needing to understand resource-specific status details. They complement more
detailed information about the observed status of an object.

## Platform

| Type      | Status | Reason                      | Description                                          |
| --------- | ------ | --------------------------- | ---------------------------------------------------- |
| Available | True   | PlatformComponentsAvailable | KubeFox Broker, HTTP Server, and NATS are available. |
|           | False  | BrokerUnavailable           | KubeFox Broker is unavailable.                       |
|           |        | HTTPSrvUnavailable          | KubeFox HTTP Server is unavailable.                  |
|           |        | NATSUnavailable             | NATS is unavailable.                                 |

## AppDeployment

| Type        | Status | Reason                         | Description                                                                       |
| ----------- | ------ | ------------------------------ | --------------------------------------------------------------------------------- |
| Available   | True   | ComponentsAvailable            | Components have minimum required Pods available.                                  |
|             | False  | ComponentUnavailable           | One or more Components do not have minimum required Pods available.               |
|             |        | ProblemsFound                  | One or more problems exist with AppDeployment, see `status.problems` for details. |
| Progressing | False  | ComponentsDeployed             | Component Deployments completed successfully.                                     |
|             |        | ComponentDeploymentFailed      | One or more Component Deployments failed.                                         |
|             |        | ProblemsFound                  | One or more problems exist with AppDeployment, see `status.problems` for details. |
|             | True   | ComponentDeploymentProgressing | One or more Component Deployments are starting, scaling, or rolling out updates.  |

## VirtualEnvironment

| Type                   | Status | Reason                  | Description                                                                                                                    |
| ---------------------- | ------ | ----------------------- | ------------------------------------------------------------------------------------------------------------------------------ |
| ActiveReleaseAvailable | True   | ContextAvailable        | Release AppDeployment and Environment are available, Routes and Adapters are valid and compatible with the VirtualEnvironment. |
|                        | False  | NoRelease               | No Release made for VirtualEnvironment.                                                                                        |
|                        |        | ReleasePending          | No active Release, Release is pending activation.                                                                              |
|                        |        | ProblemsFound           | One or more problems exist with the active Release causing it to be unavailable, see `status.activeRelease` for details.       |
|                        |        | EnvironmentNotFound     | Environment does not exist.                                                                                                    |
| ReleasePending         | False  | ReleaseActivated        | Release was activated.                                                                                                         |
|                        |        | NoRelease               | Release for the VirtualEnvironment is not set.                                                                                 |
|                        |        | PendingDeadlineExceeded | Release was not activated within pending deadline.                                                                             |
|                        | True   | ProblemsFound           | One or more problems exist with Release preventing it from being activated, see `status.pendingRelease` for details.           |
|                        |        | EnvironmentNotFound     | Environment does not exist.                                                                                                    |
