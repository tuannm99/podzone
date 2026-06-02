package outputport

import (
	"context"

	settlemententity "github.com/tuannm99/podzone/internal/backoffice/domain/settlement/entity"
)

type SettlementRecordQueryRepository interface {
	GetSettlementRecord(ctx context.Context, storeID string, orderID string) (*settlemententity.SettlementRecord, error)
}
