package models

import (
	"time"
)

type Organisation struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type HotelGroup struct {
	ID             string    `json:"id"`
	OrganisationID string    `json:"organisation_id"`
	Name           string    `json:"name"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

type Property struct {
	ID                   string    `json:"id"`
	HotelGroupID         *string   `json:"hotel_group_id"`
	OrganisationID       string    `json:"organisation_id"`
	Name                 string    `json:"name"`
	Code                 string    `json:"code"`
	Timezone             string    `json:"timezone"`
	Country              string    `json:"country"`
	Language             string    `json:"language"`
	CheckInTime          string    `json:"check_in_time"`
	CheckOutTime         string    `json:"check_out_time"`
	GracePeriodMinutes   int       `json:"grace_period_minutes"`
	AccessFallbackPolicy string    `json:"access_fallback_policy"`
	ConsentPolicyVersion string    `json:"consent_policy_version"`
	CreatedAt            time.Time `json:"created_at"`
	UpdatedAt            time.Time `json:"updated_at"`
}

type PMSConfiguration struct {
	ID                 string     `json:"id"`
	PropertyID         string     `json:"property_id"`
	PMSType            string     `json:"pms_type"`
	APIEndpoint        string     `json:"api_endpoint"`
	APIKeyEncrypted    []byte     `json:"-"`
	APISecretEncrypted []byte     `json:"-"`
	UsernameEncrypted  []byte     `json:"-"`
	PasswordEncrypted  []byte     `json:"-"`
	TokenEndpoint      string     `json:"token_endpoint"`
	ClientIDEncrypted  []byte     `json:"-"`
	SandboxMode        bool       `json:"sandbox_mode"`
	RateLimitPerSecond int        `json:"rate_limit_per_second"`
	TimeoutSeconds     int        `json:"timeout_seconds"`
	HealthStatus       string     `json:"health_status"`
	LastHealthCheckAt  *time.Time `json:"last_health_check_at"`
	CreatedAt          time.Time  `json:"created_at"`
	UpdatedAt          time.Time  `json:"updated_at"`
}

type GuestProfile struct {
	ID                   string     `json:"id"`
	PropertyID           string     `json:"property_id"`
	PMSGuestID           string     `json:"pms_guest_id"`
	FirstName            string     `json:"first_name"`
	LastNameHash         []byte     `json:"-"`
	EmailHash            []byte     `json:"-"`
	EmailEncrypted       []byte     `json:"-"`
	MobileHash           []byte     `json:"-"`
	MobileEncrypted      []byte     `json:"-"`
	Country              string     `json:"country"`
	PreferredLanguage    string     `json:"preferred_language"`
	LoyaltyStatus        string     `json:"loyalty_status"`
	MarketingConsent     bool       `json:"marketing_consent"`
	MarketingConsentAt   *time.Time `json:"marketing_consent_at"`
	PrivacyPolicyVersion string     `json:"privacy_policy_version"`
	PrivacyConsentAt     *time.Time `json:"privacy_consent_at"`
	FirstConnectedAt     *time.Time `json:"first_connected_at"`
	LastConnectedAt      *time.Time `json:"last_connected_at"`
	DeviceCount          int        `json:"device_count"`
	VisitCount           int        `json:"visit_count"`
	Source               string     `json:"source"`
	Tags                 []byte     `json:"tags"`
	CreatedAt            time.Time  `json:"created_at"`
	UpdatedAt            time.Time  `json:"updated_at"`
}

type Reservation struct {
	ID                  string     `json:"id"`
	PropertyID          string     `json:"property_id"`
	PMSReservationID    string     `json:"pms_reservation_id"`
	PMSGuestID          string     `json:"pms_guest_id"`
	GuestProfileID      *string    `json:"guest_profile_id"`
	RoomNumber          string     `json:"room_number"`
	GuestNameHash       []byte     `json:"-"`
	FirstName           string     `json:"first_name"`
	LastNameHash        []byte     `json:"-"`
	CheckInAt           time.Time  `json:"check_in_at"`
	CheckOutAt          time.Time  `json:"check_out_at"`
	Status              string     `json:"status"`
	Adults              int        `json:"adults"`
	Children            int        `json:"children"`
	BookingSource       string     `json:"booking_source"`
	BookingConfirmation string     `json:"booking_confirmation"`
	VIPStatus           bool       `json:"vip_status"`
	LoyaltyNumber       string     `json:"loyalty_number"`
	PMSData             []byte     `json:"-"`
	LastSyncedAt        *time.Time `json:"last_synced_at"`
	CreatedAt           time.Time  `json:"created_at"`
	UpdatedAt           time.Time  `json:"updated_at"`
}

type GuestSession struct {
	ID             string     `json:"id"`
	PropertyID     string     `json:"property_id"`
	ReservationID  *string    `json:"reservation_id"`
	GuestProfileID *string    `json:"guest_profile_id"`
	MACAddress     string     `json:"mac_address"`
	PacketFenceMAC string     `json:"packetfence_node_mac"`
	PacketFencePID string     `json:"packetfence_pid"`
	WifiPlanID     *string    `json:"wifi_plan_id"`
	IPAddress      *string    `json:"ip_address"`
	SSID           string     `json:"ssid"`
	Status         string     `json:"status"`
	ExpiresAt      time.Time  `json:"expires_at"`
	StartedAt      time.Time  `json:"started_at"`
	EndedAt        *time.Time `json:"ended_at"`
	AuthMethod     string     `json:"auth_method"`
	DeviceCount    int        `json:"device_count"`
	MaxDevices     int        `json:"max_devices"`
	DataUsageBytes int64      `json:"data_usage_bytes"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
}

type WiFiPlan struct {
	ID                 string    `json:"id"`
	PropertyID         string    `json:"property_id"`
	Name               string    `json:"name"`
	Code               string    `json:"code"`
	Description        string    `json:"description"`
	DownloadSpeedMbps  *int      `json:"download_speed_mbps"`
	UploadSpeedMbps    *int      `json:"upload_speed_mbps"`
	DataAllowanceBytes *int64    `json:"data_allowance_bytes"`
	SessionDurationMin *int      `json:"session_duration_minutes"`
	MaxDevices         int       `json:"max_devices"`
	ConcurrentSessions int       `json:"concurrent_sessions"`
	PacketFenceRole    string    `json:"packetfence_role"`
	VLANID             *int      `json:"vlan_id"`
	BandwidthPolicy    []byte    `json:"bandwidth_policy"`
	PriceCents         int       `json:"price_cents"`
	TaxPercent         float64   `json:"tax_percent"`
	Currency           string    `json:"currency"`
	IsPaid             bool      `json:"is_paid"`
	IsComplimentary    bool      `json:"is_complimentary"`
	IsPublic           bool      `json:"is_public"`
	SortOrder          int       `json:"sort_order"`
	CreatedAt          time.Time `json:"created_at"`
	UpdatedAt          time.Time `json:"updated_at"`
}

type WiFiEntitlement struct {
	ID             string    `json:"id"`
	GuestSessionID *string   `json:"guest_session_id"`
	WifiPlanID     string    `json:"wifi_plan_id"`
	PropertyID     string    `json:"property_id"`
	GuestProfileID *string   `json:"guest_profile_id"`
	Status         string    `json:"status"`
	PriceCents     int       `json:"price_cents"`
	Currency       string    `json:"currency"`
	Source         string    `json:"source"`
	PaymentID      *string   `json:"payment_id"`
	StartsAt       time.Time `json:"starts_at"`
	ExpiresAt      time.Time `json:"expires_at"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

type Payment struct {
	ID                string     `json:"id"`
	PropertyID        string     `json:"property_id"`
	GuestProfileID    *string    `json:"guest_profile_id"`
	GuestSessionID    *string    `json:"guest_session_id"`
	WifiEntitlementID *string    `json:"wifi_entitlement_id"`
	AmountCents       int        `json:"amount_cents"`
	Currency          string     `json:"currency"`
	Status            string     `json:"status"`
	PaymentMethod     string     `json:"payment_method"`
	FolioChargeID     string     `json:"folio_charge_id"`
	PaymentGateway    string     `json:"payment_gateway"`
	PaymentGatewayRef string     `json:"payment_gateway_ref"`
	IdempotencyKey    string     `json:"idempotency_key"`
	ReceiptReference  string     `json:"receipt_reference"`
	FailureReason     string     `json:"failure_reason"`
	RefundedAt        *time.Time `json:"refunded_at"`
	CreatedAt         time.Time  `json:"created_at"`
	UpdatedAt         time.Time  `json:"updated_at"`
}

type Event struct {
	ID                    string    `json:"id"`
	PropertyID            string    `json:"property_id"`
	Name                  string    `json:"name"`
	Description           string    `json:"description"`
	Venue                 string    `json:"venue"`
	Organiser             string    `json:"organiser"`
	OrganiserEmail        string    `json:"organiser_email"`
	StartsAt              time.Time `json:"starts_at"`
	EndsAt                time.Time `json:"ends_at"`
	SSID                  string    `json:"ssid"`
	VLANID                *int      `json:"vlan_id"`
	PacketFenceRole       string    `json:"packetfence_role"`
	MaxAttendees          *int      `json:"max_attendees"`
	MaxDevicesPerAttendee int       `json:"max_devices_per_attendee"`
	SharedBandwidthMbps   *int      `json:"shared_bandwidth_mbps"`
	PerUserBandwidthMbps  *int      `json:"per_user_bandwidth_mbps"`
	AccessCode            string    `json:"access_code"`
	SplashPageTheme       string    `json:"splash_page_theme"`
	BannerURL             string    `json:"banner_url"`
	AcceptableUsePolicy   string    `json:"acceptable_use_policy"`
	Status                string    `json:"status"`
	CreatedAt             time.Time `json:"created_at"`
	UpdatedAt             time.Time `json:"updated_at"`
}

type EventAttendee struct {
	ID           string     `json:"id"`
	EventID      string     `json:"event_id"`
	PropertyID   string     `json:"property_id"`
	EmailHash    []byte     `json:"-"`
	FirstName    string     `json:"first_name"`
	LastNameHash []byte     `json:"-"`
	VoucherCode  string     `json:"voucher_code"`
	QRCode       string     `json:"qr_code"`
	SessionID    *string    `json:"session_id"`
	Status       string     `json:"status"`
	CheckedInAt  *time.Time `json:"checked_in_at"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
}

type Promotion struct {
	ID              string     `json:"id"`
	PropertyID      string     `json:"property_id"`
	Name            string     `json:"name"`
	Type            string     `json:"type"`
	Title           string     `json:"title"`
	Description     string     `json:"description"`
	ImageURL        string     `json:"image_url"`
	CTAText         string     `json:"cta_text"`
	CTAURL          string     `json:"cta_url"`
	TargetSegment   string     `json:"target_segment"`
	TargetLanguage  string     `json:"target_language"`
	TargetCountry   string     `json:"target_country"`
	StartsAt        *time.Time `json:"starts_at"`
	EndsAt          *time.Time `json:"ends_at"`
	DisplayLocation string     `json:"display_location"`
	SortOrder       int        `json:"sort_order"`
	IsActive        bool       `json:"is_active"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
}

type PromotionInteraction struct {
	ID                      string    `json:"id"`
	PromotionID             string    `json:"promotion_id"`
	GuestProfileID          *string   `json:"guest_profile_id"`
	GuestSessionID          *string   `json:"guest_session_id"`
	InteractionType         string    `json:"interaction_type"`
	BookingReference        string    `json:"booking_reference"`
	RevenueAttributionCents int       `json:"revenue_attribution_cents"`
	CreatedAt               time.Time `json:"created_at"`
}

type ReviewCampaign struct {
	ID                    string    `json:"id"`
	PropertyID            string    `json:"property_id"`
	Name                  string    `json:"name"`
	DelayHours            int       `json:"delay_hours"`
	Channel               string    `json:"channel"`
	TemplateID            string    `json:"template_id"`
	GoogleReviewURL       string    `json:"google_review_url"`
	TripAdvisorReviewURL  string    `json:"tripadvisor_review_url"`
	InternalSurveyEnabled bool      `json:"internal_survey_enabled"`
	EscalationThreshold   *int      `json:"escalation_threshold"`
	EscalationEmail       string    `json:"escalation_email"`
	IsActive              bool      `json:"is_active"`
	CreatedAt             time.Time `json:"created_at"`
	UpdatedAt             time.Time `json:"updated_at"`
}

type ReviewRequest struct {
	ID                      string     `json:"id"`
	CampaignID              string     `json:"campaign_id"`
	GuestSessionID          *string    `json:"guest_session_id"`
	GuestProfileID          *string    `json:"guest_profile_id"`
	ReservationID           *string    `json:"reservation_id"`
	PropertyID              string     `json:"property_id"`
	Status                  string     `json:"status"`
	SentAt                  *time.Time `json:"sent_at"`
	OpenedAt                *time.Time `json:"opened_at"`
	Rating                  *int       `json:"rating"`
	FeedbackText            string     `json:"feedback_text"`
	ExternalReviewSubmitted bool       `json:"external_review_submitted"`
	CreatedAt               time.Time  `json:"created_at"`
	UpdatedAt               time.Time  `json:"updated_at"`
}

type Consent struct {
	ID             string    `json:"id"`
	GuestProfileID *string   `json:"guest_profile_id"`
	PropertyID     string    `json:"property_id"`
	ConsentType    string    `json:"consent_type"`
	Granted        bool      `json:"granted"`
	PolicyVersion  string    `json:"policy_version"`
	PolicyText     string    `json:"policy_text"`
	Language       string    `json:"language"`
	Source         string    `json:"source"`
	IPAddress      *string   `json:"ip_address"`
	UserAgent      string    `json:"user_agent"`
	CreatedAt      time.Time `json:"created_at"`
}

type PolicyVersion struct {
	ID          string    `json:"id"`
	PropertyID  string    `json:"property_id"`
	PolicyType  string    `json:"policy_type"`
	Version     string    `json:"version"`
	Content     string    `json:"content"`
	Language    string    `json:"language"`
	PublishedAt time.Time `json:"published_at"`
}

type AuditEvent struct {
	ID            string    `json:"id"`
	PropertyID    *string   `json:"property_id"`
	ActorType     string    `json:"actor_type"`
	ActorID       string    `json:"actor_id"`
	Action        string    `json:"action"`
	ResourceType  string    `json:"resource_type"`
	ResourceID    string    `json:"resource_id"`
	Details       []byte    `json:"details"`
	IPAddress     *string   `json:"ip_address"`
	CorrelationID string    `json:"correlation_id"`
	CreatedAt     time.Time `json:"created_at"`
}

type IntegrationHealth struct {
	ID               string     `json:"id"`
	PropertyID       string     `json:"property_id"`
	IntegrationType  string     `json:"integration_type"`
	Status           string     `json:"status"`
	LatencyMs        int        `json:"latency_ms"`
	ErrorCount       int        `json:"error_count"`
	LastSuccessAt    *time.Time `json:"last_success_at"`
	LastFailureAt    *time.Time `json:"last_failure_at"`
	LastErrorMessage string     `json:"last_error_message"`
	IsAuthenticated  bool       `json:"is_authenticated"`
	CreatedAt        time.Time  `json:"created_at"`
	UpdatedAt        time.Time  `json:"updated_at"`
}

type WebhookEvent struct {
	ID               string     `json:"id"`
	PropertyID       *string    `json:"property_id"`
	EventType        string     `json:"event_type"`
	Payload          []byte     `json:"payload"`
	Signature        string     `json:"signature"`
	Status           string     `json:"status"`
	DestinationURL   string     `json:"destination_url"`
	Attempts         int        `json:"attempts"`
	LastResponseCode int        `json:"last_response_code"`
	LastResponseBody string     `json:"last_response_body"`
	NextRetryAt      *time.Time `json:"next_retry_at"`
	CreatedAt        time.Time  `json:"created_at"`
	UpdatedAt        time.Time  `json:"updated_at"`
}
