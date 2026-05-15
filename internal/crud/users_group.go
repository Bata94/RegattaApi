package crud

import (
	"github.com/bata94/RegattaApi/internal/db"
	"github.com/bata94/RegattaApi/internal/handlers/api"
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

func GetUsersGroupsMinimal(id uuid.UUID) (sqlc.UsersGroup, error) {
	ctx, cancel := getCtxWithTo()
	defer cancel()

	ug, err := DB.Queries.GetUserGroupMinimal(ctx, id)
	if err != nil {
		if isNoRowError(err) {
			return sqlc.UsersGroup{}, &api.NOT_FOUND
		}
		return sqlc.UsersGroup{}, err
	}

	return ug, nil
}

func UGwUsersFromSQLC(q []sqlc.GetUserGroupRow, id uuid.UUID) (UsersGroupWithUsers, error) {
	users := []ReturnUserMinimal{}
	var (
		ug  sqlc.UsersGroup
		err error
	)
	if len(q) == 0 {
		ug, err = GetUsersGroupsMinimal(id)
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

func GetUsersGroup(id uuid.UUID) (UsersGroupWithUsers, error) {
	ctx, cancel := getCtxWithTo()
	defer cancel()

	q, err := DB.Queries.GetUserGroup(ctx, id)
	if err != nil {
		if isNoRowError(err) {
			return UsersGroupWithUsers{}, &api.NOT_FOUND
		}
		return UsersGroupWithUsers{}, err
	}

	return UGwUsersFromSQLC(q, id)
}

func GetUsersGroupByName(name string) (UsersGroupWithUsers, error) {
	ctx, cancel := getCtxWithTo()
	defer cancel()

	id, err := DB.Queries.GetUserGroupUuidByName(ctx, name)
	if err != nil {
		return UsersGroupWithUsers{}, err
	}

	return GetUsersGroup(id)
}

func CreateUserGroup(ugParams sqlc.CreateUserGroupParams) (sqlc.UsersGroup, error) {
	ctx, cancel := getCtxWithTo()
	defer cancel()

	ug, err := DB.Queries.CreateUserGroup(ctx, ugParams)
	if err != nil {
		return sqlc.UsersGroup{}, err
	}

	return ug, nil
}

func UpdateUserGroup(uuid uuid.UUID, uParams sqlc.UpdateUserGroupParams) error {
	ctx, cancel := getCtxWithTo()
	defer cancel()

	err := DB.Queries.UpdateUserGroup(ctx, sqlc.UpdateUserGroupParams{
		Uuid: uuid,
		Name: uParams.Name,
		AllowedAdmin: uParams.AllowedAdmin,
		AllowedZeitnahme: uParams.AllowedZeitnahme,
		AllowedStartlisten: uParams.AllowedStartlisten,
		AllowedRegattabuero: uParams.AllowedRegattabuero,
		AllowedRegattaleitung: uParams.AllowedRegattaleitung,
	})
	if err != nil {
		return err
	}

	return nil
}
