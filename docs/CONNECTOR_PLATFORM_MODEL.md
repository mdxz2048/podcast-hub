# Connector Platform Model

## 1) Connector

Connector is an administrator-uploaded package, not a built-in website feature.

Core attributes:

- package artifact (ZIP in future platform flow)
- immutable version
- `manifest.yaml`
- runtime profile declaration
- review and enablement status (for governance)

Connector itself is an integration tool, not user-facing content.

## 2) Source

Source binds runtime configuration to one specific `ConnectorVersion`.

Source stores:

- connector version reference
- non-sensitive startup parameters
- secret references only (not plaintext)
- execution dimensions:
  - trigger_type
  - auth_mode
  - execution_mode

Source is the operational unit for job triggering.

## 3) Program

Program is user-visible, subscribable content.

Program is created/matched/updated by import results from Sources and jobs.

Program is **not** Connector.

Users subscribe to Program/RSS outputs, never to Connector.

## 4) ImportJob

ImportJob represents one execution attempt of a Source.

It records:

- status transitions
- sanitized logs/events
- output artifact references
- failure classification and reason

ImportJob is execution history, not the Program entity itself.

## 5) ImportedEpisode

ImportedEpisode is per-episode metadata produced by Connector output contract.

Default lifecycle:

- created in staging
- initial state `review_pending`
- blocked from publish/RSS until approved

## 6) Secret Model

Secrets must follow strict boundary rules:

- not inside Connector ZIP
- not inside Git
- not inside logs/events
- injected only through controlled runtime secret mount/injection
- Source stores only secret reference IDs, never secret plaintext

## 7) First-Version Security Limits

Initial platform constraints:

- Python runtime only
- fixed runtime profile
- Connector package does not bring its own Dockerfile as execution authority
- no privileged runtime
- no Docker socket access
- no host filesystem mount
- no default unrestricted network
- no direct database writes by Connector
- no direct publish action by Connector
- no direct RSS generation by Connector

## 8) duoting Positioning

duoting is planned as a future **administrator-uploaded Connector** example.

It is:

- not main-repo built-in business code
- not a built-in Program
- not a user-facing product module

Only after first approved run can real Program/Source/Episode records be created from actual import results.
