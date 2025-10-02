package domain

import (
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/tuannm99/podzone/internal/auth/domain/entity"
	outputmocks "github.com/tuannm99/podzone/internal/auth/domain/outputport/mocks"
)

func TestUserUsecase_CreateNew_DelegatesToRepo(t *testing.T) {
	repo := &outputmocks.MockUserRepository{}
	uc := NewUserUsecase(repo)

	in := entity.User{Email: "x@x.io", Username: "x"}
	out := &entity.User{Id: 10, Email: "x@x.io", Username: "x"}

	repo.
		On("Create", in).
		Return(out, nil)

	got, err := uc.CreateNew(in)
	require.NoError(t, err)
	require.Equal(t, out, got)

	repo.AssertExpectations(t)
}

func TestUserUsecase_CreateNewAfterAuthCallback_UsesRepoCreateByEmailIfNotExisted(t *testing.T) {
	repo := &outputmocks.MockUserRepository{}
	uc := NewUserUsecase(repo)

	in := entity.User{Email: "neo@mx.io"}
	out := &entity.User{Id: 99, Email: "neo@mx.io"}

	repo.
		On("CreateByEmailIfNotExisted", "neo@mx.io").
		Return(out, nil)

	got, err := uc.CreateNewAfterAuthCallback(in)
	require.NoError(t, err)
	require.Equal(t, out, got)

	repo.AssertExpectations(t)
}

func TestUserUsecase_UpdateOne_DelegatesToRepo(t *testing.T) {
	repo := &outputmocks.MockUserRepository{}
	uc := NewUserUsecase(repo)

	id := uint(7)
	update := entity.User{FullName: "John Doe"}

	repo.
		On("UpdateById", id, mock.MatchedBy(func(u entity.User) bool {
			return u.FullName == "John Doe"
		})).
		Return(nil)

	err := uc.UpdateOne(id, update)
	require.NoError(t, err)

	repo.AssertExpectations(t)
}
