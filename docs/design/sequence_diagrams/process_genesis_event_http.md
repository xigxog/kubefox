# Process Genesis Event

The diagrams below show [Genesis Event](../../concepts/index.md#genesis-event)
processing for [HTTP and HTTPS
Server](process_genesis_event_http.md#http-server) requests, and for the
[KubeFox Broker](process_genesis_event_http.md#broker).

### HTTP Server

<!-- 1. HTTP request comes in to http server and is converted to event.
2. If request contains deployment and environment headers/query params they are set on the event.
3. The default TTL, which can be set as a flag on broker, is set on event.
4. http server records request event id in map.
5. sends event to broker for routing
6. http server then blocks and waits either for ttl to expire or response to come from broker
8. if ttl expires 503 is returned to client
9. if response comes back it is converted to http response and sent back to
   client. -->

```mermaid
  sequenceDiagram
    autonumber
    actor client as Client
    participant http as HTTP/HTTPS Server
    participant map as Map
    participant broker as Broker

    client->>http: Client sends HTTP/HTTP request
    
    rect rgba(100, 100, 255, .5)
        Note right of client: Inbound Request
    end

    http->>http: Create Event
    opt Contains query parameters
        rect rgba(100, 100, 255, .5)
           Note right of http: If kf-dep and kf-ve query<br/>parameters present, set them in the event
        end
        http->>http: Augment Event
    end
    
    http->>http: Set ttl on Event

    http->>map: Record Event id in Map

    alt Process Event
        http->>broker: Send Event to Broker
    else ttl expires
        rect rgba(100, 100, 255, .5)
           Note left of http: If ttl expires, HTTP Server <br/> wakes up and returns 503 to Client
        end
        http->>client: Return HTTP 503 to Client
    else response from broker
        broker->>http: Event response sent by Broker
        http->>http: Convert Event to HTTP response
        http->>client: Return HTTP response
    end

```

### Broker

<!-- 1. Broker receives http request event from http server.
2. Event will either have no context which indicates it should match a release or have the deployment and environment set.
3. If environment or deployment is set without the other the request is rejected.
4. Based on the context a matcher for the specified deployment or for all released components is used. These are built by inspecting the deployed and released components and looking up their routes in the registration stored in the NATS.
5. If request does not match any routes a not found error is returned.
6. If matched the event target is updated with the matched component's name, commit, and matched route id.
7. broker checks if a matching component is subscribed locally via grpc and if so sends the event to that component.
8. otherwise the event is publish onto the component group's NATS subject so that another broker with a subscription for that component can process it.
9. ttl is updated before sending
 -->

```mermaid
  sequenceDiagram
    autonumber
    participant http as HTTP/HTTPS Server
    participant broker as Broker
    participant grpcs as gRPC Server
    participant nats as NATS
    participant component as Matched Component

    http->>broker: HTTP sends request event to Broker
        rect rgba(100, 100, 255, .5)
            Note over broker: Check for Query Parameters
        end
        alt Deployment (kf-dep) set<br/>but VE (kf-ve) not set
            broker->>http: Request Rejected
        else VE (kf-ve) set but<br/>Deployment (kf-dep) not set
            broker->>http: Request Rejected
        else
            rect rgba(100, 100, 255, .5)
                Note right of broker: Either:<br/>1. No query parameters set (Release) or<br/>2. kf-dep AND kf-ve set
            end
        end

    rect rgba(100, 100, 255, .5)
        Note right of broker: Match request event context in NATS<br/>registry to obtain route from<br/>deployment or release.
    end

    broker->>nats: Search NATS registry to identify route
    alt Matching route not found
        nats->>broker: Route not found
        broker->>http: Request Rejected
    else Matching route Found
        nats->>broker: Route found
        rect rgba(100, 100, 255, .5)
            Note right of broker: Update event target with:<br/>1. matched component's name,<br/>2. commit, and<br/>3. matched route id
        end
        broker->>broker: Update Event Target
    end

    rect rgba(100, 100, 255, .5)
        Note right of broker: Check if matched component<br/>is locally subscribed via gRPC
    end

    broker->>grpcs: Determine whether matched component<br/>is locally subscribed
    alt Matched component locally subscribed
        grpcs-->component: Local subscription
        grpcs->>broker: Component subscription found
        rect rgba(100, 100, 255, .5)
            Note left of grpcs: Update ttl and send event <br/> to locally-subscribed component
        end
        broker->>broker: Update ttl
        broker->>grpcs: Send event to component via gRPC
    else Matched component not locally subscribed
        grpcs->>broker: Component subscription not found
        rect rgba(100, 100, 255, .5)
            Note left of nats: Update ttl and publish to <br/>component group's NATS subject
        end
        broker->>broker: Update ttl
        broker->>nats: Publish event to NATS
    end

```

