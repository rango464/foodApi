package userService

import (
	st "github.com/RangoCoder/foodApi/internal/structs"
	"gorm.io/gorm"
)

type UserRepository interface {
	RegisterUser(user st.User) (st.User, error)
	CreateUser(user st.User) (uint64, error)
	LoginUser(user st.User) (st.User, error)
	GetAllUsers(offset, count int) ([]st.UserShow, error)
	GetUserById(uid uint64) (st.User, error)
	DeleteUser(uid uint64) error
	SetVarifEmail(uid uint64, varif bool) error
	CreateUserParams(uid uint64) (st.ParamsUser, error)
	GetUserParams(uid uint64) (st.ParamsUser, error)
	UpdateUserParams(params st.ParamsUser) error
	CreateUserAuth(uid uint64) (st.AuthUser, error)
	GetUserAuth(uid uint64) (st.AuthUser, error) //
	GetUserAuthByRefresh(uid uint64, rToken string) (st.AuthUser, error)
	UpdateUserAuth(auth st.AuthUser) (st.AuthUser, error) //
	UpdateEmailCodeConfirm(uid uint64, confirmLine st.UserMailConfirm) (st.UserMailConfirm, error)
	GetEmailCodeConfirm(confirmLine st.UserMailConfirm) (st.UserMailConfirm, error)
	DeleteEmailCodeConfirm(uid uint64) error
}

type userRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) UserRepository {
	return &userRepository{db: db}
}

// создадим пользователя, его пустые параметры и пустую авторизацию - применим транзакцию
func (r *userRepository) RegisterUser(user st.User) (st.User, error) {
	//пробовал транзакции, но работает через раз - по этому по старинке но с гарантией
	// fmt.Printf("/n prepare user create, user income = %v /n ", user)
	// создаем пользователя
	uid, err := r.CreateUser(user)
	if err != nil {
		return st.User{}, err
	}
	// fmt.Printf("/n create ok, user id = %v /n ", uid)
	if uid == 0 { // юзер создалсz криво - откатим удалив пользователя - связанное каскадно удальтся по зависимостям бд
		err := r.DeleteUser(uid)
		if err != nil {
			return st.User{}, err
		}
	}
	// создаем параметры пользователя
	uparams, err := r.CreateUserParams(uid)
	if err != nil {
		return st.User{}, err
	}
	if uparams.UserID != uid { // параметры юзера создались криво - откатим удалив пользователя - связанное каскадно удальтся по зависимостям бд
		err := r.DeleteUser(uid)
		if err != nil {
			return st.User{}, err
		}
	}
	// fmt.Printf("/n newuserparams created, user id = %v, userparamsId = %v/n ", uid, uparams.ID)
	// создаем запись авторизации пользователя
	uauth, err := r.CreateUserAuth(uid)
	if err != nil {
		return st.User{}, err
	}
	if uauth.UserID != uid { // аутентификация юзера создались криво - откатим удалив пользователя - связанное каскадно удальтся по зависимостям бд
		err := r.DeleteUser(uid)
		if err != nil {
			return st.User{}, err
		}
	}
	// fmt.Printf("/n newuserauth created, user id = %v, userauthId = %v /n ", uid, uauth.ID)
	if uauth.UserID == uid && uparams.UserID == uid { // все создалось отправим выбранного пользователя
		regUser, err := r.GetUserById(uid)
		if err != nil {
			return st.User{}, err
		}
		return regUser, err
	}

	return st.User{}, err
}

// создадим пользователя
func (r *userRepository) CreateUser(user st.User) (uint64, error) {
	result := r.db.Create(&user)
	return user.ID, result.Error
}

func (r *userRepository) LoginUser(u st.User) (st.User, error) {
	var user st.User
	err := r.db.Preload("ParamsUser").Preload("AuthUser").Where("email = ? AND password = ?", u.Email, u.Password).First(&user).Error
	return user, err
}

// показывает данные всех пользователей
func (r *userRepository) GetAllUsers(offset, count int) ([]st.UserShow, error) {
	var userShow []st.UserShow
	query := r.db.Table("users").Select("users.id, users.email, pu.name, pu.root").Joins("left join params_users pu on pu.user_id = users.id").Limit(count).Offset(offset).Scan(&userShow)
	err := query.Error
	return userShow, err
}

// показывает данные пользователя с ид
func (r *userRepository) GetUserById(uid uint64) (st.User, error) {
	var user st.User
	query := r.db.Preload("ParamsUser").Preload("AuthUser").Take(&user, uid)
	err := query.Error
	return user, err
}

func (r *userRepository) DeleteUser(uid uint64) error {
	var user st.User
	return r.db.Delete(&user, uid).Error
}

func (r *userRepository) SetVarifEmail(uid uint64, varif bool) error {
	user, err := r.GetUserById(uid)
	if err != nil {
		return err
	}
	// меняем парамет и сохраняем
	user.VarifEmail = true
	return r.db.Save(&user).Error
}

/*................................................*/

func (r *userRepository) CreateUserParams(uid uint64) (st.ParamsUser, error) {
	var params = st.ParamsUser{UserID: uid}
	err := r.db.Create(&params).Error
	return params, err
}

func (r *userRepository) GetUserParams(uid uint64) (st.ParamsUser, error) {
	var params = st.ParamsUser{}
	err := r.db.Where(&st.ParamsUser{UserID: uid}).First(&params).Error
	return params, err
}

func (r *userRepository) UpdateUserParams(params st.ParamsUser) error {
	return r.db.Save(&params).Error
}

/*....................................................*/

func (r *userRepository) CreateUserAuth(uid uint64) (st.AuthUser, error) {
	var auth = st.AuthUser{UserID: uid}
	err := r.db.Create(&auth).Error
	return auth, err
}

func (r *userRepository) GetUserAuth(uid uint64) (st.AuthUser, error) {
	var auth = st.AuthUser{}
	err := r.db.Where(&st.AuthUser{UserID: uid}).First(&auth).Error
	return auth, err
}

func (r *userRepository) GetUserAuthByRefresh(uid uint64, rToken string) (st.AuthUser, error) {
	var auth = st.AuthUser{}
	err := r.db.Where(&st.AuthUser{UserID: uid, RefreshToken: rToken}).First(&auth).Error
	return auth, err
}

func (r *userRepository) UpdateUserAuth(auth st.AuthUser) (st.AuthUser, error) {
	err := r.db.Save(&auth).Error
	return auth, err
}

/*....................................................*/
// обновим код подтверждения, если его еще нет то создадим
func (r *userRepository) UpdateEmailCodeConfirm(uid uint64, confirmLine st.UserMailConfirm) (st.UserMailConfirm, error) {
	err := r.db.Save(&confirmLine).Error
	return confirmLine, err
}

func (r *userRepository) GetEmailCodeConfirm(confirmLine st.UserMailConfirm) (st.UserMailConfirm, error) {
	var confirm = st.UserMailConfirm{}
	err := r.db.Where(&st.UserMailConfirm{UserID: confirmLine.UserID, VarificationCode: confirmLine.VarificationCode}).First(&confirm).Error
	// if err != nil { // пытался прредотвратить вывод ошибки об отсутствии записи (record not found) - пока не получилось
	// 	if errors.Is(err, gorm.ErrRecordNotFound) {
	// 		return st.UserMailConfirm{}, nil
	// 	}
	// }
	return confirm, err
}

func (r *userRepository) DeleteEmailCodeConfirm(uid uint64) error {
	var confirmLine st.UserMailConfirm
	return r.db.Delete(&confirmLine, uid).Error
}
