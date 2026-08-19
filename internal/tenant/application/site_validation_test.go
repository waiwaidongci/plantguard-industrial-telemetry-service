package application

import (
	"context"
	"errors"
	"testing"

	shareddomain "github.com/acme/plantguard/internal/shared/domain"
	"github.com/acme/plantguard/internal/tenant/domain"
)

func TestCreateSiteRejectsEmptyNameWithoutPanic(t *testing.T) {
	svc := NewService(nil, nil, nil, nil)
	t.Run("empty name", func(t *testing.T) {
		_, _, err := svc.CreateSite(context.Background(), "tenant-1", "", CreateSiteInput{Name: "   ", Location: "Line 1"})
		if !errors.Is(err, domain.ErrSiteNameRequired) {
			t.Fatalf("expected ErrSiteNameRequired, got %T: %v", err, err)
		}
	})
	t.Run("empty location", func(t *testing.T) {
		_, _, err := svc.CreateSite(context.Background(), "tenant-1", "", CreateSiteInput{Name: "Pump", Location: "   "})
		if !errors.Is(err, shareddomain.ErrBadRequest) {
			t.Fatalf("expected ErrBadRequest, got %T: %v", err, err)
		}
	})
	t.Run("empty timezone", func(t *testing.T) {
		_, _, err := svc.CreateSite(context.Background(), "tenant-1", "", CreateSiteInput{Name: "Pump", Location: "Line 1", Timezone: "   "})
		if !errors.Is(err, shareddomain.ErrBadRequest) {
			t.Fatalf("expected ErrBadRequest, got %T: %v", err, err)
		}
	})
}

func TestValidateSiteRejectsInvalidInput(t *testing.T) {
	t.Run("name", func(t *testing.T) {
		if err := validateSite(CreateSiteInput{Name: "   ", Location: "Line 1", Timezone: "UTC"}); !errors.Is(err, domain.ErrSiteNameRequired) {
			t.Fatalf("expected name error, got %T: %v", err, err)
		}
	})
	t.Run("location", func(t *testing.T) {
		if err := validateSite(CreateSiteInput{Name: "Pump", Location: "   ", Timezone: "UTC"}); !errors.Is(err, shareddomain.ErrBadRequest) {
			t.Fatalf("expected bad request, got %T: %v", err, err)
		}
	})
	t.Run("timezone", func(t *testing.T) {
		if err := validateSite(CreateSiteInput{Name: "Pump", Location: "Line 1", Timezone: "   "}); !errors.Is(err, shareddomain.ErrBadRequest) {
			t.Fatalf("expected bad request, got %T: %v", err, err)
		}
	})
}
