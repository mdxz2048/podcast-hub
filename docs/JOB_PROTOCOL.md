# Job Protocol

## 1. Scope

The Job Protocol defines how the platform communicates with Connectors and how Connectors return events and outputs.

The protocol is intentionally file and stream based:

- Platform gives one JSON input document to one job.
- Connector emits JSON Lines events to stdout.
- Connector writes standardized episode JSON files and allowed artifacts to its assigned output directory.
- Connector exits after one import attempt.

## 2. Protocol Guarantees

The platform guarantees:

- A stable job ID.
- A single Source and Program context.
- A validated Connector package version.
- A prepared temporary workspace.
- A writable output directory.
- Runtime limits and timeout.
- Redacted input values where appropriate.

The Connector guarantees:

- It treats input as read-only.
- It writes only to the assigned output directory.
- It emits only JSON objects as protocol events.
- It does not print secrets.
- It exits when the one job is done, failed, cancelled, or timed out.

## 3. Job Input JSON

Example:

```json
{
  "schema_version": "1.0",
  "job": {
    "id": "job_01JZ7Y5F1J6Q0AZR4V0N7E8E9A",
    "attempt": 1,
    "ingestion_type": "connector",
    "trigger_type": "manual",
    "auth_mode": "reusable_session",
    "execution_mode": "unattended",
    "trigger_reason": "admin_run_now",
    "created_at": "2026-06-26T07:30:00Z",
    "deadline_at": "2026-06-26T07:45:00Z"
  },
  "program": {
    "id": "program_01JZ7Y3TD2Z2D5F4M8V7K6GQXE",
    "title": "Example Program",
    "language": "zh-CN",
    "timezone": "Asia/Shanghai"
  },
  "source": {
    "id": "source_01JZ7Y48M1AVJ3E4AEKYCY922T",
    "ingestion_type": "connector",
    "name": "Example Authorized Source",
    "external_ref": "example-program-123",
    "rights_note": "Operator has documented authorization to import and publish this program."
  },
  "connector": {
    "id": "example_authorized_source",
    "version": "0.1.0",
    "runtime": "python",
    "entrypoint": "connector.py"
  },
  "auth": {
    "mode": "reusable_session",
    "session_state": "valid",
    "session_ref": "session_ref_01JZ7Y6RAGXDBCD6GQ68K5VTK2",
    "expires_at": "2026-07-26T07:30:00Z"
  },
  "input": {
    "source_base_url": "https://api.example.invalid",
    "program_external_id": "example-program-123",
    "import_limit": 20
  },
  "paths": {
    "output_dir": "/workspace/output",
    "temp_dir": "/workspace/tmp"
  },
  "limits": {
    "timeout_seconds": 900,
    "max_output_mb": 2048,
    "max_episode_count": 100
  },
  "policy": {
    "network_allowed_hosts": [
      "api.example.invalid",
      "media.example.invalid"
    ],
    "allow_media_download": true,
    "allow_external_media_reference": true,
    "redact_event_fields": [
      "authorization",
      "cookie",
      "token",
      "password",
      "session"
    ]
  }
}
```

### 3.1 Field Rules

- `job.id` is immutable.
- `job.attempt` starts at 1.
- `job.ingestion_type` must be one of `native_rss`, `connector`, or `manual_upload`.
- `job.trigger_type` must be one of `manual` or `scheduled`.
- `job.auth_mode` must be one of `none`, `reusable_session`, or `qr_each_run`.
- `job.execution_mode` must be one of `unattended` or `interactive`.
- `manual_upload` must not execute a Connector.
- `qr_each_run` must use `trigger_type: manual` and `execution_mode: interactive`.
- `paths.output_dir` is the only writable output location.
- Secret values should be passed by opaque reference where possible.

## 4. JSON Lines Event Stream

Each stdout line must be a complete JSON object.

Common envelope:

```json
{
  "schema_version": "1.0",
  "event_id": "evt_01JZ7Y75EBC8G6NTKZR4B8S9PM",
  "job_id": "job_01JZ7Y5F1J6Q0AZR4V0N7E8E9A",
  "time": "2026-06-26T07:31:02Z",
  "level": "info",
  "type": "job.started",
  "message": "Connector started."
}
```

### 4.1 Event Types

Required event types:

- `job.started`
- `job.progress`
- `episode.discovered`
- `episode.output_written`
- `job.completed`

Conditional event types:

- `auth.qr_required`
- `auth.waiting`
- `auth.succeeded`
- `auth.failed`
- `warning`
- `error`
- `rate_limited`
- `output.summary`

### 4.2 Event Examples

```jsonl
{"schema_version":"1.0","event_id":"evt_001","job_id":"job_01JZ7Y5F1J6Q0AZR4V0N7E8E9A","time":"2026-06-26T07:31:00Z","level":"info","type":"job.started","message":"Connector started."}
{"schema_version":"1.0","event_id":"evt_002","job_id":"job_01JZ7Y5F1J6Q0AZR4V0N7E8E9A","time":"2026-06-26T07:31:05Z","level":"info","type":"job.progress","message":"Fetching episode index.","data":{"step":"fetch_index","current":1,"total":4}}
{"schema_version":"1.0","event_id":"evt_003","job_id":"job_01JZ7Y5F1J6Q0AZR4V0N7E8E9A","time":"2026-06-26T07:31:20Z","level":"info","type":"episode.discovered","message":"Episode discovered.","data":{"source_episode_id":"ep-1001","title":"Pilot Episode"}}
{"schema_version":"1.0","event_id":"evt_004","job_id":"job_01JZ7Y5F1J6Q0AZR4V0N7E8E9A","time":"2026-06-26T07:31:35Z","level":"info","type":"episode.output_written","message":"Episode output written.","data":{"path":"episodes/episode-001.json"}}
{"schema_version":"1.0","event_id":"evt_005","job_id":"job_01JZ7Y5F1J6Q0AZR4V0N7E8E9A","time":"2026-06-26T07:31:40Z","level":"info","type":"job.completed","message":"Connector completed.","data":{"episodes":1,"warnings":0}}
```

QR example:

```jsonl
{"schema_version":"1.0","event_id":"evt_qr_001","job_id":"job_01JZ7Y8T4D9WF7S1YC0TYP8JSC","time":"2026-06-26T08:00:00Z","level":"info","type":"auth.qr_required","message":"QR authentication required.","data":{"qr_ref":"qr_ref_01JZ7Y8YY7B6QQMS5C8DZ7B72N","expires_at":"2026-06-26T08:05:00Z"}}
{"schema_version":"1.0","event_id":"evt_qr_002","job_id":"job_01JZ7Y8T4D9WF7S1YC0TYP8JSC","time":"2026-06-26T08:01:12Z","level":"info","type":"auth.succeeded","message":"Interactive authentication completed."}
```

### 4.3 Event Redaction

Events must not include:

- Cookies.
- Tokens.
- Passwords.
- Authorization headers.
- QR session data.
- Raw session files.
- Personal secrets.

The platform should apply a second redaction pass before persistence and display.

## 5. Episode Output JSON

Each episode output file represents one staged episode candidate.

Example:

```json
{
  "schema_version": "1.0",
  "source": {
    "source_id": "source_01JZ7Y48M1AVJ3E4AEKYCY922T",
    "source_episode_id": "ep-1001",
    "canonical_url": "https://example.invalid/program/example-program-123/episodes/ep-1001",
    "imported_at": "2026-06-26T07:31:35Z"
  },
  "episode": {
    "title": "Pilot Episode",
    "subtitle": "A first authorized sample episode.",
    "description": "This episode is imported from an authorized source for review.",
    "language": "zh-CN",
    "explicit": false,
    "published_at": "2026-06-20T00:00:00Z",
    "duration_seconds": 1830,
    "episode_number": 1,
    "season_number": 1,
    "guid": "example-authorized-source:ep-1001"
  },
  "media": {
    "type": "audio/mpeg",
    "filename": "artifacts/audio/episode-001.mp3",
    "byte_size": 29200311,
    "checksum_sha256": "0000000000000000000000000000000000000000000000000000000000000000"
  },
  "artwork": {
    "filename": "artifacts/images/episode-001-cover.jpg",
    "type": "image/jpeg"
  },
  "provenance": {
    "connector_id": "example_authorized_source",
    "connector_version": "0.1.0",
    "job_id": "job_01JZ7Y5F1J6Q0AZR4V0N7E8E9A",
    "rights_note": "Operator has documented authorization to import and publish this episode."
  },
  "review_hints": {
    "duplicate_candidates": [],
    "warnings": [],
    "suggested_publication_scope": "selected_users"
  }
}
```

### 5.1 Required Episode Fields

Required:

- `schema_version`
- `source.source_id`
- `source.source_episode_id` or `episode.guid`
- `episode.title`
- `episode.description`
- `episode.language`
- `episode.published_at`
- `media.type`
- `provenance.connector_id`
- `provenance.connector_version`
- `provenance.job_id`

Required when media file is included:

- `media.filename`
- `media.byte_size`
- `media.checksum_sha256`

Allowed when media is referenced externally by policy:

- `media.external_url`
- `media.byte_size`, if known.
- `media.checksum_sha256`, if known.

## 6. Summary Output

Optional `summary.json` example:

```json
{
  "schema_version": "1.0",
  "job_id": "job_01JZ7Y5F1J6Q0AZR4V0N7E8E9A",
  "started_at": "2026-06-26T07:31:00Z",
  "finished_at": "2026-06-26T07:31:40Z",
  "episodes_discovered": 1,
  "episodes_written": 1,
  "warnings": [],
  "output_files": [
    "episodes/episode-001.json"
  ]
}
```

## 7. Exit Codes

Recommended exit codes:

- `0`: success.
- `10`: invalid input.
- `20`: authentication required or failed.
- `30`: source unavailable.
- `40`: rate limited.
- `50`: output write failure.
- `60`: unsupported source response.
- `70`: policy violation detected by Connector.
- `90`: unexpected Connector error.

The platform may classify errors using exit code, final event, and output validation result.

## 8. Error Classification

Platform-facing categories:

- `invalid_input`
- `auth_required`
- `auth_failed`
- `source_unavailable`
- `rate_limited`
- `connector_crash`
- `timeout`
- `output_invalid`
- `policy_violation`
- `cancelled`
- `unknown`

Retry policy should be based on classification:

- Retryable: `source_unavailable`, `rate_limited`, selected `connector_crash`.
- Not retryable until fixed: `invalid_input`, `output_invalid`, `policy_violation`.
- Requires human action: `auth_required`, `auth_failed`.

## 9. Idempotency and Duplicate Handling

Connectors should provide stable identifiers:

- `source_episode_id`
- `episode.guid`
- `source.canonical_url`

The platform should use these identifiers to avoid duplicate staging. If duplicates are uncertain, the Episode Review UI should ask a reviewer to resolve the conflict.

Retries must not overwrite prior job artifacts. Each attempt receives its own job ID or attempt namespace.

## 10. Manual Import Protocol

Manual import does not execute a Connector, but it should still create job-like records.

Manual import job input contains:

- Program ID.
- Source ID.
- `ingestion_type: manual_upload`.
- `trigger_type: manual`.
- Uploaded metadata.
- Uploaded media artifact references.
- Actor ID.
- Rights note.

Manual output should be normalized into the same Imported Episode schema so review and publication remain consistent.
