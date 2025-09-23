package main

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"regexp"
	"strconv"
	"strings"

	// "github.com/labstack/echo-jwt"
	"github.com/RangoCoder/foodApi/internal/appmidleware"
	"github.com/RangoCoder/foodApi/internal/db"
	"github.com/RangoCoder/foodApi/internal/handlersUser"
	"github.com/RangoCoder/foodApi/internal/userService"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func authControl(next echo.HandlerFunc) echo.HandlerFunc { // свой midleware
	return func(c echo.Context) error {
		incomeToken := strings.TrimPrefix(c.Request().Header.Get("Authorization"), "Bearer ") // берем JWT токен
		if len(incomeToken) == 0 {                                                            // если токен не указан - запрещаем
			return c.JSON(http.StatusUnauthorized, map[string]string{"error": "invalid access token"})
		}
		if len(incomeToken) > 10 { // токен указан и длина больше 10 символов - проверим токен

			newToken, err := appmidleware.ValidateJWT(incomeToken) //валидируем и возвращаем новый или пустой или сигнал для рефреша
			if err != nil {
				return errors.New("wrong authorisation") // при проверке токена произошла ошибка
			}
			// fmt.Printf(" \n newToken= %v \n", newToken)
			if newToken == "needrefresh" { // токен просрочен - сообщаем что нужно попробовать через рефрештокен
				c.Response().Header().Set(echo.HeaderAuthorization, newToken) // добавляем токен в ответный Headers
				return c.JSON(http.StatusUnauthorized, map[string]string{"error": "invalid access token"})
			} else if len(newToken) > 10 { //токен валиден и проверен
				claims, err := appmidleware.GetClaimsFromJWT(incomeToken) //смотрим что за ид в токене
				if err != nil {
					return errors.New("can`t get params from JWT") // при проверке токена произошла ошибка
				}
				access := acessControl(c, claims.Root, claims.Sub)
				if access { // если доступ разрешен
					//добавим в контекст для удобства
					uid := strconv.FormatUint(claims.Sub, 10)   // из числа в строку
					root := strconv.FormatUint(claims.Root, 10) // из числа в строку
					c.Set("ctxUid", uid)
					c.Set("ctxRoot", root)

					c.Response().Header().Set(echo.HeaderAuthorization, newToken) // добавляем токен в ответный Header
				} else {
					return errors.New("user acсess denied") // при проверке соответствия ид токена и адреса пути выявилось, что пользователь не допущен
				}
			} else {
				return c.JSON(http.StatusUnauthorized, map[string]string{"error": "invalid access token"})
			}
		}
		// fmt.Printf("authControl context = %v", c)
		return next(c)
	}
}

func acessControl(c echo.Context, root, uid uint64) bool { // проверим чтоб пользователь получал только ту информацию к которой допущен
	// проверим юид из параметра в запросе и сравним с нашиего пользователя
	// ROOT(0-guest 1-simpleuser 3-premiumuser 5-moderator 7-admin 9-superadmin)
	contentUid, err := strconv.ParseUint(fmt.Sprint(c.Param("uid")), 10, 64) // string to uint64
	if err != nil {                                                          //
		if root == 9 { // СУПЕРАДМИНУ ВСЕ МОЖНО
			return true
		}
	}
	if root == 9 || contentUid == uid { // если суперадмин или свой пользователь то тоже можно
		return true
	}

	// fmt.Println("/n acessIdControl get uid", uid)
	// fmt.Println("/n acessIdControl get root", root)
	// fmt.Println("/n acessIdControl get context", c)

	return false
}

// allowOrigin takes the origin as an argument and returns true if the origin
// is allowed or false otherwise.
func allowOrigin(origin string) (bool, error) { // тут разрешаем корс запросы от других источников
	// In this example we use a regular expression but we can imagine various
	// kind of custom logic. For example, an external datasource could be used
	// to maintain the list of allowed origins.
	return regexp.MatchString(`^https:\/\/labstack\.(net|com)$`, origin)
}

func main() {

	database, err := db.InitDB()
	if err != nil {
		log.Fatalf("Could not connect to database: %v", err)
	}

	e := echo.New()

	userRepo := userService.NewUserRepository(database)
	userServ := userService.NewUserService(userRepo)
	userHands := handlersUser.NewUserHandler(userServ)
	// guestHands := handlersGuest.NewUserHandler(userServ)

	// Middleware
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	// CORS restricted with a custom function to allow origins
	// and with the GET, PUT, POST or DELETE methods allowed.
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOriginFunc: allowOrigin,
		AllowMethods:    []string{http.MethodGet, http.MethodPut, http.MethodPost, http.MethodDelete},
	}))

	// Routes
	// e.GET("/", userHands.GetAllUsers)

	e.POST("/guest/register", userHands.RegisterUser) // register new user
	e.POST("/guest/login", userHands.LoginUser)       // login user

	//только для авторизоанного пользователя
	protected := e.Group("/user")
	protected.Use(authControl) //кастомный midleware - проверяет токен

	protected.GET("/all/:offset", userHands.GetAllUsers)
	protected.GET("/:uid", userHands.ReadUser)
	protected.PUT("/:uid", userHands.UpdateUserParams)
	protected.DELETE("/:uid", userHands.DeleteUser)

	// Start server
	e.Logger.Fatal(e.Start(":1323"))
}
