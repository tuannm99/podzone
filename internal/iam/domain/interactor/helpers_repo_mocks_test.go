package interactor_test

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/stretchr/testify/mock"

	entity "github.com/tuannm99/podzone/internal/iam/domain/entity"
	outputportmocks "github.com/tuannm99/podzone/internal/iam/domain/outputport/mocks"
	"github.com/tuannm99/podzone/pkg/collection"
)

func configurePolicyRepoMocks(policyRepo *outputportmocks.MockPolicyRepository, state *iamTestState) {
	policyRepo.EXPECT().
		CreatePolicy(mock.Anything, mock.Anything, mock.Anything).
		RunAndReturn(func(ctx context.Context, policy entity.Policy, statements []entity.PolicyStatement) (*entity.Policy, []entity.PolicyStatement, error) {
			policy.ID = state.nextPolicyID
			state.nextPolicyID++
			policy.DefaultVersion = "v1"
			state.policiesByName[policy.Name] = policy
			state.policiesByID[policy.ID] = policy
			outStatements := make([]entity.PolicyStatement, 0, len(statements))
			for i, statement := range statements {
				statement.ID = uint64(i + 1)
				statement.PolicyID = policy.ID
				statement.PolicyName = policy.Name
				outStatements = append(outStatements, statement)
			}
			state.policyStatements[policy.ID] = append([]entity.PolicyStatement(nil), outStatements...)
			state.policyVersions[policy.ID] = []entity.PolicyVersion{{
				ID:         1,
				PolicyID:   policy.ID,
				PolicyName: policy.Name,
				Version:    "v1",
				IsDefault:  true,
				CreatedAt:  policy.CreatedAt,
			}}
			state.policyVersionStatements[fmt.Sprintf("%d:%s", policy.ID, "v1")] = append(
				[]entity.PolicyStatement(nil),
				outStatements...)
			return &policy, outStatements, nil
		}).
		Maybe()
	policyRepo.EXPECT().
		CreatePolicyVersion(mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		RunAndReturn(func(ctx context.Context, policyID uint64, policyName string, statements []entity.PolicyStatement, setAsDefault bool) (*entity.PolicyVersion, []entity.PolicyStatement, error) {
			versions := state.policyVersions[policyID]
			versionLabel := fmt.Sprintf("v%d", len(versions)+1)
			version := entity.PolicyVersion{
				ID:         uint64(len(versions) + 1),
				PolicyID:   policyID,
				PolicyName: policyName,
				Version:    versionLabel,
				IsDefault:  setAsDefault,
				CreatedAt:  time.Now().UTC(),
			}
			if setAsDefault {
				for i := range versions {
					versions[i].IsDefault = false
				}
				policy := state.policiesByID[policyID]
				policy.DefaultVersion = versionLabel
				state.policiesByID[policyID] = policy
				state.policiesByName[policy.Name] = policy
				state.policyStatements[policyID] = append([]entity.PolicyStatement(nil), statements...)
			}
			state.policyVersions[policyID] = append(versions, version)
			state.policyVersionStatements[fmt.Sprintf("%d:%s", policyID, versionLabel)] = append(
				[]entity.PolicyStatement(nil),
				statements...)
			return &version, append([]entity.PolicyStatement(nil), statements...), nil
		}).
		Maybe()
	policyRepo.EXPECT().
		DeletePolicyVersion(mock.Anything, mock.Anything, mock.Anything).
		RunAndReturn(func(ctx context.Context, policyID uint64, version string) error {
			versions := state.policyVersions[policyID]
			nextVersions := make([]entity.PolicyVersion, 0, len(versions))
			found := false
			for _, item := range versions {
				if item.Version != version {
					nextVersions = append(nextVersions, item)
					continue
				}
				if item.IsDefault {
					return entity.ErrDefaultPolicyVersion
				}
				found = true
			}
			if !found {
				return entity.ErrPolicyVersionNotFound
			}
			state.policyVersions[policyID] = nextVersions
			delete(state.policyVersionStatements, fmt.Sprintf("%d:%s", policyID, version))
			return nil
		}).
		Maybe()
	policyRepo.EXPECT().
		GetPolicyByName(mock.Anything, mock.Anything).
		RunAndReturn(func(ctx context.Context, name string) (*entity.Policy, error) {
			policy, ok := state.policiesByName[name]
			if !ok {
				return nil, entity.ErrRoleNotFound
			}
			copyPolicy := policy
			return &copyPolicy, nil
		}).
		Maybe()
	policyRepo.EXPECT().
		GetPolicyStatements(mock.Anything, mock.Anything).
		RunAndReturn(func(ctx context.Context, policyID uint64) ([]entity.PolicyStatement, error) {
			return append([]entity.PolicyStatement(nil), state.policyStatements[policyID]...), nil
		}).
		Maybe()
	policyRepo.EXPECT().
		ListPolicyVersions(mock.Anything, mock.Anything, mock.Anything).
		RunAndReturn(func(ctx context.Context, policyID uint64, policyName string) ([]entity.PolicyVersion, error) {
			return append([]entity.PolicyVersion(nil), state.policyVersions[policyID]...), nil
		}).
		Maybe()
	policyRepo.EXPECT().
		SetDefaultPolicyVersion(mock.Anything, mock.Anything, mock.Anything).
		RunAndReturn(func(ctx context.Context, policyID uint64, version string) error {
			versions := state.policyVersions[policyID]
			for i := range versions {
				versions[i].IsDefault = versions[i].Version == version
			}
			state.policyVersions[policyID] = versions
			policy := state.policiesByID[policyID]
			policy.DefaultVersion = version
			state.policiesByID[policyID] = policy
			state.policiesByName[policy.Name] = policy
			state.policyStatements[policyID] = append(
				[]entity.PolicyStatement(nil),
				state.policyVersionStatements[fmt.Sprintf("%d:%s", policyID, version)]...)
			return nil
		}).
		Maybe()
	policyRepo.EXPECT().
		ListPolicies(mock.Anything, mock.Anything, mock.Anything).
		RunAndReturn(func(
			ctx context.Context,
			scope string,
			query collection.Query,
		) (collection.Page[entity.Policy], error) {
			out := make([]entity.Policy, 0)
			for _, policy := range state.policiesByName {
				if scope == "" || policy.Scope == scope {
					out = append(out, policy)
				}
			}
			return collection.NewPage(out, int64(len(out)), query), nil
		}).
		Maybe()
	policyRepo.EXPECT().
		ListPolicyAttachments(mock.Anything, mock.Anything).
		RunAndReturn(func(ctx context.Context, policyID uint64) ([]entity.PolicyAttachment, error) {
			return append([]entity.PolicyAttachment(nil), state.policyAttachments[policyID]...), nil
		}).
		Maybe()
	policyRepo.EXPECT().
		DeletePolicy(mock.Anything, mock.Anything).
		RunAndReturn(func(ctx context.Context, policyID uint64) error {
			policy, ok := state.policiesByID[policyID]
			if !ok {
				return entity.ErrPolicyNotFound
			}
			if policy.IsSystem {
				return entity.ErrImmutablePolicy
			}
			for _, statements := range state.tenantDirect {
				for _, statement := range statements {
					if statement.PolicyID == policyID {
						return entity.ErrPolicyInUse
					}
				}
			}
			for _, statements := range state.platformDirect {
				for _, statement := range statements {
					if statement.PolicyID == policyID {
						return entity.ErrPolicyInUse
					}
				}
			}
			for _, statements := range state.roleStatements {
				for _, statement := range statements {
					if statement.PolicyID == policyID {
						return entity.ErrPolicyInUse
					}
				}
			}
			delete(state.policiesByName, policy.Name)
			delete(state.policiesByID, policyID)
			delete(state.policyStatements, policyID)
			delete(state.policyAttachments, policyID)
			return nil
		}).
		Maybe()
	policyRepo.EXPECT().
		ListRoleStatements(mock.Anything, mock.Anything).
		RunAndReturn(func(ctx context.Context, roleID uint64) ([]entity.PolicyStatement, error) {
			return append([]entity.PolicyStatement(nil), state.roleStatements[roleID]...), nil
		}).
		Maybe()
	policyRepo.EXPECT().
		ListPlatformUserStatements(mock.Anything, mock.Anything).
		RunAndReturn(func(ctx context.Context, userID uint) ([]entity.PolicyStatement, error) {
			out := append([]entity.PolicyStatement(nil), state.platformDirect[userID]...)
			for _, inline := range state.platformInlinePolicies[userID] {
				out = append(out, inline.Statements...)
			}
			return out, nil
		}).
		Maybe()
	policyRepo.EXPECT().
		ListTenantUserStatements(mock.Anything, mock.Anything, mock.Anything).
		RunAndReturn(func(ctx context.Context, tenantID string, userID uint) ([]entity.PolicyStatement, error) {
			key := membershipKey(tenantID, userID)
			out := append([]entity.PolicyStatement(nil), state.tenantDirect[key]...)
			for _, inline := range state.tenantInlinePolicies[key] {
				out = append(out, inline.Statements...)
			}
			return out, nil
		}).
		Maybe()
	policyRepo.EXPECT().
		ListPlatformGroupStatements(mock.Anything, mock.Anything).
		RunAndReturn(func(ctx context.Context, userID uint) ([]entity.PolicyStatement, error) {
			out := make([]entity.PolicyStatement, 0)
			for _, group := range state.groupsByID {
				if group.Scope != entity.PolicyScopePlatform {
					continue
				}
				if _, ok := state.memberships.items[membershipKey(fmt.Sprintf("group:%d", group.ID), userID)]; !ok {
					continue
				}
				out = append(out, state.tenantDirect[fmt.Sprintf("group-policy:%d", group.ID)]...)
				for _, inline := range state.groupInlinePolicies[group.ID] {
					out = append(out, inline.Statements...)
				}
			}
			return out, nil
		}).
		Maybe()
	policyRepo.EXPECT().
		ListTenantGroupStatements(mock.Anything, mock.Anything, mock.Anything).
		RunAndReturn(func(ctx context.Context, tenantID string, userID uint) ([]entity.PolicyStatement, error) {
			out := make([]entity.PolicyStatement, 0)
			for _, group := range state.groupsByID {
				if group.Scope != entity.PolicyScopeTenant || group.TenantID != tenantID {
					continue
				}
				if _, ok := state.memberships.items[membershipKey(fmt.Sprintf("group:%d", group.ID), userID)]; !ok {
					continue
				}
				out = append(out, state.tenantDirect[fmt.Sprintf("group-policy:%d", group.ID)]...)
				for _, inline := range state.groupInlinePolicies[group.ID] {
					out = append(out, inline.Statements...)
				}
			}
			return out, nil
		}).
		Maybe()
	policyRepo.EXPECT().
		AttachPlatformUserPolicy(mock.Anything, mock.Anything, mock.Anything).
		RunAndReturn(func(ctx context.Context, userID uint, policyID uint64) error {
			policy := state.policiesByID[policyID]
			state.platformDirect[userID] = append(state.platformDirect[userID], entity.PolicyStatement{
				PolicyID:        policy.ID,
				PolicyName:      policy.Name,
				Effect:          entity.PolicyEffectAllow,
				ActionPattern:   "*",
				ResourcePattern: "*",
			})
			state.policyAttachments[policyID] = append(state.policyAttachments[policyID], entity.PolicyAttachment{
				AttachmentType: "platform_user",
				Scope:          entity.PolicyScopePlatform,
				UserID:         userID,
			})
			return nil
		}).
		Maybe()
	policyRepo.EXPECT().DetachPlatformUserPolicy(mock.Anything, mock.Anything, mock.Anything).Return(nil).Maybe()
	policyRepo.EXPECT().
		ListPlatformUserPolicies(mock.Anything, mock.Anything).
		RunAndReturn(func(ctx context.Context, userID uint) ([]entity.Policy, error) {
			out := make([]entity.Policy, 0, len(state.platformDirect[userID]))
			for _, statement := range state.platformDirect[userID] {
				if policy, ok := state.policiesByID[statement.PolicyID]; ok {
					out = append(out, policy)
				}
			}
			return out, nil
		}).
		Maybe()
	policyRepo.EXPECT().
		PutPlatformUserInlinePolicy(mock.Anything, mock.Anything).
		RunAndReturn(func(ctx context.Context, input entity.PutPlatformUserInlinePolicyInput) error {
			if state.platformInlinePolicies[input.UserID] == nil {
				state.platformInlinePolicies[input.UserID] = map[string]entity.UserInlinePolicy{}
			}
			state.platformInlinePolicies[input.UserID][input.Name] = entity.UserInlinePolicy{
				Scope:       entity.PolicyScopePlatform,
				UserID:      input.UserID,
				Name:        input.Name,
				Description: input.Description,
				Statements:  append([]entity.PolicyStatement(nil), input.Statements...),
			}
			return nil
		}).
		Maybe()
	policyRepo.EXPECT().
		GetPlatformUserInlinePolicy(mock.Anything, mock.Anything, mock.Anything).
		RunAndReturn(func(ctx context.Context, userID uint, name string) (*entity.UserInlinePolicy, error) {
			policies := state.platformInlinePolicies[userID]
			policy, ok := policies[name]
			if !ok {
				return nil, entity.ErrPolicyNotFound
			}
			copyPolicy := policy
			return &copyPolicy, nil
		}).
		Maybe()
	policyRepo.EXPECT().
		ListPlatformUserInlinePolicies(mock.Anything, mock.Anything).
		RunAndReturn(func(ctx context.Context, userID uint) ([]entity.UserInlinePolicy, error) {
			policies := state.platformInlinePolicies[userID]
			out := make([]entity.UserInlinePolicy, 0, len(policies))
			for _, policy := range policies {
				out = append(out, policy)
			}
			return out, nil
		}).
		Maybe()
	policyRepo.EXPECT().
		DeletePlatformUserInlinePolicy(mock.Anything, mock.Anything, mock.Anything).
		RunAndReturn(func(ctx context.Context, userID uint, name string) error {
			if state.platformInlinePolicies[userID] == nil {
				return entity.ErrPolicyNotFound
			}
			delete(state.platformInlinePolicies[userID], name)
			return nil
		}).
		Maybe()
	policyRepo.EXPECT().
		PutPlatformUserPermissionBoundary(mock.Anything, mock.Anything, mock.Anything).
		RunAndReturn(func(ctx context.Context, userID uint, policyID uint64) error {
			policy := state.policiesByID[policyID]
			boundary := &entity.PermissionBoundary{
				Scope:      entity.PolicyScopePlatform,
				UserID:     userID,
				PolicyID:   policyID,
				PolicyName: policy.Name,
				CreatedAt:  time.Now().UTC(),
			}
			state.platformBoundary[userID] = boundary
			state.platformBoundaryStmts[userID] = append(
				[]entity.PolicyStatement(nil),
				state.policyStatements[policyID]...)
			return nil
		}).
		Maybe()
	policyRepo.EXPECT().
		GetPlatformUserPermissionBoundary(mock.Anything, mock.Anything).
		RunAndReturn(func(ctx context.Context, userID uint) (*entity.PermissionBoundary, error) {
			item := state.platformBoundary[userID]
			if item == nil {
				return nil, nil
			}
			copyItem := *item
			return &copyItem, nil
		}).
		Maybe()
	policyRepo.EXPECT().
		GetPlatformUserPermissionBoundaryStatements(mock.Anything, mock.Anything).
		RunAndReturn(func(ctx context.Context, userID uint) ([]entity.PolicyStatement, error) {
			return append([]entity.PolicyStatement(nil), state.platformBoundaryStmts[userID]...), nil
		}).
		Maybe()
	policyRepo.EXPECT().
		DeletePlatformUserPermissionBoundary(mock.Anything, mock.Anything).
		RunAndReturn(func(ctx context.Context, userID uint) error {
			delete(state.platformBoundary, userID)
			delete(state.platformBoundaryStmts, userID)
			return nil
		}).
		Maybe()
	policyRepo.EXPECT().
		AttachTenantUserPolicy(mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		RunAndReturn(func(ctx context.Context, tenantID string, userID uint, policyID uint64) error {
			policy := state.policiesByID[policyID]
			key := membershipKey(tenantID, userID)
			state.tenantDirect[key] = append(state.tenantDirect[key], entity.PolicyStatement{
				PolicyID:        policy.ID,
				PolicyName:      policy.Name,
				Effect:          entity.PolicyEffectAllow,
				ActionPattern:   "*",
				ResourcePattern: "*",
			})
			state.policyAttachments[policyID] = append(state.policyAttachments[policyID], entity.PolicyAttachment{
				AttachmentType: "tenant_user",
				Scope:          entity.PolicyScopeTenant,
				TenantID:       tenantID,
				UserID:         userID,
			})
			return nil
		}).
		Maybe()
	policyRepo.EXPECT().
		DetachTenantUserPolicy(mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Return(nil).
		Maybe()
	policyRepo.EXPECT().
		ListTenantUserPolicies(mock.Anything, mock.Anything, mock.Anything).
		RunAndReturn(func(ctx context.Context, tenantID string, userID uint) ([]entity.Policy, error) {
			key := membershipKey(tenantID, userID)
			out := make([]entity.Policy, 0, len(state.tenantDirect[key]))
			for _, statement := range state.tenantDirect[key] {
				if policy, ok := state.policiesByID[statement.PolicyID]; ok {
					out = append(out, policy)
				}
			}
			return out, nil
		}).
		Maybe()
	policyRepo.EXPECT().
		PutTenantUserInlinePolicy(mock.Anything, mock.Anything).
		RunAndReturn(func(ctx context.Context, input entity.PutTenantUserInlinePolicyInput) error {
			key := membershipKey(input.TenantID, input.UserID)
			if state.tenantInlinePolicies[key] == nil {
				state.tenantInlinePolicies[key] = map[string]entity.UserInlinePolicy{}
			}
			state.tenantInlinePolicies[key][input.Name] = entity.UserInlinePolicy{
				Scope:       entity.PolicyScopeTenant,
				TenantID:    input.TenantID,
				UserID:      input.UserID,
				Name:        input.Name,
				Description: input.Description,
				Statements:  append([]entity.PolicyStatement(nil), input.Statements...),
			}
			return nil
		}).
		Maybe()
	policyRepo.EXPECT().
		GetTenantUserInlinePolicy(mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		RunAndReturn(func(ctx context.Context, tenantID string, userID uint, name string) (*entity.UserInlinePolicy, error) {
			policies := state.tenantInlinePolicies[membershipKey(tenantID, userID)]
			policy, ok := policies[name]
			if !ok {
				return nil, entity.ErrPolicyNotFound
			}
			copyPolicy := policy
			return &copyPolicy, nil
		}).
		Maybe()
	policyRepo.EXPECT().
		ListTenantUserInlinePolicies(mock.Anything, mock.Anything, mock.Anything).
		RunAndReturn(func(ctx context.Context, tenantID string, userID uint) ([]entity.UserInlinePolicy, error) {
			policies := state.tenantInlinePolicies[membershipKey(tenantID, userID)]
			out := make([]entity.UserInlinePolicy, 0, len(policies))
			for _, policy := range policies {
				out = append(out, policy)
			}
			return out, nil
		}).
		Maybe()
	policyRepo.EXPECT().
		DeleteTenantUserInlinePolicy(mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		RunAndReturn(func(ctx context.Context, tenantID string, userID uint, name string) error {
			key := membershipKey(tenantID, userID)
			if state.tenantInlinePolicies[key] == nil {
				return entity.ErrPolicyNotFound
			}
			delete(state.tenantInlinePolicies[key], name)
			return nil
		}).
		Maybe()
	policyRepo.EXPECT().
		PutTenantUserPermissionBoundary(mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		RunAndReturn(func(ctx context.Context, tenantID string, userID uint, policyID uint64) error {
			policy := state.policiesByID[policyID]
			key := membershipKey(tenantID, userID)
			boundary := &entity.PermissionBoundary{
				Scope:      entity.PolicyScopeTenant,
				TenantID:   tenantID,
				UserID:     userID,
				PolicyID:   policyID,
				PolicyName: policy.Name,
				CreatedAt:  time.Now().UTC(),
			}
			state.tenantBoundary[key] = boundary
			state.tenantBoundaryStmts[key] = append([]entity.PolicyStatement(nil), state.policyStatements[policyID]...)
			return nil
		}).
		Maybe()
	policyRepo.EXPECT().
		GetTenantUserPermissionBoundary(mock.Anything, mock.Anything, mock.Anything).
		RunAndReturn(func(ctx context.Context, tenantID string, userID uint) (*entity.PermissionBoundary, error) {
			item := state.tenantBoundary[membershipKey(tenantID, userID)]
			if item == nil {
				return nil, nil
			}
			copyItem := *item
			return &copyItem, nil
		}).
		Maybe()
	policyRepo.EXPECT().
		GetTenantUserPermissionBoundaryStatements(mock.Anything, mock.Anything, mock.Anything).
		RunAndReturn(func(ctx context.Context, tenantID string, userID uint) ([]entity.PolicyStatement, error) {
			return append(
				[]entity.PolicyStatement(nil),
				state.tenantBoundaryStmts[membershipKey(tenantID, userID)]...), nil
		}).
		Maybe()
	policyRepo.EXPECT().
		DeleteTenantUserPermissionBoundary(mock.Anything, mock.Anything, mock.Anything).
		RunAndReturn(func(ctx context.Context, tenantID string, userID uint) error {
			key := membershipKey(tenantID, userID)
			delete(state.tenantBoundary, key)
			delete(state.tenantBoundaryStmts, key)
			return nil
		}).
		Maybe()
}

func configureGroupRepoMocks(groupRepo *outputportmocks.MockGroupRepository, state *iamTestState) {
	groupRepo.EXPECT().
		CreateGroup(mock.Anything, mock.Anything).
		RunAndReturn(func(ctx context.Context, group entity.Group) (*entity.Group, error) {
			group.ID = state.nextGroupID
			state.nextGroupID++
			state.groupsByID[group.ID] = group
			copyGroup := group
			return &copyGroup, nil
		}).
		Maybe()
	groupRepo.EXPECT().
		ListGroups(mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		RunAndReturn(func(
			ctx context.Context,
			scope string,
			tenantID string,
			query collection.Query,
		) (collection.Page[entity.Group], error) {
			out := make([]entity.Group, 0)
			for _, group := range state.groupsByID {
				if scope != "" && group.Scope != scope {
					continue
				}
				if tenantID != "" && group.TenantID != tenantID {
					continue
				}
				out = append(out, group)
			}
			return collection.NewPage(out, int64(len(out)), query), nil
		}).
		Maybe()
	groupRepo.EXPECT().
		GetByID(mock.Anything, mock.Anything).
		RunAndReturn(func(ctx context.Context, groupID uint64) (*entity.Group, error) {
			group, ok := state.groupsByID[groupID]
			if !ok {
				return nil, entity.ErrGroupNotFound
			}
			copyGroup := group
			return &copyGroup, nil
		}).
		Maybe()
	groupRepo.EXPECT().
		DeleteGroup(mock.Anything, mock.Anything).
		RunAndReturn(func(ctx context.Context, groupID uint64) error {
			group, ok := state.groupsByID[groupID]
			if !ok {
				return entity.ErrGroupNotFound
			}
			if group.IsSystem {
				return entity.ErrImmutableGroup
			}
			delete(state.groupsByID, groupID)
			delete(state.groupInlinePolicies, groupID)
			for key := range state.memberships.items {
				if strings.HasPrefix(key, fmt.Sprintf("group:%d:", groupID)) {
					delete(state.memberships.items, key)
				}
			}
			delete(state.tenantDirect, fmt.Sprintf("group-policy:%d", groupID))
			return nil
		}).
		Maybe()
	groupRepo.EXPECT().
		PutInlinePolicy(mock.Anything, mock.Anything).
		RunAndReturn(func(ctx context.Context, input entity.PutGroupInlinePolicyInput) error {
			if state.groupInlinePolicies[input.GroupID] == nil {
				state.groupInlinePolicies[input.GroupID] = map[string]entity.GroupInlinePolicy{}
			}
			state.groupInlinePolicies[input.GroupID][input.Name] = entity.GroupInlinePolicy{
				GroupID:     input.GroupID,
				Name:        input.Name,
				Description: input.Description,
				Statements:  append([]entity.PolicyStatement(nil), input.Statements...),
			}
			return nil
		}).
		Maybe()
	groupRepo.EXPECT().
		GetInlinePolicy(mock.Anything, mock.Anything, mock.Anything).
		RunAndReturn(func(ctx context.Context, groupID uint64, name string) (*entity.GroupInlinePolicy, error) {
			policies := state.groupInlinePolicies[groupID]
			policy, ok := policies[name]
			if !ok {
				return nil, entity.ErrPolicyNotFound
			}
			copyPolicy := policy
			return &copyPolicy, nil
		}).
		Maybe()
	groupRepo.EXPECT().
		ListInlinePolicies(mock.Anything, mock.Anything).
		RunAndReturn(func(ctx context.Context, groupID uint64) ([]entity.GroupInlinePolicy, error) {
			policies := state.groupInlinePolicies[groupID]
			out := make([]entity.GroupInlinePolicy, 0, len(policies))
			for _, policy := range policies {
				out = append(out, policy)
			}
			return out, nil
		}).
		Maybe()
	groupRepo.EXPECT().
		DeleteInlinePolicy(mock.Anything, mock.Anything, mock.Anything).
		RunAndReturn(func(ctx context.Context, groupID uint64, name string) error {
			if state.groupInlinePolicies[groupID] == nil {
				return entity.ErrPolicyNotFound
			}
			delete(state.groupInlinePolicies[groupID], name)
			return nil
		}).
		Maybe()
	groupRepo.EXPECT().
		AddMember(mock.Anything, mock.Anything, mock.Anything).
		RunAndReturn(func(ctx context.Context, groupID uint64, userID uint) error {
			state.memberships.items[membershipKey(fmt.Sprintf("group:%d", groupID), userID)] = entity.Membership{
				TenantID: fmt.Sprintf("group:%d", groupID),
				UserID:   userID,
			}
			return nil
		}).
		Maybe()
	groupRepo.EXPECT().
		AttachPolicy(mock.Anything, mock.Anything, mock.Anything).
		RunAndReturn(func(ctx context.Context, groupID uint64, policyID uint64) error {
			group := state.groupsByID[groupID]
			state.tenantDirect[fmt.Sprintf("group-policy:%d", groupID)] = append(
				state.tenantDirect[fmt.Sprintf("group-policy:%d", groupID)],
				state.policyStatements[policyID]...,
			)
			state.policyAttachments[policyID] = append(state.policyAttachments[policyID], entity.PolicyAttachment{
				AttachmentType: "group",
				Scope:          group.Scope,
				TenantID:       group.TenantID,
				GroupID:        group.ID,
				GroupName:      group.Name,
			})
			return nil
		}).
		Maybe()
}

func configurePlatformRepoMocks(platformRepo *outputportmocks.MockPlatformMembershipRepository, state *iamTestState) {
	platformRepo.EXPECT().
		Upsert(mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		RunAndReturn(func(ctx context.Context, userID uint, roleID uint64, status string) error {
			state.platformRoleIDs[userID] = append(state.platformRoleIDs[userID], roleID)
			return nil
		}).
		Maybe()
	platformRepo.EXPECT().
		ListRoleIDsByUser(mock.Anything, mock.Anything).
		RunAndReturn(func(ctx context.Context, userID uint) ([]uint64, error) {
			return append([]uint64(nil), state.platformRoleIDs[userID]...), nil
		}).
		Maybe()
	platformRepo.EXPECT().
		ListByUser(mock.Anything, mock.Anything).
		RunAndReturn(func(ctx context.Context, userID uint) ([]entity.PlatformMembership, error) {
			roleIDs := state.platformRoleIDs[userID]
			out := make([]entity.PlatformMembership, 0, len(roleIDs))
			for _, roleID := range roleIDs {
				out = append(out, entity.PlatformMembership{
					UserID:   userID,
					RoleID:   roleID,
					RoleName: fmt.Sprintf("role-%d", roleID),
					Status:   entity.MembershipStatusActive,
				})
			}
			return out, nil
		}).
		Maybe()
	platformRepo.EXPECT().
		Delete(mock.Anything, mock.Anything, mock.Anything).
		RunAndReturn(func(ctx context.Context, userID uint, roleID uint64) error {
			roleIDs := state.platformRoleIDs[userID]
			next := make([]uint64, 0, len(roleIDs))
			found := false
			for _, id := range roleIDs {
				if id == roleID {
					found = true
					continue
				}
				next = append(next, id)
			}
			if !found {
				return entity.ErrMembershipNotFound
			}
			state.platformRoleIDs[userID] = next
			return nil
		}).
		Maybe()
}

func configureMembershipRepoMocks(membershipRepo *outputportmocks.MockMembershipRepository, state *iamTestState) {
	membershipRepo.EXPECT().
		Upsert(mock.Anything, mock.Anything).
		RunAndReturn(func(ctx context.Context, membership entity.Membership) error {
			state.memberships.items[membershipKey(membership.TenantID, membership.UserID)] = membership
			return nil
		}).
		Maybe()
	membershipRepo.EXPECT().
		GetByTenantAndUser(mock.Anything, mock.Anything, mock.Anything).
		RunAndReturn(func(ctx context.Context, tenantID string, userID uint) (*entity.Membership, error) {
			return state.memberships.GetByTenantAndUser(ctx, tenantID, userID)
		}).
		Maybe()
	membershipRepo.EXPECT().
		ListByTenant(mock.Anything, mock.Anything).
		RunAndReturn(func(ctx context.Context, tenantID string) ([]entity.Membership, error) {
			out := make([]entity.Membership, 0)
			for _, item := range state.memberships.items {
				if item.TenantID == tenantID {
					out = append(out, item)
				}
			}
			return out, nil
		}).
		Maybe()
	membershipRepo.EXPECT().
		ListByUser(mock.Anything, mock.Anything).
		RunAndReturn(func(ctx context.Context, userID uint) ([]entity.Membership, error) {
			out := make([]entity.Membership, 0)
			for _, item := range state.memberships.items {
				if item.UserID == userID {
					out = append(out, item)
				}
			}
			return out, nil
		}).
		Maybe()
	membershipRepo.EXPECT().
		Delete(mock.Anything, mock.Anything, mock.Anything).
		RunAndReturn(func(ctx context.Context, tenantID string, userID uint) error {
			key := membershipKey(tenantID, userID)
			if _, ok := state.memberships.items[key]; !ok {
				return entity.ErrMembershipNotFound
			}
			delete(state.memberships.items, key)
			return nil
		}).
		Maybe()
}

func configureInviteRepoMocks(inviteRepo *outputportmocks.MockInviteRepository, state *iamTestState) {
	inviteRepo.EXPECT().
		Create(mock.Anything, mock.Anything).
		RunAndReturn(func(ctx context.Context, invite entity.TenantInvite) error {
			state.invites.items[invite.ID] = invite
			state.invites.tokenIndex[invite.TokenHash] = invite.ID
			return nil
		}).
		Maybe()
	inviteRepo.EXPECT().
		GetByID(mock.Anything, mock.Anything).
		RunAndReturn(func(ctx context.Context, inviteID string) (*entity.TenantInvite, error) {
			return state.invites.GetByID(ctx, inviteID)
		}).
		Maybe()
	inviteRepo.EXPECT().
		GetByTokenHash(mock.Anything, mock.Anything).
		RunAndReturn(func(ctx context.Context, tokenHash string) (*entity.TenantInvite, error) {
			inviteID, ok := state.invites.tokenIndex[tokenHash]
			if !ok {
				return nil, entity.ErrInviteNotFound
			}
			return state.invites.GetByID(ctx, inviteID)
		}).
		Maybe()
	inviteRepo.EXPECT().
		ListByTenant(mock.Anything, mock.Anything).
		RunAndReturn(func(ctx context.Context, tenantID string) ([]entity.TenantInvite, error) {
			out := make([]entity.TenantInvite, 0)
			for _, item := range state.invites.items {
				if item.TenantID == tenantID {
					out = append(out, item)
				}
			}
			return out, nil
		}).
		Maybe()
	inviteRepo.EXPECT().
		MarkAccepted(mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		RunAndReturn(func(ctx context.Context, inviteID string, acceptedByUserID uint, acceptedAt time.Time) error {
			item, ok := state.invites.items[inviteID]
			if !ok {
				return entity.ErrInviteNotFound
			}
			item.Status = entity.InviteStatusAccepted
			item.AcceptedByUserID = &acceptedByUserID
			item.AcceptedAt = &acceptedAt
			item.UpdatedAt = acceptedAt
			state.invites.items[inviteID] = item
			return nil
		}).
		Maybe()
	inviteRepo.EXPECT().
		MarkRevoked(mock.Anything, mock.Anything, mock.Anything).
		RunAndReturn(func(ctx context.Context, inviteID string, revokedAt time.Time) error {
			item, ok := state.invites.items[inviteID]
			if !ok {
				return entity.ErrInviteNotFound
			}
			item.Status = entity.InviteStatusRevoked
			item.RevokedAt = &revokedAt
			item.UpdatedAt = revokedAt
			state.invites.items[inviteID] = item
			return nil
		}).
		Maybe()
}
