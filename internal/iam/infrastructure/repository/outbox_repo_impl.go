package repository

import (
	"github.com/jmoiron/sqlx"
	"github.com/tuannm99/podzone/internal/iam/domain/outputport"
	"github.com/tuannm99/podzone/pkg/messaging"
	messagingsqlstore "github.com/tuannm99/podzone/pkg/messaging/sqlstore"
	"go.uber.org/fx"
)

const iamOutboxTableName = "message_outbox"

type outboxRepoParams struct {
	fx.In
	DB *sqlx.DB `name:"sql-iam"`
}

func NewOutboxRepository(p outboxRepoParams) (*messagingsqlstore.OutboxStore, error) {
	return messagingsqlstore.NewOutboxStore(p.DB, iamOutboxTableName)
}

var (
	_ outputport.OutboxRepository = (*messagingsqlstore.OutboxStore)(nil)
	_ messaging.OutboxStore       = (*messagingsqlstore.OutboxStore)(nil)
)
