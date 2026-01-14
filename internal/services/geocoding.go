package services

import (
	"fmt"
)

// AddressDetails contains address components
type AddressDetails struct {
	StreetAddress string
	District      string // Kecamatan
	City          string // Kabupaten
	Village       string // Kelurahan
	PostalCode    string
}

// GeocodingService defines the interface for geocoding operations
type GeocodingService interface {
	ReverseGeocode(lat, lng float64) (*AddressDetails, error)
}

// MockGeocodingService implements a mock geocoding service
type MockGeocodingService struct{}

// NewMockGeocodingService creates a new mock geocoding service
func NewMockGeocodingService() *MockGeocodingService {
	return &MockGeocodingService{}
}

// ReverseGeocode returns a mock address based on coordinates
func (s *MockGeocodingService) ReverseGeocode(lat, lng float64) (*AddressDetails, error) {
	// TODO: Integrate with Google Maps or Mapbox API
	// For now, return a fixed location (Tebet, Jakarta Selatan) for testing
	// In a real implementation, this would call an external API

	// Basic logic to vary result based on coordinates to show "dynamic" behavior if needed
	// For simplicity, just return Tebet
	return &AddressDetails{
		StreetAddress: fmt.Sprintf("Jalan Simulated %f,%f", lat, lng),
		District:      "Kecamatan Tebet",
		City:          "Jakarta Selatan",
		Village:       "Tebet Timur",
		PostalCode:    "12820",
	}, nil
}
