# Doodal Broker: A Broker Implemented in GO

ChizBroker is a fast and simple GRPC based implementation of kafka.
Features:
- Ready to be deployed on kubernetes
- Prometheus metrics
- All message get stored in DB (supporting PostgreSQL, Cassandra and local memory)

# Architecture
DoodalBroker can have several topics and each message published to certain topic will be broadcasted
to all subscribers to that topic.

