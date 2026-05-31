# CDC Runtime

Local CDC uses Debezium Kafka Connect against the existing Kafka broker.

## Recommendation

- Use Debezium PostgreSQL CDC for service-owned Postgres outbox tables.
- Keep the existing polling relay as a local fallback only.
- Use MongoDB change streams for Mongo CDC after the local Mongo deployment is changed to a replica set.

The current IAM connector template streams raw inserts from `iam.public.message_outbox` into Kafka.
It does not use Debezium's Outbox Event Router yet because the current table stores the complete
Podzone envelope in `envelope_json` instead of the default Debezium outbox columns
`aggregatetype`, `aggregateid`, `type`, and `payload`.

## Start

```sh
docker compose -f deployments/docker/infras.yml up -d postgres kafka cdc-connect
```

## Register IAM Outbox CDC

```sh
curl -sS -X POST http://localhost:8083/connectors \
  -H 'Content-Type: application/json' \
  --data @deployments/docker/cdc/connectors/iam-message-outbox-raw.json
```

The raw connector topic is:

```text
podzone.cdc.iam.public.message_outbox
```

## Next Schema Step

For first-class CDC outbox routing, add a v2 outbox table shape:

```text
id
aggregate_type
aggregate_id
event_type
payload
topic
message_key
created_at
```

Then the connector can enable Debezium `io.debezium.transforms.outbox.EventRouter` and route
directly to the event topic stored in `topic`.
