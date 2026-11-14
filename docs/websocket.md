# WebSocket events

Chirp pushes events to signed-in clients over one WebSocket per app session.
The socket is a transport, not a source of truth. Durable state stays in
Postgres and screens still reconcile through the HTTPS API after a reconnect.

Delivery is WebSocket only. There is no outbound HTTP webhook, but a webhook
dispatcher can reuse the same envelope from `server/internal/event`.

## Endpoint

```text
GET /ws
```

The server upgrades only after it verifies the access token and loads the
User. A missing, expired, revoked, or pending-Username token gets the normal
HTTP error envelope and no upgrade.

### Native handshake

```http
GET /ws HTTP/1.1
Upgrade: websocket
Connection: Upgrade
Authorization: Bearer <access_token>
Sec-WebSocket-Protocol: chirp.v1
```

### Browser handshake

Browsers cannot set handshake headers, so Expo web offers the token as a
second subprotocol entry. The server echoes `chirp.v1` and never echoes the
token entry.

```http
Sec-WebSocket-Protocol: chirp.v1, bearer.<access_token>
```

The token never goes in the query string. Query strings leak into access logs
and crash reports.

## Envelope

Every server frame is one JSON object.

```json
{
  "v": 1,
  "id": "evt_9f2c4a7d1b8e3c60a1b2c3d4e5f60718",
  "type": "connection.ready",
  "topic": "user:usr_123",
  "ts": "2026-11-14T12:00:00Z",
  "data": {
    "connection_id": "conn_4a91c3d2e8f07b65",
    "username": "ada",
    "heartbeat_seconds": 25,
    "access_expires_at": "2026-11-14T12:15:00Z"
  }
}
```

| Key | Type | Meaning |
| --- | --- | --- |
| `v` | number | Envelope version. Current value is `1`. |
| `id` | string | Stable event ID, unique per publish. Use it to deduplicate. |
| `type` | string | Catalog type, for example `connection.ready`. |
| `topic` | string | Delivery scope. `user:<user_id>` today. |
| `ts` | string | RFC 3339 creation time in UTC. |
| `data` | object | Type-specific payload. |

A client drops a frame with an unknown `v`, a missing required key, or
invalid JSON. It does not close the socket for a bad frame.

## Event catalog

| Type | Topic | Data |
| --- | --- | --- |
| `connection.ready` | `user:<self>` | `connection_id`, `username`, `heartbeat_seconds`, `access_expires_at` |
| `session.ended` | `user:<self>` | `reason`: `expired`, `revoked`, `signed_out`, or `server_shutdown` |

`connection.ready` is the first frame on every accepted socket. It confirms
the server accepted the token and tells the client which connection ID it
holds.

`session.ended` precedes a server-initiated close. The close frame carries
the machine-readable code in the table below.

Inbound frames are reserved in v1. The server reads them to keep the close
handshake and pong handler alive, enforces a 4096-byte frame limit, and
ignores their content.

## Close codes

| Code | Meaning | Client action |
| --- | --- | --- |
| `1000` | Normal close. | Stop. |
| `1001` | Server shutdown. | Reconnect with backoff. |
| `4001` | Access token expired. | Refresh once through `/auth/refresh`, then reconnect. |
| `4003` | Session ended on the server. | Stop and return to sign-in. |
| `4009` | Client fell behind and its queue filled. | Reconnect with backoff and reconcile over HTTPS. |
| `1011` | Server could not encode or write a frame. | Reconnect with backoff. |

## Connection rules

- One socket per signed-in app session. Multiple devices for the same User
  each get their own socket, and a publish reaches all of them.
- The server sends a ping every 25 seconds and drops a peer that misses the
  60-second pong deadline.
- The outbound queue is bounded. A full queue closes the connection with
  `4009` rather than growing memory.
- The connection closes with `4001` when the access token expires. The client
  refreshes and reconnects.
- Revoking a User's sessions closes every live socket for that User with a
  `session.ended` frame whose reason is `revoked` or `signed_out`, followed by
  close code `4003`.
- Reconnect uses jittered exponential backoff from 1 second to 30 seconds.

## Publishing an event

```go
envelope, err := event.New(
    event.UserTopic(user.ID),
    "message.created",
    payload,
    time.Now(),
)
if err != nil {
    return err
}
hub.Publish(realtime.Delivery{
    UserID: user.ID,
    Events: []event.Event{envelope},
})
```

Publish after the domain write commits. Do not publish inside a transaction
that can roll back.

## Configuration

| Variable | Where | Purpose |
| --- | --- | --- |
| `WS_ALLOWED_ORIGINS` | server | Comma-separated browser origins allowed to upgrade. A handshake carrying an `Authorization` header is treated as native and skips this filter. |
| `EXPO_PUBLIC_WS_URL` | mobile | Optional socket URL override. Defaults to `EXPO_PUBLIC_API_URL` with `http` swapped for `ws` and `/ws` appended. |

Outside release mode an empty `WS_ALLOWED_ORIGINS` allows the Expo web dev
origins `http://localhost:8081` and `http://localhost:19006`. Release mode
with an empty list allows only same-host browser origins. Set the variable
explicitly in production.

## Verification

```sh
cd server
go vet ./...
go test ./internal/event ./internal/realtime

cd ../chirp-mobile
npx tsc --noEmit
```
