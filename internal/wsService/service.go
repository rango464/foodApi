package wsService

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/RangoCoder/foodApi/internal/structs"
	"github.com/gorilla/websocket"
)

type WsService interface {
	WriteJSON(conn *websocket.Conn, v any) error
	WsManager(conn *websocket.Conn, req structs.ClientRequest)
	WsLiveStreamer(conn *websocket.Conn, fps float64, symbols []string, stop chan string) error
}

type wsService struct {
	repo WsRepository
}

func NewWsService(repo WsRepository) WsService {
	return &wsService{repo: repo}
}

func (s *wsService) WriteJSON(conn *websocket.Conn, v any) error {
	conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
	return conn.WriteJSON(v)
}

/*
менеджер решает что делать с пришедшей командой
*/
func (s *wsService) WsManager(conn *websocket.Conn, req structs.ClientRequest) {
	stop := make(chan string) //канал для остановки горутины
	//создаем слайс с нужными пользователю валютными парами (их названия)
	// symbols := []string{"BTCUSDT", "ETHUSDT","BTCUSDT", "BNBUSDT", "BNBBTC", "LTCBTC"} //BTCUSDT, BNBUSDT, BNBBTC, LTCBTC
	// в менеджер приходит собщение - он выделяет из него команду решает что делать
	// например получено сообщение  structs.CommandRequest
	//  с коммандой Command: "livestream" и Data: { "fps":0.5} { "command": "livestream", "data": {"fps":0.5} }
	//  с коммандой Command: "addsymbol" и Data: { "fps":0.5} { "command": "addsymbol", "data": {"symbol":"LTCBTC"} }
	switch req.Command {
	case "livestream":
		log.Println("начинаем стрим по команде")
		//в сообщении могут случайно задублироваться symbols (человеческий фактор) предотвратим попадание в стример дублей
		allKeys := make(map[string]bool)
		var symbols []string
		for _, symbol := range req.Symbols {
			if _, value := allKeys[symbol]; !value {
				allKeys[symbol] = true
				symbols = append(symbols, symbol)
			}
		}
		//направим в стример
		s.WsLiveStreamer(conn, req.Fps, symbols, stop)
	case "stopstream": // останавливаем стрим
		log.Println("останавливаем стрим отдельных пар  по команде")
		for _, simbol := range req.Symbols {
			stop <- simbol
		}

	case "other": // тест
	case "test": // тест
	}
}

/*
функция стримера получает данные о валютной паре и частоте оновления данных
*/

func (s *wsService) WsLiveStreamer(conn *websocket.Conn, fps float64, symbols []string, stop chan string) error {
	cnl := make(chan map[string]float64) // общий канал для сбора данных от горутин
	// go func() {                          // таймер до отключения стримера
	// 	time.Sleep(10 * time.Second)
	// 	stop <- "BTCUSDT"
	// }()
	// var w sync.WaitGroup
	// w.Add(len(symbols))
	go func(cnl chan map[string]float64, stop chan string) {
		for i := 0; i < len(symbols); i++ { //для каждой пары запускаем горутину, цель которой брать данные и стримить в сокет

			sbl := symbols[i]
			go func(sbl string, cnl chan map[string]float64, stop chan string) {
				name := sbl
				log.Println("запущена горутина - " + sbl)
				for {
					select {
					case stp := <-stop: // остановка
						fmt.Println("stp=", name, "sbl=", name)
						if stp == name {
							fmt.Println("остановили горутину ", stp, "принудительно")
							return
						}
					default: // работаем в штатном режиме
						url := fmt.Sprintf("https://api.binance.com/api/v3/ticker/price?symbol=%v", sbl)
						resp, err := http.Get(url)
						if err != nil || resp.StatusCode != 200 { // если неудалось - завершаем горутину
							log.Fatal("остановили горутину из за ошибки", err)
							return
						}
						body, err := io.ReadAll(resp.Body) //прочитаем ответ
						if err != nil {
							log.Fatalf("ошибка чтения ответа ws сервера: %v", err)
						}
						// fmt.Println("данные body", body)
						var AR structs.ARSymbolPrise // (AR- apirequest ) ответ сервера ждем в формате такой структуры
						err = json.Unmarshal(body, &AR)
						if err != nil {
							log.Fatalf("Ошибка десериализации JSON: %v", err)
						}
						price, err := strconv.ParseFloat(AR.Price, 64)
						if err != nil {
							log.Fatalf("ошибка в ходе преобразования price в float64: %v", err)
						}
						// fmt.Println("данные структуры после десерелизации", AR)
						cnl <- map[string]float64{
							AR.Symbol: price,
						}
						time.Sleep(time.Duration(fps * float64(time.Second))) // обновление валютных пар происходит с задержкой, указанной в fps
					}

				}
			}(sbl, cnl, stop)
			// w.Done()
			// s.WsLiveStreamer(symbol string, cnl string) // менеджер отдал задачу стримеру показывать значение выбранных пар каждые 0.5 секунд
		}
	}(cnl, stop)

	rsAgr := map[string]float64{} // наполняем общий ответ
	for data := range cnl {
		for ismb, ival := range data {
			rsAgr[ismb] = ival
		}
		s.WriteJSON(conn, rsAgr) // отвечаем пользователю при каждом обновлении данных
	}
	defer close(cnl)
	// w.Wait()
	return nil
}
