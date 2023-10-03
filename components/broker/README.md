# Broker

## Workflow

### gRPC Recv

- if evt has expired
  - drop evt
- else if evt has compId
  - if brk has sub for compId
    - put evt on sub's sendCh
  - else
    - put evt on jetstream's sendCh
- else if brk has sub group for comp
  - put evt on sub group's sendCh
- else
  - put evt on jetstream's sendCh
- put evt on comp's archive jetstream subject (disk)

### gRPC Send

- if evt has expired
  - drop evt
- else
  - send evt to component via gRPC stream.Send()

### JetStream Recv

- if evt has expired
  - drop evt
  - ack jetstream
- else if evt has compId
  - if brk has sub for compId
    - put evt on sub's sendCh
    - wait for evt to be processed
    - ack jetstream
- else if brk has sub group for comp
  - put evt on sub group's sendCh
  - wait for evt to be processed
  - ack jetstream
- else
  - nack jetstream
  - drop evt (error state, broker should not have been subscribed, log warn, and
    clean jetstream subs)

### JetStream Send

- if evt has expired
  - drop evt
- else if evt has compId
  - put evt on compId's stream (memory)
- else
  - put evt on comp's stream (memory)

### HTTP Srv Recv

HTTP Srv creates a subscription as if it is a component.

- convert http req to evt
- if evt has compId
  - if brk has sub for compId
    - put evt on sub's sendCh
  - else
    - put evt on jetstream's sendCh
- else if brk has sub group for comp
  - put evt on sub group's sendCh
- else
  - put evt on jetstream's sendCh
- put evt on comp's archive jetstream subject (disk)
- wait for resp
- send resp http resp

### HTTP Client Recv

HTTP Client creates a subscription as if it is a component.

- convert evt to http req
- send http req to external http srv
- wait for response
- convert resp to evt
- if brk has sub for compId
  - put evt on sub's sendCh
- else
  - drop evt (the src of req should always be on same node, if gone means it
    unsubscribed while http request was being processed)
- put evt on comp's archive jetstream subject (disk)

## Component Auth

<!--
component -> grpc recv
grpc recv -> grpc send
grpc recv -> js send
grpc recv -> js archive
gprc recv -> http cli
grpc send -> component (by id, by name/gitHash)
js send -> js subj (by id, by name/gitHash)
js archive -> js subj
js recv -> grpc send
http srv -> grpc send
http srv -> js send
http srv -> js archive
http cli -> grpc send
http cli -> js archive

gRPC Recv
  - gRPCRecvCh
gRPC Send
  - gRPCSendByIdCh (id)
  - gRPCSendByNameCh (compName.gitHash)
JS Send
  - jsSendByIdCh
  - jsSendByNameCh
  - jsArchiveCh
JS Recv
  - jsRecvByIdCh
  - jsRecvByNameCh
HTTP Srv
  - httpRecvCh
HTTP Cli
  - httpSendCh

gRPCRecvCh
httpRecvCh
jsRecvByIdCh
jsRecvByNameCh
-> recvCh -> engine -> *Send*Ch

Ch Payload
  - Event
  - Context
    - Source Service (enum)
    - Resp Callback (channel?)

per sub
  -

archive should be k/v
-->
