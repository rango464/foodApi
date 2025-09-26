package handlersUser

import (
	"fmt"
	"net/http"
	"strconv"

	st "github.com/RangoCoder/foodApi/internal/structs"
	"github.com/RangoCoder/foodApi/internal/userService"
	"github.com/labstack/echo/v4"
)

type UserHandler struct {
	service userService.UserService
}

func NewUserHandler(s userService.UserService) *UserHandler {
	return &UserHandler{service: s}
}

func (h *UserHandler) RegisterUser(c echo.Context) error { // register new user
	var req st.User
	if err := c.Bind(&req); err != nil { // если передаваемые данные не соответствуют модели
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid request"})
	}

	userNew := st.User{
		Email:    req.Email,
		Password: req.Password,
	}

	user, err := h.service.RegisterUser(userNew)

	if err != nil {
		fmt.Println(err)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Could not create user"})
	}
	respMessage := fmt.Sprintf("user saved with id %v", user.ID)
	return c.JSON(http.StatusCreated, respMessage)
}

func (h *UserHandler) RefreshUserAccess(c echo.Context) error { // restore user access by refresh token
	uid, err := strconv.ParseUint(c.Param("uid"), 10, 64) // string to uint
	if err != nil {                                       // ... handle error
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid uid in get request"})
	}

	var req st.AuthTokens
	if err := c.Bind(&req); err != nil { // если передаваемые данные не соответствуют модели
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid request"})
	}
	tokens := st.AuthTokens{
		AccessToken:  req.AccessToken,
		RefreshToken: req.RefreshToken,
	}

	auth, err := h.service.RestoreAccessByRefresh(uid, tokens) // acsessJwt, refreshJwt

	if err != nil {
		fmt.Println(err)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Could not refresh user access, wrong data"})
	}

	return c.JSON(http.StatusOK, auth)
}

func (h *UserHandler) LoginUser(c echo.Context) error { // login user (email, pass)
	var req st.User
	if err := c.Bind(&req); err != nil { // если передаваемые данные не соответствуют модели
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid request"})
	}

	userLogin := st.User{
		Email:    req.Email,
		Password: req.Password,
	}

	auth, err := h.service.LoginUser(userLogin) // acsessJwt, refreshJwt

	if err != nil {
		fmt.Println(err)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Could not login, wrong data"})
	}

	return c.JSON(http.StatusOK, auth)
}

/*
проверяем валидность входящего токена
если токен валиден - обновим и вернем свежий
если невалиден вернем пустую строку
*/
// func (h *UserHandler) ValidateJWT(key string) (string, error) {
// 	// fmt.Printf("handler validate token %v", key)
// 	validateRes, err := h.service.ValidateAcessKey(key)
// 	if err != nil {
// 		fmt.Printf("handler validate token err %v", err)
// 		return "", errors.New("wrong authorisation")
// 	}
// 	return validateRes, nil
// }

func (h *UserHandler) GetAllUsers(c echo.Context) error {
	offset, err := strconv.Atoi(c.Param("offset")) //получим смещение string to int
	if err != nil {                                // ... handle error
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid offset in get request"})
	}

	count, err := strconv.Atoi(c.QueryParam("count")) //получим смещение string to int
	if err != nil {                                   // ... handle error
		count = 50
	}

	users, err := h.service.GetAllUsers(offset, count)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Could not get all users"})
	}
	return c.JSON(http.StatusOK, users)
}

func (h *UserHandler) ReadUser(c echo.Context) error { //read one user by id
	// fmt.Printf("handler read user with context %v", c)

	uid, err := strconv.ParseUint(c.Param("uid"), 10, 64) // string to uint
	if err != nil {                                       // ... handle error
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid id in get request"})
	}

	user, err := h.service.GetUserById(uid)
	if err != nil {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "user not found"})
	}
	return c.JSON(http.StatusOK, user)
}

func (h *UserHandler) UpdateUserParams(c echo.Context) error {
	uid, err := strconv.ParseUint(c.Param("uid"), 10, 64) // string to uint
	if err != nil {                                       // ... handle error
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid id in update request"})
	}

	var req st.ParamsUser
	if err := c.Bind(&req); err != nil { // если передаваемые данные не соответствуют модели
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid request"})
	}

	//прочитаем root из контекста
	root, err := strconv.ParseInt(c.Get("ctxRoot").(string), 10, 8) // строку в uint8
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "can`t get correct root from ctx"})
	}

	fmt.Printf("geted from context root=%v", root)

	newUserParams := st.ParamsUser{
		UserID: uid,
		Name:   req.Name, // может менять пользователь
	}
	if root == 9 {
		newUserParams.Root = req.Root // может менять только суперадмин
	}

	updated, err := h.service.UpdateUserParams(newUserParams)

	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Could not update user params"})
	}
	return c.JSON(http.StatusCreated, updated)
}

func (h *UserHandler) DeleteUser(c echo.Context) error {
	uid, err := strconv.ParseUint(c.Param("uid"), 10, 64) // string to uint
	if err != nil {                                       // ... handle error
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid id in update request"})
	}

	if err := h.service.DeleteUser(uid); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Could not delete user"})
	}
	return c.NoContent(http.StatusNoContent)
}

func (h *UserHandler) ExitUser(c echo.Context) error {
	uid, err := strconv.ParseUint(c.Param("uid"), 10, 64) // string to uint
	if err != nil {                                       // ... handle error
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid id in exit request"})
	}

	exit, err := h.service.ExitUser(uid)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Could not exit"})
	}
	return c.JSON(http.StatusOK, exit)
}
