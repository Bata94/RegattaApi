package crud

import (
	"github.com/bata94/RegattaApi/internal/db"
	"github.com/bata94/RegattaApi/internal/handlers/api"
	"github.com/bata94/RegattaApi/internal/sqlc"
	"github.com/oklog/ulid/v2"
)

type UsersGroupWithUsers struct {
	sqlc.UsersGroup
	Users []ReturnUserMinimal
}

func GetAllUsersGroups() ([]sqlc.UsersGroup, error) {
	ctx, cancel := getCtxWithTo()
	defer cancel()

	ugLs, err := DB.Queries.GetAllUserGroup(ctx)
	if err != nil {
		return nil, err
	}

	if len(ugLs) == 0 {
		ugLs = []sqlc.UsersGroup{}
	}

	return ugLs, nil
}

func GetUsersGroupsMinimal(ulid ulid.ULID) (sqlc.UsersGroup, error) {
	ctx, cancel := getCtxWithTo()
	defer cancel()

	ug, err := DB.Queries.GetUserGroupMinimal(ctx, ulid.String())
	if err != nil {
		if isNoRowError(err) {
			return sqlc.UsersGroup{}, &api.NOT_FOUND
		}
		return sqlc.UsersGroup{}, err
	}

	return ug, nil
}

func UGwUsersFromSQLC(q []sqlc.GetUserGroupRow, ulid ulid.ULID) (UsersGroupWithUsers, error) {
	users := []ReturnUserMinimal{}
	var (
		ug  sqlc.UsersGroup
		err error
	)
	if len(q) == 0 {
		ug, err = GetUsersGroupsMinimal(ulid)
		if err != nil {
			return UsersGroupWithUsers{}, err
		}
	} else {
		ug = q[0].UsersGroup
		for _, u := range q {
			users = append(users, ReturnUserMinimal{
				Ulid:     u.User.Ulid,
				Username: u.User.Username,
			})
		}
	}

	return UsersGroupWithUsers{
		UsersGroup: ug,
		Users:      users,
	}, nil
}

func GetUsersGroup(ulid ulid.ULID) (UsersGroupWithUsers, error) {
	ctx, cancel := getCtxWithTo()
	defer cancel()

	q, err := DB.Queries.GetUserGroup(ctx, ulid.String())
	if err != nil {
		if isNoRowError(err) {
			return UsersGroupWithUsers{}, &api.NOT_FOUND
		}
		return UsersGroupWithUsers{}, err
	}

	return UGwUsersFromSQLC(q, ulid)
}

func GetUsersGroupByName(name string) (UsersGroupWithUsers, error) {
	ctx, cancel := getCtxWithTo()
	defer cancel()

	ulidStr, err := DB.Queries.GetUserGroupUlidByName(ctx, name)
	if err != nil {
		return UsersGroupWithUsers{}, err
	}

	ulid, err := ulid.Parse(ulidStr)
	if err != nil {
		return UsersGroupWithUsers{}, err
	}

	return GetUsersGroup(ulid)
}
