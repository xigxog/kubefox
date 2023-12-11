# Conditions

Conditions provide a standard mechanism for higher-level status reporting from
the KubeFox operator. Conditions provide summary information about resources
without needing to understand resource-specific status details. They complement
more detailed information about the observed status of an object

## Platform

| Type      | Status | Reason              | Description                                              |
| --------- | ------ | ------------------- | -------------------------------------------------------- |
| Available | True   | ComponentsAvailable | The KubeFox Broker, HTTP Server, and NATS are available. |
|           | False  | BrokerUnavailable   | The KubeFox Broker is not available.                     |
|           |        | HTTPSrvUnavailable  | The KubeFox HTTP Server is not available.                |
|           |        | NATSUnavailable     | The NATS is not available.                               |

## AppDeployment

| Type      | Status | Reason             | Description                                   |
| --------- | ------ | ------------------ | --------------------------------------------- |
| Available | True   | ComponentsReady    | All component Pods have Ready condition.      |
|           | False  | ComponentsNotReady | One or more component Pods is not ready.      |
| Deployed  | True   | ComponentsDeployed | All components were successfully deployed.    |
|           | False  | TODO               | One or more components could not be deployed. |

## Release

| Type                   | Status | Reason                   | Description                                                                                                                  |
| ---------------------- | ------ | ------------------------ | ---------------------------------------------------------------------------------------------------------------------------- |
| Available              | True   | Active                   | The AppDeployment and VirtualEnv specified are available and Release is active.                                              |
|                        | False  | AppDeploymentUnavailable | The AppDeployment is not available.                                                                                          |
|                        |        | VirtualEnvUnavailable    | The VirtualEnv is not available.                                                                                             |
| AppDeploymentAvailable | True   | ComponentsReady          | The AppDeployment specified exists, matches the Release version, and is available.                                           |
|                        | False  | ComponentsNotReady       | The AppDeployment has one or more component Pods that are not ready.                                                         |
|                        |        | NotFound                 | The AppDeployment does not exist.                                                                                            |
|                        |        | VersionMismatch          | The AppDeployment and Release versions do not match.                                                                         |
| VirtualEnvAvailable    | True   | Valid                    | The VirtualEnv specified has correctly typed variables, all required variables, and all required adapters present and valid. |
|                        | False  | AdapterInvalid           | The VirtualEnv has one or more Adapters that could not be validated.                                                         |
|                        |        | AdapterMissing           | The VirtualEnv has one or more Adapters missing.                                                                             |
|                        |        | NotFound                 | The VirtualEnv does not exist.                                                                                               |
|                        |        | PolicyViolation          | The Release violates the VirtualEnv release policy.                                                                          |
|                        |        | VarConflict              | The VirtualEnv has one or more unique variables in conflict.                                                                 |
|                        |        | VarMissing               | The VirtualEnv has one or more required variables missing.                                                                   |
|                        |        | VarWrongType             | The VirtualEnv has one or more variables of the incorrect type.                                                              |
