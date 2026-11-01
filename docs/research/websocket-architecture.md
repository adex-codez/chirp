# WebSocket Architecture for Chirp

> Research note based on repository context and primary sources: IETF RFCs, official Go/Gin/React Native/Redis/PostgreSQL documentation, and first-party project documentation. Statements marked **Protocol fact** describe specified behavior; **Recommendation** is a design choice for Chirp; **Inference** applies those facts to Chirp's current stack.

## Repository context

Chirp is a social mobile app. The backend is Go 1.25 with Gin and Postgres (`server/go.mod`); the mobile app uses Expo and React Native (`chirp-mobile/package.json`). The authentication ADR specifies 15-minute JWT access tokens and rotating opaque refresh tokens, with refresh tokens stored hashed in Postgres (`docs/adr/0002-access-15m-rotating-refresh.md`). No WebSocket implementation or message-domain model is present yet. This means real-time delivery should be introduced as a transport over domain services, not as a second source of truth.

## 1. What opening a connection means

**Protocol fact.** A WebSocket begins as an HTTP/1.1 GET Upgrade handshake. The client requests `Upgrade: websocket`; a successful server response is `101 Switching Protocols`, after which both endpoints exchange framed messages independently over the connection. `wss://` runs this over TLS. RFC 6455 defines optional subprotocol negotiation through `Sec-WebSocket-Protocol`, the browser `Origin` header, and Ping/Pong control frames. [RFC 6455, §§1.2, 4, 5.5, 10](https://www.rfc-editor.org/rfc/rfc6455.html)

**Recommendation.** Expose one authenticated endpoint, for example `GET /ws`, and use a versioned application envelope with a small set of message types. Keep the endpoint and envelope independent of individual features; authorize each command and subscription separately. Negotiate a subprotocol such as `chirp.v1` only if compatibility/version negotiation is useful; it does not itself provide authentication or authorization. Gin handlers have access to the underlying `http.Request` and `ResponseWriter`, and Gin documents upgrading a route using a WebSocket implementation. [Gin WebSocket support](https://gin-gonic.com/en/docs/server-config/websocket/)

### Authentication at open

**Protocol fact.** RFC 6455 allows authentication-related HTTP headers in the handshake, and permits the server to return an HTTP error before switching protocols. The RFC's browser origin model is separate from client authentication; native clients can supply arbitrary Origin values, so Origin is not identity. [RFC 6455, §§4, 10. Origin guidance](https://www.rfc-editor.org/rfc/rfc6455.html)

**Recommendation.** Authenticate before upgrading: validate Chirp's short-lived access JWT, resolve the User and session claims, check revocation policy, then upgrade. For React Native, use the WebSocket API's supported handshake headers if the shipped runtime allows it; React Native documents WebSocket support but its API page is still marked work in progress, so verify the actual Expo/iOS/Android runtime behavior before settling the client contract. [React Native networking](https://reactnative.dev/docs/network), [React Native WebSocket API](https://reactnative.dev/docs/global-WebSocket)

If client header support is unreliable, issue a short-lived, single-use WebSocket ticket over authenticated HTTPS, then pass that ticket in the connection URL and consume it atomically at handshake. Do not put a long-lived access or refresh token in a URL: URLs are commonly retained in access logs and diagnostics. This ticket approach is a **recommendation**, not a WebSocket protocol requirement. Check Origin against an explicit allowlist for browser clients; for native app origins, define and test the expected behavior rather than trusting arbitrary Origin. Always use `wss://` outside local development. [RFC 6455, §§4, 10](https://www.rfc-editor.org/rfc/rfc6455.html)

## 2. Connection lifecycle and reconnect

**Protocol fact.** A connection is bidirectional until a close handshake or transport failure. Ping/Pong can serve as keepalive or responsiveness checks; RFC 6455 requires a Pong response to a received Ping unless Close has already been received. [RFC 6455, §§5.5, 7](https://www.rfc-editor.org/rfc/rfc6455.html)

**Recommendation.** Model each live connection explicitly: `connecting → authenticated → active → closing → closed`. On open, register it with a connection manager keyed by User ID and connection ID; on close, remove it idempotently and leave rooms/subscriptions. Use a heartbeat interval shorter than the smallest relevant proxy/load-balancer idle timeout, and expire peers that fail a Pong deadline. Coordinate graceful shutdown by refusing new upgrades, notifying existing clients, allowing a bounded drain, then closing remaining sockets.

The mobile client should reconnect with exponential backoff and jitter, reset the delay after stable operation, refresh access credentials through the existing HTTPS refresh flow when needed, and reconcile missed data after reconnect. Treat disconnects as expected: mobile networks, app backgrounding, and process restarts make permanent sockets impossible to guarantee. Do not rely on a successful connection as proof that events were delivered.

## 3. One transport, distinct feature semantics

**Recommendation.** Use one connection per logged-in app session, with typed commands/events, rather than one socket per screen or feature. Give each event a stable ID, type, schema version, creation time, and relevant aggregate/room ID. Keep request/response commands distinct from server-pushed events. The domain service remains responsible for validating state changes; the WebSocket layer authenticates, decodes bounded messages, invokes those services, and routes authorized events.

| Use case | Suggested delivery model | Durability and reconnect behavior |
| --- | --- | --- |
| Chat messages | Persist message first; publish an event to conversation members | Durable history is authoritative. On reconnect fetch after last seen message/event cursor; deduplicate by ID. |
| Presence | Ephemeral per-User online state with leases/heartbeat and multiple-device counts | Presence may be approximate and rebuilt from active connections; expire stale leases. |
| Typing/read indicators | Ephemeral, scoped to conversation and sender, with short expiry/coalescing | Drop when congested or disconnected; current state can be re-emitted. |
| Notifications | Persist notification and unread state, then send a wake-up/update event | On reconnect fetch notification state from HTTPS/API; socket is a low-latency hint, not the inbox. |
| Rooms/groups | Server-authorized subscription to a conversation or topic | Verify membership on subscribe and on sensitive commands; never accept arbitrary room membership from the client. |

These are architecture recommendations. The split follows the different loss tolerance of durable chat/notification state versus transient presence/typing. Keeping durable records in Postgres also fits Chirp's current stack and avoids coupling correctness to a live connection.

## 4. Go connection handling and backpressure

**First-party library guidance.** Gorilla WebSocket documents one concurrent reader and one concurrent writer per connection; applications must serialize writes and reads accordingly. Its upgrader also has an Origin check, with a same-host default when no custom check is set. [Gorilla WebSocket package documentation](https://pkg.go.dev/github.com/gorilla/websocket)

**Recommendation.** Use one read loop and one write loop per connection. Producers enqueue to a bounded outbound channel; only the write loop touches socket write methods. Set read limits, read deadlines, write deadlines, and handlers for Ping/Pong/Close. Validate decoded payload size, schema, and command rate before invoking business logic. A full outbound queue must have an explicit policy: drop/coalesce transient events (typing/presence), or disconnect a slow client so it can recover from durable state. Do not allow an unbounded per-client queue; a slow peer otherwise turns fanout into growing memory use.

For performance, avoid one database query per recipient and avoid serializing the same event repeatedly when possible. Batch recipient lookup, encode once per compatible payload, and measure before adding compression: compression costs CPU and may be inappropriate for secret-bearing mixed payloads. Bound connections per User/device/IP and concurrent handshakes; measure memory per idle connection, write queue depth, bytes and messages per second, and fanout size before tuning limits.

## 5. Multi-instance scaling and delivery guarantees

With a single Go process, an in-memory map from User/room to connections is sufficient for routing. It is process-local and cannot deliver to a User connected to another instance. **Inference:** once Chirp runs multiple backend replicas, a shared fanout mechanism or deliberate connection affinity is needed; affinity alone does not solve cross-instance events.

**Recommendation.** Keep the first implementation behind a small event-bus interface. Start with an in-process bus for one instance. When replicas require cross-node delivery, add a broker. Redis Pub/Sub is a reasonable low-latency fanout path for ephemeral updates, but Redis documents at-most-once delivery: disconnected subscribers miss messages. Use it only as a hint alongside durable Postgres state. If replay/at-least-once processing is required, use a durable stream/queue and make consumers idempotent. [Redis Pub/Sub delivery semantics](https://redis.io/docs/latest/develop/pubsub/), [Redis Streams and Pub/Sub use cases](https://redis.io/docs/latest/develop/use-cases/pub-sub/)

PostgreSQL `LISTEN`/`NOTIFY` can be a lightweight signal between processes, but not a message history: notifications are delivered to currently listening sessions, and PostgreSQL documents a setup race that requires LISTEN commit followed by state inspection before relying on future notifications. Use it for simple invalidation/wakeup, not durable event delivery. [PostgreSQL LISTEN](https://www.postgresql.org/docs/current/sql-listen.html), [PostgreSQL NOTIFY](https://www.postgresql.org/docs/current/sql-notify.html)

For durable event publication, a transaction that updates domain rows and publishes separately can fail between the two operations. **Recommendation/inference:** use a transactional outbox when missing a live notification is unacceptable: commit domain state and an outbox row together, then publish asynchronously and mark/retry the outbox. Clients still reconcile from authoritative APIs, and event IDs support duplicate suppression.

## 6. Security and authorization

- Use TLS (`wss://`) in production; RFC 6455 recommends TLS and describes confidentiality/integrity as provided by WebSocket over TLS. [RFC 6455, §10.6](https://www.rfc-editor.org/rfc/rfc6455.html)
- Validate browser Origin against an exact allowlist as a defense against cross-site WebSocket use. Origin is not authentication, especially for native clients. [RFC 6455, §10.2](https://www.rfc-editor.org/rfc/rfc6455.html)
- Authenticate at handshake and authorize every command and room subscription. Re-check or close connections when the access token/session expires or is revoked; Chirp's 15-minute JWT lifetime makes token expiration handling part of the connection lifecycle. The refresh token belongs only in the existing HTTPS refresh flow.
- Treat every frame as untrusted input. Enforce message-size limits, schema/type allowlists, rate limits, and conversation membership checks; use parameterized DB operations through existing service/repository boundaries. RFC 6455 explicitly requires validation of incoming data. [RFC 6455, §10.7](https://www.rfc-editor.org/rfc/rfc6455.html)
- Avoid logging message bodies, credentials, or tokens. Log connection open/close, User/session identifiers in privacy-safe form, origin, close reason, auth/authorization failures, malformed messages, rate-limit events, queue drops, and abnormal disconnects. OWASP's WebSocket guidance recommends lifecycle/security logging and excluding sensitive payloads. [OWASP WebSocket Security Cheat Sheet](https://cheatsheetseries.owasp.org/cheatsheets/WebSocket_Security_Cheat_Sheet.html)

## 7. Observability

Record counters/gauges for active connections, handshakes accepted/rejected, close codes, reconnects, auth failures, inbound/outbound bytes, queue depth and drops, fanout recipients, and broker publish/consume failures. Add histograms for handshake duration, event enqueue-to-write latency, and command handling time. Correlate each message with an event ID and request/connection ID, but avoid high-cardinality labels such as User ID on metrics. This is a **recommendation** for diagnosing resource pressure and delivery gaps; RFC 6455 defines protocol close codes and Ping/Pong but does not prescribe application telemetry.

## Recommended first architecture

1. Add an authenticated `/ws` route in Gin using a maintained Go WebSocket implementation; authenticate before upgrade and enforce Origin policy where relevant.
2. Build a connection manager with one reader and one writer loop per socket, bounded queues, heartbeat/deadline handling, and idempotent cleanup.
3. Define a versioned typed envelope and server-authorized subscription commands. Keep domain writes in existing services and persist chat/notification state in Postgres before publishing events.
4. Initially route events in-process. Add Redis Pub/Sub only when multiple replicas need live cross-node fanout; use durable records/reconciliation for missed events, and add an outbox/stream only when event-loss requirements justify the added operations.
5. Use transient delivery for typing/presence and durable reconciliation for messages/notifications. Instrument queues, fanout, latency, connection counts, and closure causes from the start.

**Tradeoff.** This starts with low operational complexity and lets Chirp ship all feature types over a shared transport. The cost is that in-process routing does not span replicas; the bus seam and durable data model make that a deliberate later change. A broker improves cross-node fanout but introduces operations, duplicate/loss semantics, and failure modes that must be made explicit.

## Sources

- [IETF RFC 6455 — The WebSocket Protocol](https://www.rfc-editor.org/rfc/rfc6455.html)
- [Gin WebSocket support](https://gin-gonic.com/en/docs/server-config/websocket/)
- [React Native networking and WebSocket support](https://reactnative.dev/docs/network)
- [React Native WebSocket API](https://reactnative.dev/docs/global-WebSocket)
- [Gorilla WebSocket Go package documentation](https://pkg.go.dev/github.com/gorilla/websocket)
- [Redis Pub/Sub](https://redis.io/docs/latest/develop/pubsub/)
- [Redis Pub/Sub use cases and Streams guidance](https://redis.io/docs/latest/develop/use-cases/pub-sub/)
- [PostgreSQL LISTEN](https://www.postgresql.org/docs/current/sql-listen.html)
- [PostgreSQL NOTIFY](https://www.postgresql.org/docs/current/sql-notify.html)
- [OWASP WebSocket Security Cheat Sheet](https://cheatsheetseries.owasp.org/cheatsheets/WebSocket_Security_Cheat_Sheet.html)
