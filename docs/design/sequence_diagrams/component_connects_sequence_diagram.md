# Component Connects

1. Component opens connections to broker's gRPC server using root CA to verify certificate.
2. Component calls subscribe service (2-way streaming) and sends a registration event, which contains component name, commit, id, routes, and K8s SAT.
3. Broker authorizes by verifying SAT with K8s API and ensures claims match attributes provided by component.
4. The broker starts a subscription for the component that is used to pass events.
5. Broker puts component's registration into NATS k/v store using the component name + commit as key (not id).
6. Broker creates NATS consumer which listens for new messages on the component
   group's subject (evt.js.{compName}.{compCommit}).


```mermaid
  sequenceDiagram
    participant k8s as Kubernetes
    participant bgrpc as Broker gRPC Server
    participant Component
    Participant NATS

    Component->>bgrpc: Open Connection <br/>with root CA
    
    Component->>bgrpc: Subscribe to 2-way<br/>Streaming Service
    
    Component->>bgrpc: Send registration event with<br/>component name, commit<br/>id, routes, k8s SAT

    critical Validate SAT and Component Claims
        bgrpc->>k8s: Validate SAT + Claims
    option SAT invalid
        bgrpc->>Component: Drop Connection
        Note over Component: [end]
    option Claims invalid
        bgrpc->>Component: Drop Connection
        Note over Component: [end]
    option SAT validated
        bgrpc->>bgrpc: Create subscription<br/>for requesting component
    end

    %%    bgrpc->>k8s: Verify SAT with k8s<br/>and validate Component claims
     
    %% Note over bgrpc: Create subscription<br/>for requesting component
    
    bgrpc->>NATS: Insantiate component's<br/>registration using<br/>component name + commit as key
    
    bgrpc->>NATS: Create NATS consumer<br/>listening on the group's subject<br/>evt.js.{compName}.{compCommit}
```