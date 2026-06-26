# Permissions Matrix

## 1. Scope

Podcast Hub account roles are frozen as:

- `user`
- `admin`

`System Owner`, `Operator`, and `Reviewer` are not separate account roles. They are responsibility labels or permission profiles assigned to `admin` accounts.

`User` represents a normal `user` account.

## 2. Responsibility Profiles

### 2.1 System Owner

An `admin` responsibility profile for platform-wide governance, security defaults, system settings, and final operational authority.

### 2.2 Operator

An `admin` responsibility profile for day-to-day program/source/job operation, including manual triggers and QR task handling.

### 2.3 Reviewer

An `admin` responsibility profile for imported episode review, metadata checks, rights checks, and review decisions.

### 2.4 User

A normal `user` account that can browse authorized Programs, manage personal collections, and use personal RSS access.

## 3. Matrix

| Capability | System Owner | Operator | Reviewer | User |
| --- | --- | --- | --- | --- |
| Manage users | Yes | No | No | No |
| Manage programs | Yes | Yes | No | No |
| Manage sources | Yes | Yes | No | No |
| Manage Connector | Yes | Limited | No | No |
| Upload Connector | Yes | Yes, if granted | No | No |
| Manually trigger jobs | Yes | Yes | No | No |
| Handle QR tasks | Yes | Yes | No | No |
| View job logs | Yes | Yes | Yes, for review context | No |
| Review episodes | Yes | No | Yes | No |
| Publish RSS | Yes | Yes, if granted | No | No |
| Revoke RSS Token | Yes | Yes, if granted | No | Own token only |
| Modify system settings | Yes | No | No | No |

## 4. Permission Notes

- All admin actions are server-side permission checks in later implementation phases.
- M0 and M0.1 may display static permission states only; they must not implement real authorization.
- Permission-denied UI must use the shared `PermissionDenied` pattern.
- Admin responsibility labels can be represented in Mock data but must not be treated as account roles.
- Invitations are not implemented in M0. If shown, they are future static placeholders only and do not introduce an Invitation API or User status.

## 5. User Status Rules

Only these User statuses are valid:

- `pending_verification`
- `active`
- `suspended`
- `deleted`

Do not use `Pending invite` or `Disabled User` as User statuses.

