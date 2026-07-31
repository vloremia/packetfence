# PacketFence Hospitality Platform — Architecture Assessment

## 1. Current PacketFence Architecture

### 1.1 Language Stack

| Layer | Language | Framework | Location |
|-------|----------|-----------|----------|
| Captive Portal | Perl | Catalyst MVC | `html/captive-portal/lib/captiveportal/` |
| Admin API | Perl | Mojolicious | `lib/pf/UnifiedApi.pm` + `lib/pf/UnifiedApi/` |
| API Gateway | Go | Gin + Gorilla Mux + Chi | `go/api-frontend/` |
| Admin UI | JavaScript | Vue.js 2 | `html/pfappserver/` |
| RADIUS | C/Perl | FreeRADIUS + Perl modules | `raddb/` |
| Background Jobs | Perl + Go | pfqueue + cron | `lib/pf/pfqueue/`, `go/cron/` |
| DNS | Go | CoreDNS + plugin | `go/plugin/coredns/` |
| DHCP | Perl + Go | Custom + pool | `lib/pf/dhcp/`, `go/dhcp/` |
| Database | SQL | MySQL/MariaDB via GORM (Go) + RoseDB (Perl) | `db/` |
| Cache/Queue | - | Redis + Kafka | `conf/redis_*.conf`, `conf/kafka/` |
| Config | Go + Perl | pfconfig driver | `go/pfconfigdriver/`, `lib/pfconfig/` |

### 1.2 Key Source Sizes

- Perl core (`lib/pf/`): 1021 files
- Go services (`go/`): 453 files
- Switch modules (`lib/pf/Switch/`): 207 modules across 60+ vendors
- Captive portal (`html/captive-portal/`): 299 files
- Admin UI (`html/pfappserver/`): 386 lib + 596 root (Vue.js SPA)
- Configuration (`conf/`): 258 files
- Tests (`t/`): 1336 files

### 1.3 Existing Authentication Sources (35 types)

AD, AdminProxy, Authorization, AzureAD, Billing, Blackhole, Clickatell, EAP-TLS, EDIR, Eduroam, Email, Facebook, Github, Google, GoogleWorkspaceLDAP, HTTP, Htpasswd, Kerberos, Kickbox, LDAP, LinkedIn, Null, OAuth, OpenID, PayPal, Potd, RADIUS, SAML, SMS, SQL, SponsorEmail, Stripe, Twilio, WindowsLive

### 1.4 Network Device Support (tested/untested matrix)

Cisco (23 modules), Aruba (9), HP (13), Nortel (10), Avaya (6), Dlink (7), Extreme (6), Juniper (8), Meraki (4), Ruckus (5), Ubiquiti (2), Fortinet (2), Mikrotik, Cambium, PaloAlto — 60+ vendors via `pf::SwitchSupports` capability system.

Capabilities tracked per module: `supportsDot1x`, `supportsMacAuth`, `supportsCoA`, `supportsDynamicVlan`, `supportsQoS`, `supportsCaptivePortal`, etc.

### 1.5 Database Schema (current — v15.1)

Core tables (MySQL):

| Table | Purpose | Relevance |
|-------|---------|-----------|
| `person` | User/guest profile — pid, firstname, lastname, email, telephone, room_number, lang, etc. | **Directly reusable** |
| `password` | Credentials — pid, password, access_duration, access_level, category, unregdate | **Reusable** |
| `node` | Device — mac, pid, category_id, status, unregdate, time_balance, bandwidth_balance | **Directly reusable** |
| `node_category` | Roles — category_id, name, max_nodes_per_pid, acls | **Reusable** |
| `locationlog` | Location tracking — mac, switch, port, vlan, role, ssid, session_id | **Reusable** |
| `radacct` | RADIUS accounting | **Reusable** |
| `billing` | Payment records | **Extendable** |
| `auth_log` | Authentication attempts | **Reusable** |
| `admin_api_audit_log` | Admin actions | **Reusable** |
| `activation` | Activation codes | **Reusable** |
| `security_event` | Security events | **Reusable** |
| `node_current_session` | Online session tracking | **Reusable** |
| `chi_cache` | Key-value cache | **Reusable** |

### 1.6 Portal Module System

Profiles (`profiles.conf`) map connection profiles to portal modules. Portal modules (`portal_modules.conf`) define a tree of policy modules: Root → Choice → Authentication::Login, Authentication::Billing, Authentication::Choice (OAuth/SAML/social), Provisioning, MFA.

Already has: billing module (Stripe, PayPal), social login, SMS/email OTP, self-registration, sponsor approval.

### 1.7 Container Architecture

Docker containers for all services: radiusd, pfacct, pfdetect, pfdhcp, pfdns, pfqueue, pfconfig, pfcron, pfperl-api, pfpki, pfsso, httpd.portal, httpd.webservices, haproxy, redis, kafka, fingerbank, proxySQL.

## 2. Proposed Hospitality Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                      PacketFence Core                       │
│  (RADIUS, NAC, Captive Portal, VLAN, Roles, Enforcement)    │
└───────────────────┬─────────────────────────────────────────┘
                    │ REST API │ RADIUS │ Webhooks │ CoA
┌───────────────────▼─────────────────────────────────────────┐
│               PacketFence Gateway (Adapter)                  │
│  createGuestIdentity | assignRole | grantAccess | revokeAccess│
│  registerDevice | disconnectSession | applyBandwidthPolicy   │
│  getSessionStatus | setExpiration | getNetworkHealth         │
└───────────────────┬─────────────────────────────────────────┘
                    │
┌───────────────────▼─────────────────────────────────────────┐
│               Hospitality Integration Service                │
│  (Go/NestJS — new separate service)                          │
│                                                              │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────────┐  │
│  │ PMS Connector │  │ Auth Service │  │ Guest Lifecycle  │  │
│  │ Framework     │  │ (Room+Name,  │  │ (Check-in/out    │  │
│  │               │  │  Email, SMS, │  │  sync, expiry)   │  │
│  │ • OPERA       │  │  Voucher,    │  │                  │  │
│  │ • Cloudbeds   │  │  Social)     │  │ Reconcile jobs   │  │
│  │ • Mews        │  └──────────────┘  └──────────────────┘  │
│  │ • Mock PMS    │                                           │
│  └──────────────┘  ┌──────────────┐  ┌──────────────────┐  │
│                     │ Wi-Fi Plans  │  │ Marketing & CRM  │  │
│                     │ & Billing    │  │ • Consents       │  │
│                     │ • Tiers      │  │ • Promotions     │  │
│                     │ • Upgrade    │  │ • Reviews        │  │
│                     │ • Folio      │  │ • Guest profile  │  │
│                     └──────────────┘  └──────────────────┘  │
│                                                              │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────────┐  │
│  │ Events       │  │ Multi-Prop   │  │ Webhooks &       │  │
│  │ • Conference │  │ • Tenants    │  │ Integrations     │  │
│  │ • Vouchers   │  │ • Isolation  │  │ • Webhook svc    │  │
│  │ • QR codes   │  │ • RBAC       │  │ • Job queue      │  │
│  └──────────────┘  └──────────────┘  └──────────────────┘  │
└─────────────────────────────────────────────────────────────┘
                    │ PostgreSQL │ Redis │ BullMQ
┌───────────────────▼─────────────────────────────────────────┐
│                  Hospitality Database                         │
│  (Separate PostgreSQL instance)                              │
│  properties, pms_configs, reservations, guest_profiles,      │
│  wifi_plans, entitlements, payments, events, promotions,     │
│  consents, review_campaigns, audit_events, integration_health│
└─────────────────────────────────────────────────────────────┘
                    │
┌───────────────────▼─────────────────────────────────────────┐
│              Hospitality Admin Portal (Vue.js SPA)            │
│  Properties | PMS | Sessions | Plans | Events | Promotions   │
│  Reviews | Consents | Reports | Users/Roles | Audit          │
└─────────────────────────────────────────────────────────────┘
```

## 3. Components to Reuse

### 3.1 Directly Reusable (no modifications needed)

| Component | Path | Purpose |
|-----------|------|---------|
| RADIUS auth/acct | `raddb/` | Guest authentication, accounting |
| RADIUS CoA | `raddb/` + Switch modules | Session disconnect, VLAN changes |
| Switch modules | `lib/pf/Switch/` | Network device integration |
| Node/device registration | `lib/pf/node.pm` | Device management |
| Role/category system | `lib/pf/nodecategory.pm`, `node_category` table | Role assignment |
| VLAN assignment | Various | Traffic isolation |
| Location log | `locationlog` table | Guest location tracking |
| Audit logging | `admin_api_audit_log` table | Admin action audit trail |
| Rate limiter | `lib/pf/rate_limiter.pm` | Brute force protection |
| Password hashing | `lib/pf/password.pm` | Secure credential storage |
| I18N framework | `lib/pf/I18N/` | Multi-language support |
| Configuration system | `conf/` + pfconfig driver | Service configuration |
| Monitoring | Netdata, Prometheus | Infrastructure monitoring |
| iptables/ip6tables | `lib/pf/iptables.pm` | Firewall rules |

### 3.2 Extendable with Configuration

| Component | Extension Method |
|-----------|-----------------|
| Connection profiles (`profiles.conf`) | Add new profile + portal modules |
| Portal modules (`portal_modules.conf`) | Add hospitality login modules |
| Authentication sources | Add external auth source type |
| Billing tiers (`billing_tiers.conf`) | Extend with Wi-Fi plan tiers |
| Person table (`room_number` field) | Extend through hospitality service |
| Email templates (`templates/emails/`) | Add marketing/review email templates |
| Captive portal content (`content/`) | Add splash page templates |

### 3.3 Reusable with Configuration Only

- FreeRADIUS virtual servers in `raddb/sites-available/`
- Redis cache/queue infrastructure
- Kafka message bus (existing but unused in parts)
- DNS audit logging
- DHCP fingerprinting
- Provisioning framework

## 4. Components to Extend (Minimal Changes)

| Component | Change Required | Risk |
|-----------|----------------|------|
| `lib/pf/web.pm` | Add `stash_template_vars` hooks for hospitality portal data | Low — hook already exists |
| `person` table | Already has `room_number` — no change needed | None |
| `portal_modules.conf` | Add `default_hotel_login_policy` and `default_hotel_policy` | Low |
| `profiles.conf` | Add hotel profile with custom modules | Low |
| Captive portal login template | Branded hotel login page | Low |

## 5. Components to Isolate Behind Adapters

| PacketFence Component | Adapter Interface | Purpose |
|----------------------|-------------------|---------|
| RADIUS per-user config | `PacketFenceGateway.assignRole()` / `applyBandwidthPolicy()` | Dynamic role/VLAN via RADIUS or API |
| Node management | `PacketFenceGateway.createGuestIdentity()` / `registerDevice()` | Device registration |
| Session management | `PacketFenceGateway.getSessionStatus()` / `disconnectSession()` | CoA for session control |
| Node expiration | `PacketFenceGateway.setExpiration()` / `revokeAccess()` | Checkout-based expiry |
| Portal templates | Custom templates in profile | Branded captive portal |
| Authentication sources | External auth source via RADIUS or API | Room auth via hospitality service |

## 6. New Services to Introduce

### 6.1 Hospitality Integration Service
- **Stack**: Go (matching existing Go microservices) with PostgreSQL
- **Rationale**: Go is already the established microservices language in PacketFence (api-frontend, cron, pfqueue, etc.); adding NestJS/TypeScript would introduce a second runtime and increase operational complexity
- **Alternative**: Perl Mojolicious (matching `pf::UnifiedApi`) — possible but Go is preferred for performance and type safety
- **Database**: Separate PostgreSQL instance (not PacketFence's MySQL) — clean separation, independent upgrade cycle
- **Job Queue**: Redis + Kafka (already deployed in PacketFence infrastructure)

### 6.2 Hospitality Admin Portal
- **Stack**: Vue.js SPA extending existing `pfappserver` or standalone
- **Preference**: Extend `html/pfappserver/` with new modules for property/guest management
- **Authentication**: SSO through existing PacketFence admin auth

### 6.3 Guest Captive Portal
- **Stack**: PacketFence's existing Catalyst portal + custom templates
- **Customization**: Add hotel login flows as new dynamic routing modules
- **Mobile-first**: Use existing responsive template system

## 7. Upgrade-Safe Extension Strategy

### 7.1 Golden Rules

1. **Never modify** PacketFence core database tables directly from hospitality code
2. **Never patch** PacketFence Perl/Go core files — use hooks, plugins, config, and adapters
3. **All hospitality data** lives in separate PostgreSQL database
4. **All integration** goes through `PacketFenceGateway` adapter (REST API + RADIUS attributes)
5. **Configuration** for hospitality features lives in hospitality service config, not PacketFence conf files

### 7.2 Upgrade-Safe Integration Points

| Integration Point | Mechanism | Upgrade-Safe? |
|------------------|-----------|--------------|
| Guest identity creation | PF REST API: `POST /api/v1/nodes` (existing) | Yes |
| Role assignment | PF REST API: `PATCH /api/v1/nodes/:mac` (existing) | Yes |
| Session disconnect | PF REST API: device actions endpoint | Yes |
| Bandwidth policy | RADIUS attributes via `radreply` | Yes |
| Auth validation | External auth source (RADIUS proxy or HTTP API) | Yes |
| Portal login flow | New portal module (config-driven) | Yes |
| Email/SMS | Existing activation/send system or new provider | Yes |
| Configuration | `pfconfig` custom namespace | Yes |

## 8. Database Schema — Hospitality Entities (PostgreSQL)

```sql
-- Multi-tenant hierarchy
CREATE TABLE organisations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE hotel_groups (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organisation_id UUID NOT NULL REFERENCES organisations(id),
    name VARCHAR(255) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE properties (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    hotel_group_id UUID REFERENCES hotel_groups(id),
    organisation_id UUID NOT NULL REFERENCES organisations(id),
    name VARCHAR(255) NOT NULL,
    code VARCHAR(50) NOT NULL,           -- PMS property code
    timezone VARCHAR(50) NOT NULL DEFAULT 'UTC',
    country VARCHAR(2) NOT NULL,
    language VARCHAR(10) NOT NULL DEFAULT 'en',
    address TEXT,
    phone VARCHAR(50),
    email VARCHAR(255),
    check_in_time TIME NOT NULL DEFAULT '14:00',
    check_out_time TIME NOT NULL DEFAULT '11:00',
    grace_period_minutes INT NOT NULL DEFAULT 120,
    access_fallback_policy VARCHAR(50) NOT NULL DEFAULT 'deny',
    pms_outage_grace_minutes INT NOT NULL DEFAULT 30,
    consent_policy_version VARCHAR(50),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- PMS Configuration (encrypted)
CREATE TABLE pms_configurations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    property_id UUID NOT NULL REFERENCES properties(id),
    pms_type VARCHAR(50) NOT NULL,        -- opera, opera_cloud, cloudbeds, mews, mock
    api_endpoint VARCHAR(512) NOT NULL,
    api_key_encrypted BYTEA,             -- encrypted at application level
    api_secret_encrypted BYTEA,          -- encrypted at application level
    username_encrypted BYTEA,
    password_encrypted BYTEA,
    token_endpoint VARCHAR(512),
    client_id_encrypted BYTEA,
    sandbox_mode BOOLEAN NOT NULL DEFAULT false,
    rate_limit_per_second INT NOT NULL DEFAULT 10,
    timeout_seconds INT NOT NULL DEFAULT 30,
    health_status VARCHAR(50) NOT NULL DEFAULT 'unknown',
    last_health_check_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Guest Profile (minimal — data minimisation)
CREATE TABLE guest_profiles (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    property_id UUID NOT NULL REFERENCES properties(id),
    pms_guest_id VARCHAR(255),            -- PMS internal guest ID
    first_name VARCHAR(255),
    last_name_hash BYTEA,                -- hashed surname for lookup
    email_hash BYTEA,                    -- hashed email for lookup
    email_encrypted BYTEA,               -- reversibly encrypted for operational use
    mobile_hash BYTEA,
    mobile_encrypted BYTEA,
    country VARCHAR(2),
    preferred_language VARCHAR(10),
    loyalty_status VARCHAR(50),
    marketing_consent BOOLEAN NOT NULL DEFAULT false,
    marketing_consent_at TIMESTAMPTZ,
    privacy_policy_version VARCHAR(50),
    privacy_consent_at TIMESTAMPTZ,
    first_connected_at TIMESTAMPTZ,
    last_connected_at TIMESTAMPTZ,
    device_count INT NOT NULL DEFAULT 0,
    visit_count INT NOT NULL DEFAULT 0,
    source VARCHAR(50),                   -- how they authenticated
    tags JSONB,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Reservation (from PMS)
CREATE TABLE reservations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    property_id UUID NOT NULL REFERENCES properties(id),
    pms_reservation_id VARCHAR(255) NOT NULL,
    pms_guest_id VARCHAR(255),
    guest_profile_id UUID REFERENCES guest_profiles(id),
    room_number VARCHAR(50) NOT NULL,
    guest_name_hash BYTEA NOT NULL,       -- hashed surname for lookup
    first_name VARCHAR(255),
    last_name_hash BYTEA NOT NULL,
    check_in_at TIMESTAMPTZ NOT NULL,
    check_out_at TIMESTAMPTZ NOT NULL,
    status VARCHAR(50) NOT NULL,          -- reserved, checked_in, checked_out, cancelled, no_show
    adults INT NOT NULL DEFAULT 1,
    children INT NOT NULL DEFAULT 0,
    booking_source VARCHAR(255),
    booking_confirmation VARCHAR(255),
    vip_status BOOLEAN NOT NULL DEFAULT false,
    loyalty_number VARCHAR(255),
    pms_data JSONB,                       -- raw PMS data for debugging
    last_synced_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(property_id, pms_reservation_id)
);

-- Guest Session (maps to PacketFence node)
CREATE TABLE guest_sessions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    property_id UUID NOT NULL REFERENCES properties(id),
    reservation_id UUID REFERENCES reservations(id),
    guest_profile_id UUID REFERENCES guest_profiles(id),
    mac_address VARCHAR(17) NOT NULL,
    packetfence_node_mac VARCHAR(17),      -- PF node MAC
    packetfence_pid VARCHAR(255),          -- PF person ID
    wifi_plan_id UUID,
    ip_address INET,
    ssid VARCHAR(255),
    status VARCHAR(50) NOT NULL DEFAULT 'active',
    -- active, expired, revoked, disconnected
    expires_at TIMESTAMPTZ NOT NULL,
    started_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    ended_at TIMESTAMPTZ,
    auth_method VARCHAR(50) NOT NULL,      -- room, email, sms, voucher, social
    device_count INT NOT NULL DEFAULT 1,
    max_devices INT NOT NULL DEFAULT 2,
    data_usage_bytes BIGINT NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Wi-Fi Plans
CREATE TABLE wifi_plans (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    property_id UUID NOT NULL REFERENCES properties(id),
    name VARCHAR(255) NOT NULL,
    code VARCHAR(50) NOT NULL,             -- basic, premium, vip, conference, staff
    description TEXT,
    download_speed_mbps INT,
    upload_speed_mbps INT,
    data_allowance_bytes BIGINT,
    session_duration_minutes INT,
    max_devices INT NOT NULL DEFAULT 2,
    concurrent_sessions INT NOT NULL DEFAULT 1,
    packetfence_role VARCHAR(255),         -- PF category/role
    vlan_id INT,
    bandwidth_policy JSONB,
    price_cents INT NOT NULL DEFAULT 0,
    tax_percent DECIMAL(5,2) NOT NULL DEFAULT 0,
    currency VARCHAR(3) NOT NULL DEFAULT 'USD',
    is_paid BOOLEAN NOT NULL DEFAULT false,
    is_complimentary BOOLEAN NOT NULL DEFAULT false,
    is_public BOOLEAN NOT NULL DEFAULT true,
    sort_order INT NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Guest Entitlements (actual assignment of plan to guest)
CREATE TABLE wifi_entitlements (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    guest_session_id UUID REFERENCES guest_sessions(id),
    wifi_plan_id UUID NOT NULL REFERENCES wifi_plans(id),
    property_id UUID NOT NULL REFERENCES properties(id),
    guest_profile_id UUID REFERENCES guest_profiles(id),
    status VARCHAR(50) NOT NULL DEFAULT 'active',
    -- active, expired, revoked, pending_payment
    price_cents INT NOT NULL DEFAULT 0,
    currency VARCHAR(3) NOT NULL DEFAULT 'USD',
    source VARCHAR(50) NOT NULL,           -- complimentary, purchased, upgrade, staff
    payment_id UUID,
    starts_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    expires_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Payments
CREATE TABLE payments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    property_id UUID NOT NULL REFERENCES properties(id),
    guest_profile_id UUID REFERENCES guest_profiles(id),
    guest_session_id UUID,
    wifi_entitlement_id UUID,
    amount_cents INT NOT NULL,
    currency VARCHAR(3) NOT NULL DEFAULT 'USD',
    status VARCHAR(50) NOT NULL,           -- pending, completed, failed, refunded
    payment_method VARCHAR(50),            -- folio, card, promotion, staff
    folio_charge_id VARCHAR(255),         -- PMS folio reference
    payment_gateway VARCHAR(50),
    payment_gateway_ref VARCHAR(255),
    idempotency_key VARCHAR(255) UNIQUE,
    receipt_reference VARCHAR(255),
    failure_reason TEXT,
    refunded_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Events (Conferences)
CREATE TABLE events (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    property_id UUID NOT NULL REFERENCES properties(id),
    name VARCHAR(255) NOT NULL,
    description TEXT,
    venue VARCHAR(255),
    organiser VARCHAR(255),
    organiser_email VARCHAR(255),
    starts_at TIMESTAMPTZ NOT NULL,
    ends_at TIMESTAMPTZ NOT NULL,
    ssid VARCHAR(255),
    vlan_id INT,
    packetfence_role VARCHAR(255),
    max_attendees INT,
    max_devices_per_attendee INT NOT NULL DEFAULT 2,
    shared_bandwidth_mbps INT,
    per_user_bandwidth_mbps INT,
    access_code VARCHAR(50) UNIQUE NOT NULL,
    splash_page_theme VARCHAR(255),
    banner_url TEXT,
    acceptable_use_policy TEXT,
    status VARCHAR(50) NOT NULL DEFAULT 'draft',
    -- draft, active, completed, cancelled
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Event Attendees
CREATE TABLE event_attendees (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    event_id UUID NOT NULL REFERENCES events(id),
    property_id UUID NOT NULL REFERENCES properties(id),
    email_hash BYTEA,
    first_name VARCHAR(255),
    last_name_hash BYTEA,
    voucher_code VARCHAR(50),
    qr_code TEXT,
    session_id UUID,
    status VARCHAR(50) NOT NULL DEFAULT 'registered',
    -- registered, checked_in, expired
    checked_in_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Promotions
CREATE TABLE promotions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    property_id UUID NOT NULL REFERENCES properties(id),
    name VARCHAR(255) NOT NULL,
    type VARCHAR(50) NOT NULL,            -- offer, upsell, direct_booking
    title VARCHAR(255),
    description TEXT,
    image_url TEXT,
    cta_text VARCHAR(255),
    cta_url TEXT,
    target_segment VARCHAR(50),           -- all, returning, loyalty, conference
    target_language VARCHAR(10),
    target_country VARCHAR(2),
    starts_at TIMESTAMPTZ,
    ends_at TIMESTAMPTZ,
    display_location VARCHAR(50),         -- portal_login, post_login, email, sms
    sort_order INT NOT NULL DEFAULT 0,
    is_active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Promotion interactions
CREATE TABLE promotion_interactions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    promotion_id UUID NOT NULL REFERENCES promotions(id),
    guest_profile_id UUID REFERENCES guest_profiles(id),
    guest_session_id UUID,
    interaction_type VARCHAR(50) NOT NULL, -- impression, click, conversion, dismissal
    booking_reference VARCHAR(255),
    revenue_attribution_cents INT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Review campaigns
CREATE TABLE review_campaigns (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    property_id UUID NOT NULL REFERENCES properties(id),
    name VARCHAR(255) NOT NULL,
    delay_hours INT NOT NULL DEFAULT 24,
    channel VARCHAR(50) NOT NULL,          -- email, sms
    template_id VARCHAR(255),
    google_review_url TEXT,
    tripadvisor_review_url TEXT,
    internal_survey_enabled BOOLEAN NOT NULL DEFAULT false,
    escalation_threshold INT,             -- rating ≤ this triggers escalation
    escalation_email VARCHAR(255),
    is_active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Review requests
CREATE TABLE review_requests (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    campaign_id UUID NOT NULL REFERENCES review_campaigns(id),
    guest_session_id UUID,
    guest_profile_id UUID REFERENCES guest_profiles(id),
    reservation_id UUID REFERENCES reservations(id),
    property_id UUID NOT NULL REFERENCES properties(id),
    status VARCHAR(50) NOT NULL DEFAULT 'pending',
    -- pending, sent, opened, completed, opted_out
    sent_at TIMESTAMPTZ,
    opened_at TIMESTAMPTZ,
    rating INT,
    feedback_text TEXT,
    external_review_submitted BOOLEAN NOT NULL DEFAULT false,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Consent records
CREATE TABLE consents (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    guest_profile_id UUID REFERENCES guest_profiles(id),
    property_id UUID NOT NULL REFERENCES properties(id),
    consent_type VARCHAR(50) NOT NULL,     -- terms, privacy, marketing, analytics
    granted BOOLEAN NOT NULL,
    policy_version VARCHAR(50) NOT NULL,
    policy_text TEXT,
    language VARCHAR(10) NOT NULL,
    source VARCHAR(50) NOT NULL,           -- portal_login, registration, admin
    ip_address INET,
    user_agent TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Policy versions
CREATE TABLE policy_versions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    property_id UUID NOT NULL REFERENCES properties(id),
    policy_type VARCHAR(50) NOT NULL,      -- terms, privacy, marketing
    version VARCHAR(50) NOT NULL,
    content TEXT NOT NULL,
    language VARCHAR(10) NOT NULL,
    published_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(property_id, policy_type, version, language)
);

-- Audit events
CREATE TABLE audit_events (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    property_id UUID,
    actor_type VARCHAR(50) NOT NULL,       -- guest, admin, system, pms
    actor_id VARCHAR(255),
    action VARCHAR(255) NOT NULL,
    resource_type VARCHAR(255),
    resource_id VARCHAR(255),
    details JSONB,
    ip_address INET,
    correlation_id VARCHAR(255),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Integration health
CREATE TABLE integration_health (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    property_id UUID NOT NULL REFERENCES properties(id),
    integration_type VARCHAR(50) NOT NULL, -- pms, packetfence, payment, email, sms
    status VARCHAR(50) NOT NULL,           -- healthy, degraded, down
    latency_ms INT,
    error_count INT NOT NULL DEFAULT 0,
    last_success_at TIMESTAMPTZ,
    last_failure_at TIMESTAMPTZ,
    last_error_message TEXT,
    is_authenticated BOOLEAN NOT NULL DEFAULT false,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Webhook events
CREATE TABLE webhook_events (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    property_id UUID REFERENCES properties(id),
    event_type VARCHAR(255) NOT NULL,
    payload JSONB NOT NULL,
    signature VARCHAR(512),
    status VARCHAR(50) NOT NULL DEFAULT 'pending',
    -- pending, delivered, failed, retrying
    destination_url TEXT,
    attempts INT NOT NULL DEFAULT 0,
    last_response_code INT,
    last_response_body TEXT,
    next_retry_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
```

## 9. API Specification

### 9.1 Versioned REST API Prefix: `/api/v1`

| Endpoint | Methods | Purpose |
|----------|---------|---------|
| `/api/v1/properties` | GET, POST | List/create properties |
| `/api/v1/properties/:id` | GET, PATCH, DELETE | Property CRUD |
| `/api/v1/properties/:id/health` | GET | Property integration health |
| `/api/v1/pms/configs` | GET, POST | PMS configurations |
| `/api/v1/pms/configs/:id` | GET, PATCH, DELETE | PMS config CRUD |
| `/api/v1/pms/configs/:id/test` | POST | Test PMS connection |
| `/api/v1/pms/configs/:id/sync` | POST | Trigger PMS data sync |
| `/api/v1/guest-auth/room` | POST | Room + surname auth |
| `/api/v1/guest-auth/email` | POST | Email verification auth |
| `/api/v1/guest-auth/sms` | POST | SMS OTP auth |
| `/api/v1/guest-auth/voucher` | POST | Voucher code auth |
| `/api/v1/guest-auth/event` | POST | Event access code auth |
| `/api/v1/guest-sessions` | GET | List active sessions |
| `/api/v1/guest-sessions/:id` | GET | Session details |
| `/api/v1/guest-sessions/:id/revoke` | POST | Revoke session |
| `/api/v1/guest-sessions/:id/extend` | POST | Extend session |
| `/api/v1/guest-profiles` | GET, POST | Guest profiles |
| `/api/v1/guest-profiles/:id` | GET, PATCH | Profile CRUD |
| `/api/v1/guest-profiles/:id/consents` | GET | Guest consent history |
| `/api/v1/guest-profiles/:id/sessions` | GET | Guest session history |
| `/api/v1/guest-profiles/:id/export` | GET | GDPR data export |
| `/api/v1/guest-profiles/:id/anonymize` | POST | GDPR anonymization |
| `/api/v1/wifi-plans` | GET, POST | Wi-Fi plan CRUD |
| `/api/v1/wifi-plans/:id` | GET, PATCH, DELETE | Plan management |
| `/api/v1/entitlements` | GET | Guest entitlements |
| `/api/v1/entitlements/:id/upgrade` | POST | Plan upgrade |
| `/api/v1/payments` | POST | Create payment |
| `/api/v1/payments/:id` | GET | Payment status |
| `/api/v1/payments/:id/refund` | POST | Refund payment |
| `/api/v1/folio-charges` | POST | Charge to room folio |
| `/api/v1/folio-charges/:id` | GET | Folio charge status |
| `/api/v1/events` | GET, POST | Event CRUD |
| `/api/v1/events/:id` | GET, PATCH, DELETE | Event management |
| `/api/v1/events/:id/attendees` | GET, POST | Event attendees |
| `/api/v1/events/:id/attendees/import` | POST | Import attendees |
| `/api/v1/events/:id/qr` | GET | Generate QR codes |
| `/api/v1/events/:id/vouchers` | GET, POST | Event vouchers |
| `/api/v1/promotions` | GET, POST | Promotion CRUD |
| `/api/v1/promotions/:id` | GET, PATCH, DELETE | Promotion management |
| `/api/v1/promotions/:id/stats` | GET | Promotion analytics |
| `/api/v1/review-campaigns` | GET, POST | Review campaign CRUD |
| `/api/v1/review-campaigns/:id` | GET, PATCH, DELETE | Campaign management |
| `/api/v1/review-requests` | GET | Review requests |
| `/api/v1/review-requests/:id` | GET, PATCH | Review details/response |
| `/api/v1/consents` | POST | Record consent |
| `/api/v1/consents/:id/withdraw` | POST | Withdraw consent |
| `/api/v1/consents/policies` | GET | Policy versions |
| `/api/v1/reports/:type` | GET | Analytics reports |
| `/api/v1/reports/:type/export` | GET | CSV/PDF export |
| `/api/v1/webhooks` | GET, POST | Webhook configuration |
| `/api/v1/webhooks/:id` | GET, PATCH, DELETE | Webhook management |
| `/api/v1/audit-events` | GET | Audit log query |
| `/api/v1/health` | GET | Service health |
| `/api/v1/health/integrations` | GET | Integration health status |

### 9.2 Common API Conventions

- Authentication: Bearer JWT or API key
- Authorization: Tenant-scoped RBAC
- Pagination: `?page=1&per_page=50`
- Sorting: `?sort=created_at&order=desc`
- Filtering: `?filter[status]=active`
- Idempotency: `Idempotency-Key` header on POST
- Correlation ID: `X-Correlation-ID` header
- Errors: `{ "error": { "code": "VALIDATION_ERROR", "message": "...", "details": {} } }`
- Rate limiting: `X-RateLimit-*` headers
- Versioning: Accept header or URL prefix

## 10. PacketFence Gateway Adapter Specification

```
interface PacketFenceGateway {
  // Guest identity
  createGuestIdentity(propertyId, mac, pid, role, expiresAt): Promise<{ pfNodeMac, pfPid }>
  updateGuestIdentity(mac, updates): Promise<void>

  // Access control
  assignRole(mac, roleId): Promise<void>
  grantAccess(mac): Promise<void>
  revokeAccess(mac): Promise<void>
  setExpiration(mac, expiresAt): Promise<void>

  // Device management
  registerDevice(mac, pid, role): Promise<void>
  listGuestDevices(pid): Promise<Device[]>

  // Session management
  disconnectSession(mac): Promise<void>
  getSessionStatus(mac): Promise<SessionStatus>

  // Bandwidth enforcement
  applyBandwidthPolicy(mac, downloadMbps, uploadMbps): Promise<void>

  // Network health
  getNetworkHealth(propertyId): Promise<NetworkHealth>

  // CoA
  sendCoA(mac, attributes): Promise<void>
}
```

### Implementation Approaches

| Approach | Pros | Cons | Recommended For |
|----------|------|------|-----------------|
| PF REST API | Official, upgrade-safe | Rate limits, latency | Default — all operations |
| RADIUS attributes | Real-time, low latency | Limited data | Role/VLAN assignment |
| External auth source | Flexible auth flows | More complex | Guest authentication |
| Database integration | Fast, direct | Not upgrade-safe | Never — explicitly forbidden |
| Custom portal module | Full UI control | Perl required | Captive portal flows |

### Primary: PF REST API Adapter

Use PacketFence's existing REST API endpoints:

- `POST /api/v1/nodes` → create device
- `PATCH /api/v1/nodes/:mac` → update role, status, expiration
- `POST /api/v1/nodes/:mac/register` → register device
- `POST /api/v1/nodes/:mac/deregister` → revoke access
- `POST /api/v1/nodes/:mac/reevaluate_access` → trigger re-auth
- `POST /api/v1/nodes` with `bulk_register` → batch operations
- `GET /api/v1/nodes/:mac` → device status
- `GET /api/v1/locationlogs` → session location
- `POST /api/v1/authentication/admin_authentication` → admin auth for API calls

### Secondary: RADIUS Attributes for Real-Time Enforcement

Set RADIUS reply attributes via PacketFence's role system:
- `Filter-Id` → ACL/bandwidth policy
- `PacketFence-Role` → role assignment
- `Session-Timeout` → session expiry
- `Idle-Timeout` → idle timeout
- `MikroTik-Rate-Limit` → bandwidth shaping (MikroTik)
- `WISPr-Bandwidth-Max-Down` / `WISPr-Bandwidth-Max-Up` → bandwidth limits

## 11. Security Threat Model

| Threat | Risk | Mitigation |
|--------|------|------------|
| Room number enumeration | High | Generic error messages, rate limiting, lockout |
| Surname brute force | High | Rate limit per room, per IP, per device; lockout after N attempts |
| PMS credential theft | Critical | Encrypted storage, secret manager, audit logging, never in logs |
| Cross-tenant data access | Critical | Row-level security, tenant ID on every query, automated isolation tests |
| Webhook forgery | High | HMAC signatures, IP allowlisting, replay detection |
| Session hijacking | High | TLS everywhere, secure cookies, short-lived tokens, MAC binding |
| MAC spoofing | Medium | Additional auth factors, rate limiting, anomaly detection |
| Voucher guessing | High | Cryptographically random codes, rate limiting, usage limits |
| OTP abuse | High | Rate limiting, TTL, one-time use, lockout |
| Consent bypass | Medium | Server-side enforcement, audit trail |
| Captive portal bypass | Medium | Firewall rules, DNS enforcement, client isolation |
| Log injection | Medium | Input validation, structured logging, output encoding |
| XSS in portal | Medium | CSP headers, output encoding, no eval |
| SSRF in PMS connector | Medium | URL allowlisting, private IP blocking, timeout |
| Insecure direct object reference | High | Tenant-scoped authorization on every resource |
| Sensitive data in logs | High | Structured logging, PII masking, audit log separation |
| RADIUS replay | Low | FreeRADIUS built-in protections |
| Supply chain | Medium | Dependency scanning, container scanning, lock files |

## 12. PMS Connector Specification

### Common Interface

```go
type PMSConnector interface {
    ValidateGuestStay(ctx context.Context, propertyID, roomNumber, lastName string) (*Reservation, error)
    FindReservation(ctx context.Context, propertyID, reservationID string) (*Reservation, error)
    FindInHouseGuest(ctx context.Context, propertyID, roomNumber string) (*Guest, error)
    GetGuestProfile(ctx context.Context, guestID string) (*Guest, error)
    GetRoomStatus(ctx context.Context, propertyID, roomNumber string) (string, error)
    GetCheckInDate(ctx context.Context, reservationID string) (time.Time, error)
    GetCheckOutDate(ctx context.Context, reservationID string) (time.Time, error)
    GetReservationStatus(ctx context.Context, reservationID string) (string, error)
    PostChargeToFolio(ctx context.Context, reservationID string, amountCents int, description string) (*FolioCharge, error)
    UpdateGuestContactDetails(ctx context.Context, guestID string, contact *Contact) error
    SubscribeToReservationUpdates(ctx context.Context, propertyID string) (<-chan ReservationUpdate, error)
    HandleWebhook(ctx context.Context, payload []byte, headers http.Header) (*WebhookEvent, error)
    HealthCheck(ctx context.Context) (*HealthStatus, error)
}
```

### Connector Targets (Priority Order)

| PMS | API Type | Auth | Documentation | Recommendation |
|-----|----------|------|---------------|----------------|
| **Oracle OPERA** | SOAP/XML | Token | Proprietary | **First** — most common in hotels |
| **Oracle OPERA Cloud** | REST/JSON | OAuth 2.0 | Oracle Hospitality API docs | **First** — growing adoption |
| **Cloudbeds** | REST/JSON | API Key | Public docs available | Second |
| **Mews** | REST/JSON | OAuth 2.0 | Public docs available | Second |

**Recommendation**: Start with Oracle OPERA Cloud (REST API, OAuth 2.0, good documentation, cloud-based). Build Mock PMS first for development.

## 13. Implementation Backlog

### Milestone 1: Foundation (Weeks 1-3)
- [x] Repository analysis (complete)
- [x] Architecture document (complete)
- [ ] Local development environment (Docker Compose)
- [ ] Hospitality service skeleton (Go)
- [ ] PostgreSQL schema (migrations)
- [ ] Authentication & RBAC (admin API keys)
- [ ] Property model CRUD
- [ ] PacketFence gateway adapter (REST API client)
- [ ] Mock PacketFence adapter
- [ ] Test framework setup (Go tests + integration tests)

### Milestone 2: Core Hotel Login (Weeks 3-5)
- [ ] Mock PMS connector
- [ ] PMS connector interface + factory
- [ ] Room + surname authentication flow
- [ ] Guest identity creation in PacketFence
- [ ] Device registration through PF API
- [ ] Role assignment (PacketFence node_category)
- [ ] Stay-based expiration (unregdate)
- [ ] Audit logging
- [ ] Rate limiting (room, IP, device)
- [ ] Admin session management UI

### Milestone 3: Wi-Fi Policies (Weeks 5-7)
- [ ] Wi-Fi plan CRUD
- [ ] Device limit enforcement
- [ ] Bandwidth limit via RADIUS attributes
- [ ] VLAN/role mapping per plan
- [ ] Plan upgrade flow
- [ ] Session management API
- [ ] Admin plan UI

### Milestone 4: Real PMS Connector (Weeks 7-9)
- [ ] Oracle OPERA Cloud connector
- [ ] API authentication + token refresh
- [ ] Rate limiting + retries
- [ ] Reservation lookup
- [ ] Guest validation
- [ ] Folio charging
- [ ] Webhook handling
- [ ] Health monitoring

### Milestone 5: Marketing & Engagement (Weeks 9-11)
- [ ] Consent management
- [ ] Promotions engine
- [ ] Review automation
- [ ] Email/SMS delivery
- [ ] Guest landing page
- [ ] Reporting

### Milestone 6: Events & Multi-Property (Weeks 11-13)
- [ ] Conference event management
- [ ] Attendee vouchers + QR codes
- [ ] Event SSID + VLAN mapping
- [ ] Multi-property isolation
- [ ] Admin RBAC
- [ ] Tenant isolation tests

### Milestone 7: Production Readiness (Weeks 13-15)
- [ ] Full threat model review
- [ ] Load tests
- [ ] Security audit
- [ ] Monitoring + dashboards
- [ ] Backup/restore procedures
- [ ] Deployment automation
- [ ] Documentation (admin, front desk, onboarding)

## 14. Risks and Technical Constraints

| Risk | Impact | Likelihood | Mitigation |
|------|--------|------------|------------|
| PacketFence REST API rate limits | Medium | Medium | Implement client-side rate limiting, queue updates |
| PMS API instability | High | Medium | Circuit breakers, fallback policies, graceful degradation |
| FreeRADIUS config conflicts | Medium | Low | Use separate virtual servers, custom dictionaries |
| Vue.js pfappserver extensibility | Medium | Low | Can extend with new route modules, no core changes needed |
| Perl learning curve | Medium | Low | Go for new service, minimize Perl modifications |
| Guest data privacy violations | Critical | Low | Data minimisation, encryption, audit, consent framework |
| PMS field mapping variance | Medium | High | Flexible schema, per-connector field mapping, mock mode |

## 15. Vendor Compatibility Matrix (Network)

| Vendor | RADIUS | Accounting | CoA | Dynamic VLAN | QoS | Captive Portal | Client Isolation |
|--------|--------|------------|-----|--------------|-----|----------------|------------------|
| Cisco Meraki | ✅ Tested | ✅ Tested | ✅ Tested | ✅ Tested | ✅ Tested | ✅ Tested | ✅ |
| Ubiquiti UniFi | ✅ Tested | ✅ Tested | ✅ Tested | ✅ Tested | ❌ | ✅ Tested | ✅ |
| Ruckus | ✅ Tested | ✅ Tested | ✅ Tested | ✅ Tested | ✅ Tested | ✅ Tested | ✅ |
| Aruba | ✅ Tested | ✅ Tested | ✅ Tested | ✅ Tested | ✅ Tested | ✅ Tested | ✅ |
| TP-Link Omada | ❌ Not tested | ❌ Not tested | ❌ Not tested | ❌ Not tested | ❌ | ❌ | ❌ |
| MikroTik | ✅ Tested | ✅ Tested | ✅ Tested | ✅ Tested | ✅ Tested | ✅ Tested | ✅ |
| Cambium | ✅ Tested | ❌ Not tested | ❌ Not tested | ❌ Not tested | ❌ | ❌ | ❌ |
| HP/Aruba ProCurve | ✅ Tested | ✅ Tested | ✅ Tested | ✅ Tested | ✅ Tested | ✅ Tested | ✅ |

NOTE: Verify PacketFence version for current support status. Modules change between versions.

## 16. Document Deliverables Checklist

- [x] #1 Repository assessment (this document)
- [ ] #2 Architecture diagram (Mermaid in this doc)
- [ ] #3 Data-flow diagrams (design phase)
- [x] #4 Threat model (section 11)
- [x] #5 Database schema (section 8)
- [x] #6 PMS connector specification (section 12)
- [x] #7 PacketFence adapter specification (section 10)
- [x] #8 API specification (section 9)
- [ ] #9 Working implementation (per milestone)
- [ ] #10-24 Various documentation (per milestone)
- [x] #25 Vendor compatibility matrix (section 15)
