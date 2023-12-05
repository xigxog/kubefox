# Conditions

Conditions provide a standard mechanism for higher-level status reporting from
the KubeFox operator. Conditions provide summary information about resources
without needing to understand resource-specific status details. They complement
more detailed information about the observed status of an object

## Platform

| Type      | Status | Reason             | Description                                              |
| --------- | ------ | ------------------ | -------------------------------------------------------- |
| Available | True   |                    | The KubeFox Broker, HTTP Server, and NATS are available. |
|           | False  | BrokerUnavailable  | The KubeFox Broker is not available.                     |
|           |        | HTTPSrvUnavailable | The KubeFox HTTP Server is not available.                |
|           |        | NATSUnavailable    | The NATS is not available.                               |

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
| Available              | True   |                          | The AppDeployment and VirtualEnv specified are available.                                                                    |
|                        | False  | AppDeploymentUnavailable | The AppDeployment is not available.                                                                                          |
|                        |        | VirtualEnvUnavailable    | The VirtualEnv is not available.                                                                                             |
| AppDeploymentAvailable | True   |                          | The AppDeployment specified exists, matches the Release version, and is available.                                           |
|                        | False  | AppDeploymentNotFound    | The AppDeployment does not exist.                                                                                            |
|                        |        | AppDeploymentUnavailable | The AppDeployment has one or more component Pods that are not ready.                                                         |
|                        |        | VersionMismatch          | The AppDeployment and Release versions do not match.                                                                         |
| VirtualEnvAvailable    | True   |                          | The VirtualEnv specified has correctly typed variables, all required variables, and all required adapters present and valid. |
|                        | False  | VarMissing               | The VirtualEnv has one or more required variables missing.                                                                   |
|                        |        | VarConflict              | The VirtualEnv has one or more unique variables in conflict.                                                                 |
|                        |        | VarWrongType             | The VirtualEnv has one or more variables of the incorrect type.                                                              |
|                        |        | AdapterMissing           | The VirtualEnv has one or more Adapters missing.                                                                             |
|                        |        | AdapterInvalid           | The VirtualEnv has one or more Adapters that could not be validated.                                                         |
