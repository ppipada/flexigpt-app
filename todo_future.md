# Project Future notes

## Laundry list

- [ ] Inter font files need to be cached in repo
- [ ] Signing for Mac dmg and Win
- [ ] Better title deduce
- [ ] Mac install shows no title bar, better add a plain titlebar.
- [ ] Mac: Add how to install non signed pkg in readme. click > ok > settings > privacy and security > blocked > open anyway
- [ ] Mac icns file not present. Need proper icns file added
- [ ] Mac pkg install is add the .app in ~/build/bin. It should be in application. Check the settings, conversation and log location
- [ ] Test Windows build

## Tasks: Application Performance Management (APM)

- User/Org stats
- Usage stats
- Quota stats
- Dashboards
- Reports
- Mails

## Tasks: Org server - Entities

### Org

- Features

  - Create, read, update, delete (CRUD) operations
  - Manage teams and users within the organization
  - Set and manage preferences
  - View and manage expenditure records
  - Set and manage quotas

- Attributes: `org_id`, `name`, `address`, `contact_info`, `created_at`, `updated_at`
- Relationships: Has many `Teams`, Has many `Users`, Has many `Preferences`, Has many `ExpenditureRecords`

### Team

- Features

  - CRUD operations
  - Manage users within the team
  - Set and manage preferences
  - View and manage expenditure records
  - Set and manage quotas
  - Manage hierarchical team structures.

- Attributes: `team_id`, `name`, `org_id`, `created_at`, `updated_at`, `parent_team_id`
- Relationships: Belongs to `Org`, Has many `Users`, Has many `Preferences`, Has many `ExpenditureRecords`, Can have
  many `Teams` as members (self-referential relationship)

### User

- Features

  - CRUD operations
  - Set and manage individual preferences
  - View and manage personal expenditure records
  - Password reset and account recovery
  - Manage user authentication details
  - Implement robust authentication mechanisms using Okta/SAML/Direct/Social/SSO
  - Support for MFA

- Attributes: `user_id`, `username`, `email`, `org_id`, `team_id`, `created_at`, `updated_at`, `last_login`, `status`
  (active/inactive), `password_hash`, `mfa_enabled`, `role_id`
- Relationships: Belongs to `Org`, Belongs to `Team`, Has many `Preferences`, Has many `ExpenditureRecords`

## Tasks: Org server - Authorization

- Implement authorization using RBAC and ACLs
- Support MFA and maintain audit trails for security

- Role

  - Features

    - CRUD operations
    - Assign permissions to roles

  - Attributes: `role_id`, `name`, `description`, `created_at`, `updated_at`
  - Relationships: Has many `UserAuthorization`, Has many `RolePermission`

- Permission

  - Features

    - CRUD operations
    - Define actions that can be taken within the system

  - Attributes: `permission_id`, `name`, `description`, `created_at`, `updated_at`
  - Relationships: Has many `RolePermission`

- RolePermission

  - Features

    - Assign permissions to roles
    - Manage role-permission relationships

  - Attributes: `role_permission_id`, `role_id`, `permission_id`, `created_at`, `updated_at`
  - Relationships: Belongs to `Role`, Belongs to `Permission`

## Tasks: Org server - Notifications

- Features

  - Configure notifications for various events (e.g., new messages, quota breaches)
  - Support for email, SMS, and in-app notifications

- Attributes: `notification_id`, `user_id`, `type` (email/SMS/in-app), `message`, `status`, `created_at`, `updated_at`
- Relationships: Belongs to `User`

## Tasks: Org server - Preferences

- Features

  - CRUD operations
  - Store and retrieve settings for orgs, teams, and users
  - Provide entity-specific customizable dashboards
  - Export and import configuration templates

- Attributes: `preference_id`, `entity_type` (Org/Team/User), `entity_id`, `settings_json`, `created_at`, `updated_at`
- Relationships: Belongs to `Org`/`Team`/`User`

## Tasks: Org server - Secrets

- Features

  - CRUD operations
  - Store and manage secrets with access control
  - Integrate with external systems for secure access to secrets

- Attributes: `secret_id`, `entity_type` (Org/Team/User), `entity_id`, `key`, `value`, `created_at`, `updated_at`
- Relationships: Belongs to `Org`/`Team`/`User`

## Tasks: Org server - Expenditure

### ExpenditureRecord

- Features

  - Log and retrieve expenditure data
  - Historical data analysis
  - Expenditure forecasting
  - Alerts for quota breaches or unusual activities
  - Detailed dashboards and reports

- Attributes: `record_id`, `entity_type` (Org/Team/User), `entity_id`, `amount`, `currency`, `item_id`, `item_type`
  (Agent/Tool/Assistant/Prompt), `timestamp`
- Relationships: Belongs to `Org`/`Team`/`User`, Belongs to `Agent`/`Tool`/`Assistant`/`Prompt`

### Quota

- Features

  - CRUD operations
  - Set and manage quotas for orgs, teams, and users
  - Granular quotas for specific items (agents, tools, assistants, prompts)
  - Monitor and enforce quotas
  - Alerts and notifications for quota status

- Attributes: `quota_id`, `entity_type` (Org/Team/User), `entity_id`, `item_id`, `item_type`
  (Agent/Tool/Assistant/Prompt), `limit`, `created_at`, `updated_at`
- Relationships: Belongs to `Org`/`Team`/`User`, Belongs to `Agent`/`Tool`/`Assistant`/`Prompt`

## Pushed out list

- [ ] ~~logger is imported in securejsondb before it is set as ipc file logger in appimage~~
- [ ] ~~Electron: currently supported only via appimagelauncher. Better do snap I suppose, but explore later.~~
- [ ] ~~Electron: see about package distribution and updates~~
- [ ] Code interpreter for few languages baked in. This can be provided as a normal tool so that
- [ ] Container build and publish
