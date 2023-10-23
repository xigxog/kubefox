# Broker

## Startup

1. On start up K8s service account token (SAT) is used to authenticate with
   Vault.
2. Vault verifies SAT with K8s API and issues certificate and private key.
3. Broker connects to NATS using mTLS with provided certificate and key.
4. Start gRPC server used by components using certificate.
5. Starts HTTP/HTTPS servers adapters.

## Component Connects

1. Component opens connections to broker's gRPC server using root CA to verify
   certificate.
2. Component calls subscribe service (2-way streaming) and sends a registration
   event, which contains component name, commit, id, routes, and K8s SAT.
3. Broker authorizes by verifying SAT with K8s API and ensures claims match
   attributes provided by component.
4. The broker starts a subscription for the component that is used to pass
   events.
5. Broker puts component's registration into NATS k/v store using the component
   name + commit as key (not id).
6. Broker creates JetStream consumer which listens for new messages on the
   component group's subject (evt.js.{compName}.{compCommit}).

## Process Genesis Event

## HTTP Server

1. Currently only HTTP server receives genesis events.
2. HTTP request comes in to http server and is converted to event.
3. If request contains deployment and environment headers/query params they are
   set on the event.
4. The default TTL, which can be set as a flag on broker, is set on event.
5. http server records request event id in map.
6. sends event to broker for routing
7. http server then blocks and waits either for ttl to expire or response to
   come from broker
8. if ttl expires 503 is returned to client
9. if response comes back it is converted to http response and sent back to
   client.

### Broker

1. Broker receives http request event from http server.
2. Event will either have no context which indicates it should match a release
   or have the deployment and environment set.
3. If environment or deployment is set without the other the request is
   rejected.
4. Based on the context a matcher for the specified deployment or for all
   released components is used. These are built by inspecting the deployed and
   released components and looking up their routes in the registration stored in
   the NATS.
5. If request does not match any routes a not found error is returned.
6. If matched the event target is updated with the matched component's name,
   commit, and matched route id.
7. broker checks if a matching component is subscribed locally via grpc and if
   so sends the event to that component.
8. otherwise the event is publish onto the component group's jetstream subject
   so that another broker with a subscription for that component can process it.
9. ttl is updated before sending

## Process Event from Component

### gRPC Server

1. Event is received by grpc server for an existing subscription.
2. grpc server verifies the event source matches the subscribed component, if it
   does not the broker drops the event and closes the subscription.
3. The grpc server passes the event to the broker for routing.

### Broker

1. The event should already have context set as it is not a genesis event, if it
   doesn't it is rejected and dropped.
2. The event source should be complete with component name, commit, and id
3. The event target should either contain only the component's name or be
   complete.
4. Either is not true event is dropped
5. If only target name is set then event is matched based on context and target
   commit and route id are added
6. If event is not matched it is dropped (currently there is no response sent to
   component)
7. the deployment or release are then pulled and broker verifies both the source
   and target components are part of the deployment spec.
8. if they are not event is dropped.
9. broker checks if a matching component is subscribed locally via grpc and if
   so sends the event to that component.
10. otherwise the event is publish onto the component's jetstream subject so
    that another broker with a subscription for that component can process it.
11. ttl is updated before sending

## Process Event from JetStream

### JetStream Client

1. Receives and unmarshals event, ttl is updated based on when msg was created.
2. Sends event to broker for routing.

### Broker

1. Workflow is same as for grpc but target should have both name and commit set.
2. If there is matching component it means that the component got unsubscribed
   while process the event, otherwise the broker would not have been listening
   on the subject. Is this case the event is republished onto the same subject.
