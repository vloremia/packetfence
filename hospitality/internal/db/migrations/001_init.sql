-- 001_init.sql
-- Hospitality platform initial schema
-- All tables are tenant-aware via organisation_id / property_id columns.

CREATE EXTENSION IF NOT EXISTS pgcrypto;

-- ============ Tenant hierarchy ============

CREATE TABLE organisations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE hotel_groups (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organisation_id UUID NOT NULL REFERENCES organisations(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX hotel_groups_organisation_idx ON hotel_groups (organisation_id);

CREATE TABLE properties (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    hotel_group_id UUID REFERENCES hotel_groups(id) ON DELETE SET NULL,
    organisation_id UUID NOT NULL REFERENCES organisations(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    code VARCHAR(50) NOT NULL,
    timezone VARCHAR(50) NOT NULL DEFAULT 'UTC',
    country VARCHAR(2) NOT NULL DEFAULT 'US',
    language VARCHAR(10) NOT NULL DEFAULT 'en',
    check_in_time TIME NOT NULL DEFAULT '14:00',
    check_out_time TIME NOT NULL DEFAULT '11:00',
    grace_period_minutes INT NOT NULL DEFAULT 120,
    access_fallback_policy VARCHAR(50) NOT NULL DEFAULT 'deny',
    consent_policy_version VARCHAR(50),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (organisation_id, code)
);
CREATE INDEX properties_organisation_idx ON properties (organisation_id);
CREATE INDEX properties_group_idx ON properties (hotel_group_id);

-- ============ PMS configuration (secrets encrypted at app level) ============

CREATE TABLE pms_configurations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    property_id UUID NOT NULL REFERENCES properties(id) ON DELETE CASCADE,
    pms_type VARCHAR(50) NOT NULL,
    api_endpoint VARCHAR(512) NOT NULL,
    api_key_encrypted BYTEA,
    api_secret_encrypted BYTEA,
    username_encrypted BYTEA,
    password_encrypted BYTEA,
    token_endpoint VARCHAR(512),
    client_id_encrypted BYTEA,
    sandbox_mode BOOLEAN NOT NULL DEFAULT FALSE,
    rate_limit_per_second INT NOT NULL DEFAULT 10,
    timeout_seconds INT NOT NULL DEFAULT 30,
    health_status VARCHAR(50) NOT NULL DEFAULT 'unknown',
    last_health_check_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (property_id, pms_type)
);
CREATE INDEX pms_configurations_property_idx ON pms_configurations (property_id);

-- ============ Guests, reservations ============

CREATE TABLE guest_profiles (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    property_id UUID NOT NULL REFERENCES properties(id) ON DELETE CASCADE,
    pms_guest_id VARCHAR(255),
    first_name VARCHAR(255),
    last_name_hash BYTEA,
    email_hash BYTEA,
    email_encrypted BYTEA,
    mobile_hash BYTEA,
    mobile_encrypted BYTEA,
    country VARCHAR(2),
    preferred_language VARCHAR(10),
    loyalty_status VARCHAR(50),
    marketing_consent BOOLEAN NOT NULL DEFAULT FALSE,
    marketing_consent_at TIMESTAMPTZ,
    privacy_policy_version VARCHAR(50),
    privacy_consent_at TIMESTAMPTZ,
    first_connected_at TIMESTAMPTZ,
    last_connected_at TIMESTAMPTZ,
    device_count INT NOT NULL DEFAULT 0,
    visit_count INT NOT NULL DEFAULT 0,
    source VARCHAR(50),
    tags JSONB,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX guest_profiles_property_idx ON guest_profiles (property_id);
CREATE INDEX guest_profiles_email_hash_idx ON guest_profiles (email_hash);
CREATE INDEX guest_profiles_last_name_hash_idx ON guest_profiles (last_name_hash);
CREATE INDEX guest_profiles_pms_guest_idx ON guest_profiles (property_id, pms_guest_id);

CREATE TABLE reservations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    property_id UUID NOT NULL REFERENCES properties(id) ON DELETE CASCADE,
    pms_reservation_id VARCHAR(255) NOT NULL,
    pms_guest_id VARCHAR(255),
    guest_profile_id UUID REFERENCES guest_profiles(id) ON DELETE SET NULL,
    room_number VARCHAR(50) NOT NULL,
    guest_name_hash BYTEA NOT NULL,
    first_name VARCHAR(255),
    last_name_hash BYTEA NOT NULL,
    check_in_at TIMESTAMPTZ NOT NULL,
    check_out_at TIMESTAMPTZ NOT NULL,
    status VARCHAR(50) NOT NULL,
    adults INT NOT NULL DEFAULT 1,
    children INT NOT NULL DEFAULT 0,
    booking_source VARCHAR(255),
    booking_confirmation VARCHAR(255),
    vip_status BOOLEAN NOT NULL DEFAULT FALSE,
    loyalty_number VARCHAR(255),
    pms_data JSONB,
    last_synced_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (property_id, pms_reservation_id)
);
CREATE INDEX reservations_property_status_idx ON reservations (property_id, status);
CREATE INDEX reservations_room_idx ON reservations (property_id, room_number, status);
CREATE INDEX reservations_name_hash_idx ON reservations (property_id, last_name_hash, status);

-- ============ Wi-Fi plans, sessions, entitlements ============

CREATE TABLE wifi_plans (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    property_id UUID NOT NULL REFERENCES properties(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    code VARCHAR(50) NOT NULL,
    description TEXT,
    download_speed_mbps INT,
    upload_speed_mbps INT,
    data_allowance_bytes BIGINT,
    session_duration_minutes INT,
    max_devices INT NOT NULL DEFAULT 2,
    concurrent_sessions INT NOT NULL DEFAULT 1,
    packetfence_role VARCHAR(255),
    vlan_id INT,
    bandwidth_policy JSONB,
    price_cents INT NOT NULL DEFAULT 0,
    tax_percent NUMERIC(5,2) NOT NULL DEFAULT 0,
    currency VARCHAR(3) NOT NULL DEFAULT 'USD',
    is_paid BOOLEAN NOT NULL DEFAULT FALSE,
    is_complimentary BOOLEAN NOT NULL DEFAULT FALSE,
    is_public BOOLEAN NOT NULL DEFAULT TRUE,
    sort_order INT NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (property_id, code)
);
CREATE INDEX wifi_plans_property_idx ON wifi_plans (property_id);

CREATE TABLE guest_sessions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    property_id UUID NOT NULL REFERENCES properties(id) ON DELETE CASCADE,
    reservation_id UUID REFERENCES reservations(id) ON DELETE SET NULL,
    guest_profile_id UUID REFERENCES guest_profiles(id) ON DELETE SET NULL,
    mac_address VARCHAR(17) NOT NULL,
    packetfence_node_mac VARCHAR(17),
    packetfence_pid VARCHAR(255),
    wifi_plan_id UUID REFERENCES wifi_plans(id) ON DELETE SET NULL,
    ip_address INET,
    ssid VARCHAR(255),
    status VARCHAR(50) NOT NULL DEFAULT 'active',
    expires_at TIMESTAMPTZ NOT NULL,
    started_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    ended_at TIMESTAMPTZ,
    auth_method VARCHAR(50) NOT NULL,
    device_count INT NOT NULL DEFAULT 1,
    max_devices INT NOT NULL DEFAULT 2,
    data_usage_bytes BIGINT NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX guest_sessions_property_status_idx ON guest_sessions (property_id, status);
CREATE INDEX guest_sessions_mac_idx ON guest_sessions (mac_address);
CREATE INDEX guest_sessions_expires_idx ON guest_sessions (expires_at);

CREATE TABLE wifi_entitlements (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    guest_session_id UUID REFERENCES guest_sessions(id) ON DELETE CASCADE,
    wifi_plan_id UUID NOT NULL REFERENCES wifi_plans(id),
    property_id UUID NOT NULL REFERENCES properties(id) ON DELETE CASCADE,
    guest_profile_id UUID REFERENCES guest_profiles(id) ON DELETE SET NULL,
    status VARCHAR(50) NOT NULL DEFAULT 'active',
    price_cents INT NOT NULL DEFAULT 0,
    currency VARCHAR(3) NOT NULL DEFAULT 'USD',
    source VARCHAR(50) NOT NULL,
    payment_id UUID,
    starts_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    expires_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX wifi_entitlements_property_idx ON wifi_entitlements (property_id);

CREATE TABLE payments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    property_id UUID NOT NULL REFERENCES properties(id) ON DELETE CASCADE,
    guest_profile_id UUID REFERENCES guest_profiles(id) ON DELETE SET NULL,
    guest_session_id UUID REFERENCES guest_sessions(id) ON DELETE SET NULL,
    wifi_entitlement_id UUID REFERENCES wifi_entitlements(id) ON DELETE SET NULL,
    amount_cents INT NOT NULL,
    currency VARCHAR(3) NOT NULL DEFAULT 'USD',
    status VARCHAR(50) NOT NULL DEFAULT 'pending',
    payment_method VARCHAR(50),
    folio_charge_id VARCHAR(255),
    payment_gateway VARCHAR(50),
    payment_gateway_ref VARCHAR(255),
    idempotency_key VARCHAR(255) UNIQUE,
    receipt_reference VARCHAR(255),
    failure_reason TEXT,
    refunded_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX payments_property_idx ON payments (property_id);

-- ============ Events ============

CREATE TABLE events (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    property_id UUID NOT NULL REFERENCES properties(id) ON DELETE CASCADE,
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
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX events_property_idx ON events (property_id);

CREATE TABLE event_attendees (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    event_id UUID NOT NULL REFERENCES events(id) ON DELETE CASCADE,
    property_id UUID NOT NULL REFERENCES properties(id) ON DELETE CASCADE,
    email_hash BYTEA,
    first_name VARCHAR(255),
    last_name_hash BYTEA,
    voucher_code VARCHAR(50),
    qr_code TEXT,
    session_id UUID,
    status VARCHAR(50) NOT NULL DEFAULT 'registered',
    checked_in_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX event_attendees_event_idx ON event_attendees (event_id);

-- ============ Marketing ============

CREATE TABLE promotions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    property_id UUID NOT NULL REFERENCES properties(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    type VARCHAR(50) NOT NULL,
    title VARCHAR(255),
    description TEXT,
    image_url TEXT,
    cta_text VARCHAR(255),
    cta_url TEXT,
    target_segment VARCHAR(50),
    target_language VARCHAR(10),
    target_country VARCHAR(2),
    starts_at TIMESTAMPTZ,
    ends_at TIMESTAMPTZ,
    display_location VARCHAR(50),
    sort_order INT NOT NULL DEFAULT 0,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX promotions_property_idx ON promotions (property_id);

CREATE TABLE promotion_interactions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    promotion_id UUID NOT NULL REFERENCES promotions(id) ON DELETE CASCADE,
    guest_profile_id UUID REFERENCES guest_profiles(id) ON DELETE SET NULL,
    guest_session_id UUID REFERENCES guest_sessions(id) ON DELETE SET NULL,
    interaction_type VARCHAR(50) NOT NULL,
    booking_reference VARCHAR(255),
    revenue_attribution_cents INT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX promotion_interactions_promotion_idx ON promotion_interactions (promotion_id);

CREATE TABLE review_campaigns (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    property_id UUID NOT NULL REFERENCES properties(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    delay_hours INT NOT NULL DEFAULT 24,
    channel VARCHAR(50) NOT NULL DEFAULT 'email',
    template_id VARCHAR(255),
    google_review_url TEXT,
    tripadvisor_review_url TEXT,
    internal_survey_enabled BOOLEAN NOT NULL DEFAULT FALSE,
    escalation_threshold INT,
    escalation_email VARCHAR(255),
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX review_campaigns_property_idx ON review_campaigns (property_id);

CREATE TABLE review_requests (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    campaign_id UUID NOT NULL REFERENCES review_campaigns(id) ON DELETE CASCADE,
    guest_session_id UUID REFERENCES guest_sessions(id) ON DELETE SET NULL,
    guest_profile_id UUID REFERENCES guest_profiles(id) ON DELETE SET NULL,
    reservation_id UUID REFERENCES reservations(id) ON DELETE SET NULL,
    property_id UUID NOT NULL REFERENCES properties(id) ON DELETE CASCADE,
    status VARCHAR(50) NOT NULL DEFAULT 'pending',
    sent_at TIMESTAMPTZ,
    opened_at TIMESTAMPTZ,
    rating INT,
    feedback_text TEXT,
    external_review_submitted BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX review_requests_property_idx ON review_requests (property_id);

-- ============ Consent & privacy ============

CREATE TABLE policy_versions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    property_id UUID NOT NULL REFERENCES properties(id) ON DELETE CASCADE,
    policy_type VARCHAR(50) NOT NULL,
    version VARCHAR(50) NOT NULL,
    content TEXT NOT NULL,
    language VARCHAR(10) NOT NULL,
    published_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (property_id, policy_type, version, language)
);
CREATE INDEX policy_versions_property_idx ON policy_versions (property_id);

CREATE TABLE consents (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    guest_profile_id UUID REFERENCES guest_profiles(id) ON DELETE SET NULL,
    property_id UUID NOT NULL REFERENCES properties(id) ON DELETE CASCADE,
    consent_type VARCHAR(50) NOT NULL,
    granted BOOLEAN NOT NULL,
    policy_version VARCHAR(50) NOT NULL,
    policy_text TEXT,
    language VARCHAR(10) NOT NULL,
    source VARCHAR(50) NOT NULL,
    ip_address INET,
    user_agent TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX consents_profile_idx ON consents (guest_profile_id);
CREATE INDEX consents_property_idx ON consents (property_id);

-- ============ Ops ============

CREATE TABLE audit_events (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    property_id UUID,
    actor_type VARCHAR(50) NOT NULL,
    actor_id VARCHAR(255),
    action VARCHAR(255) NOT NULL,
    resource_type VARCHAR(255),
    resource_id VARCHAR(255),
    details JSONB,
    ip_address INET,
    correlation_id VARCHAR(255),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX audit_events_property_created_idx ON audit_events (property_id, created_at);

CREATE TABLE integration_health (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    property_id UUID NOT NULL REFERENCES properties(id) ON DELETE CASCADE,
    integration_type VARCHAR(50) NOT NULL,
    status VARCHAR(50) NOT NULL DEFAULT 'unknown',
    latency_ms INT NOT NULL DEFAULT 0,
    error_count INT NOT NULL DEFAULT 0,
    last_success_at TIMESTAMPTZ,
    last_failure_at TIMESTAMPTZ,
    last_error_message TEXT,
    is_authenticated BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (property_id, integration_type)
);
CREATE INDEX integration_health_property_idx ON integration_health (property_id);

CREATE TABLE webhook_events (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    property_id UUID,
    event_type VARCHAR(255) NOT NULL,
    payload JSONB NOT NULL,
    signature VARCHAR(512),
    status VARCHAR(50) NOT NULL DEFAULT 'pending',
    destination_url TEXT,
    attempts INT NOT NULL DEFAULT 0,
    last_response_code INT,
    last_response_body TEXT,
    next_retry_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX webhook_events_status_idx ON webhook_events (status);

-- ============ Seed data ============

-- Default organisation and property so the platform boots in a usable state
INSERT INTO organisations (id, name) VALUES ('00000000-0000-0000-0000-000000000001', 'Default Organisation');
INSERT INTO properties (id, organisation_id, name, code, country, language)
VALUES ('00000000-0000-0000-0000-000000000002', '00000000-0000-0000-0000-000000000001', 'Demo Hotel', 'DEMO', 'US', 'en');

-- Default Wi-Fi plans
INSERT INTO wifi_plans (property_id, name, code, download_speed_mbps, upload_speed_mbps, max_devices, packetfence_role, price_cents, is_paid, is_complimentary, sort_order) VALUES
  ('00000000-0000-0000-0000-000000000002', 'Complimentary Basic', 'basic', 10, 5, 2, 'guest', 0, FALSE, TRUE, 1),
  ('00000000-0000-0000-0000-000000000002', 'Premium', 'premium', 50, 20, 5, 'guest_premium', 990, TRUE, FALSE, 2),
  ('00000000-0000-0000-0000-000000000002', 'VIP', 'vip', 100, 40, 8, 'guest_vip', 0, FALSE, TRUE, 3);

-- Default privacy/terms policies
INSERT INTO policy_versions (property_id, policy_type, version, content, language) VALUES
  ('00000000-0000-0000-0000-000000000002', 'terms', '1.0', 'Terms of use placeholder — review by legal required.', 'en'),
  ('00000000-0000-0000-0000-000000000002', 'privacy', '1.0', 'Privacy notice placeholder — review by legal required.', 'en'),
  ('00000000-0000-0000-0000-000000000002', 'marketing', '1.0', 'Marketing consent notice placeholder — review by legal required.', 'en');
