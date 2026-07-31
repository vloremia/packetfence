# Hospitality Service

Standalone hospitality integration service for PacketFence. The current slice provides:

- PostgreSQL migrations for tenant, property, guest, reservation, session, plan, event, promotion, consent, review, and audit entities.
- Mock PMS connector with room and surname validation.
- PacketFence gateway interface, mock gateway, and REST client.
- Room-authentication service with keyed surname hashing, generic errors, checkout expiry, and PacketFence access grant.
- Versioned health, guest-auth, property, plan, and session endpoints.

## Local development

```sh
docker compose up --build
curl http://localhost:8080/health/live
```

Admin endpoints require `X-API-Key` and `X-Organisation-ID` where applicable. The mock PMS is empty by default; populate it through tests or replace it with a configured connector before enabling guest login.

Set `PACKETFENCE_API_URL` and `PACKETFENCE_API_TOKEN` to use the REST gateway. If the URL is unset, the service uses an in-memory gateway for development only. The current executable wires the mock PMS intentionally; production PMS connectors must be configured and selected before deployment.

## Verification

```sh
go test ./...
go vet ./...
```

The production deployment must use a secret manager, TLS termination, a real PMS connector, and the PacketFence REST gateway. The seeded legal-policy text is placeholder content and requires legal review.
