# Broker Startup

- On start up K8s service account token (SAT) is used to authenticate with Vault.
- Vault verifies SAT with K8s API and issues certificate and private key.
- Broker connects to NATS using mTLS with provided certificate and key.
- Start gRPC server used by components using certificate.
- Starts HTTP/HTTPS servers adapters.

```mermaid
  sequenceDiagram
    participant k8s as Kubernetes
    participant Vault
    participant Broker
    Participant NATS
    participant grpc as gRPC Server
    participant http(s) as HTTP/HTTPS Server Adapter

    Broker->>k8s: Obtain k8s SAT

    %% rect rgb(100, 223, 255) 
    %% Note over Broker: Here I am
    %% end

    critical Broker Authenticates with Vault
        Broker->>Vault: Authenticate with Vault<br/>using k8s SAT
        critical Validate SAT from Broker
          Vault->>k8s: Validate SAT with k8s
        option SAT invalid
            Vault->>Broker: Authentication failed
            Note over Broker: [end]
        option SAT validated
            Vault->>Broker: Issue Cert +<br/>Private Key
        end
    end
        
    Broker->>NATS: Connect via mTLS with<br/>Cert + Private Key
    
    Broker->>grpc: Start gRPC Server<br/>with Cert
    
    Broker->>http(s): Start HTTP/HTTPS Server<br/>Adapter with Cert
```







    