package routing

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	catalogctx "github.com/tuannm99/podzone/internal/backoffice/domain/catalog"
)

func TestRoutingDecisionSelectPreferredEmitsEvent(t *testing.T) {
	t.Parallel()

	decision, err := NewRoutingDecision(
		"cand-1",
		"Vintage Tee",
		"Print Partner A",
		"tshirt",
		"us",
		[]RoutingPartnerOption{
			{
				Partner:                  PartnerRoutingProfile{Name: "Fulfill Fast", Code: "fulfill-fast"},
				Eligible:                 true,
				EstimatedUnitMargin:      "$11.00",
				EstimatedShippingCost:    "$2.00",
				EstimatedFulfillmentCost: "$7.00",
			},
		},
	)
	require.NoError(t, err)

	selected := decision.SelectPreferred("fulfill-fast", time.Date(2026, 6, 4, 10, 30, 0, 0, time.UTC))

	require.True(t, selected)
	snapshot := decision.Snapshot()
	require.Equal(t, "Fulfill Fast", snapshot.SelectedPartner)
	require.Empty(t, snapshot.BlockedReasonCode)

	events := decision.PullEvents()
	require.Len(t, events, 1)
	require.Equal(t, "RoutingPartnerSelected", events[0].EventType())
	require.Empty(t, decision.PullEvents())
}

func TestRoutingDecisionBlockEmitsEvent(t *testing.T) {
	t.Parallel()

	decision, err := NewRoutingDecision("cand-1", "Poster", "", "poster", "us", nil)
	require.NoError(t, err)

	decision.Block(
		"negative_margin",
		"all eligible partners have negative expected margin",
		"No auto-route partner selected.",
		time.Date(2026, 6, 4, 10, 30, 0, 0, time.UTC),
	)

	snapshot := decision.Snapshot()
	require.Empty(t, snapshot.SelectedPartner)
	require.Equal(t, "negative_margin", snapshot.BlockedReasonCode)
	require.Equal(t, "all eligible partners have negative expected margin", snapshot.BlockedReason)

	events := decision.PullEvents()
	require.Len(t, events, 1)
	require.Equal(t, "RoutingBlocked", events[0].EventType())
}

func TestBuildRoutingRecommendationUsesDecisionForBlockedResult(t *testing.T) {
	t.Parallel()

	recommendation := BuildRoutingRecommendation(
		&catalogctx.ProductSetupCandidate{
			ID:          "cand-1",
			Title:       "Poster",
			BaseCost:    "$8.00",
			RetailPrice: "$8.00",
		},
		[]PartnerRoutingProfile{
			{
				Name:                  "Fulfill Fast",
				Status:                "active",
				SupportedProductTypes: []string{"poster"},
				SupportedRegions:      []string{"us"},
				BaseFulfillmentCost:   "$7.00",
				ShippingCostRules:     []PartnerShippingCostRule{{Region: "us", Cost: "$2.00"}},
			},
		},
		"poster",
		"us",
		"",
		time.Date(2026, 6, 4, 10, 30, 0, 0, time.UTC),
	)

	require.Empty(t, recommendation.SelectedPartner)
	require.Equal(t, "negative_margin", recommendation.BlockedReasonCode)
	require.NotEmpty(t, recommendation.Options)
}
