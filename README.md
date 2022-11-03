# Doodal Broker: A Broker Implemented in GO

ChizBroker is a fast and simple GRPC based implementation of kafka.
Features:
- Ready to be deployed on kubernetes
- Prometheus metrics
- All message get stored in DB (supporting PostgreSQL, Cassandra and local memory)

# Architecture
DoodalBroker can have several topics and each message published to certain topic will be broadcasted
to all subscribers to that topic.

# RPCs Description
- Publish Requst
```protobuf
message PublishRequest {
  string subject = 1;
  bytes body = 2;
  int32 expirationSeconds = 3;
}
```
- Fetch Request
```protobuf
message FetchRequest {
  string subject = 1;
  int32 id = 2;
}
```
- Subscribe Request
```protobuf
message SubscribeRequest {
  string subject = 1;
}
```
- RPC Service
```protobuf
service Broker {
  rpc Publish (PublishRequest) returns (PublishResponse);
  rpc Subscribe(SubscribeRequest) returns (stream MessageResponse);
  rpc Fetch(FetchRequest) returns (MessageResponse);
}
```
