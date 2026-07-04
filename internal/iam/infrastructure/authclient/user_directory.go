package authclient

import (
	"context"
	"fmt"

	iamconfig "github.com/tuannm99/podzone/internal/iam/config"
	"github.com/tuannm99/podzone/internal/iam/domain/entity"
	"github.com/tuannm99/podzone/internal/iam/domain/outputport"
	pbauthv1 "github.com/tuannm99/podzone/pkg/api/proto/auth/v1"
	pbcommonv1 "github.com/tuannm99/podzone/pkg/api/proto/common/v1"
	"github.com/tuannm99/podzone/pkg/collection"
	"github.com/tuannm99/podzone/pkg/pdlog"
	"go.uber.org/fx"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
)

type UserDirectory struct {
	client pbauthv1.AuthServiceClient
}

var _ outputport.UserDirectory = (*UserDirectory)(nil)

type UserDirectoryParams struct {
	fx.In
	Lifecycle fx.Lifecycle
	Logger    pdlog.Logger
	Config    iamconfig.ServerConfig
}

func NewUserDirectory(p UserDirectoryParams) (outputport.UserDirectory, error) {
	addr := fmt.Sprintf("%s:%s", p.Config.Auth.GRPCHost, p.Config.Auth.GRPCPort)
	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}
	p.Lifecycle.Append(fx.Hook{
		OnStop: func(ctx context.Context) error {
			p.Logger.Info("Closing Auth user-directory gRPC client connection")
			return conn.Close()
		},
	})
	return &UserDirectory{client: pbauthv1.NewAuthServiceClient(conn)}, nil
}

func (d *UserDirectory) GetByIdentity(ctx context.Context, identity string) (*entity.User, error) {
	resp, err := d.client.GetUserByIdentity(ctx, &pbauthv1.GetUserByIdentityRequest{Identity: identity})
	if err != nil {
		if status.Code(err) == codes.NotFound {
			return nil, entity.ErrUserNotFound
		}
		return nil, err
	}
	return toIAMUser(resp.GetUserInfo()), nil
}

func (d *UserDirectory) EnsureByEmail(ctx context.Context, email string) (*entity.User, bool, error) {
	resp, err := d.client.EnsureUserByEmail(ctx, &pbauthv1.EnsureUserByEmailRequest{Email: email})
	if err != nil {
		return nil, false, err
	}
	return toIAMUser(resp.GetUserInfo()), resp.GetCreated(), nil
}

func (d *UserDirectory) GetByID(ctx context.Context, userID uint) (*entity.User, error) {
	resp, err := d.client.GetUserByID(ctx, &pbauthv1.GetUserByIDRequest{UserId: uint64(userID)})
	if err != nil {
		if status.Code(err) == codes.NotFound {
			return nil, entity.ErrUserNotFound
		}
		return nil, err
	}
	return toIAMUser(resp.GetUserInfo()), nil
}

func (d *UserDirectory) List(
	ctx context.Context,
	query collection.Query,
) (collection.Page[entity.User], error) {
	resp, err := d.client.ListUsers(ctx, &pbauthv1.ListUsersRequest{
		Collection: toPBCollectionRequest(query),
	})
	if err != nil {
		return collection.Page[entity.User]{}, err
	}
	users := make([]entity.User, 0, len(resp.GetUsers()))
	for _, user := range resp.GetUsers() {
		if mapped := toIAMUser(user); mapped != nil {
			users = append(users, *mapped)
		}
	}
	pageInfo := resp.GetPageInfo()
	return collection.Page[entity.User]{
		Items:       users,
		Total:       pageInfo.GetTotal(),
		Page:        int(pageInfo.GetPage()),
		PageSize:    int(pageInfo.GetPageSize()),
		TotalPages:  int(pageInfo.GetTotalPages()),
		HasNext:     pageInfo.GetHasNext(),
		HasPrevious: pageInfo.GetHasPrevious(),
	}, nil
}

func toIAMUser(in *pbauthv1.UserInfo) *entity.User {
	if in == nil {
		return nil
	}
	return &entity.User{
		ID:          uint(in.Id),
		Email:       in.Email,
		Username:    in.Username,
		DisplayName: in.FullName,
	}
}

func toPBCollectionRequest(query collection.Query) *pbcommonv1.CollectionRequest {
	normalized := query.Normalize()
	filters := make([]*pbcommonv1.CollectionFilter, 0, len(normalized.Filters))
	for _, filter := range normalized.Filters {
		filters = append(filters, &pbcommonv1.CollectionFilter{
			Field:    filter.Field,
			Operator: toPBFilterOperator(filter.Operator),
			Values:   append([]string(nil), filter.Values...),
		})
	}
	return &pbcommonv1.CollectionRequest{
		Page:          int32(normalized.Page),
		PageSize:      int32(normalized.PageSize),
		Search:        normalized.Search,
		Filters:       filters,
		SortBy:        normalized.SortBy,
		SortDirection: toPBSortDirection(normalized.SortDirection),
	}
}

func toPBSortDirection(direction collection.SortDirection) pbcommonv1.SortDirection {
	if direction == collection.SortAscending {
		return pbcommonv1.SortDirection_SORT_DIRECTION_ASC
	}
	return pbcommonv1.SortDirection_SORT_DIRECTION_DESC
}

func toPBFilterOperator(operator collection.FilterOperator) pbcommonv1.FilterOperator {
	switch operator {
	case collection.FilterEqual:
		return pbcommonv1.FilterOperator_FILTER_OPERATOR_EQ
	case collection.FilterNotEqual:
		return pbcommonv1.FilterOperator_FILTER_OPERATOR_NEQ
	case collection.FilterContains:
		return pbcommonv1.FilterOperator_FILTER_OPERATOR_CONTAINS
	case collection.FilterStartsWith:
		return pbcommonv1.FilterOperator_FILTER_OPERATOR_STARTS_WITH
	case collection.FilterGreaterThan:
		return pbcommonv1.FilterOperator_FILTER_OPERATOR_GT
	case collection.FilterGreaterThanOrEqual:
		return pbcommonv1.FilterOperator_FILTER_OPERATOR_GTE
	case collection.FilterLessThan:
		return pbcommonv1.FilterOperator_FILTER_OPERATOR_LT
	case collection.FilterLessThanOrEqual:
		return pbcommonv1.FilterOperator_FILTER_OPERATOR_LTE
	case collection.FilterIn:
		return pbcommonv1.FilterOperator_FILTER_OPERATOR_IN
	default:
		return pbcommonv1.FilterOperator_FILTER_OPERATOR_UNSPECIFIED
	}
}
