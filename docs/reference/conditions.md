# Conditions

Conditions provide a standard mechanism for higher-level status reporting from
the KubeFox operator. Conditions provide summary information about resources
without needing to understand resource-specific status details. They complement
more detailed information about the observed status of an object

## Platform

| Type      | Status | Reason              | Description                                              |
| --------- | ------ | ------------------- | -------------------------------------------------------- |
| Available | True   | ComponentsAvailable | KubeFox Broker, HTTP Server, and NATS are available. |
|           | False  | BrokerUnavailable   | KubeFox Broker is not available.                     |
|           |        | HTTPSrvUnavailable  | KubeFox HTTP Server is not available.                |
|           |        | NATSUnavailable     | NATS is not available.                               |

## AppDeployment

| Type      | Status | Reason             | Description                                   |
| --------- | ------ | ------------------ | --------------------------------------------- |
| Available | True   | ComponentsReady    | All component Pods have Ready condition.      |
|           | False  | ComponentsNotReady | One or more component Pods is not ready.      |
| Deployed  | True   | ComponentsDeployed | All components were successfully deployed.    |
|           | False  | TODO               | One or more components could not be deployed. |

## Release

| Type                   | Status | Reason                   | Description                                                                                                              |
| ---------------------- | ------ | ------------------------ | ------------------------------------------------------------------------------------------------------------------------ |
| Available              | True   | Active                   | AppDeployment and VirtualEnv specified are available and Release is active.                                              |
|                        | False  | AppDeploymentUnavailable | AppDeployment is not available.                                                                                          |
|                        |        | VirtualEnvUnavailable    | VirtualEnv is not available.                                                                                             |
| AppDeploymentAvailable | True   | ComponentsReady          | AppDeployment specified exists, matches the Release version, and is available.                                           |
|                        | False  | ComponentsNotReady       | AppDeployment has one or more component Pods that are not ready.                                                         |
|                        |        | NotFound                 | AppDeployment does not exist.                                                                                            |
|                        |        | VersionMismatch          | AppDeployment and Release versions do not match.                                                                         |
| VirtualEnvAvailable    | True   | Valid                    | VirtualEnv specified has correctly typed variables, all required variables, and all required adapters present and valid. |
|                        | False  | AdapterInvalid           | VirtualEnv has one or more Adapters that could not be validated.                                                         |
|                        |        | AdapterMissing           | VirtualEnv has one or more Adapters missing.                                                                             |
|                        |        | NotFound                 | VirtualEnv does not exist.                                                                                               |
|                        |        | PolicyViolation          | Release violates VirtualEnv release policy.                                                                              |
|                        |        | VarConflict              | VirtualEnv has one or more unique variables in conflict.                                                                 |
|                        |        | VarMissing               | VirtualEnv has one or more required variables missing.                                                                   |
|                        |        | VarWrongType             | VirtualEnv has one or more variables of the incorrect type.                                                              |
