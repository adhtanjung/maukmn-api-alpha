package models

import "github.com/google/uuid"

// Address represents an Indonesia administrative address
type Address struct {
	AddressID     uuid.UUID `db:"address_id" json:"address_id"`
	StreetAddress *string   `db:"street_address" json:"street_address,omitempty"`
	Kelurahan     *string   `db:"kelurahan" json:"kelurahan,omitempty"`
	Kecamatan     *string   `db:"kecamatan" json:"kecamatan,omitempty"`
	Kabupaten     *string   `db:"kabupaten" json:"kabupaten,omitempty"`
	Provinsi      *string   `db:"provinsi" json:"provinsi,omitempty"`
	PostalCode    *string   `db:"postal_code" json:"postal_code,omitempty"`
	DisplayName   *string   `db:"display_name" json:"display_name,omitempty"`
}
