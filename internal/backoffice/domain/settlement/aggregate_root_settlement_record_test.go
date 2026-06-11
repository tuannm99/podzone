package settlement

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestUpdateSettlementRecalculatesMarginAndEmitsEvent(t *testing.T) {
	t.Parallel()

	record, err := RehydrateSettlementRecord(SettlementRecordSnapshot{
		OrderID:         "ord-1",
		Total:           "$40.00",
		FulfillmentCost: "$10.00",
		ShippingCost:    "$4.00",
		IssueCost:       "$3.00",
		Status:          StatusPending,
	})
	require.NoError(t, err)

	systemChange, noteChange, err := record.UpdateSettlement(
		"$12.00",
		"$5.50",
		StatusReconciled,
		"Supplier invoice matched",
		time.Date(2026, 6, 4, 10, 30, 0, 0, time.UTC),
	)

	require.NoError(t, err)
	require.Contains(t, systemChange.Message, "Settlement reconciled")
	require.NotNil(t, noteChange)

	snapshot := record.Snapshot()
	require.Equal(t, "$19.50", snapshot.RealizedMargin)
	require.Equal(t, StatusReconciled, snapshot.Status)

	events := record.PullEvents()
	require.Len(t, events, 1)
	require.Equal(t, "SettlementUpdated", events[0].EventType())
	require.Empty(t, record.PullEvents())
}

func TestUpdateIssueHandlingRequiresActiveIssue(t *testing.T) {
	t.Parallel()

	record, err := RehydrateSettlementRecord(SettlementRecordSnapshot{
		OrderID:         "ord-1",
		Total:           "$40.00",
		FulfillmentCost: "$10.00",
		ShippingCost:    "$5.00",
		IssueCost:       "$0.00",
		Status:          StatusPending,
	})
	require.NoError(t, err)

	_, _, err = record.UpdateIssueHandling(
		"$6.00",
		IssueResolutionReprint,
		"Needs reprint",
		time.Date(2026, 6, 4, 10, 30, 0, 0, time.UTC),
	)

	require.Error(t, err)
	require.Contains(t, err.Error(), "active exception or delivery issue")
}

func TestUpdateIssueHandlingRecalculatesMarginAndEmitsEvent(t *testing.T) {
	t.Parallel()

	record, err := RehydrateSettlementRecord(SettlementRecordSnapshot{
		OrderID:         "ord-1",
		Total:           "$40.00",
		FulfillmentCost: "$10.00",
		ShippingCost:    "$5.00",
		IssueCost:       "$0.00",
		Status:          StatusPending,
		ExceptionType:   "reprint_request",
	})
	require.NoError(t, err)

	systemChange, noteChange, err := record.UpdateIssueHandling(
		"$6.00",
		IssueResolutionReprint,
		"Reprint approved",
		time.Date(2026, 6, 4, 10, 30, 0, 0, time.UTC),
	)

	require.NoError(t, err)
	require.Contains(t, systemChange.Message, "Issue handling")
	require.NotNil(t, noteChange)

	snapshot := record.Snapshot()
	require.Equal(t, "$19.00", snapshot.RealizedMargin)
	require.Equal(t, IssueResolutionReprint, snapshot.IssueResolution)

	events := record.PullEvents()
	require.Len(t, events, 1)
	require.Equal(t, "IssueHandlingUpdated", events[0].EventType())
}

func TestRehydrateSettlementRecordDoesNotEmitEvents(t *testing.T) {
	t.Parallel()

	record, err := RehydrateSettlementRecord(SettlementRecordSnapshot{
		OrderID: "ord-1",
		Status:  StatusPending,
	})

	require.NoError(t, err)
	require.Empty(t, record.PullEvents())
}
