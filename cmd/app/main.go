package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/RangoCoder/foodApi/internal/db"
	"github.com/RangoCoder/foodApi/internal/env"
	"github.com/RangoCoder/foodApi/internal/handlersHome"
	"github.com/RangoCoder/foodApi/internal/handlersUser"
	"github.com/RangoCoder/foodApi/internal/handlersWs"
	"github.com/RangoCoder/foodApi/internal/userService"
	"github.com/RangoCoder/foodApi/internal/wsService"

	"github.com/RangoCoder/foodApi/internal/homeService"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func main() {

	database, err := db.InitDB()
	if err != nil {
		log.Fatalf("Could not connect to database: %v", err)
	}

	e := echo.New()
	// общедоступные разделы
	homeRepo := homeService.NewHomeRepository(database)
	homeServ := homeService.NewHomeService(homeRepo)
	homeHands := handlersHome.NewHomeHandler(homeServ)

	// работа пользователя пользователя через HTTP
	userRepo := userService.NewUserRepository(database)
	userServ := userService.NewUserService(userRepo)
	userHands := handlersUser.NewUserHandler(userServ)

	// работа пользователя пользователя через WS
	wsRepo := wsService.NewWsRepository(database)
	wsServ := wsService.NewWsService(wsRepo)
	wsHands := handlersWs.NewWsHandler(wsServ)

	// Middleware
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	// CORS restricted with a custom function to allow origins
	// and with the GET, PUT, POST or DELETE methods allowed.
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOriginFunc: allowOrigin,
		AllowMethods:    []string{http.MethodGet, http.MethodPut, http.MethodPost, http.MethodDelete},
	}))
	///////////////////////////////////////////////////////////////////////////
	e.Static("/", "../public")
	/////////////////////////////////////////////////////////////////////////////
	// Routes
	e.GET("/", homeHands.Home)

	// соединение через Websoket
	e.GET("/ws/ticker/:uid", wsHands.WsGOTickers) // котировки через пул в горутинах
	// e.GET("/wschat", wsHands.WSChat)

	//работа с пользователями по HTTP
	e.POST("/guest/register", userHands.RegisterUser) // register new user
	e.POST("/guest/login", userHands.LoginUser)       // login user

	addr := fmt.Sprintf("/%v/:uid", env.GetEnvVar("EMAILCONFIRM_URL"))
	e.GET(addr, userHands.EmailConfirm) // confirm user Email

	e.POST("/guest/refresh/:uid", userHands.RefreshUserAccess) // restore user access by refresh token
	//only authorized users
	protected := e.Group("/user")
	protected.Use(authControl) //custom midleware - use Access token to control correct aceess user by root

	protected.GET("/all/:offset", userHands.GetAllUsers)
	protected.GET("/:uid", userHands.ReadUser)
	protected.PUT("/:uid", userHands.UpdateUserParams)
	protected.DELETE("/:uid", userHands.DeleteUser)
	protected.POST("/exit/:uid", userHands.ExitUser)

	// Start server
	e.Logger.Fatal(e.Start(":1323"))
}
