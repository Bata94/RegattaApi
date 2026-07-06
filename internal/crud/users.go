package crud

import (
	"encoding/json"
	"errors"
	"os"
	"time"

	"github.com/bata94/RegattaApi/internal/db"
	"github.com/bata94/RegattaApi/internal/handlers/api"
	"github.com/bata94/RegattaApi/internal/sqlc"
	"github.com/google/uuid"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

type User struct {
	sqlc.User
	UserGroup *sqlc.UsersGroup `json:"user_group,omitempty"`
}

func UserFromSqlc(u sqlc.User) User {
	return User{User: u}
}

func (u *User) GetUserGroup() (*sqlc.UsersGroup, error) {
	if u.UserGroup != nil {
		return u.UserGroup, nil
	}
	return nil, nil
}

type userJSON struct {
	Uuid      uuid.UUID        `json:"uuid"`
	Username  string           `json:"username"`
	IsActive  bool             `json:"is_active"`
	UserGroup *sqlc.UsersGroup `json:"user_group,omitempty"`
}

func (u User) MarshalJSON() ([]byte, error) {
	j := userJSON{
		Uuid:      u.Uuid,
		Username:  u.Username,
		IsActive:  u.IsActive,
		UserGroup: u.UserGroup,
	}
	return json.Marshal(j)
}

type JWT struct {
	Token      string    `json:"token"`
	Expiration time.Time `json:"expiration"`
}

type ReturnUserWithJWT struct {
	Uuid      uuid.UUID        `json:"uuid"`
	Jwt       JWT              `json:"jwt"`
	Username  string           `json:"username"`
	UserGroup *sqlc.UsersGroup `json:"user_group"`
}

type ReturnUserMinimal struct {
	Uuid     uuid.UUID `json:"uuid"`
	Username string    `json:"username"`
}

type LoginParams struct {
	Username string
	Password string
}

type CreateUserParams struct {
	GroupUuid uuid.UUID `json:"group_uuid"`
	Username  string    `json:"username"`
	Password  string    `json:"password"`
}

type UpdateUserParams struct {
	Username  string    `json:"username"`
	IsActive  bool      `json:"is_active"`
	GroupUuid uuid.UUID `json:"group_uuid"`
}

func CheckPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

func genJWT(u sqlc.User, ug *sqlc.UsersGroup) (string, time.Time, error) {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		secret = "DO_NOT_USE_IN_PROD"
	}

	token := jwt.New(jwt.SigningMethodHS256)
	exp := time.Now().Add(time.Hour * 72)

	claims := token.Claims.(jwt.MapClaims)
	claims["username"] = u.Username
	claims["user_id"] = u.Uuid.String()
	claims["exp"] = exp.Unix()

	if ug != nil {
		claims["user_group_name"] = ug.Name
		claims["allowed_admin"] = ug.AllowedAdmin
		claims["allowed_zeitnahme"] = ug.AllowedZeitnahme
		claims["allowed_startlisten"] = ug.AllowedStartlisten
		claims["allowed_regattabuero"] = ug.AllowedRegattabuero
		claims["allowed_regattaleitung"] = ug.AllowedRegattaleitung
	}
	claims["allowed_logged_in"] = true

	jwtStr, err := token.SignedString([]byte(secret))
	return jwtStr, exp, err
}

func hashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	return string(bytes), err
}

func ParseUUID(s string) (uuid.UUID, error) {
	return uuid.Parse(s)
}

func validUser(id uuid.UUID, p string) bool {
	user, err := GetUser(id)
	if err != nil {
		return false
	}
	if !CheckPasswordHash(p, user.HashedPassword) {
		return false
	}
	return true
}

func GetAllUsers() ([]sqlc.User, error) {
	ctx, cancel := getCtxWithTo()
	defer cancel()

	uLs, err := DB.Queries.GetAllUser(ctx)
	if err != nil {
		return nil, err
	}

	if len(uLs) == 0 {
		uLs = []sqlc.User{}
	}

	return uLs, nil
}

func GetUser(id uuid.UUID) (*User, error) {
	ctx, cancel := getCtxWithTo()
	defer cancel()

	u, err := DB.Queries.GetUser(ctx, id)
	if err != nil {
		if isNoRowError(err) {
			return nil, &api.NOT_FOUND
		}
		return nil, err
	}

	return &User{
		User:      u.User,
		UserGroup: &u.UsersGroup,
	}, err
}

func GetUserByUsername(name string) (*User, error) {
	ctx, cancel := getCtxWithTo()
	defer cancel()

	id, err := DB.Queries.GetUserUuidByName(ctx, name)
	if err != nil {
		return nil, err
	}

	return GetUser(id)
}

func CreateUser(uInp CreateUserParams) (User, error) {
	ctx, cancel := getCtxWithTo()
	defer cancel()

	hashedPW, err := hashPassword(uInp.Password)
	if err != nil {
		return User{}, err
	}

	uParams := sqlc.CreateUserParams{
		GroupUuid:      uInp.GroupUuid,
		Username:       uInp.Username,
		HashedPassword: hashedPW,
	}

	u, err := DB.Queries.CreateUser(ctx, uParams)
	if err != nil {
		return User{}, err
	}

	return User{User: u}, nil
}

func UpdateUser(u uuid.UUID, uParams UpdateUserParams) error {
	ctx, cancel := getCtxWithTo()
	defer cancel()

	err := DB.Queries.UpdateUser(ctx, sqlc.UpdateUserParams{
		Uuid:      u,
		Username:  uParams.Username,
		IsActive:  uParams.IsActive,
		GroupUuid: uParams.GroupUuid,
	})
	if err != nil {
		return err
	}

	return nil
}

func UpdatePassword(u uuid.UUID, p string) error {
	ctx, cancel := getCtxWithTo()
	defer cancel()

	hp, err := hashPassword(p)
	if err != nil {
		return err
	}

	err = DB.Queries.UpdatePassword(ctx,  sqlc.UpdatePasswordParams{
		Uuid: u,
		HashedPassword: hp,
	})
	if err != nil {
		return err
	}

	return nil
}

func AuthLogin(l LoginParams) (*ReturnUserWithJWT, error) {
	u, err := GetUserByUsername(l.Username)
	if err != nil {
		return nil, err
	}

	if u.IsActive == false {
		return nil, &api.AUTH_LOGIN_USER_NOT_ACTIVE
	}

	tokenStr := ""
	tokenExp := time.Now()
	if CheckPasswordHash(l.Password, u.HashedPassword) {
		tokenStr, tokenExp, err = genJWT(u.User, u.UserGroup)
		if err != nil {
			retErr := &api.TOKEN_GENERATION_ERROR
			retErr.Details = err.Error()
			return nil, retErr
		}
	} else {
		return nil, &api.AUTH_LOGIN_WRONG_PASSWORD
	}

	if tokenStr == "" {
		return nil, errors.New("Unkown Error!")
	}

	return &ReturnUserWithJWT{
		Uuid: u.User.Uuid,
		Jwt: JWT{
			Token:      tokenStr,
			Expiration: tokenExp,
		},
		Username:  u.User.Username,
		UserGroup: u.UserGroup,
	}, nil
}
