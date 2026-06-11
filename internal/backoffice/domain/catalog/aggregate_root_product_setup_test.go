package catalog

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestNewProductSetupDraftValidatesAndEmitsEvent(t *testing.T) {
	t.Parallel()

	draft, events, err := NewProductSetupDraft("draft-1", CreateProductSetupDraftCmd{
		StoreID:     "store-1",
		Name:        "Vintage Tee",
		Partner:     "Print Partner A",
		BaseCost:    "$8.00",
		RetailPrice: "$20.00",
	}, ProductSetupDraftStatusDraft, time.Date(2026, 6, 4, 10, 30, 0, 0, time.UTC))

	require.NoError(t, err)
	snapshot := draft.Snapshot()
	require.NotEmpty(t, snapshot.ID)
	require.Equal(t, "Vintage Tee", snapshot.Name)
	require.Len(t, events, 1)
	require.Equal(t, "ProductSetupDraftCreated", events[0].EventType())
}

func TestNewProductSetupDraftRequiresStoreAndName(t *testing.T) {
	t.Parallel()

	_, _, err := NewProductSetupDraft(
		"draft-1",
		CreateProductSetupDraftCmd{Name: "Vintage Tee"},
		ProductSetupDraftStatusDraft,
		time.Date(2026, 6, 4, 10, 30, 0, 0, time.UTC),
	)
	require.Error(t, err)

	_, _, err = NewProductSetupDraft(
		"draft-1",
		CreateProductSetupDraftCmd{StoreID: "store-1"},
		ProductSetupDraftStatusDraft,
		time.Date(2026, 6, 4, 10, 30, 0, 0, time.UTC),
	)
	require.Error(t, err)
}

func TestPromoteProductSetupCandidateEmitsEvent(t *testing.T) {
	t.Parallel()

	draft, _, err := NewProductSetupDraft("draft-1", CreateProductSetupDraftCmd{
		StoreID:     "store-1",
		Name:        "Vintage Tee",
		Partner:     "Print Partner A",
		BaseCost:    "$8.00",
		RetailPrice: "$20.00",
	}, ProductSetupDraftStatusDraft, time.Date(2026, 6, 4, 10, 30, 0, 0, time.UTC))
	require.NoError(t, err)

	candidate, events, err := draft.PromoteCandidate("candidate-1", "variant-1", PromoteProductSetupCandidateCmd{
		VariantColor: "Black",
		VariantSize:  "M",
	}, time.Date(2026, 6, 4, 11, 0, 0, 0, time.UTC))

	require.NoError(t, err)
	require.Equal(t, draft.AggregateID().String(), candidate.DraftID)
	require.Equal(t, "$12.00", candidate.EstimatedMargin)
	require.Len(t, events, 1)
	require.Equal(t, "ProductSetupCandidatePromoted", events[0].EventType())
}
