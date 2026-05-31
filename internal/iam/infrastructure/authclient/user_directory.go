package authclient

import (
	"context"
	"fmt"

	iamconfig "github.com/tuannm99/podzone/internal/iam/config"
	"github.com/tuannm99/podzone/internal/iam/domain/entity"
	"github.com/tuannm99/podzone/internal/iam/domain/outputport"
	pbauthv1 "github.com/tuannm99/podzone/pkg/api/proto/auth/v1"
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

func toIAMUser(in *pbauthv1.UserInfo) *entity.User {
	if in == nil {
		return nil
	}
	return &entity.User{
		ID:       uint(in.Id),
		Email:    in.Email,
		Username: in.Username,
	}
}
