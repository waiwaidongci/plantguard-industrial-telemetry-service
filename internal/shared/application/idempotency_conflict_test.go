package application

import (
	"testing"

	shareddomain "github.com/acme/plantguard/internal/shared/domain"
)

func TestIsIdempotencyConflict(t *testing.T) {
	if !IsIdempotencyConflict(shareddomain.ErrConflict) {
		t.Fatal("expected conflict to be recognized")
	}
}
