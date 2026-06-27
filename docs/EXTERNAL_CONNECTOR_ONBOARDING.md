# External Connector Onboarding

This document defines how future external Connectors may be prepared for Podcast Hub without putting source-specific code, secrets, or unauthorized content into the platform repository.

## Repository Boundaries

Podcast Hub main repository owns:

- platform domain model
- Connector package specification
- upload and static validation
- review and approval workflow
- Source configuration
- Import Job metadata
- Runner execution boundary
- staging intake
- review and publication state
- user authorization
- private media and RSS publishing

External Connector Workbench owns:

- source-specific adapter development
- local source-code inspection output that is safe to keep outside the platform repo
- local dependency experiments
- local fixture generation with non-sensitive test data

Recommended workbench location:

```text
/Users/lvzhipeng/Code/podcast-connectors-workbench/duoting/
```

A future duoting ZIP is an administrator-uploaded Connector package. It is not part of the Podcast Hub main repository and must not be committed here.

## Onboarding Flow

1. Confirm the operator owns, licenses, or is otherwise authorized to import and redistribute the selected content.
2. Create or use an external workbench outside this repository.
3. Inspect only non-sensitive source code and dependency descriptions.
4. Exclude all secrets, sessions, tokens, cookies, media, databases, caches, and logs.
5. Adapt the source-specific code to the standard Connector package contract.
6. Configure secrets locally through Podcast Hub Secret records, not through ZIP files or manifests.
7. Upload the Connector ZIP through the admin Connector Registry.
8. Review validation output and approve only after security review.
9. Create a Source manually.
10. Run the first Import Job for one confirmed authorized Program and at most one Episode.
11. Intake only into staging.
12. Publish only after admin review.
13. Grant user access explicitly before private RSS can include the content.

## Prohibited

- Bypassing copyright, paywalls, DRM, CAPTCHA, authentication controls, platform restrictions, or access controls.
- Importing or republishing content without authorization.
- Connecting to real source services from this repository.
- Adding duoting-specific code, scripts, or analysis documents to this repository.
- Including `.env`, cookies, sessions, tokens, API hashes, media, logs, databases, caches, or secrets in a Connector ZIP.
- Putting Secret values in `manifest.yaml`.
- Letting a Connector write directly to PostgreSQL, Redis, object storage control plane, Docker socket, or production secrets.
- Letting a Connector publish content or generate final RSS XML directly.
