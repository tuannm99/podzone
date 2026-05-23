package iam

import (
	"go.uber.org/fx"

	"github.com/tuannm99/podzone/internal/iam/infrastructure/repository"
	"github.com/tuannm99/podzone/internal/iam/inputport"
	"github.com/tuannm99/podzone/internal/iam/interactor"
	"github.com/tuannm99/podzone/internal/iam/outputport"
	"github.com/tuannm99/podzone/pkg/messaging"
)

var Module = fx.Options(
	fx.Provide(
		fx.Annotate(repository.NewTenantRepository, fx.As(new(outputport.TenantRepository))),
		fx.Annotate(repository.NewRoleRepository, fx.As(new(outputport.RoleRepository))),
		fx.Annotate(repository.NewPolicyRepository, fx.As(new(outputport.PolicyRepository))),
		fx.Annotate(repository.NewGroupRepository, fx.As(new(outputport.GroupRepository))),
		fx.Annotate(repository.NewOrganizationRepository, fx.As(new(outputport.OrganizationRepository))),
		fx.Annotate(repository.NewPlatformMembershipRepository, fx.As(new(outputport.PlatformMembershipRepository))),
		fx.Annotate(repository.NewMembershipRepository, fx.As(new(outputport.MembershipRepository))),
		fx.Annotate(repository.NewInviteRepository, fx.As(new(outputport.InviteRepository))),
		fx.Annotate(repository.NewOutboxRepository, fx.As(new(outputport.OutboxRepository), new(messaging.OutboxStore))),
		fx.Annotate(interactor.NewIAMUsecase, fx.As(new(inputport.IAMUsecase))),
	),
)
