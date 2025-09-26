package userService

import (
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"math/rand"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/RangoCoder/foodApi/internal/appmidleware"
	"github.com/RangoCoder/foodApi/internal/env"
	st "github.com/RangoCoder/foodApi/internal/structs"
)

type UserService interface {
	RegisterUser(user st.User) (st.User, error)
	LoginUser(user st.User) (st.AuthTokens, error)
	RestoreAccessByRefresh(uid uint64, tokens st.AuthTokens) (st.AuthTokens, error)
	GetAllUsers(offset, count int) ([]st.UserShow, error)
	GetUserById(uid uint64) (st.User, error)
	UpdateUserParams(userParams st.ParamsUser) (st.ParamsUser, error)
	DeleteUser(uid uint64) error
	ExitUser(uid uint64) (bool, error)
}

type userService struct {
	repo UserRepository
}

func NewUserService(repo UserRepository) UserService {
	return &userService{repo: repo}
}

func (s *userService) ValidateEmail(email string) bool { // проверка валидности email
	emailRegex := regexp.MustCompile(`^[a-z0-9._%+\-]+@[a-z0-9.\-]+\.[a-z]{2,4}$`)
	// result := emailRegex.MatchString(email)
	// fmt.Printf("результат проверки емейла %v = %v ", email, result)
	return emailRegex.MatchString(email)
}

// хэшируем соленый пароль
func (s *userService) ConvertToSha256(incomeStr string) string {
	soult := env.GetEnvVar("USER_PASS_SOULT") // строка соли генерируем наобум для защиты в случае утечки
	soult2 := env.GetEnvVar("USER_PASS_SOULT2")
	soultString := fmt.Sprintf("%v%v%v", soult2, incomeStr, soult)
	bv := []byte(soultString)
	hasher := sha256.New()
	hasher.Write(bv)
	sha := base64.URLEncoding.EncodeToString(hasher.Sum(nil))
	return sha
}

// создаем RefreshToken
func (s *userService) MakeRefreshToken(claims st.TokenClaims, uid uint64) (string, error) {
	var authUser st.AuthUser
	charset := env.GetEnvVar("SIMBOLS_CHAR")
	sb := strings.Builder{}
	n := 32 // длина refrehToken
	sb.Grow(n)
	for i := 0; i < n; i++ {
		sb.WriteByte(charset[rand.Intn(len(charset))])
	}

	authUser, err := s.repo.GetUserAuth(uid)
	if err != nil { // если нет пользователя с таким id
		return "", err
		// return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid email"})
	}

	//готовим запись нового refresfToken в бд
	authUser.ID = uid
	RefreshLiveDaysVar := env.GetEnvVar("REFRESH_LIVE_DAYS") // время жизни refresh токена в днях (RefreshLiveDays)
	RefreshLiveDays, err := strconv.Atoi(RefreshLiveDaysVar)
	if err != nil { //
		return "", err
	}
	expireRtTime := time.Now().AddDate(0, 0, RefreshLiveDays).Unix() // uinx выдачи refresh токена
	authUser.ExpareTime = expireRtTime                               // renew expiration time
	authUser.LastRefreshTime = time.Now().Unix()                     // unix token create
	authUser.RefreshToken = sb.String()                              // renew refreshToken

	saved, err := s.repo.UpdateUserAuth(authUser)
	if err != nil {
		return "", err
		// return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Could not create user"})
	}

	return saved.RefreshToken, nil
}

/*
renew user tokens - AccessToken RefreshToken
*/
func (s *userService) RestoreAccessByRefresh(uid uint64, tokens st.AuthTokens) (st.AuthTokens, error) {
	var authResp st.AuthTokens
	var tokenClaims st.TokenClaims
	// confirm uid & refresh -> get LastEntryRoot as root
	authUser, err := s.repo.GetUserAuth(uid)
	if err != nil { // если нет пользователя с таким id
		return st.AuthTokens{}, err
	}
	// make new jwt with root param geted from last user login
	nowTimeUnix := time.Now().Unix()
	tokenClaims.Sub = uid
	tokenClaims.Iat = uint64(nowTimeUnix)
	tokenClaims.Root = uint64(authUser.LastEntryRoot)
	accessToken, err := appmidleware.MakeAccessJWT(tokenClaims)
	if err != nil {
		return authResp, err
	}
	authResp.AccessToken = accessToken

	refreshToken, err := s.MakeRefreshToken(tokenClaims, uid)
	if err != nil {
		return authResp, err
	}
	authResp.RefreshToken = refreshToken

	return authResp, nil
}

/*
make tokens - AccessToken RefreshToken
RefreshToken save in table with uid, uRoot, expireTime
*/
func (s *userService) GenerateAndSaveJWT(user st.User) (st.AuthTokens, error) {
	var authResp st.AuthTokens
	var tokenClaims st.TokenClaims
	// fmt.Println("GenerateAndSaveJWT user", user)
	// generate jwt
	nowTimeUnix := time.Now().Unix()
	tokenClaims.Sub = user.ID
	tokenClaims.Iat = uint64(nowTimeUnix)
	tokenClaims.Root = uint64(user.ParamsUser.Root)
	accessToken, err := appmidleware.MakeAccessJWT(tokenClaims)
	if err != nil {
		return authResp, err
	}
	authResp.AccessToken = accessToken

	refreshToken, err := s.MakeRefreshToken(tokenClaims, user.ID)
	if err != nil {
		return authResp, err
	}
	authResp.RefreshToken = refreshToken

	return authResp, nil
}

// make line in table users - save login, pass
func (s *userService) RegisterUser(user st.User) (st.User, error) {
	var err error
	if !s.ValidateEmail(user.Email) { // if email not valid
		return st.User{}, errors.New("invalid email")
	}

	newuser := st.User{
		Email:    user.Email,
		Password: s.ConvertToSha256(user.Password),
	}

	resp, err := s.repo.RegisterUser(newuser)
	if err != nil {
		return st.User{}, err
		// return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Could not create user"})
	}
	return resp, err
}

// аутентификация пользователя по логину
func (s *userService) LoginUser(user st.User) (st.AuthTokens, error) {
	var err error
	if !s.ValidateEmail(user.Email) { // если указанная почта не валидна
		return st.AuthTokens{}, errors.New("invalid email")
	}

	loguser := st.User{
		Email:    user.Email,
		Password: s.ConvertToSha256(user.Password),
	}

	respuser, err := s.repo.LoginUser(loguser)

	if err != nil {
		return st.AuthTokens{}, err
		// return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Could not create user"})
	}

	auth, err := s.GenerateAndSaveJWT(respuser) // генерим параметры авторизации, сохраняем в базу и отдаем токены

	if err != nil {
		return st.AuthTokens{}, err
		// return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Could not create user"})
	}

	return auth, err
}

// выборка всех пользователей из бд
func (s *userService) GetAllUsers(offset, count int) ([]st.UserShow, error) {
	return s.repo.GetAllUsers(offset, count)
}

// выборка пользователя из бд по ид
func (s *userService) GetUserById(uid uint64) (st.User, error) {
	user, err := s.repo.GetUserById(uid)
	if err != nil { // если нет пользователя с таким id
		return st.User{}, err
	}
	return user, nil
}

// редактирование параметров пользователя
func (s *userService) UpdateUserParams(newParams st.ParamsUser) (st.ParamsUser, error) {
	userParam, err := s.repo.GetUserParams(newParams.UserID)
	if err != nil { // если нет пользователя с таким id
		return st.ParamsUser{}, err
	}

	userParam.Name = newParams.Name
	if newParams.Root > 0 {
		userParam.Root = newParams.Root
	}

	fmt.Println(userParam)

	if err := s.repo.UpdateUserParams(userParam); err != nil {
		return st.ParamsUser{}, err
		// return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Could not update user params"})
	}
	return userParam, nil
}

func (s *userService) DeleteUser(uid uint64) error {
	return s.repo.DeleteUser(uid)
}

func (s *userService) ExitUser(uid uint64) (bool, error) {
	var authUser st.AuthUser

	authUser, err := s.repo.GetUserAuth(uid)
	if err != nil { // если нет пользователя с таким id
		return false, err
	}
	authUser.RefreshToken = ""

	exit, err := s.repo.UpdateUserAuth(authUser)
	if err != nil {
		return false, err
		// return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Could not create user"})
	}

	return exit.RefreshToken == "", nil
}
