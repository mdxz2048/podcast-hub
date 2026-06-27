# Real Connector Pilot Checklist

This checklist is for a future pilot only. M1.4A does not perform any real duoting access, server login, Connector ZIP creation, Source creation, Secret injection, Import Job execution, media download, publication, or RSS release.

## Before Any Source Access

- Confirm written authorization for the content to be imported and republished.
- Confirm the first pilot scope: one Program and at most one Episode.
- Confirm the workbench is outside the Podcast Hub repository.
- Confirm no public deployment or real user distribution is part of the pilot setup.

## Future Duoting Pilot Sequence

1. Perform a server read-only inventory only after operator approval.
2. Download only non-sensitive source code and dependency descriptions.
3. Exclude:
   - `.env`
   - cookies
   - sessions
   - tokens
   - API hashes
   - media
   - logs
   - databases
   - caches
4. Adapt code in the external workbench:
   `/Users/lvzhipeng/Code/podcast-connectors-workbench/duoting/`
5. Reconfigure Secrets locally through Podcast Hub admin Secret records.
6. Create a Source manually.
7. Run the first Import Job for one confirmed authorized Program and at most one Episode.
8. Keep imported data in staging.
9. Do not auto-publish.
10. Do not auto-generate RSS.
11. Admin reviews staged metadata and media.
12. Admin publishes only authorized content.
13. Admin grants selected user access.
14. Private RSS includes content only after current authorization exists.

## Stop Conditions

Stop the pilot if any step requires:

- bypassing copyright, payment, DRM, CAPTCHA, authentication, platform restrictions, or access controls
- use of real Telegram sessions or QR login in the platform repository
- committing secrets or source-specific private materials
- running arbitrary Connector code on the host
- publishing before review
- giving private RSS access before user authorization

## Evidence To Collect Later

- Connector package manifest and validation report.
- Security review decision.
- Source configuration metadata without secrets.
- Import Job event log with redactions.
- Staging review record.
- Publication decision.
- User access grant record.

Do not collect or commit raw cookies, tokens, sessions, media files, private URLs, or secret payloads.
