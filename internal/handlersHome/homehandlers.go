package handlersHome

import (
	"net/http"

	"github.com/RangoCoder/foodApi/internal/homeService"
	"github.com/gorilla/websocket"
	"github.com/labstack/echo/v4"
)

type HomeHandler struct {
	service homeService.HomeService
}

func NewHomeHandler(s homeService.HomeService) *HomeHandler {
	return &HomeHandler{service: s}
}

var (
	upgrader = websocket.Upgrader{}
)

func (h *HomeHandler) Home(c echo.Context) error {
	return c.JSON(http.StatusOK, "Homepage httpServer + WS")
}
