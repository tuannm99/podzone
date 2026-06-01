package iam

import (
	"go.uber.org/fx"

	"github.com/tuannm99/podzone/internal/iam/domain/inputport"
	"github.com/tuannm99/podzone/internal/iam/domain/interactor"
	"github.com/tuannm99/podzone/internal/iam/domain/outputport"
	"github.com/tuannm99/podzone/internal/iam/infrastructure/repository"
	"github.com/tuannm99/podzone/pkg/messaging"
)

var Module = fx.Options(
	RepositoryModule,
	UsecaseModule,
)

var CommandModule = fx.Options(
	CommandRepositoryModule,
	CommandUsecaseModule,
)

var QueryModule = fx.Options(
	QueryRepositoryModule,
	QueryUsecaseModule,
)

var UsecaseModule = fx.Provide(
	fx.Annotate(
		interactor.NewInteractor,
		fx.As(new(inputport.IAMCommandUsecase), new(inputport.IAMQueryUsecase)),
	),
)

var CommandUsecaseModule = fx.Provide(
	fx.Annotate(interactor.NewCommandInteractor, fx.As(new(inputport.IAMCommandUsecase))),
)

var QueryUsecaseModule = fx.Provide(
	fx.Annotate(interactor.NewQueryInteractor, fx.As(new(inputport.IAMQueryUsecase))),
)

var RepositoryModule = fx.Provide(
	tenantRepositoryProvider(new(outputport.TenantCommandRepository), new(outputport.TenantQueryRepository)),
	roleRepositoryProvider(new(outputport.RoleCommandRepository), new(outputport.RoleQueryRepository)),
	policyRepositoryProvider(new(outputport.PolicyCommandRepository), new(outputport.PolicyQueryRepository)),
	groupRepositoryProvider(new(outputport.GroupCommandRepository), new(outputport.GroupQueryRepository)),
	organizationRepositoryProvider(
		new(outputport.OrganizationCommandRepository),
		new(outputport.OrganizationQueryRepository),
	),
	platformMembershipRepositoryProvider(
		new(outputport.PlatformMembershipCommandRepository),
		new(outputport.PlatformMembershipQueryRepository),
	),
	membershipRepositoryProvider(new(outputport.MembershipCommandRepository), new(outputport.MembershipQueryRepository)),
	inviteRepositoryProvider(new(outputport.InviteCommandRepository), new(outputport.InviteQueryRepository)),
	outboxRepositoryProvider(new(outputport.OutboxRepository), new(messaging.OutboxStore)),
)

var CommandRepositoryModule = fx.Provide(
	tenantRepositoryProvider(new(outputport.TenantCommandRepository), new(outputport.TenantQueryRepository)),
	roleRepositoryProvider(new(outputport.RoleCommandRepository), new(outputport.RoleQueryRepository)),
	policyRepositoryProvider(new(outputport.PolicyCommandRepository), new(outputport.PolicyQueryRepository)),
	groupRepositoryProvider(new(outputport.GroupCommandRepository), new(outputport.GroupQueryRepository)),
	organizationRepositoryProvider(
		new(outputport.OrganizationCommandRepository),
		new(outputport.OrganizationQueryRepository),
	),
	platformMembershipRepositoryProvider(
		new(outputport.PlatformMembershipCommandRepository),
		new(outputport.PlatformMembershipQueryRepository),
	),
	membershipRepositoryProvider(new(outputport.MembershipCommandRepository), new(outputport.MembershipQueryRepository)),
	inviteRepositoryProvider(new(outputport.InviteCommandRepository), new(outputport.InviteQueryRepository)),
	outboxRepositoryProvider(new(outputport.OutboxRepository), new(messaging.OutboxStore)),
)

var QueryRepositoryModule = fx.Provide(
	tenantRepositoryProvider(new(outputport.TenantQueryRepository)),
	roleRepositoryProvider(new(outputport.RoleQueryRepository)),
	policyRepositoryProvider(new(outputport.PolicyQueryRepository)),
	groupRepositoryProvider(new(outputport.GroupQueryRepository)),
	organizationRepositoryProvider(new(outputport.OrganizationQueryRepository)),
	platformMembershipRepositoryProvider(new(outputport.PlatformMembershipQueryRepository)),
	membershipRepositoryProvider(new(outputport.MembershipQueryRepository)),
	inviteRepositoryProvider(new(outputport.InviteQueryRepository)),
)

func tenantRepositoryProvider(interfaces ...any) any {
	return fx.Annotate(repository.NewTenantRepository, fx.As(interfaces...))
}

func roleRepositoryProvider(interfaces ...any) any {
	return fx.Annotate(repository.NewRoleRepository, fx.As(interfaces...))
}

func policyRepositoryProvider(interfaces ...any) any {
	return fx.Annotate(repository.NewPolicyRepository, fx.As(interfaces...))
}

func groupRepositoryProvider(interfaces ...any) any {
	return fx.Annotate(repository.NewGroupRepository, fx.As(interfaces...))
}

func organizationRepositoryProvider(interfaces ...any) any {
	return fx.Annotate(repository.NewOrganizationRepository, fx.As(interfaces...))
}

func platformMembershipRepositoryProvider(interfaces ...any) any {
	return fx.Annotate(repository.NewPlatformMembershipRepository, fx.As(interfaces...))
}

func membershipRepositoryProvider(interfaces ...any) any {
	return fx.Annotate(repository.NewMembershipRepository, fx.As(interfaces...))
}

func inviteRepositoryProvider(interfaces ...any) any {
	return fx.Annotate(repository.NewInviteRepository, fx.As(interfaces...))
}

func outboxRepositoryProvider(interfaces ...any) any {
	return fx.Annotate(repository.NewOutboxRepository, fx.As(interfaces...))
}
