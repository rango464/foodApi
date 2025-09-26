package main

import (
	"log"
	"net/http"

	// "github.com/labstack/echo-jwt"

	"github.com/RangoCoder/foodApi/internal/db"
	"github.com/RangoCoder/foodApi/internal/handlersUser"
	"github.com/RangoCoder/foodApi/internal/userService"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

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

	e.POST("/guest/reresh", userHands.RefreshUserAccess) // restore user access by refresh token

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
