//go:build ignore

package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"

	"github.com/tuannm99/podzone/internal/backoffice/migrations"
	"github.com/tuannm99/podzone/pkg/toolkit"
)

type cfg struct {
	TenantID       string
	StoreName      string
	StoreSubdomain string
	SchemaName     string
	BackofficeDB   string
	PGHost         string
	PGPort         string
	PGUser         string
	PGPassword     string
	PGSSLMode      string
	OnboardingURL  string
}

type activityDetail struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

type orderActivity struct {
	Type      string           `json:"type"`
	Actor     string           `json:"actor"`
	Message   string           `json:"message"`
	Details   []activityDetail `json:"details"`
	CreatedAt time.Time        `json:"createdAt"`
}

func main() {
	cfg := loadCfg()
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := createOnboardingStore(ctx, cfg); err != nil {
		fmt.Fprintf(os.Stderr, "onboarding store seed failed: %v\n", err)
	}

	backofficeDB, err := openPostgres(cfg, cfg.BackofficeDB)
	if err != nil {
		fail("open backoffice db", err)
	}
	defer backofficeDB.Close()

	if err := seedBackoffice(ctx, backofficeDB, cfg); err != nil {
		fail("seed backoffice", err)
	}

	partnerDB, err := openPostgres(cfg, "partner")
	if err != nil {
		fail("open partner db", err)
	}
	defer partnerDB.Close()

	if err := seedPartners(ctx, partnerDB, cfg); err != nil {
		fail("seed partners", err)
	}

	fmt.Printf("Seeded sample POD data for tenant=%s schema=%s\n", cfg.TenantID, cfg.SchemaName)
}

func loadCfg() cfg {
	tenantID := envOr("TENANT_ID", "tenant-dev")
	return cfg{
		TenantID:       tenantID,
		StoreName:      envOr("STORE_NAME", "Demo POD Store"),
		StoreSubdomain: envOr("STORE_SUBDOMAIN", "demo-pod-store"),
		SchemaName:     envOr("SCHEMA_NAME", toolkit.SchemaName("t_", tenantID)),
		BackofficeDB:   envOr("DB_NAME", "podzone_tenants"),
		PGHost:         envOr("PG_HOST", "localhost"),
		PGPort:         envOr("PG_PORT", "5432"),
		PGUser:         envOr("PG_USER", "postgres"),
		PGPassword:     envOr("PG_PASSWORD", "postgres"),
		PGSSLMode:      envOr("PG_SSL_MODE", "disable"),
		OnboardingURL:  strings.TrimRight(envOr("ONBOARDING_URL", "http://localhost:8800"), "/"),
	}
}

func envOr(key, fallback string) string {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}
	return value
}

func openPostgres(cfg cfg, dbName string) (*sqlx.DB, error) {
	if dbName != "postgres" {
		if err := ensureDatabase(cfg, dbName); err != nil {
			return nil, err
		}
	}

	dsn := (&url.URL{
		Scheme: "postgres",
		User:   url.UserPassword(cfg.PGUser, cfg.PGPassword),
		Host:   fmt.Sprintf("%s:%s", cfg.PGHost, cfg.PGPort),
		Path:   "/" + dbName,
		RawQuery: url.Values{
			"sslmode": []string{cfg.PGSSLMode},
		}.Encode(),
	}).String()

	db, err := sqlx.Open("postgres", dsn)
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(4)
	db.SetMaxIdleConns(2)
	db.SetConnMaxLifetime(5 * time.Minute)
	if err := db.Ping(); err != nil {
		_ = db.Close()
		return nil, err
	}
	return db, nil
}

func ensureDatabase(cfg cfg, dbName string) error {
	admin, err := openPostgres(cfg, "postgres")
	if err != nil {
		return fmt.Errorf("open postgres admin database: %w", err)
	}
	defer admin.Close()

	var exists bool
	if err := admin.QueryRow(`SELECT EXISTS (SELECT 1 FROM pg_database WHERE datname = $1)`, dbName).Scan(&exists); err != nil {
		return fmt.Errorf("check database %q: %w", dbName, err)
	}
	if exists {
		return nil
	}
	if _, err := admin.Exec(`CREATE DATABASE ` + quoteIdent(dbName)); err != nil {
		return fmt.Errorf("create database %q: %w", dbName, err)
	}
	return nil
}

func seedBackoffice(ctx context.Context, db *sqlx.DB, cfg cfg) error {
	if _, err := db.ExecContext(ctx, fmt.Sprintf(`CREATE SCHEMA IF NOT EXISTS %s`, quoteIdent(cfg.SchemaName))); err != nil {
		return err
	}

	tx, err := db.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	if _, err := tx.ExecContext(ctx, fmt.Sprintf(`SET LOCAL search_path TO %s, public`, quoteIdent(cfg.SchemaName))); err != nil {
		return err
	}
	if err := migrations.ApplyTx(ctx, tx); err != nil {
		return err
	}

	now := time.Now().UTC()
	storeID := "seed-store-main"
	if _, err := tx.ExecContext(ctx, `
		INSERT INTO stores (id, name, description, owner_id, status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		ON CONFLICT (id) DO UPDATE SET
			name = EXCLUDED.name,
			description = EXCLUDED.description,
			owner_id = EXCLUDED.owner_id,
			status = EXCLUDED.status,
			updated_at = EXCLUDED.updated_at
	`, storeID, cfg.StoreName, "Seeded POD operations store", "dev-owner", "active", now.Add(-72*time.Hour), now); err != nil {
		return err
	}

	type draft struct {
		id, name, partner, baseCost, retailPrice, status, notes string
		createdAt, updatedAt                                    time.Time
	}
	drafts := []draft{
		{
			id:          "seed-draft-classic-tee",
			name:        "Classic POD Tee",
			partner:     "Print Partner A",
			baseCost:    "$12.40",
			retailPrice: "$29.99",
			status:      "publish_candidate",
			notes:       "Approved for launch and reprint-safe artwork specs.",
			createdAt:   now.Add(-48 * time.Hour),
			updatedAt:   now.Add(-26 * time.Hour),
		},
		{
			id:          "seed-draft-heavy-hoodie",
			name:        "Heavy Hoodie",
			partner:     "Rapid Fulfillment",
			baseCost:    "$24.10",
			retailPrice: "$54.00",
			status:      "ready_for_review",
			notes:       "Margin review pending for EU lane.",
			createdAt:   now.Add(-36 * time.Hour),
			updatedAt:   now.Add(-18 * time.Hour),
		},
	}
	for _, item := range drafts {
		if _, err := tx.ExecContext(ctx, `
			INSERT INTO product_setup_drafts (id, name, partner, base_cost, retail_price, status, notes, created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
			ON CONFLICT (id) DO UPDATE SET
				name = EXCLUDED.name,
				partner = EXCLUDED.partner,
				base_cost = EXCLUDED.base_cost,
				retail_price = EXCLUDED.retail_price,
				status = EXCLUDED.status,
				notes = EXCLUDED.notes,
				updated_at = EXCLUDED.updated_at
		`, item.id, item.name, item.partner, item.baseCost, item.retailPrice, item.status, item.notes, item.createdAt, item.updatedAt); err != nil {
			return err
		}
	}

	type candidate struct {
		id, draftID, title, sku, partner, baseCost, retailPrice, estimatedMargin, status, channel, variantsJSON, checklistJSON, merchandisingNotes string
		updatedAt                                                                                                                                  time.Time
	}
	candidates := []candidate{
		{
			id:                 "seed-candidate-classic-tee",
			draftID:            "seed-draft-classic-tee",
			title:              "Classic POD Tee",
			sku:                "TEE-CLASSIC-BLK-M",
			partner:            "Print Partner A",
			baseCost:           "$12.40",
			retailPrice:        "$29.99",
			estimatedMargin:    "$17.59",
			status:             "published_mock",
			channel:            "etsy",
			variantsJSON:       `[{"id":"variant-tee-black-m","label":"Black / M","color":"Black","size":"M","status":"ready"},{"id":"variant-tee-white-l","label":"White / L","color":"White","size":"L","status":"ready"}]`,
			checklistJSON:      `{"frontArtwork":true,"backArtwork":false,"mockupReady":true,"printSpecChecked":true}`,
			merchandisingNotes: "Primary bestseller lane for quick-turn POD samples.",
			updatedAt:          now.Add(-20 * time.Hour),
		},
		{
			id:                 "seed-candidate-heavy-hoodie",
			draftID:            "seed-draft-heavy-hoodie",
			title:              "Heavy Hoodie",
			sku:                "HOOD-HEAVY-ASH-XL",
			partner:            "Rapid Fulfillment",
			baseCost:           "$24.10",
			retailPrice:        "$54.00",
			estimatedMargin:    "$29.90",
			status:             "published_mock",
			channel:            "shopify",
			variantsJSON:       `[{"id":"variant-hoodie-ash-xl","label":"Ash / XL","color":"Ash","size":"XL","status":"ready"}]`,
			checklistJSON:      `{"frontArtwork":true,"backArtwork":true,"mockupReady":true,"printSpecChecked":true}`,
			merchandisingNotes: "Use this candidate for higher AOV routes and settlement testing.",
			updatedAt:          now.Add(-12 * time.Hour),
		},
	}
	for _, item := range candidates {
		if _, err := tx.ExecContext(ctx, `
			INSERT INTO product_setup_candidates (id, draft_id, title, sku, partner, base_cost, retail_price, estimated_margin, status, channel, variants_json, artwork_checklist_json, merchandising_notes, updated_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)
			ON CONFLICT (id) DO UPDATE SET
				draft_id = EXCLUDED.draft_id,
				title = EXCLUDED.title,
				sku = EXCLUDED.sku,
				partner = EXCLUDED.partner,
				base_cost = EXCLUDED.base_cost,
				retail_price = EXCLUDED.retail_price,
				estimated_margin = EXCLUDED.estimated_margin,
				status = EXCLUDED.status,
				channel = EXCLUDED.channel,
				variants_json = EXCLUDED.variants_json,
				artwork_checklist_json = EXCLUDED.artwork_checklist_json,
				merchandising_notes = EXCLUDED.merchandising_notes,
				updated_at = EXCLUDED.updated_at
		`, item.id, item.draftID, item.title, item.sku, item.partner, item.baseCost, item.retailPrice, item.estimatedMargin, item.status, item.channel, item.variantsJSON, item.checklistJSON, item.merchandisingNotes, item.updatedAt); err != nil {
			return err
		}
	}

	type order struct {
		id, candidateID, productTitle, partner, total, customerName, status, exceptionType, exceptionStatus string
		quantity                                                                                            int
		timelineJSON                                                                                        string
		shipmentStatus, shipmentCarrier, shipmentTrackingNumber, shipmentTrackingURL, shipmentNotes         string
		operatorAssignee, baseCostSnapshot, fulfillmentCost, shippingCost, issueCost, issueResolution       string
		issueNotes, realizedMargin, settlementStatus, settlementNotes                                       string
		shipmentSlaDueAt, issueSlaDueAt, shippedAt, deliveredAt                                             *time.Time
		createdAt, updatedAt                                                                                time.Time
		activities                                                                                          []orderActivity
	}
	order1ShipmentSLA := now.Add(-4 * time.Hour)
	order2IssueSLA := now.Add(-2 * time.Hour)
	order3ShipmentSLA := now.Add(6 * time.Hour)
	order3ShippedAt := now.Add(-8 * time.Hour)
	order4IssueSLA := now.Add(-10 * time.Hour)
	order4ShippedAt := now.Add(-30 * time.Hour)
	order5ShippedAt := now.Add(-50 * time.Hour)
	order5DeliveredAt := now.Add(-20 * time.Hour)
	orders := []order{
		{
			id:               "seed-order-001",
			candidateID:      "seed-candidate-classic-tee",
			productTitle:     "Classic POD Tee",
			partner:          "Print Partner A",
			quantity:         1,
			total:            "$29.99",
			customerName:     "Ava Martinez",
			status:           "queued",
			timelineJSON:     `["Order imported from Etsy","Artwork approved","Waiting label handoff"]`,
			exceptionType:    "",
			exceptionStatus:  "",
			shipmentStatus:   "awaiting_label",
			shipmentCarrier:  "",
			shipmentNotes:    "Label queue waiting on morning pickup batch.",
			operatorAssignee: "unassigned",
			shipmentSlaDueAt: &order1ShipmentSLA,
			baseCostSnapshot: "$12.40",
			fulfillmentCost:  "$0.00",
			shippingCost:     "$0.00",
			issueCost:        "$0.00",
			issueResolution:  "monitor",
			issueNotes:       "",
			realizedMargin:   "$17.59",
			settlementStatus: "pending",
			settlementNotes:  "Awaiting carrier label cost.",
			createdAt:        now.Add(-10 * time.Hour),
			updatedAt:        now.Add(-90 * time.Minute),
			activities: []orderActivity{
				{
					Type:      "system",
					Actor:     "system",
					Message:   "Seeded queued order for SLA testing",
					CreatedAt: now.Add(-10 * time.Hour),
				},
				{
					Type:      "shipment_note",
					Actor:     "user:ops.seed",
					Message:   "Queued for label creation",
					Details:   []activityDetail{{Key: "shipment_status", Value: "awaiting_label"}},
					CreatedAt: now.Add(-2 * time.Hour),
				},
				{
					Type:      "settlement_note",
					Actor:     "user:finance.seed",
					Message:   "Settlement still pending shipping rate",
					Details:   []activityDetail{{Key: "settlement_status", Value: "pending"}},
					CreatedAt: now.Add(-90 * time.Minute),
				},
			},
		},
		{
			id:                     "seed-order-002",
			candidateID:            "seed-candidate-heavy-hoodie",
			productTitle:           "Heavy Hoodie",
			partner:                "Rapid Fulfillment",
			quantity:               2,
			total:                  "$108.00",
			customerName:           "Mia Carter",
			status:                 "in_production",
			timelineJSON:           `["Order routed to partner","QC exception opened","Reprint approved"]`,
			exceptionType:          "reprint_request",
			exceptionStatus:        "open",
			shipmentStatus:         "label_ready",
			shipmentCarrier:        "UPS",
			shipmentTrackingNumber: "1Z999SEED002",
			shipmentTrackingURL:    "https://tracking.example.com/1Z999SEED002",
			shipmentNotes:          "Label prepared while reprint unit clears final QC.",
			operatorAssignee:       "ops.linh",
			issueSlaDueAt:          &order2IssueSLA,
			baseCostSnapshot:       "$48.20",
			fulfillmentCost:        "$52.00",
			shippingCost:           "$12.00",
			issueCost:              "$26.00",
			issueResolution:        "reprint",
			issueNotes:             "Front print density failed first pass.",
			realizedMargin:         "$18.00",
			settlementStatus:       "disputed",
			settlementNotes:        "Reprint surcharge under review with partner.",
			createdAt:              now.Add(-28 * time.Hour),
			updatedAt:              now.Add(-75 * time.Minute),
			activities: []orderActivity{
				{
					Type:      "system",
					Actor:     "system",
					Message:   "Seeded production order with exception",
					CreatedAt: now.Add(-28 * time.Hour),
				},
				{
					Type:    "issue_note",
					Actor:   "user:ops.linh",
					Message: "Reprint approved after QC failure",
					Details: []activityDetail{
						{Key: "issue_resolution", Value: "reprint"},
						{Key: "issue_cost", Value: "$26.00"},
					},
					CreatedAt: now.Add(-3 * time.Hour),
				},
				{
					Type:      "settlement_note",
					Actor:     "user:finance.seed",
					Message:   "Partner surcharge disputed",
					Details:   []activityDetail{{Key: "settlement_status", Value: "disputed"}},
					CreatedAt: now.Add(-75 * time.Minute),
				},
			},
		},
		{
			id:                     "seed-order-003",
			candidateID:            "seed-candidate-classic-tee",
			productTitle:           "Classic POD Tee",
			partner:                "Print Partner A",
			quantity:               2,
			total:                  "$59.98",
			customerName:           "Noah Kim",
			status:                 "shipped",
			timelineJSON:           `["Order packed","Carrier handoff complete","In transit to customer"]`,
			exceptionType:          "",
			exceptionStatus:        "",
			shipmentStatus:         "in_transit",
			shipmentCarrier:        "USPS",
			shipmentTrackingNumber: "9400SEED003",
			shipmentTrackingURL:    "https://tracking.example.com/9400SEED003",
			shipmentNotes:          "Tracking updated from carrier scan.",
			operatorAssignee:       "ops.pack",
			shipmentSlaDueAt:       &order3ShipmentSLA,
			baseCostSnapshot:       "$24.80",
			fulfillmentCost:        "$24.80",
			shippingCost:           "$7.20",
			issueCost:              "$0.00",
			issueResolution:        "monitor",
			issueNotes:             "",
			realizedMargin:         "$27.98",
			settlementStatus:       "reconciled",
			settlementNotes:        "Carrier cost posted cleanly.",
			shippedAt:              &order3ShippedAt,
			createdAt:              now.Add(-22 * time.Hour),
			updatedAt:              now.Add(-4 * time.Hour),
			activities: []orderActivity{
				{
					Type:      "system",
					Actor:     "system",
					Message:   "Seeded in-transit order",
					CreatedAt: now.Add(-22 * time.Hour),
				},
				{
					Type:    "shipment_note",
					Actor:   "user:ops.pack",
					Message: "Handed off to USPS",
					Details: []activityDetail{
						{Key: "shipment_status", Value: "in_transit"},
						{Key: "carrier", Value: "USPS"},
					},
					CreatedAt: now.Add(-8 * time.Hour),
				},
				{
					Type:      "settlement_note",
					Actor:     "user:finance.seed",
					Message:   "Settlement reconciled",
					Details:   []activityDetail{{Key: "settlement_status", Value: "reconciled"}},
					CreatedAt: now.Add(-4 * time.Hour),
				},
			},
		},
		{
			id:                     "seed-order-004",
			candidateID:            "seed-candidate-heavy-hoodie",
			productTitle:           "Heavy Hoodie",
			partner:                "Rapid Fulfillment",
			quantity:               1,
			total:                  "$54.00",
			customerName:           "Elena Rossi",
			status:                 "shipped",
			timelineJSON:           `["Order shipped","Carrier exception detected","Claims follow-up opened"]`,
			exceptionType:          "carrier_delay",
			exceptionStatus:        "escalated",
			shipmentStatus:         "delivery_issue",
			shipmentCarrier:        "DHL eCommerce",
			shipmentTrackingNumber: "DHLSEED004",
			shipmentTrackingURL:    "https://tracking.example.com/DHLSEED004",
			shipmentNotes:          "Carrier reports delay at regional hub.",
			operatorAssignee:       "ops.claims",
			issueSlaDueAt:          &order4IssueSLA,
			baseCostSnapshot:       "$24.10",
			fulfillmentCost:        "$24.10",
			shippingCost:           "$8.40",
			issueCost:              "$8.50",
			issueResolution:        "carrier_claim",
			issueNotes:             "Claim packet prepared for late delivery rebate.",
			realizedMargin:         "$13.00",
			settlementStatus:       "disputed",
			settlementNotes:        "Hold payout until carrier claim resolves.",
			shippedAt:              &order4ShippedAt,
			createdAt:              now.Add(-44 * time.Hour),
			updatedAt:              now.Add(-40 * time.Minute),
			activities: []orderActivity{
				{
					Type:      "system",
					Actor:     "system",
					Message:   "Seeded delivery issue order",
					CreatedAt: now.Add(-44 * time.Hour),
				},
				{
					Type:    "shipment_note",
					Actor:   "user:ops.claims",
					Message: "Carrier marked shipment delayed",
					Details: []activityDetail{
						{Key: "shipment_status", Value: "delivery_issue"},
						{Key: "carrier", Value: "DHL eCommerce"},
					},
					CreatedAt: now.Add(-6 * time.Hour),
				},
				{
					Type:    "issue_note",
					Actor:   "user:ops.claims",
					Message: "Carrier claim opened",
					Details: []activityDetail{
						{Key: "issue_resolution", Value: "carrier_claim"},
						{Key: "issue_cost", Value: "$8.50"},
					},
					CreatedAt: now.Add(-40 * time.Minute),
				},
			},
		},
		{
			id:                     "seed-order-005",
			candidateID:            "seed-candidate-classic-tee",
			productTitle:           "Classic POD Tee",
			partner:                "Print Partner A",
			quantity:               1,
			total:                  "$29.99",
			customerName:           "Jordan Lee",
			status:                 "shipped",
			timelineJSON:           `["Order shipped","Delivered","Settlement paid"]`,
			exceptionType:          "",
			exceptionStatus:        "resolved",
			shipmentStatus:         "delivered",
			shipmentCarrier:        "USPS",
			shipmentTrackingNumber: "9400SEED005",
			shipmentTrackingURL:    "https://tracking.example.com/9400SEED005",
			shipmentNotes:          "Delivered with no incident.",
			operatorAssignee:       "ops.pack",
			baseCostSnapshot:       "$12.40",
			fulfillmentCost:        "$12.40",
			shippingCost:           "$4.20",
			issueCost:              "$0.00",
			issueResolution:        "monitor",
			issueNotes:             "",
			realizedMargin:         "$13.39",
			settlementStatus:       "paid",
			settlementNotes:        "Closed and paid.",
			shippedAt:              &order5ShippedAt,
			deliveredAt:            &order5DeliveredAt,
			createdAt:              now.Add(-60 * time.Hour),
			updatedAt:              now.Add(-18 * time.Hour),
			activities: []orderActivity{
				{
					Type:      "system",
					Actor:     "system",
					Message:   "Seeded delivered order",
					CreatedAt: now.Add(-60 * time.Hour),
				},
				{
					Type:      "shipment_note",
					Actor:     "user:ops.pack",
					Message:   "Delivered successfully",
					Details:   []activityDetail{{Key: "shipment_status", Value: "delivered"}},
					CreatedAt: now.Add(-20 * time.Hour),
				},
				{
					Type:      "settlement_note",
					Actor:     "user:finance.seed",
					Message:   "Settlement paid out",
					Details:   []activityDetail{{Key: "settlement_status", Value: "paid"}},
					CreatedAt: now.Add(-18 * time.Hour),
				},
			},
		},
	}

	orderIDs := make([]string, 0, len(orders))
	for _, item := range orders {
		orderIDs = append(orderIDs, item.id)
	}
	if err := deleteActivities(ctx, tx, orderIDs); err != nil {
		return err
	}

	for _, item := range orders {
		if _, err := tx.ExecContext(ctx, `
			INSERT INTO routed_orders (
				id, candidate_id, product_title, partner, quantity, total, customer_name, status, timeline_json,
				exception_type, exception_status, shipment_status, shipment_carrier, shipment_tracking_number,
				shipment_tracking_url, shipment_notes, operator_assignee, shipment_sla_due_at, issue_sla_due_at,
				base_cost_snapshot, fulfillment_cost, shipping_cost, issue_cost, issue_resolution, issue_notes,
				realized_margin, settlement_status, settlement_notes, shipped_at, delivered_at, created_at, updated_at
			) VALUES (
				$1, $2, $3, $4, $5, $6, $7, $8, $9,
				$10, $11, $12, $13, $14,
				$15, $16, $17, $18, $19,
				$20, $21, $22, $23, $24, $25,
				$26, $27, $28, $29, $30, $31, $32
			)
			ON CONFLICT (id) DO UPDATE SET
				candidate_id = EXCLUDED.candidate_id,
				product_title = EXCLUDED.product_title,
				partner = EXCLUDED.partner,
				quantity = EXCLUDED.quantity,
				total = EXCLUDED.total,
				customer_name = EXCLUDED.customer_name,
				status = EXCLUDED.status,
				timeline_json = EXCLUDED.timeline_json,
				exception_type = EXCLUDED.exception_type,
				exception_status = EXCLUDED.exception_status,
				shipment_status = EXCLUDED.shipment_status,
				shipment_carrier = EXCLUDED.shipment_carrier,
				shipment_tracking_number = EXCLUDED.shipment_tracking_number,
				shipment_tracking_url = EXCLUDED.shipment_tracking_url,
				shipment_notes = EXCLUDED.shipment_notes,
				operator_assignee = EXCLUDED.operator_assignee,
				shipment_sla_due_at = EXCLUDED.shipment_sla_due_at,
				issue_sla_due_at = EXCLUDED.issue_sla_due_at,
				base_cost_snapshot = EXCLUDED.base_cost_snapshot,
				fulfillment_cost = EXCLUDED.fulfillment_cost,
				shipping_cost = EXCLUDED.shipping_cost,
				issue_cost = EXCLUDED.issue_cost,
				issue_resolution = EXCLUDED.issue_resolution,
				issue_notes = EXCLUDED.issue_notes,
				realized_margin = EXCLUDED.realized_margin,
				settlement_status = EXCLUDED.settlement_status,
				settlement_notes = EXCLUDED.settlement_notes,
				shipped_at = EXCLUDED.shipped_at,
				delivered_at = EXCLUDED.delivered_at,
				updated_at = EXCLUDED.updated_at
		`, item.id, item.candidateID, item.productTitle, item.partner, item.quantity, item.total, item.customerName, item.status, item.timelineJSON,
			item.exceptionType, item.exceptionStatus, item.shipmentStatus, item.shipmentCarrier, item.shipmentTrackingNumber,
			item.shipmentTrackingURL, item.shipmentNotes, item.operatorAssignee, item.shipmentSlaDueAt, item.issueSlaDueAt,
			item.baseCostSnapshot, item.fulfillmentCost, item.shippingCost, item.issueCost, item.issueResolution, item.issueNotes,
			item.realizedMargin, item.settlementStatus, item.settlementNotes, item.shippedAt, item.deliveredAt, item.createdAt, item.updatedAt); err != nil {
			return err
		}
		for _, activity := range item.activities {
			detailsJSON, err := json.Marshal(activity.Details)
			if err != nil {
				return err
			}
			if _, err := tx.ExecContext(ctx, `
				INSERT INTO routed_order_activities (
					order_id, product_title, partner, operator_assignee, activity_type, actor, message, details_json, created_at
				) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
			`, item.id, item.productTitle, item.partner, item.operatorAssignee, activity.Type, activity.Actor, activity.Message, string(detailsJSON), activity.CreatedAt); err != nil {
				return err
			}
		}
	}

	return tx.Commit()
}

func deleteActivities(ctx context.Context, tx *sqlx.Tx, orderIDs []string) error {
	if len(orderIDs) == 0 {
		return nil
	}
	args := make([]any, 0, len(orderIDs))
	placeholders := make([]string, 0, len(orderIDs))
	for i, id := range orderIDs {
		args = append(args, id)
		placeholders = append(placeholders, fmt.Sprintf("$%d", i+1))
	}
	_, err := tx.ExecContext(ctx,
		fmt.Sprintf(`DELETE FROM routed_order_activities WHERE order_id IN (%s)`, strings.Join(placeholders, ", ")),
		args...,
	)
	return err
}

func seedPartners(ctx context.Context, db *sqlx.DB, cfg cfg) error {
	if _, err := db.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS partners (
			id TEXT PRIMARY KEY,
			tenant_id TEXT NOT NULL,
			code TEXT NOT NULL,
			name TEXT NOT NULL,
			contact_name TEXT NOT NULL DEFAULT '',
			contact_email TEXT NOT NULL DEFAULT '',
			notes TEXT NOT NULL DEFAULT '',
			partner_type TEXT NOT NULL DEFAULT 'print_on_demand',
			status TEXT NOT NULL,
			created_at TIMESTAMPTZ NOT NULL,
			updated_at TIMESTAMPTZ NOT NULL,
			CONSTRAINT uq_partners_tenant_code UNIQUE (tenant_id, code)
		)
	`); err != nil {
		return err
	}

	now := time.Now().UTC()
	type partner struct {
		id, code, name, contactName, contactEmail, notes, partnerType, status string
	}
	partners := []partner{
		{
			id:           "seed-partner-print-a",
			code:         "print-partner-a",
			name:         "Print Partner A",
			contactName:  "Alicia Tran",
			contactEmail: "alicia@printpartnera.dev",
			notes:        "Primary POD printer for apparel fast lane.",
			partnerType:  "print_on_demand",
			status:       "active",
		},
		{
			id:           "seed-partner-rapid-fulfillment",
			code:         "rapid-fulfillment",
			name:         "Rapid Fulfillment",
			contactName:  "Marco Silva",
			contactEmail: "marco@rapidfulfillment.dev",
			notes:        "Escalation partner for premium blanks and reprints.",
			partnerType:  "fulfillment",
			status:       "active",
		},
		{
			id:           "seed-partner-carrier-claims",
			code:         "carrier-claims-desk",
			name:         "Carrier Claims Desk",
			contactName:  "Daria Nguyen",
			contactEmail: "claims@carrierdesk.dev",
			notes:        "Used for delay and refund follow-up examples.",
			partnerType:  "dropship_supplier",
			status:       "active",
		},
	}

	for _, item := range partners {
		if _, err := db.ExecContext(ctx, `
			INSERT INTO partners (
				id, tenant_id, code, name, contact_name, contact_email, notes, partner_type, status, created_at, updated_at
			) VALUES (
				$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11
			)
			ON CONFLICT (id) DO UPDATE SET
				tenant_id = EXCLUDED.tenant_id,
				code = EXCLUDED.code,
				name = EXCLUDED.name,
				contact_name = EXCLUDED.contact_name,
				contact_email = EXCLUDED.contact_email,
				notes = EXCLUDED.notes,
				partner_type = EXCLUDED.partner_type,
				status = EXCLUDED.status,
				updated_at = EXCLUDED.updated_at
		`, item.id, cfg.TenantID, item.code, item.name, item.contactName, item.contactEmail, item.notes, item.partnerType, item.status, now.Add(-72*time.Hour), now); err != nil {
			return err
		}
	}
	return nil
}

func createOnboardingStore(ctx context.Context, cfg cfg) error {
	payload := map[string]string{
		"workspace_id": cfg.TenantID,
		"requested_by": "dev-bootstrap",
		"name":         cfg.StoreName,
		"subdomain":    cfg.StoreSubdomain,
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		cfg.OnboardingURL+"/onboarding/v1/stores",
		bytes.NewReader(body),
	)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusCreated || resp.StatusCode == http.StatusConflict {
		return nil
	}
	return fmt.Errorf("unexpected onboarding store status: %s", resp.Status)
}

func quoteIdent(s string) string {
	return `"` + strings.ReplaceAll(s, `"`, `""`) + `"`
}

func fail(step string, err error) {
	fmt.Fprintf(os.Stderr, "%s: %v\n", step, err)
	os.Exit(1)
}
