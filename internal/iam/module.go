package iam

import (
	"go.uber.org/fx"

	iamdomain "github.com/tuannm99/podzone/internal/iam/domain"
	"github.com/tuannm99/podzone/internal/iam/infrastructure/repository"
)

var Module = fx.Options(
	fx.Provide(
		fx.Annotate(repository.NewTenantRepository, fx.As(new(iamdomain.TenantRepository))),
		fx.Annotate(repository.NewRoleRepository, fx.As(new(iamdomain.RoleRepository))),
		fx.Annotate(repository.NewMembershipRepository, fx.As(new(iamdomain.MembershipRepository))),
		fx.Annotate(iamdomain.NewIAMUsecase, fx.As(new(iamdomain.IAMUsecase))),
	),
)
