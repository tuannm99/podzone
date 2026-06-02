package outputport

import (
	"context"

	settlemententity "github.com/tuannm99/podzone/internal/backoffice/domain/settlement/entity"
)

type SettlementRecordCommandRepository interface {
	SaveSettlementRecord(ctx context.Context, record *settlemententity.SettlementRecord) error
}
