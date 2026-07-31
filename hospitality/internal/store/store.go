package store

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/inverse-inc/packetfence/hospitality/internal/db"
	"github.com/inverse-inc/packetfence/hospitality/internal/models"
)

type Store struct {
	DB *db.DB
}

func New(database *db.DB) *Store {
	return &Store{DB: database}
}

func newID() string {
	return uuid.NewString()
}

func now() time.Time {
	return time.Now().UTC()
}

// ---- Properties ----

func (s *Store) CreateProperty(ctx context.Context, orgID string, p *models.Property) (*models.Property, error) {
	p.ID = newID()
	p.CreatedAt = now()
	p.UpdatedAt = now()
	if p.OrganisationID == "" {
		p.OrganisationID = orgID
	}
	err := s.DB.Pool.QueryRow(ctx, `
		INSERT INTO properties (id, hotel_group_id, organisation_id, name, code, timezone, country, language,
			check_in_time, check_out_time, grace_period_minutes, access_fallback_policy, consent_policy_version)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13)
		RETURNING created_at, updated_at`,
		p.ID, p.HotelGroupID, p.OrganisationID, p.Name, p.Code, p.Timezone, p.Country, p.Language,
		p.CheckInTime, p.CheckOutTime, p.GracePeriodMinutes, p.AccessFallbackPolicy, p.ConsentPolicyVersion,
	).Scan(&p.CreatedAt, &p.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return p, nil
}

func (s *Store) GetProperty(ctx context.Context, id string) (*models.Property, error) {
	p := &models.Property{}
	err := s.DB.Pool.QueryRow(ctx, `
		SELECT id, hotel_group_id, organisation_id, name, code, timezone, country, language,
			check_in_time, check_out_time, grace_period_minutes, access_fallback_policy, consent_policy_version,
			created_at, updated_at
		FROM properties WHERE id=$1`, id).Scan(
		&p.ID, &p.HotelGroupID, &p.OrganisationID, &p.Name, &p.Code, &p.Timezone, &p.Country, &p.Language,
		&p.CheckInTime, &p.CheckOutTime, &p.GracePeriodMinutes, &p.AccessFallbackPolicy, &p.ConsentPolicyVersion,
		&p.CreatedAt, &p.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return p, nil
}

func (s *Store) ListProperties(ctx context.Context, orgID string) ([]models.Property, error) {
	rows, err := s.DB.Pool.Query(ctx, `
		SELECT id, hotel_group_id, organisation_id, name, code, timezone, country, language,
			check_in_time, check_out_time, grace_period_minutes, access_fallback_policy, consent_policy_version,
			created_at, updated_at
		FROM properties WHERE organisation_id=$1 ORDER BY name`, orgID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	props := []models.Property{}
	for rows.Next() {
		p := models.Property{}
		if err := rows.Scan(
			&p.ID, &p.HotelGroupID, &p.OrganisationID, &p.Name, &p.Code, &p.Timezone, &p.Country, &p.Language,
			&p.CheckInTime, &p.CheckOutTime, &p.GracePeriodMinutes, &p.AccessFallbackPolicy, &p.ConsentPolicyVersion,
			&p.CreatedAt, &p.UpdatedAt); err != nil {
			return nil, err
		}
		props = append(props, p)
	}
	return props, rows.Err()
}

func (s *Store) ListAllProperties(ctx context.Context) ([]models.Property, error) {
	rows, err := s.DB.Pool.Query(ctx, `
		SELECT id, hotel_group_id, organisation_id, name, code, timezone, country, language,
			check_in_time, check_out_time, grace_period_minutes, access_fallback_policy, consent_policy_version,
			created_at, updated_at
		FROM properties ORDER BY id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	properties := []models.Property{}
	for rows.Next() {
		p := models.Property{}
		if err := rows.Scan(&p.ID, &p.HotelGroupID, &p.OrganisationID, &p.Name, &p.Code, &p.Timezone, &p.Country, &p.Language,
			&p.CheckInTime, &p.CheckOutTime, &p.GracePeriodMinutes, &p.AccessFallbackPolicy, &p.ConsentPolicyVersion,
			&p.CreatedAt, &p.UpdatedAt); err != nil {
			return nil, err
		}
		properties = append(properties, p)
	}
	return properties, rows.Err()
}

func (s *Store) UpdateProperty(ctx context.Context, p *models.Property) error {
	p.UpdatedAt = now()
	_, err := s.DB.Pool.Exec(ctx, `
		UPDATE properties SET name=$2, code=$3, timezone=$4, country=$5, language=$6,
			check_in_time=$7, check_out_time=$8, grace_period_minutes=$9, access_fallback_policy=$10,
			consent_policy_version=$11, updated_at=NOW()
		WHERE id=$1`,
		p.ID, p.Name, p.Code, p.Timezone, p.Country, p.Language,
		p.CheckInTime, p.CheckOutTime, p.GracePeriodMinutes, p.AccessFallbackPolicy, p.ConsentPolicyVersion)
	return err
}

func (s *Store) DeleteProperty(ctx context.Context, id string) error {
	_, err := s.DB.Pool.Exec(ctx, `DELETE FROM properties WHERE id=$1`, id)
	return err
}

// ---- Wi-Fi plans ----

func (s *Store) CreatePlan(ctx context.Context, p *models.WiFiPlan) (*models.WiFiPlan, error) {
	p.ID = newID()
	p.CreatedAt = now()
	p.UpdatedAt = now()
	err := s.DB.Pool.QueryRow(ctx, `
		INSERT INTO wifi_plans (id, property_id, name, code, description, download_speed_mbps, upload_speed_mbps,
			data_allowance_bytes, session_duration_minutes, max_devices, concurrent_sessions, packetfence_role,
			vlan_id, bandwidth_policy, price_cents, tax_percent, currency, is_paid, is_complimentary, is_public, sort_order)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18,$19,$20,$21)
		RETURNING created_at, updated_at`,
		p.ID, p.PropertyID, p.Name, p.Code, p.Description, p.DownloadSpeedMbps, p.UploadSpeedMbps,
		p.DataAllowanceBytes, p.SessionDurationMin, p.MaxDevices, p.ConcurrentSessions, p.PacketFenceRole,
		p.VLANID, p.BandwidthPolicy, p.PriceCents, p.TaxPercent, p.Currency, p.IsPaid, p.IsComplimentary, p.IsPublic, p.SortOrder,
	).Scan(&p.CreatedAt, &p.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return p, nil
}

func (s *Store) GetPlan(ctx context.Context, id string) (*models.WiFiPlan, error) {
	p := &models.WiFiPlan{}
	err := s.DB.Pool.QueryRow(ctx, `
		SELECT id, property_id, name, code, description, download_speed_mbps, upload_speed_mbps,
			data_allowance_bytes, session_duration_minutes, max_devices, concurrent_sessions, packetfence_role,
			vlan_id, bandwidth_policy, price_cents, tax_percent, currency, is_paid, is_complimentary, is_public,
			sort_order, created_at, updated_at
		FROM wifi_plans WHERE id=$1`, id).Scan(
		&p.ID, &p.PropertyID, &p.Name, &p.Code, &p.Description, &p.DownloadSpeedMbps, &p.UploadSpeedMbps,
		&p.DataAllowanceBytes, &p.SessionDurationMin, &p.MaxDevices, &p.ConcurrentSessions, &p.PacketFenceRole,
		&p.VLANID, &p.BandwidthPolicy, &p.PriceCents, &p.TaxPercent, &p.Currency, &p.IsPaid, &p.IsComplimentary,
		&p.IsPublic, &p.SortOrder, &p.CreatedAt, &p.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return p, nil
}

func (s *Store) ListPlans(ctx context.Context, propertyID string, publicOnly bool) ([]models.WiFiPlan, error) {
	q := `SELECT id, property_id, name, code, description, download_speed_mbps, upload_speed_mbps,
			data_allowance_bytes, session_duration_minutes, max_devices, concurrent_sessions, packetfence_role,
			vlan_id, bandwidth_policy, price_cents, tax_percent, currency, is_paid, is_complimentary, is_public,
			sort_order, created_at, updated_at
		FROM wifi_plans WHERE property_id=$1`
	args := []any{propertyID}
	if publicOnly {
		q += ` AND is_public=TRUE`
	}
	q += ` ORDER BY sort_order`
	rows, err := s.DB.Pool.Query(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	plans := []models.WiFiPlan{}
	for rows.Next() {
		p := models.WiFiPlan{}
		if err := rows.Scan(
			&p.ID, &p.PropertyID, &p.Name, &p.Code, &p.Description, &p.DownloadSpeedMbps, &p.UploadSpeedMbps,
			&p.DataAllowanceBytes, &p.SessionDurationMin, &p.MaxDevices, &p.ConcurrentSessions, &p.PacketFenceRole,
			&p.VLANID, &p.BandwidthPolicy, &p.PriceCents, &p.TaxPercent, &p.Currency, &p.IsPaid, &p.IsComplimentary,
			&p.IsPublic, &p.SortOrder, &p.CreatedAt, &p.UpdatedAt); err != nil {
			return nil, err
		}
		plans = append(plans, p)
	}
	return plans, rows.Err()
}

func (s *Store) UpdatePlan(ctx context.Context, p *models.WiFiPlan) error {
	_, err := s.DB.Pool.Exec(ctx, `
		UPDATE wifi_plans SET name=$2, description=$3, download_speed_mbps=$4, upload_speed_mbps=$5,
			data_allowance_bytes=$6, session_duration_minutes=$7, max_devices=$8, concurrent_sessions=$9,
			packetfence_role=$10, vlan_id=$11, bandwidth_policy=$12, price_cents=$13, tax_percent=$14,
			currency=$15, is_paid=$16, is_complimentary=$17, is_public=$18, sort_order=$19, updated_at=NOW()
		WHERE id=$1`,
		p.ID, p.Name, p.Description, p.DownloadSpeedMbps, p.UploadSpeedMbps,
		p.DataAllowanceBytes, p.SessionDurationMin, p.MaxDevices, p.ConcurrentSessions,
		p.PacketFenceRole, p.VLANID, p.BandwidthPolicy, p.PriceCents, p.TaxPercent,
		p.Currency, p.IsPaid, p.IsComplimentary, p.IsPublic, p.SortOrder)
	return err
}

func (s *Store) DeletePlan(ctx context.Context, id string) error {
	_, err := s.DB.Pool.Exec(ctx, `DELETE FROM wifi_plans WHERE id=$1`, id)
	return err
}

// ---- Guest profiles ----

func (s *Store) UpsertGuestProfile(ctx context.Context, g *models.GuestProfile) (*models.GuestProfile, error) {
	if g.ID == "" {
		g.ID = newID()
	}
	g.UpdatedAt = now()
	_, err := s.DB.Pool.Exec(ctx, `
		INSERT INTO guest_profiles (id, property_id, pms_guest_id, first_name, last_name_hash, email_hash,
			email_encrypted, mobile_hash, mobile_encrypted, country, preferred_language, loyalty_status,
			marketing_consent, marketing_consent_at, privacy_policy_version, privacy_consent_at, source, tags)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18)
		ON CONFLICT (id) DO UPDATE SET first_name=EXCLUDED.first_name, last_name_hash=EXCLUDED.last_name_hash,
			email_hash=EXCLUDED.email_hash, email_encrypted=EXCLUDED.email_encrypted,
			mobile_hash=EXCLUDED.mobile_hash, mobile_encrypted=EXCLUDED.mobile_encrypted,
			country=EXCLUDED.country, preferred_language=EXCLUDED.preferred_language,
			loyalty_status=EXCLUDED.loyalty_status, marketing_consent=EXCLUDED.marketing_consent,
			marketing_consent_at=EXCLUDED.marketing_consent_at,
			privacy_policy_version=EXCLUDED.privacy_policy_version, privacy_consent_at=EXCLUDED.privacy_consent_at,
			source=EXCLUDED.source, tags=EXCLUDED.tags, updated_at=NOW()`,
		g.ID, g.PropertyID, g.PMSGuestID, g.FirstName, g.LastNameHash, g.EmailHash,
		g.EmailEncrypted, g.MobileHash, g.MobileEncrypted, g.Country, g.PreferredLanguage, g.LoyaltyStatus,
		g.MarketingConsent, g.MarketingConsentAt, g.PrivacyPolicyVersion, g.PrivacyConsentAt, g.Source, g.Tags)
	return g, err
}

func (s *Store) GetGuestProfile(ctx context.Context, id string) (*models.GuestProfile, error) {
	g := &models.GuestProfile{}
	err := s.DB.Pool.QueryRow(ctx, `
		SELECT id, property_id, pms_guest_id, first_name, last_name_hash, email_hash, email_encrypted,
			mobile_hash, mobile_encrypted, country, preferred_language, loyalty_status, marketing_consent,
			marketing_consent_at, privacy_policy_version, privacy_consent_at, first_connected_at,
			last_connected_at, device_count, visit_count, source, tags, created_at, updated_at
		FROM guest_profiles WHERE id=$1`, id).Scan(
		&g.ID, &g.PropertyID, &g.PMSGuestID, &g.FirstName, &g.LastNameHash, &g.EmailHash, &g.EmailEncrypted,
		&g.MobileHash, &g.MobileEncrypted, &g.Country, &g.PreferredLanguage, &g.LoyaltyStatus, &g.MarketingConsent,
		&g.MarketingConsentAt, &g.PrivacyPolicyVersion, &g.PrivacyConsentAt, &g.FirstConnectedAt,
		&g.LastConnectedAt, &g.DeviceCount, &g.VisitCount, &g.Source, &g.Tags, &g.CreatedAt, &g.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return g, nil
}

func (s *Store) FindGuestByLastNameHash(ctx context.Context, propertyID string, hash []byte) ([]models.GuestProfile, error) {
	rows, err := s.DB.Pool.Query(ctx, `
		SELECT id, property_id, pms_guest_id, first_name, last_name_hash, email_hash, email_encrypted,
			mobile_hash, mobile_encrypted, country, preferred_language, loyalty_status, marketing_consent,
			marketing_consent_at, privacy_policy_version, privacy_consent_at, first_connected_at,
			last_connected_at, device_count, visit_count, source, tags, created_at, updated_at
		FROM guest_profiles WHERE property_id=$1 AND last_name_hash=$2`, propertyID, hash)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanGuestProfiles(rows)
}

func scanGuestProfiles(rows interface {
	Next() bool
	Err() error
	Scan(dest ...any) error
}) ([]models.GuestProfile, error) {
	out := []models.GuestProfile{}
	for rows.Next() {
		g := models.GuestProfile{}
		if err := rows.Scan(
			&g.ID, &g.PropertyID, &g.PMSGuestID, &g.FirstName, &g.LastNameHash, &g.EmailHash, &g.EmailEncrypted,
			&g.MobileHash, &g.MobileEncrypted, &g.Country, &g.PreferredLanguage, &g.LoyaltyStatus, &g.MarketingConsent,
			&g.MarketingConsentAt, &g.PrivacyPolicyVersion, &g.PrivacyConsentAt, &g.FirstConnectedAt,
			&g.LastConnectedAt, &g.DeviceCount, &g.VisitCount, &g.Source, &g.Tags, &g.CreatedAt, &g.UpdatedAt); err != nil {
			return nil, err
		}
		out = append(out, g)
	}
	return out, rows.Err()
}

// ---- Reservations ----

func (s *Store) UpsertReservation(ctx context.Context, r *models.Reservation) (*models.Reservation, error) {
	if r.ID == "" {
		r.ID = newID()
	}
	r.UpdatedAt = now()
	_, err := s.DB.Pool.Exec(ctx, `
		INSERT INTO reservations (id, property_id, pms_reservation_id, pms_guest_id, guest_profile_id, room_number,
			guest_name_hash, first_name, last_name_hash, check_in_at, check_out_at, status, adults, children,
			booking_source, booking_confirmation, vip_status, loyalty_number, pms_data, last_synced_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18,$19,$20)
		ON CONFLICT (property_id, pms_reservation_id) DO UPDATE SET
			pms_guest_id=EXCLUDED.pms_guest_id, guest_profile_id=EXCLUDED.guest_profile_id, room_number=EXCLUDED.room_number,
			guest_name_hash=EXCLUDED.guest_name_hash, first_name=EXCLUDED.first_name, last_name_hash=EXCLUDED.last_name_hash,
			check_in_at=EXCLUDED.check_in_at, check_out_at=EXCLUDED.check_out_at, status=EXCLUDED.status,
			adults=EXCLUDED.adults, children=EXCLUDED.children, booking_source=EXCLUDED.booking_source,
			booking_confirmation=EXCLUDED.booking_confirmation, vip_status=EXCLUDED.vip_status,
			loyalty_number=EXCLUDED.loyalty_number, pms_data=EXCLUDED.pms_data, last_synced_at=EXCLUDED.last_synced_at,
			updated_at=NOW()`,
		r.ID, r.PropertyID, r.PMSReservationID, r.PMSGuestID, r.GuestProfileID, r.RoomNumber,
		r.GuestNameHash, r.FirstName, r.LastNameHash, r.CheckInAt, r.CheckOutAt, r.Status, r.Adults, r.Children,
		r.BookingSource, r.BookingConfirmation, r.VIPStatus, r.LoyaltyNumber, r.PMSData, r.LastSyncedAt)
	return r, err
}

func (s *Store) FindReservationByRoomAndName(ctx context.Context, propertyID, roomNumber string, lastNameHash []byte) (*models.Reservation, error) {
	r := &models.Reservation{}
	err := s.DB.Pool.QueryRow(ctx, `
		SELECT id, property_id, pms_reservation_id, pms_guest_id, guest_profile_id, room_number,
			guest_name_hash, first_name, last_name_hash, check_in_at, check_out_at, status, adults, children,
			booking_source, booking_confirmation, vip_status, loyalty_number, pms_data, last_synced_at, created_at, updated_at
		FROM reservations
		WHERE property_id=$1 AND room_number=$2 AND last_name_hash=$3 AND status IN ('checked_in','reserved')
		ORDER BY check_in_at DESC LIMIT 1`,
		propertyID, roomNumber, lastNameHash).Scan(
		&r.ID, &r.PropertyID, &r.PMSReservationID, &r.PMSGuestID, &r.GuestProfileID, &r.RoomNumber,
		&r.GuestNameHash, &r.FirstName, &r.LastNameHash, &r.CheckInAt, &r.CheckOutAt, &r.Status, &r.Adults, &r.Children,
		&r.BookingSource, &r.BookingConfirmation, &r.VIPStatus, &r.LoyaltyNumber, &r.PMSData, &r.LastSyncedAt,
		&r.CreatedAt, &r.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return r, nil
}

func (s *Store) GetReservation(ctx context.Context, id string) (*models.Reservation, error) {
	r := &models.Reservation{}
	err := s.DB.Pool.QueryRow(ctx, `
		SELECT id, property_id, pms_reservation_id, pms_guest_id, guest_profile_id, room_number,
			guest_name_hash, first_name, last_name_hash, check_in_at, check_out_at, status, adults, children,
			booking_source, booking_confirmation, vip_status, loyalty_number, pms_data, last_synced_at, created_at, updated_at
		FROM reservations WHERE id=$1`, id).Scan(
		&r.ID, &r.PropertyID, &r.PMSReservationID, &r.PMSGuestID, &r.GuestProfileID, &r.RoomNumber,
		&r.GuestNameHash, &r.FirstName, &r.LastNameHash, &r.CheckInAt, &r.CheckOutAt, &r.Status, &r.Adults, &r.Children,
		&r.BookingSource, &r.BookingConfirmation, &r.VIPStatus, &r.LoyaltyNumber, &r.PMSData, &r.LastSyncedAt,
		&r.CreatedAt, &r.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return r, nil
}

func (s *Store) ListActiveReservations(ctx context.Context, propertyID string) ([]models.Reservation, error) {
	rows, err := s.DB.Pool.Query(ctx, `
		SELECT id, property_id, pms_reservation_id, pms_guest_id, guest_profile_id, room_number,
			guest_name_hash, first_name, last_name_hash, check_in_at, check_out_at, status, adults, children,
			booking_source, booking_confirmation, vip_status, loyalty_number, pms_data, last_synced_at, created_at, updated_at
		FROM reservations WHERE property_id=$1 AND status='checked_in'`, propertyID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []models.Reservation{}
	for rows.Next() {
		r := models.Reservation{}
		if err := rows.Scan(
			&r.ID, &r.PropertyID, &r.PMSReservationID, &r.PMSGuestID, &r.GuestProfileID, &r.RoomNumber,
			&r.GuestNameHash, &r.FirstName, &r.LastNameHash, &r.CheckInAt, &r.CheckOutAt, &r.Status, &r.Adults, &r.Children,
			&r.BookingSource, &r.BookingConfirmation, &r.VIPStatus, &r.LoyaltyNumber, &r.PMSData, &r.LastSyncedAt,
			&r.CreatedAt, &r.UpdatedAt); err != nil {
			return nil, err
		}
		out = append(out, r)
	}
	return out, rows.Err()
}

// ---- Guest sessions ----

func (s *Store) CreateSession(ctx context.Context, g *models.GuestSession) (*models.GuestSession, error) {
	g.ID = newID()
	g.CreatedAt = now()
	g.UpdatedAt = now()
	err := s.DB.Pool.QueryRow(ctx, `
		INSERT INTO guest_sessions (id, property_id, reservation_id, guest_profile_id, mac_address,
			packetfence_node_mac, packetfence_pid, wifi_plan_id, ip_address, ssid, status, expires_at,
			started_at, auth_method, device_count, max_devices)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16)
		RETURNING created_at, updated_at`,
		g.ID, g.PropertyID, g.ReservationID, g.GuestProfileID, g.MACAddress,
		g.PacketFenceMAC, g.PacketFencePID, g.WifiPlanID, g.IPAddress, g.SSID, g.Status, g.ExpiresAt,
		g.StartedAt, g.AuthMethod, g.DeviceCount, g.MaxDevices).Scan(&g.CreatedAt, &g.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return g, nil
}

func (s *Store) GetSession(ctx context.Context, id string) (*models.GuestSession, error) {
	g := &models.GuestSession{}
	err := s.DB.Pool.QueryRow(ctx, `
		SELECT id, property_id, reservation_id, guest_profile_id, mac_address, packetfence_node_mac,
			packetfence_pid, wifi_plan_id, ip_address, ssid, status, expires_at, started_at, ended_at,
			auth_method, device_count, max_devices, data_usage_bytes, created_at, updated_at
		FROM guest_sessions WHERE id=$1`, id).Scan(
		&g.ID, &g.PropertyID, &g.ReservationID, &g.GuestProfileID, &g.MACAddress, &g.PacketFenceMAC,
		&g.PacketFencePID, &g.WifiPlanID, &g.IPAddress, &g.SSID, &g.Status, &g.ExpiresAt, &g.StartedAt, &g.EndedAt,
		&g.AuthMethod, &g.DeviceCount, &g.MaxDevices, &g.DataUsageBytes, &g.CreatedAt, &g.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return g, nil
}

func (s *Store) ListSessions(ctx context.Context, propertyID, status string) ([]models.GuestSession, error) {
	q := `SELECT id, property_id, reservation_id, guest_profile_id, mac_address, packetfence_node_mac,
			packetfence_pid, wifi_plan_id, ip_address, ssid, status, expires_at, started_at, ended_at,
			auth_method, device_count, max_devices, data_usage_bytes, created_at, updated_at
		FROM guest_sessions WHERE property_id=$1`
	args := []any{propertyID}
	if status != "" {
		q += ` AND status=$2`
		args = append(args, status)
	}
	q += ` ORDER BY started_at DESC`
	rows, err := s.DB.Pool.Query(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []models.GuestSession{}
	for rows.Next() {
		g := models.GuestSession{}
		if err := rows.Scan(
			&g.ID, &g.PropertyID, &g.ReservationID, &g.GuestProfileID, &g.MACAddress, &g.PacketFenceMAC,
			&g.PacketFencePID, &g.WifiPlanID, &g.IPAddress, &g.SSID, &g.Status, &g.ExpiresAt, &g.StartedAt, &g.EndedAt,
			&g.AuthMethod, &g.DeviceCount, &g.MaxDevices, &g.DataUsageBytes, &g.CreatedAt, &g.UpdatedAt); err != nil {
			return nil, err
		}
		out = append(out, g)
	}
	return out, rows.Err()
}

func (s *Store) FindSessionByMAC(ctx context.Context, propertyID, mac string) (*models.GuestSession, error) {
	g := &models.GuestSession{}
	err := s.DB.Pool.QueryRow(ctx, `
		SELECT id, property_id, reservation_id, guest_profile_id, mac_address, packetfence_node_mac,
			packetfence_pid, wifi_plan_id, ip_address, ssid, status, expires_at, started_at, ended_at,
			auth_method, device_count, max_devices, data_usage_bytes, created_at, updated_at
		FROM guest_sessions WHERE property_id=$1 AND mac_address=$2 AND status='active' LIMIT 1`, propertyID, mac).Scan(
		&g.ID, &g.PropertyID, &g.ReservationID, &g.GuestProfileID, &g.MACAddress, &g.PacketFenceMAC,
		&g.PacketFencePID, &g.WifiPlanID, &g.IPAddress, &g.SSID, &g.Status, &g.ExpiresAt, &g.StartedAt, &g.EndedAt,
		&g.AuthMethod, &g.DeviceCount, &g.MaxDevices, &g.DataUsageBytes, &g.CreatedAt, &g.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return g, nil
}

func (s *Store) UpdateSession(ctx context.Context, g *models.GuestSession) error {
	g.UpdatedAt = now()
	_, err := s.DB.Pool.Exec(ctx, `
		UPDATE guest_sessions SET status=$2, expires_at=$3, ended_at=$4, wifi_plan_id=$5, device_count=$6,
			max_devices=$7, updated_at=NOW()
		WHERE id=$1`,
		g.ID, g.Status, g.ExpiresAt, g.EndedAt, g.WifiPlanID, g.DeviceCount, g.MaxDevices)
	return err
}

func (s *Store) ListSessionsExpiringBefore(ctx context.Context, propertyID string, before time.Time) ([]models.GuestSession, error) {
	rows, err := s.DB.Pool.Query(ctx, `
		SELECT id, property_id, reservation_id, guest_profile_id, mac_address, packetfence_node_mac,
			packetfence_pid, wifi_plan_id, ip_address, ssid, status, expires_at, started_at, ended_at,
			auth_method, device_count, max_devices, data_usage_bytes, created_at, updated_at
		FROM guest_sessions WHERE property_id=$1 AND status='active' AND expires_at < $2`, propertyID, before)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []models.GuestSession{}
	for rows.Next() {
		g := models.GuestSession{}
		if err := rows.Scan(
			&g.ID, &g.PropertyID, &g.ReservationID, &g.GuestProfileID, &g.MACAddress, &g.PacketFenceMAC,
			&g.PacketFencePID, &g.WifiPlanID, &g.IPAddress, &g.SSID, &g.Status, &g.ExpiresAt, &g.StartedAt, &g.EndedAt,
			&g.AuthMethod, &g.DeviceCount, &g.MaxDevices, &g.DataUsageBytes, &g.CreatedAt, &g.UpdatedAt); err != nil {
			return nil, err
		}
		out = append(out, g)
	}
	return out, rows.Err()
}

// ---- Entitlements ----

func (s *Store) CreateEntitlement(ctx context.Context, e *models.WiFiEntitlement) (*models.WiFiEntitlement, error) {
	e.ID = newID()
	e.CreatedAt = now()
	e.UpdatedAt = now()
	err := s.DB.Pool.QueryRow(ctx, `
		INSERT INTO wifi_entitlements (id, guest_session_id, wifi_plan_id, property_id, guest_profile_id,
			status, price_cents, currency, source, payment_id, starts_at, expires_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12)
		RETURNING created_at, updated_at`,
		e.ID, e.GuestSessionID, e.WifiPlanID, e.PropertyID, e.GuestProfileID,
		e.Status, e.PriceCents, e.Currency, e.Source, e.PaymentID, e.StartsAt, e.ExpiresAt).Scan(&e.CreatedAt, &e.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return e, nil
}

func (s *Store) ListEntitlementsBySession(ctx context.Context, sessionID string) ([]models.WiFiEntitlement, error) {
	rows, err := s.DB.Pool.Query(ctx, `
		SELECT id, guest_session_id, wifi_plan_id, property_id, guest_profile_id, status, price_cents,
			currency, source, payment_id, starts_at, expires_at, created_at, updated_at
		FROM wifi_entitlements WHERE guest_session_id=$1 ORDER BY starts_at DESC`, sessionID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []models.WiFiEntitlement{}
	for rows.Next() {
		e := models.WiFiEntitlement{}
		if err := rows.Scan(
			&e.ID, &e.GuestSessionID, &e.WifiPlanID, &e.PropertyID, &e.GuestProfileID, &e.Status, &e.PriceCents,
			&e.Currency, &e.Source, &e.PaymentID, &e.StartsAt, &e.ExpiresAt, &e.CreatedAt, &e.UpdatedAt); err != nil {
			return nil, err
		}
		out = append(out, e)
	}
	return out, rows.Err()
}

// ---- Payments ----

func (s *Store) CreatePayment(ctx context.Context, p *models.Payment) (*models.Payment, error) {
	p.ID = newID()
	p.CreatedAt = now()
	p.UpdatedAt = now()
	err := s.DB.Pool.QueryRow(ctx, `
		INSERT INTO payments (id, property_id, guest_profile_id, guest_session_id, wifi_entitlement_id,
			amount_cents, currency, status, payment_method, folio_charge_id, payment_gateway, payment_gateway_ref,
			idempotency_key, receipt_reference, failure_reason)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15)
		RETURNING created_at, updated_at`,
		p.ID, p.PropertyID, p.GuestProfileID, p.GuestSessionID, p.WifiEntitlementID,
		p.AmountCents, p.Currency, p.Status, p.PaymentMethod, p.FolioChargeID, p.PaymentGateway, p.PaymentGatewayRef,
		p.IdempotencyKey, p.ReceiptReference, p.FailureReason).Scan(&p.CreatedAt, &p.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return p, nil
}

func (s *Store) GetPayment(ctx context.Context, id string) (*models.Payment, error) {
	p := &models.Payment{}
	err := s.DB.Pool.QueryRow(ctx, `
		SELECT id, property_id, guest_profile_id, guest_session_id, wifi_entitlement_id,
			amount_cents, currency, status, payment_method, folio_charge_id, payment_gateway, payment_gateway_ref,
			idempotency_key, receipt_reference, failure_reason, refunded_at, created_at, updated_at
		FROM payments WHERE id=$1`, id).Scan(
		&p.ID, &p.PropertyID, &p.GuestProfileID, &p.GuestSessionID, &p.WifiEntitlementID,
		&p.AmountCents, &p.Currency, &p.Status, &p.PaymentMethod, &p.FolioChargeID, &p.PaymentGateway, &p.PaymentGatewayRef,
		&p.IdempotencyKey, &p.ReceiptReference, &p.FailureReason, &p.RefundedAt, &p.CreatedAt, &p.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return p, nil
}

func (s *Store) GetPaymentByIdempotencyKey(ctx context.Context, propertyID, key string) (*models.Payment, error) {
	p := &models.Payment{}
	err := s.DB.Pool.QueryRow(ctx, `
		SELECT id, property_id, guest_profile_id, guest_session_id, wifi_entitlement_id,
			amount_cents, currency, status, payment_method, folio_charge_id, payment_gateway, payment_gateway_ref,
			idempotency_key, receipt_reference, failure_reason, refunded_at, created_at, updated_at
		FROM payments WHERE property_id=$1 AND idempotency_key=$2`, propertyID, key).Scan(
		&p.ID, &p.PropertyID, &p.GuestProfileID, &p.GuestSessionID, &p.WifiEntitlementID,
		&p.AmountCents, &p.Currency, &p.Status, &p.PaymentMethod, &p.FolioChargeID, &p.PaymentGateway, &p.PaymentGatewayRef,
		&p.IdempotencyKey, &p.ReceiptReference, &p.FailureReason, &p.RefundedAt, &p.CreatedAt, &p.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return p, nil
}

func (s *Store) UpdatePayment(ctx context.Context, p *models.Payment) error {
	p.UpdatedAt = now()
	_, err := s.DB.Pool.Exec(ctx, `
		UPDATE payments SET status=$2, folio_charge_id=$3, payment_gateway_ref=$4, receipt_reference=$5,
			failure_reason=$6, refunded_at=$7, updated_at=NOW()
		WHERE id=$1`,
		p.ID, p.Status, p.FolioChargeID, p.PaymentGatewayRef, p.ReceiptReference,
		p.FailureReason, p.RefundedAt)
	return err
}

// ---- Consents ----

func (s *Store) CreateConsent(ctx context.Context, c *models.Consent) (*models.Consent, error) {
	c.ID = newID()
	c.CreatedAt = now()
	err := s.DB.Pool.QueryRow(ctx, `
		INSERT INTO consents (id, guest_profile_id, property_id, consent_type, granted, policy_version,
			policy_text, language, source, ip_address, user_agent)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)
		RETURNING created_at`,
		c.ID, c.GuestProfileID, c.PropertyID, c.ConsentType, c.Granted, c.PolicyVersion,
		c.PolicyText, c.Language, c.Source, c.IPAddress, c.UserAgent).Scan(&c.CreatedAt)
	if err != nil {
		return nil, err
	}
	return c, nil
}

func (s *Store) ListConsentsByProfile(ctx context.Context, profileID string) ([]models.Consent, error) {
	rows, err := s.DB.Pool.Query(ctx, `
		SELECT id, guest_profile_id, property_id, consent_type, granted, policy_version, policy_text,
			language, source, ip_address, user_agent, created_at
		FROM consents WHERE guest_profile_id=$1 ORDER BY created_at DESC`, profileID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []models.Consent{}
	for rows.Next() {
		c := models.Consent{}
		if err := rows.Scan(
			&c.ID, &c.GuestProfileID, &c.PropertyID, &c.ConsentType, &c.Granted, &c.PolicyVersion, &c.PolicyText,
			&c.Language, &c.Source, &c.IPAddress, &c.UserAgent, &c.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, c)
	}
	return out, rows.Err()
}

// ---- Policy versions ----

func (s *Store) GetPolicyVersion(ctx context.Context, propertyID, policyType, language string) (*models.PolicyVersion, error) {
	p := &models.PolicyVersion{}
	err := s.DB.Pool.QueryRow(ctx, `
		SELECT id, property_id, policy_type, version, content, language, published_at
		FROM policy_versions
		WHERE property_id=$1 AND policy_type=$2 AND language=$3
		ORDER BY published_at DESC LIMIT 1`, propertyID, policyType, language).Scan(
		&p.ID, &p.PropertyID, &p.PolicyType, &p.Version, &p.Content, &p.Language, &p.PublishedAt)
	if err != nil {
		return nil, err
	}
	return p, nil
}

// ---- Audit ----

func (s *Store) CreateAuditEvent(ctx context.Context, e *models.AuditEvent) (*models.AuditEvent, error) {
	e.ID = newID()
	e.CreatedAt = now()
	_, err := s.DB.Pool.Exec(ctx, `
		INSERT INTO audit_events (id, property_id, actor_type, actor_id, action, resource_type, resource_id,
			details, ip_address, correlation_id)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)`,
		e.ID, e.PropertyID, e.ActorType, e.ActorID, e.Action, e.ResourceType, e.ResourceID,
		e.Details, e.IPAddress, e.CorrelationID)
	return e, err
}

// ---- Integration health ----

func (s *Store) UpsertIntegrationHealth(ctx context.Context, h *models.IntegrationHealth) (*models.IntegrationHealth, error) {
	h.UpdatedAt = now()
	_, err := s.DB.Pool.Exec(ctx, `
		INSERT INTO integration_health (id, property_id, integration_type, status, latency_ms, error_count,
			last_success_at, last_failure_at, last_error_message, is_authenticated)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)
		ON CONFLICT (property_id, integration_type) DO UPDATE SET
			status=EXCLUDED.status, latency_ms=EXCLUDED.latency_ms, error_count=EXCLUDED.error_count,
			last_success_at=EXCLUDED.last_success_at, last_failure_at=EXCLUDED.last_failure_at,
			last_error_message=EXCLUDED.last_error_message, is_authenticated=EXCLUDED.is_authenticated,
			updated_at=NOW()`,
		newID(), h.PropertyID, h.IntegrationType, h.Status, h.LatencyMs, h.ErrorCount,
		h.LastSuccessAt, h.LastFailureAt, h.LastErrorMessage, h.IsAuthenticated)
	return h, err
}

// ---- Events ----

func (s *Store) CreateEvent(ctx context.Context, e *models.Event) (*models.Event, error) {
	e.ID = newID()
	e.CreatedAt = now()
	e.UpdatedAt = now()
	err := s.DB.Pool.QueryRow(ctx, `
		INSERT INTO events (id, property_id, name, description, venue, organiser, organiser_email, starts_at,
			ends_at, ssid, vlan_id, packetfence_role, max_attendees, max_devices_per_attendee, shared_bandwidth_mbps,
			per_user_bandwidth_mbps, access_code, splash_page_theme, banner_url, acceptable_use_policy, status)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18,$19,$20,$21)
		RETURNING created_at, updated_at`,
		e.ID, e.PropertyID, e.Name, e.Description, e.Venue, e.Organiser, e.OrganiserEmail, e.StartsAt,
		e.EndsAt, e.SSID, e.VLANID, e.PacketFenceRole, e.MaxAttendees, e.MaxDevicesPerAttendee, e.SharedBandwidthMbps,
		e.PerUserBandwidthMbps, e.AccessCode, e.SplashPageTheme, e.BannerURL, e.AcceptableUsePolicy, e.Status).Scan(&e.CreatedAt, &e.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return e, nil
}

func (s *Store) GetEvent(ctx context.Context, id string) (*models.Event, error) {
	e := &models.Event{}
	err := s.DB.Pool.QueryRow(ctx, `
		SELECT id, property_id, name, description, venue, organiser, organiser_email, starts_at,
			ends_at, ssid, vlan_id, packetfence_role, max_attendees, max_devices_per_attendee, shared_bandwidth_mbps,
			per_user_bandwidth_mbps, access_code, splash_page_theme, banner_url, acceptable_use_policy, status,
			created_at, updated_at
		FROM events WHERE id=$1`, id).Scan(
		&e.ID, &e.PropertyID, &e.Name, &e.Description, &e.Venue, &e.Organiser, &e.OrganiserEmail, &e.StartsAt,
		&e.EndsAt, &e.SSID, &e.VLANID, &e.PacketFenceRole, &e.MaxAttendees, &e.MaxDevicesPerAttendee, &e.SharedBandwidthMbps,
		&e.PerUserBandwidthMbps, &e.AccessCode, &e.SplashPageTheme, &e.BannerURL, &e.AcceptableUsePolicy, &e.Status,
		&e.CreatedAt, &e.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return e, nil
}

func (s *Store) ListEvents(ctx context.Context, propertyID string) ([]models.Event, error) {
	rows, err := s.DB.Pool.Query(ctx, `
		SELECT id, property_id, name, description, venue, organiser, organiser_email, starts_at,
			ends_at, ssid, vlan_id, packetfence_role, max_attendees, max_devices_per_attendee, shared_bandwidth_mbps,
			per_user_bandwidth_mbps, access_code, splash_page_theme, banner_url, acceptable_use_policy, status,
			created_at, updated_at
		FROM events WHERE property_id=$1 ORDER BY starts_at DESC`, propertyID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []models.Event{}
	for rows.Next() {
		e := models.Event{}
		if err := rows.Scan(
			&e.ID, &e.PropertyID, &e.Name, &e.Description, &e.Venue, &e.Organiser, &e.OrganiserEmail, &e.StartsAt,
			&e.EndsAt, &e.SSID, &e.VLANID, &e.PacketFenceRole, &e.MaxAttendees, &e.MaxDevicesPerAttendee, &e.SharedBandwidthMbps,
			&e.PerUserBandwidthMbps, &e.AccessCode, &e.SplashPageTheme, &e.BannerURL, &e.AcceptableUsePolicy, &e.Status,
			&e.CreatedAt, &e.UpdatedAt); err != nil {
			return nil, err
		}
		out = append(out, e)
	}
	return out, rows.Err()
}

func (s *Store) UpdateEvent(ctx context.Context, e *models.Event) error {
	e.UpdatedAt = now()
	_, err := s.DB.Pool.Exec(ctx, `
		UPDATE events SET name=$2, description=$3, venue=$4, organiser=$5, organiser_email=$6, starts_at=$7,
			ends_at=$8, ssid=$9, vlan_id=$10, packetfence_role=$11, max_attendees=$12, max_devices_per_attendee=$13,
			shared_bandwidth_mbps=$14, per_user_bandwidth_mbps=$15, splash_page_theme=$16, banner_url=$17,
			acceptable_use_policy=$18, status=$19, updated_at=NOW()
		WHERE id=$1`,
		e.ID, e.Name, e.Description, e.Venue, e.Organiser, e.OrganiserEmail, e.StartsAt,
		e.EndsAt, e.SSID, e.VLANID, e.PacketFenceRole, e.MaxAttendees, e.MaxDevicesPerAttendee, e.SharedBandwidthMbps,
		e.PerUserBandwidthMbps, e.SplashPageTheme, e.BannerURL, e.AcceptableUsePolicy, e.Status)
	return err
}

func (s *Store) DeleteEvent(ctx context.Context, id string) error {
	_, err := s.DB.Pool.Exec(ctx, `DELETE FROM events WHERE id=$1`, id)
	return err
}

func (s *Store) CreateEventAttendee(ctx context.Context, a *models.EventAttendee) (*models.EventAttendee, error) {
	a.ID = newID()
	a.CreatedAt = now()
	a.UpdatedAt = now()
	err := s.DB.Pool.QueryRow(ctx, `
		INSERT INTO event_attendees (id, event_id, property_id, email_hash, first_name, last_name_hash,
			voucher_code, qr_code, session_id, status, checked_in_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)
		RETURNING created_at, updated_at`,
		a.ID, a.EventID, a.PropertyID, a.EmailHash, a.FirstName, a.LastNameHash,
		a.VoucherCode, a.QRCode, a.SessionID, a.Status, a.CheckedInAt).Scan(&a.CreatedAt, &a.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return a, nil
}

func (s *Store) GetEventAttendeeByVoucher(ctx context.Context, eventID, voucher string) (*models.EventAttendee, error) {
	a := &models.EventAttendee{}
	err := s.DB.Pool.QueryRow(ctx, `
		SELECT id, event_id, property_id, email_hash, first_name, last_name_hash, voucher_code, qr_code,
			session_id, status, checked_in_at, created_at, updated_at
		FROM event_attendees WHERE event_id=$1 AND voucher_code=$2`, eventID, voucher).Scan(
		&a.ID, &a.EventID, &a.PropertyID, &a.EmailHash, &a.FirstName, &a.LastNameHash, &a.VoucherCode, &a.QRCode,
		&a.SessionID, &a.Status, &a.CheckedInAt, &a.CreatedAt, &a.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return a, nil
}

func (s *Store) ListEventAttendees(ctx context.Context, eventID string) ([]models.EventAttendee, error) {
	rows, err := s.DB.Pool.Query(ctx, `
		SELECT id, event_id, property_id, email_hash, first_name, last_name_hash, voucher_code, qr_code,
			session_id, status, checked_in_at, created_at, updated_at
		FROM event_attendees WHERE event_id=$1`, eventID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []models.EventAttendee{}
	for rows.Next() {
		a := models.EventAttendee{}
		if err := rows.Scan(
			&a.ID, &a.EventID, &a.PropertyID, &a.EmailHash, &a.FirstName, &a.LastNameHash, &a.VoucherCode, &a.QRCode,
			&a.SessionID, &a.Status, &a.CheckedInAt, &a.CreatedAt, &a.UpdatedAt); err != nil {
			return nil, err
		}
		out = append(out, a)
	}
	return out, rows.Err()
}

func (s *Store) UpdateEventAttendee(ctx context.Context, a *models.EventAttendee) error {
	a.UpdatedAt = now()
	_, err := s.DB.Pool.Exec(ctx, `
		UPDATE event_attendees SET status=$2, session_id=$3, checked_in_at=$4, updated_at=NOW()
		WHERE id=$1`,
		a.ID, a.Status, a.SessionID, a.CheckedInAt)
	return err
}

// ---- Promotions ----

func (s *Store) CreatePromotion(ctx context.Context, p *models.Promotion) (*models.Promotion, error) {
	p.ID = newID()
	p.CreatedAt = now()
	p.UpdatedAt = now()
	err := s.DB.Pool.QueryRow(ctx, `
		INSERT INTO promotions (id, property_id, name, type, title, description, image_url, cta_text, cta_url,
			target_segment, target_language, target_country, starts_at, ends_at, display_location, sort_order, is_active)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17)
		RETURNING created_at, updated_at`,
		p.ID, p.PropertyID, p.Name, p.Type, p.Title, p.Description, p.ImageURL, p.CTAText, p.CTAURL,
		p.TargetSegment, p.TargetLanguage, p.TargetCountry, p.StartsAt, p.EndsAt, p.DisplayLocation, p.SortOrder, p.IsActive).Scan(&p.CreatedAt, &p.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return p, nil
}

func (s *Store) ListPromotions(ctx context.Context, propertyID string, active bool) ([]models.Promotion, error) {
	q := `SELECT id, property_id, name, type, title, description, image_url, cta_text, cta_url,
			target_segment, target_language, target_country, starts_at, ends_at, display_location, sort_order,
			is_active, created_at, updated_at
		FROM promotions WHERE property_id=$1`
	args := []any{propertyID}
	if active {
		q += ` AND is_active=TRUE`
	}
	q += ` ORDER BY sort_order`
	rows, err := s.DB.Pool.Query(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []models.Promotion{}
	for rows.Next() {
		p := models.Promotion{}
		if err := rows.Scan(
			&p.ID, &p.PropertyID, &p.Name, &p.Type, &p.Title, &p.Description, &p.ImageURL, &p.CTAText, &p.CTAURL,
			&p.TargetSegment, &p.TargetLanguage, &p.TargetCountry, &p.StartsAt, &p.EndsAt, &p.DisplayLocation, &p.SortOrder,
			&p.IsActive, &p.CreatedAt, &p.UpdatedAt); err != nil {
			return nil, err
		}
		out = append(out, p)
	}
	return out, rows.Err()
}

func (s *Store) CreatePromotionInteraction(ctx context.Context, i *models.PromotionInteraction) error {
	i.ID = newID()
	i.CreatedAt = now()
	_, err := s.DB.Pool.Exec(ctx, `
		INSERT INTO promotion_interactions (id, promotion_id, guest_profile_id, guest_session_id, interaction_type,
			booking_reference, revenue_attribution_cents)
		VALUES ($1,$2,$3,$4,$5,$6,$7)`,
		i.ID, i.PromotionID, i.GuestProfileID, i.GuestSessionID, i.InteractionType,
		i.BookingReference, i.RevenueAttributionCents)
	return err
}

// ---- Review campaigns & requests ----

func (s *Store) CreateReviewCampaign(ctx context.Context, c *models.ReviewCampaign) (*models.ReviewCampaign, error) {
	c.ID = newID()
	c.CreatedAt = now()
	c.UpdatedAt = now()
	err := s.DB.Pool.QueryRow(ctx, `
		INSERT INTO review_campaigns (id, property_id, name, delay_hours, channel, template_id, google_review_url,
			tripadvisor_review_url, internal_survey_enabled, escalation_threshold, escalation_email, is_active)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12)
		RETURNING created_at, updated_at`,
		c.ID, c.PropertyID, c.Name, c.DelayHours, c.Channel, c.TemplateID, c.GoogleReviewURL,
		c.TripAdvisorReviewURL, c.InternalSurveyEnabled, c.EscalationThreshold, c.EscalationEmail, c.IsActive).Scan(&c.CreatedAt, &c.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return c, nil
}

func (s *Store) ListReviewCampaigns(ctx context.Context, propertyID string) ([]models.ReviewCampaign, error) {
	rows, err := s.DB.Pool.Query(ctx, `
		SELECT id, property_id, name, delay_hours, channel, template_id, google_review_url,
			tripadvisor_review_url, internal_survey_enabled, escalation_threshold, escalation_email, is_active,
			created_at, updated_at
		FROM review_campaigns WHERE property_id=$1`, propertyID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []models.ReviewCampaign{}
	for rows.Next() {
		c := models.ReviewCampaign{}
		if err := rows.Scan(
			&c.ID, &c.PropertyID, &c.Name, &c.DelayHours, &c.Channel, &c.TemplateID, &c.GoogleReviewURL,
			&c.TripAdvisorReviewURL, &c.InternalSurveyEnabled, &c.EscalationThreshold, &c.EscalationEmail, &c.IsActive,
			&c.CreatedAt, &c.UpdatedAt); err != nil {
			return nil, err
		}
		out = append(out, c)
	}
	return out, rows.Err()
}

func (s *Store) CreateReviewRequest(ctx context.Context, r *models.ReviewRequest) (*models.ReviewRequest, error) {
	r.ID = newID()
	r.CreatedAt = now()
	r.UpdatedAt = now()
	err := s.DB.Pool.QueryRow(ctx, `
		INSERT INTO review_requests (id, campaign_id, guest_session_id, guest_profile_id, reservation_id,
			property_id, status)
		VALUES ($1,$2,$3,$4,$5,$6,$7)
		RETURNING created_at, updated_at`,
		r.ID, r.CampaignID, r.GuestSessionID, r.GuestProfileID, r.ReservationID,
		r.PropertyID, r.Status).Scan(&r.CreatedAt, &r.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return r, nil
}

func (s *Store) ListReviewRequests(ctx context.Context, propertyID, status string) ([]models.ReviewRequest, error) {
	q := `SELECT id, campaign_id, guest_session_id, guest_profile_id, reservation_id, property_id, status,
			sent_at, opened_at, rating, feedback_text, external_review_submitted, created_at, updated_at
		FROM review_requests WHERE property_id=$1`
	args := []any{propertyID}
	if status != "" {
		q += ` AND status=$2`
		args = append(args, status)
	}
	q += ` ORDER BY created_at DESC`
	rows, err := s.DB.Pool.Query(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []models.ReviewRequest{}
	for rows.Next() {
		r := models.ReviewRequest{}
		if err := rows.Scan(
			&r.ID, &r.CampaignID, &r.GuestSessionID, &r.GuestProfileID, &r.ReservationID, &r.PropertyID, &r.Status,
			&r.SentAt, &r.OpenedAt, &r.Rating, &r.FeedbackText, &r.ExternalReviewSubmitted, &r.CreatedAt, &r.UpdatedAt); err != nil {
			return nil, err
		}
		out = append(out, r)
	}
	return out, rows.Err()
}

func (s *Store) GetReviewRequest(ctx context.Context, id string) (*models.ReviewRequest, error) {
	r := &models.ReviewRequest{}
	err := s.DB.Pool.QueryRow(ctx, `
		SELECT id, campaign_id, guest_session_id, guest_profile_id, reservation_id, property_id, status,
			sent_at, opened_at, rating, feedback_text, external_review_submitted, created_at, updated_at
		FROM review_requests WHERE id=$1`, id).Scan(
		&r.ID, &r.CampaignID, &r.GuestSessionID, &r.GuestProfileID, &r.ReservationID, &r.PropertyID, &r.Status,
		&r.SentAt, &r.OpenedAt, &r.Rating, &r.FeedbackText, &r.ExternalReviewSubmitted, &r.CreatedAt, &r.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return r, nil
}

func (s *Store) UpdateReviewRequest(ctx context.Context, r *models.ReviewRequest) error {
	r.UpdatedAt = now()
	_, err := s.DB.Pool.Exec(ctx, `
		UPDATE review_requests SET status=$2, sent_at=$3, opened_at=$4, rating=$5, feedback_text=$6,
			external_review_submitted=$7, updated_at=NOW()
		WHERE id=$1`,
		r.ID, r.Status, r.SentAt, r.OpenedAt, r.Rating, r.FeedbackText, r.ExternalReviewSubmitted)
	return err
}
