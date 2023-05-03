# REST API

!!! note

    All endpoints listed should be prefixed with `api/kubefox/v0`

## API Verbs

| API Verb | HTTP Verb | Description                                                                                                          |
| -------- | --------- | -------------------------------------------------------------------------------------------------------------------- |
| List     | GET       | Lists available resources of specified type.                                                                         |
| Create   | POST      | Creates resource of specified type. For resources of type object, the object should be provided in the request body. |
| Get      | GET       | Retrieves the latest resource instance and status of specified type and name.                                        |
| Delete   | DELETE    | Deletes resource of specified type and name. Often the object for the resource is also deleted.                      |
| Update   | PUT       | Updates resource of specified type and name. If the resource does not exist it is created.                           |
| Patch    | PATCH     | Patches resource of specified type and name. Can only be performed on existing resources.                            |

## Resources

### Config

A config contains secrets and component configuration required for a component
to process an event.

**Supported refs:**

- id
- tag

**Endpoints:**

| API Verb | HTTP Verb | Path                      |
| -------- | --------- | ------------------------- |
| List     | GET       | `config`                  |
| Create   | POST      | `config/{name}`           |
| Get      | GET       | `config/{name}`           |
| Delete   | DELETE    | `config/{name}`           |
| Update   | PUT       | `config/{name}/metadata`  |
| Get      | GET       | `config/{name}/metadata`  |
| List     | GET       | `config/{name}/id`        |
| Get      | GET       | `config/{name}/id/{id}`   |
| List     | GET       | `config/{name}/tag`       |
| Update   | PUT       | `config/{name}/tag/{tag}` |
| Delete   | DELETE    | `config/{name}/tag/{tag}` |

### Environment

An environment is a named set of variables. These variables can be accessed by
components during runtime or used in route templates. An environment must
reference a config so that component configuration can be looked up when
processing events assigned to that environment.

**Supported refs:**

- id
- tag

**Endpoints:**

| API Verb | HTTP Verb | Path                           |
| -------- | --------- | ------------------------------ |
| List     | GET       | `environment`                  |
| Create   | POST      | `environment/{name}`           |
| Get      | GET       | `environment/{name}`           |
| Delete   | DELETE    | `environment/{name}`           |
| Update   | PUT       | `environment/{name}/metadata`  |
| Get      | GET       | `environment/{name}/metadata`  |
| List     | GET       | `environment/{name}/id`        |
| Get      | GET       | `environment/{name}/id/{id}`   |
| List     | GET       | `environment/{name}/tag`       |
| Update   | PUT       | `environment/{name}/tag/{tag}` |
| Delete   | DELETE    | `environment/{name}/tag/{tag}` |

### Platform

Subresources:

- Deployment
- Release

**Endpoints:**

| API Verb | HTTP Verb | Path                                                  |
| -------- | --------- | ----------------------------------------------------- |
| List     | GET       | `platform`                                            |
| Update   | PUT       | `platform/{name}`                                     |
| Update   | PATCH     | `platform/{name}`                                     |
| Get      | GET       | `platform/{name}`                                     |
| Update   | PUT       | `platform/{name}/metadata`                            |
| Get      | GET       | `platform/{name}/metadata`                            |
| Create   | POST      | `platform/{name}/deployment`                          |
| List     | GET       | `platform/{name}/deployment`                          |
| Get      | GET       | `platform/{name}/deployment/{system}/{refType}/{ref}` |
| Delete   | DELETE    | `platform/{name}/deployment/{system}/{refType}/{ref}` |
| Create   | POST      | `platform/{name}/release`                             |
| List     | GET       | `platform/{name}/release`                             |
| Get      | GET       | `platform/{name}/release/{system}/{environment}`      |
| Delete   | DELETE    | `platform/{name}/release/{system}/{environment}`      |

### System

A system is a collection of applications, components, and routes. They are
generated from a KubeFox Git repository tree. The id of a system is the hash of
the Git commit for that tree.

**Supported refs:**

- Id
- Branch
- Tag

**Endpoints:**

| API Verb | HTTP Verb | Path                            |
| -------- | --------- | ------------------------------- |
| List     | GET       | `system`                        |
| Create   | POST      | `system/{name}`                 |
| Get      | GET       | `system/{name}`                 |
| Delete   | DELETE    | `system/{name}`                 |
| Update   | PUT       | `system/{name}/metadata`        |
| Get      | GET       | `system/{name}/metadata`        |
| List     | GET       | `system/{name}/id`              |
| Get      | GET       | `system/{name}/id/{id}`         |
| List     | GET       | `system/{name}/branch`          |
| Update   | PUT       | `system/{name}/branch/{branch}` |
| Delete   | DELETE    | `system/{name}/branch/{branch}` |
| List     | GET       | `system/{name}/tag`             |
| Create   | POST      | `system/{name}/tag/{tag}`       |
| Delete   | DELETE    | `system/{name}/tag/{tag}`       |
