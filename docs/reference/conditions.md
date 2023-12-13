# Conditions

Conditions provide a standard mechanism for higher-level status reporting from
the KubeFox operator. Conditions provide summary information about resources
without needing to understand resource-specific status details. They complement
more detailed information about the observed status of an object

## Platform

| Type      | Status | Reason                      | Description                                          |
| --------- | ------ | --------------------------- | ---------------------------------------------------- |
| Available | True   | ComponentsAvailable | KubeFox Broker, HTTP Server, and NATS are available. |
|           | False  | BrokerUnavailable           | KubeFox Broker is unavailable.                       |
|           |        | HTTPSrvUnavailable          | KubeFox HTTP Server is unavailable.                  |
|           |        | NATSUnavailable             | NATS is unavailable.                                 |

## AppDeployment

| Type        | Status | Reason                   | Description                                                                        |
| ----------- | ------ | ------------------------ | ---------------------------------------------------------------------------------- |
| Available   | True   | ComponentsAvailable      | Component Deployments have minimum required Pods available.                        |
|             | False  | ComponentsUnavailable    | One or more Component Deployments do not have minimum required Pods available.     |
| Progressing | True   | ComponentsProgressing    | One or more Component Deployments are scaling or rolling out updates.              |
|             | False  | ProgressDeadlineExceeded | One or more Component Deployments failed to show any progress within its deadline. |
|             |        | DeploymentError          | One or more Component Deployments failed update.                                   |

## Release

| Type        | Status | Reason                   | Description                                                                                                         |
| ----------- | ------ | ------------------------ | ------------------------------------------------------------------------------------------------------------------- |
| Available   | True   | CurrentAvailable         | Current Release AppDeployment, VirtualEnv, and Routes are available.                                                |
|             | False  | AppDeploymentUnavailable | Current Release AppDeployment is unavailable.                                                                       |
|             |        | RoutesUnavailable        | One or Component Routes of current Release are unavailable.                                                         |
|             |        | VirtualEnvUnavailable    | Current Release VirtualEnv is unavailable.                                                                          |
| Progressing | True   | ReleaseRequested         | New Release requested and is pending.                                                                               |
|             |        | ComponentsProgressing    | One or more Component Deployments of requested Release are progressing.                                             |
|             | False  | DeploymentError          | One or more Component Deployments of requested Release failed update or to show any progress within their deadline. |
|             |        | RoutesUnavailable        | One or more Component Routes of requested Release could not be parsed or VirtualEnv has invalid Route variables.    |
|             |        | VirtualEnvUnavailable    | VirtualEnv for requested Release is invalid, missing, or its release policy is violated.                            |

### Current/Requested Release

The following conditions are used to express the current state and requested
state of a Release. They are located at `status.current.conditions` and
`status.requested.conditions` paths of the Release object. If the current state
matches the requested state the `status.requested` block will not be present.

| Type                   | Status | Reason              | Description                                                                                                                                                   |
| ---------------------- | ------ | ------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| AppDeploymentAvailable | True   | ComponentsAvailable | AppDeployment exists, matches the Release version, and is available.                                                                                          |
|                        | False  | AdapterInvalid      | VirtualEnv contains one or more invalid Adapters.                                                                                                             |
|                        |        | AdapterMissing      | VirtualEnv is missing one or more Adapters.                                                                                                                   |
|                        |        | DeploymentError     | AppDeployment has one or more Component Deployments that failed.                                                                                              |
|                        |        | NotFound            | AppDeployment does not exist.                                                                                                                                 |
|                        |        | VarConflict         | VirtualEnv contains one or more conflicting unique Component variables.                                                                                       |
|                        |        | VarMissing          | VirtualEnv is missing one or more required Component variables.                                                                                               |
|                        |        | VarWrongType        | VirtualEnv contains one or more incorrectly typed Component variables.                                                                                        |
|                        |        | VersionMismatch     | AppDeployment and Release versions do not match.                                                                                                              |
| RoutesAvailable        | True   | Valid               | Routes were successfully parsed and VirtualEnv contains required Route variables, correctly typed Route variables, and no conflicting unique Route variables. |
|                        | False  | ParseError          | One or more Component Routes could not be parsed.                                                                                                             |
|                        |        | VarConflict         | VirtualEnv contains one or more conflicting unique Route variables.                                                                                           |
|                        |        | VarMissing          | VirtualEnv is missing one or more required Route variables.                                                                                                   |
|                        |        | VarWrongType        | VirtualEnv contains one or more incorrectly typed Route variables.                                                                                            |
| VirtualEnvAvailable    | True   | Valid               | Release does not violate VirtualEnv release policy and names match.                                                                                           |
|                        | False  | NotFound            | VirtualEnv does not exist.                                                                                                                                    |
|                        |        | NameMismatch        | VirtualEnv does does not match Release name.                                                                                                                  |
|                        |        | PolicyViolation     | Release violates VirtualEnv release policy.                                                                                                                   |
