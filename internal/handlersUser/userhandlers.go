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
	offset, err := strconv.Atoi(c.Param("offset")) // string to int
	if err != nil {                                // ... handle error
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid id in get request"})
	}
	count := 50
	users, err := h.service.GetAllUsers(offset, count)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Could not get all users"})
	}
	return c.JSON(http.StatusOK, users)
}

func (h *UserHandler) ReadUser(c echo.Context) error { //read one user by id
	// fmt.Printf("handler read user with context %v", c)

	id, err := strconv.ParseUint(c.Param("uid"), 10, 64) // string to uint
	if err != nil {                                      // ... handle error
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid id in get request"})
	}

	user, err := h.service.GetUserById(id)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Could not get user with id " + c.Param("id")})
	}
	return c.JSON(http.StatusOK, user)
}

func (h *UserHandler) UpdateUserParams(c echo.Context) error {
	id, err := strconv.ParseUint(c.Param("uid"), 10, 64) // string to uint
	if err != nil {                                      // ... handle error
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid id in update request"})
	}

	var req st.ParamsUser
	if err := c.Bind(&req); err != nil { // если передаваемые данные не соответствуют модели
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid request"})
	}

	newUserParams := st.ParamsUser{
		UserID: id,
		Name:   req.Name,
		Root:   req.Root,
	}

	updated, err := h.service.UpdateUserParams(newUserParams)

	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Could not update user params"})
	}
	return c.JSON(http.StatusCreated, updated)
}

func (h *UserHandler) DeleteUser(c echo.Context) error {
	id, err := strconv.ParseUint(c.Param("uid"), 10, 64) // string to uint
	if err != nil {                                      // ... handle error
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid id in update request"})
	}

	if err := h.service.DeleteUser(id); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Could not delete user with id " + c.Param("id")})
	}
	return c.NoContent(http.StatusNoContent)
}
