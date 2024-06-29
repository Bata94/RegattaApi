package crud

import (
	"errors"
	"strconv"
	"time"

	"github.com/bata94/RegattaApi/internal/db"
	"github.com/bata94/RegattaApi/internal/handlers/api"
	"github.com/bata94/RegattaApi/internal/sqlc"
	"github.com/oklog/ulid/v2"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

func checkPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

func genJWT(u sqlc.User) (string, error) {
	token := jwt.New(jwt.SigningMethodHS256)

	claims := token.Claims.(jwt.MapClaims)
	claims["username"] = u.Username
	claims["user_id"] = u.Ulid
	claims["exp"] = time.Now().Add(time.Hour * 72).Unix()

	// TODO: RM this in Prod
	return token.SignedString([]byte("DO_NOT_USE_IN_PROD"))
}

func hashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	return string(bytes), err
}

func validToken(t *jwt.Token, id string) bool {
	n, err := strconv.Atoi(id)
	if err != nil {
		return false
	}

	claims := t.Claims.(jwt.MapClaims)
	uid := int(claims["user_id"].(float64))

	return uid == n
}

func validUser(ulid ulid.ULID, p string) bool {
	user, err := GetUser(ulid)
	if err != nil {
		return false
	}
	if !checkPasswordHash(p, *user.HashedPassword) {
		return false
	}
	return true
}

type User struct {
	*sqlc.User
	UserGroup *sqlc.UsersGroup
}

type ReturnUserWithJWT struct {
	Ulid      string
	Jwt       string
	Username  string
	UserGroup *sqlc.UsersGroup
}

type ReturnUserMinimal struct {
	Ulid     string
	Username string
}

type ReturnUser struct {
	Ulid      string
	Username  string
	UserGroup *sqlc.UsersGroup
}

func (u *User) ToReturnUser() ReturnUser {
	return ReturnUser{
		Ulid:      u.Ulid,
		Username:  *u.Username,
		UserGroup: u.UserGroup,
	}
}

type LoginParams struct {
	Username string
	Password string
}

func GetAllUsers() ([]*sqlc.User, error) {
	ctx, cancel := getCtxWithTo()
	defer cancel()

	uLs, err := DB.Queries.GetAllUser(ctx)
	if err != nil {
		return nil, err
	}

	if len(uLs) == 0 {
		uLs = []*sqlc.User{}
	}

	return uLs, nil
}

func GetUser(ulid ulid.ULID) (*User, error) {
	ctx, cancel := getCtxWithTo()
	defer cancel()

	u, err := DB.Queries.GetUser(ctx, ulid.String())
	if err != nil {
		if isNoRowError(err) {
			return nil, &api.NOT_FOUND
		}
		return nil, err
	}

	return &User{
		User:      &u.User,
		UserGroup: &u.UsersGroup,
	}, err
}

func GetUserByUsername(name string) (*User, error) {
	ctx, cancel := getCtxWithTo()
	defer cancel()

	ulidStr, err := DB.Queries.GetUserUlidByName(ctx, &name)
	if err != nil {
		return nil, err
	}

	ulid, err := ulid.Parse(ulidStr)
	if err != nil {
		return nil, err
	}

	return GetUser(ulid)
}

type CreateUserParams struct {
	GroupUlid ulid.ULID `json:"group_ulid"`
	Username  string    `json:"username"`
	Password  string    `json:"password"`
}

func CreateUser(uInp CreateUserParams) (*sqlc.User, error) {
	ctx, cancel := getCtxWithTo()
	defer cancel()

	hashedPW, err := hashPassword(uInp.Password)
	if err != nil {
		return nil, err
	}

	uParams := sqlc.CreateUserParams{
		GroupUlid:      uInp.GroupUlid.String(),
		Username:       &uInp.Username,
		HashedPassword: &hashedPW,
	}

	u, err := DB.Queries.CreateUser(ctx, uParams)
	if err != nil {
		return nil, err
	}

	return u, nil
}

func AuthLogin(l LoginParams) (*ReturnUserWithJWT, error) {
	u, err := GetUserByUsername(l.Username)
	if err != nil {
		return nil, err
	}

	tokenStr := ""
	if checkPasswordHash(l.Password, *u.HashedPassword) {
		tokenStr, err = genJWT(*u.User)
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
		Ulid:      u.User.Ulid,
		Jwt:       tokenStr,
		Username:  *u.User.Username,
		UserGroup: u.UserGroup,
	}, nil
}
