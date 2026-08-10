package domain

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
)

type CatalogImportStatus string

const (
	CatalogImportPending  CatalogImportStatus = "pending"
	CatalogImportApproved CatalogImportStatus = "approved"
	CatalogImportRejected CatalogImportStatus = "rejected"
)

const (
	CatalogSourcePSStore   = "ps_store"
	CatalogSourceXboxStore = "xbox_store"
)

type CatalogImport struct {
	ID               uuid.UUID           `db:"id" json:"id"`
	ExternalID       string              `db:"external_id" json:"external_id"`
	Source           string              `db:"source" json:"source"`
	Title            string              `db:"title" json:"title"`
	TitleKey         string              `db:"title_key" json:"title_key"`
	PlatformFamily   string              `db:"platform_family" json:"platform_family"`
	Description      *string             `db:"description" json:"description"`
	ImageURL         *string             `db:"image_url" json:"image_url"`
	Platforms        pq.StringArray      `db:"platforms" json:"platforms"`
	GameSection      string              `db:"game_section" json:"game_section"`
	ReleaseYear      *int                `db:"release_year" json:"release_year"`
	ReleaseDate      *time.Time          `db:"release_date" json:"release_date"`
	Publisher        string              `db:"publisher" json:"publisher"`
	OriginalPriceRUB *float64            `db:"original_price_rub" json:"original_price_rub"`
	OriginalCurrency *string             `db:"original_currency" json:"original_currency"`
	Raw              json.RawMessage     `db:"raw" json:"raw"`
	Prices           json.RawMessage     `db:"prices" json:"prices"`
	Status           CatalogImportStatus `db:"status" json:"status"`
	ProductID        *uuid.UUID          `db:"product_id" json:"product_id"`
	CreatedAt        time.Time           `db:"created_at" json:"created_at"`
	UpdatedAt        time.Time           `db:"updated_at" json:"updated_at"`
}
