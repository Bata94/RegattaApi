package crud

import (
	"context"
	"github.com/bata94/RegattaApi/internal/db"
	apierr "github.com/bata94/RegattaApi/internal/errors"
	"github.com/bata94/RegattaApi/internal/sqlc"
	"github.com/google/uuid"
)

type UsersGroup struct {
	sqlc.UsersGroup
}

type UsersGroupWithUsers struct {
	sqlc.UsersGroup
	Users []ReturnUserMinimal
}

func GetAllUsersGroups(ctx context.Context) ([]sqlc.UsersGroup, error) {
	ctx, cancel := getCtx(ctx)
	defer cancel()

	ugLs, err := DB.QueriesFromCtx(ctx).GetAllUserGroup(ctx)
	if err != nil {
		return nil, err
	}

	if len(ugLs) == 0 {
		ugLs = []sqlc.UsersGroup{}
	}

	return ugLs, nil
}

func GetUsersGroupsMinimal(ctx context.Context, id uuid.UUID) (sqlc.UsersGroup, error) {
	ctx, cancel := getCtx(ctx)
	defer cancel()

	ug, err := DB.QueriesFromCtx(ctx).GetUserGroupMinimal(ctx, id)
	if err != nil {
		if isNoRowError(err) {
			return sqlc.UsersGroup{}, apierr.ErrNotFound
		}
		return sqlc.UsersGroup{}, err
	}

	return ug, nil
}

func UGwUsersFromSQLC(ctx context.Context, q []sqlc.GetUserGroupRow, id uuid.UUID) (UsersGroupWithUsers, error) {
	users := []ReturnUserMinimal{}
	var (
		ug  sqlc.UsersGroup
		err error
	)
	if len(q) == 0 {
		ug, err = GetUsersGroupsMinimal(ctx, id)
		if err != nil {
			return UsersGroupWithUsers{}, err
		}
	} else {
		ug = q[0].UsersGroup
		for _, u := range q {
			users = append(users, ReturnUserMinimal{
				Uuid:     u.User.Uuid,
				Username: u.User.Username,
			})
		}
	}

	return UsersGroupWithUsers{
		UsersGroup: ug,
		Users:      users,
	}, nil
}

func GetUsersGroup(ctx context.Context, id uuid.UUID) (UsersGroupWithUsers, error) {
	ctx, cancel := getCtx(ctx)
	defer cancel()

	q, err := DB.QueriesFromCtx(ctx).GetUserGroup(ctx, id)
	if err != nil {
		if isNoRowError(err) {
			return UsersGroupWithUsers{}, apierr.ErrNotFound
		}
		return UsersGroupWithUsers{}, err
	}

	return UGwUsersFromSQLC(ctx, q, id)
}

func GetUsersGroupByName(ctx context.Context, name string) (UsersGroupWithUsers, error) {
	ctx, cancel := getCtx(ctx)
	defer cancel()

	id, err := DB.QueriesFromCtx(ctx).GetUserGroupUuidByName(ctx, name)
	if err != nil {
		return UsersGroupWithUsers{}, err
	}

	return GetUsersGroup(ctx, id)
}

func CreateUserGroup(ctx context.Context, ugParams sqlc.CreateUserGroupParams) (sqlc.UsersGroup, error) {
	ctx, cancel := getCtx(ctx)
	defer cancel()

	ug, err := DB.QueriesFromCtx(ctx).CreateUserGroup(ctx, ugParams)
	if err != nil {
		return sqlc.UsersGroup{}, err
	}

	return ug, nil
}

func UpdateUserGroup(ctx context.Context, uuid uuid.UUID, uParams sqlc.UpdateUserGroupParams) error {
	ctx, cancel := getCtx(ctx)
	defer cancel()

	err := DB.QueriesFromCtx(ctx).UpdateUserGroup(ctx, sqlc.UpdateUserGroupParams{
		Uuid:         uuid,
		Name:         uParams.Name,
		Capabilities: uParams.Capabilities,
	})
	if err != nil {
		return err
	}

	return nil
}
