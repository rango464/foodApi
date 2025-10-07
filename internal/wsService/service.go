package wsService

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"

	"github.com/RangoCoder/foodApi/internal/structs"
)

type WsService interface {
	GetTickerPrice(ticker string) float64
}

type wsService struct {
	repo WsRepository
}

func NewWsService(repo WsRepository) WsService {
	return &wsService{repo: repo}
}

/*
getTickerPrice(ticker string) float64  - получает от api.binance.com текущее значение тикера и возвращает его
*/
func (s *wsService) GetTickerPrice(ticker string) float64 {
	url := fmt.Sprintf("https://api.binance.com/api/v3/ticker/price?symbol=%s", ticker)
	resp, err := http.Get(url)
	if err != nil || resp.StatusCode != 200 { // если неудалось - завершаем горутину
		log.Fatal("остановили getTickerPrice из за ошибки", err)
		return 0
	}
	body, err := io.ReadAll(resp.Body) //прочитаем ответ
	if err != nil {
		log.Fatalf("ошибка чтения ответа ws сервера: %v", err)
	}
	SP := structs.SymbolPrice{}
	err = json.Unmarshal(body, &SP) //ответ сервера ждем в формате такой структуры
	if err != nil {
		log.Fatalf("ошибка десериализации JSON: %v", err)
	}
	price, err := strconv.ParseFloat(SP.Price, 64)
	if err != nil {
		log.Fatalf("ошибка в ходе преобразования price в float64: %v", err)
	}
	return price
}
