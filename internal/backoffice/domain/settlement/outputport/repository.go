package outputport

import (
	"context"

	settlemententity "github.com/tuannm99/podzone/internal/backoffice/domain/settlement/entity"
)

type SettlementRecordRepository interface {
	GetSettlementRecord(ctx context.Context, storeID string, orderID string) (*settlemententity.SettlementRecord, error)
	SaveSettlementRecord(ctx context.Context, record *settlemententity.SettlementRecord) error
}
