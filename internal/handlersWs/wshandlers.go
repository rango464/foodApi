package handlersWs

import (
	"encoding/json"
	"fmt"
	"log"
	"maps"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/RangoCoder/foodApi/internal/structs"
	"github.com/RangoCoder/foodApi/internal/wsService"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/labstack/echo/v4"
)

type (
	WsHandler struct {
		service wsService.WsService
	}
	WSConnection struct { // соединение с сокетом
		Conn *websocket.Conn
	}
)

func NewWsHandler(s wsService.WsService) *WsHandler {
	return &WsHandler{service: s}
}

var (
	conectionPool = make(map[string]*WSConnection) // пул активных соединений пользователя ключ- uгid соединения, val - само соединение
	upgrader      = websocket.Upgrader{            // меняем формат общения клиен-сервер
		CheckOrigin: func(r *http.Request) bool { return true }, // для примера разрешаем всем
	}
)

/*
WsGOTickers(c echo.Context) error
пользователь подключается и направляет команду  - команда передается менеджеру, который решает что с этим делать дальше
используем пул соединений, работаем в горутинах
*/
func (h *WsHandler) WsGOTickers(c echo.Context) error {
	// пользователю прошедшему проверку позволяем пройти берем его ид
	uid, err := strconv.ParseUint(c.Param("uid"), 10, 64) // string to uint
	if err != nil {                                       // ... handle error
		log.Println("не удалось получить uid(", uid, ") пользователя", err)
	}
	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		//организуем соединение
		conn, err := upgrader.Upgrade(c.Response(), c.Request(), nil)
		if err != nil {
			c.Logger().Errorf("Error with websocket connectoin, (after try Upgrade Wsocket) - %v", err)
			return
		}

		wsid := uuid.New().String()    // сгенерировали ид ws сессии
		newConn := &WSConnection{conn} //
		conectionPool[wsid] = newConn  // вносим соединение в пул

		defer delete(conectionPool, wsid) // удаляем соединение при закрытии

		for {
			_, message, err := conn.ReadMessage() //слушаем что появляется на входе
			if err != nil {                       // Клиент отключился или ошибка чтения
				log.Println("read:", err)
				if strings.Contains(fmt.Sprint(err), "websocket: close") {
					delete(conectionPool, wsid)
				}
				break
			}
			h.GoWsManager(wsid, string(message)) //отправим входящую команду менеджеру - он там сам ответит
		}
		wg.Done()
	}()
	wg.Wait()
	return nil
}

/*
goWsManager(connID, msg string) - менеджер
получая сообщение менеджер определяет евляется ли оно командой, если да - продолжает работать
выбирая команду - передает задачу исполнителю
*/
func (h *WsHandler) GoWsManager(connID, msg string) {
	var cr structs.ClientComands
	err := json.Unmarshal([]byte(msg), &cr) //
	if err != nil {
		fmt.Printf("goWsManager: сообщение %s не является командой...\n", msg)
		return
	}

	switch cr.Command {
	case "stream": // { "command": "stream", "symbols":["BTCUSDT", "ETHUSDT","BTCUSDT", "BNBUSDT", "BNBBTC", "LTCBTC"], "fps":0.5 }- стримить текущее значение тикеров
		if err := h.GoWsLiveStreamer(connID, cr); err != nil { // запускаем стример для конкретного канала клиента
			fmt.Printf("GoWsLiveStreamer ответил ошибкой %s \n", err)
		}

	case "unstream": // останавливаем стрим
	case "broadcast": // общий канал
		h.GobroadcastToAll(cr.Msg)
	case "show":
	case "test": // тест
	}
}

/*
	goWsLiveStreamer(cid string, cr structs.ClientComands) error  - один из исполнителей

получая команду исполнитель запускает горутину с непрерывным циклом сбора данных по тикерам из задания
цикл асинхронной работы производится в горутинах по этому чтобы дождаться корректного коллективного ответа - используем  WaitGroup
собрав коллективный ответ из буферищированного канала - убедимся что получаель информации все еще на связи и отправим ему информацию
если получатель отключился - останавливаем работу цикла, а планировщик сам подчистит отработавшие горутины
*/
func (h *WsHandler) GoWsLiveStreamer(cid string, cr structs.ClientComands) error {
	symbols := cr.Symbols // перечень тикеров
	go func() {

		for {
			cnl := make(chan map[string]float64, len(symbols))
			result := map[string]float64{} // результат работы всех горутин
			wg := sync.WaitGroup{}
			wg.Add(len(symbols))

			for _, ticker := range symbols { // стартуем горутины
				go func() {
					cnl <- map[string]float64{ticker: h.service.GetTickerPrice(ticker)}
					wg.Done()
				}()
			}

			wg.Wait()  // дождались всех
			close(cnl) //закрыли канал
			//собираем ответ из бферезированного канала
			for spdata := range cnl {
				maps.Copy(result, spdata)
				// for s, p := range spdata {
				// 	result[s] = p
				// }

			}
			// убедимся что соединение еще в пуле (клиент не отключился и тд) - это дополнительная проверка (первая была при прослушивании того, что появляется на входе)
			if conectionPool[cid] == nil {
				fmt.Printf("соединение %s отключено", cid) // ничего не направляем и останавливаем горутину
				log.Println("состояние conectionPool", conectionPool)
				break
			}
			h.GohandleMessage(cid, fmt.Sprintf("%v", result))
			time.Sleep(time.Duration(cr.Fps * float64(time.Second))) // обновление валютных пар происходит с задержкой, указанной в fps
		}
	}()
	return nil
}

/*
gohandleMessage(connID, msg string)  - отпарвляет сообщение либо всем либо конкретному клиенту, если указан uuid его соединения
*/
func (h *WsHandler) GohandleMessage(connID, msg string) {
	fmt.Printf("сообщение %s направляется клиенту %s\n", msg, connID)
	if connID != "" {
		h.GobroadcastToClient(connID, msg)
	} else {
		h.GobroadcastToAll(msg)
	}
}

/*
gobroadcatsToClient(cid, message string)  - отпарвляет сообщение  конкретному клиенту
*/
func (h *WsHandler) GobroadcastToClient(cid, message string) {
	for wsid, conn := range conectionPool {
		if cid == wsid { //маршрутируем ответ указанному пользователю
			conn.Conn.WriteMessage(websocket.TextMessage, []byte(message))
			break
		}
	}
}

/*
gobroadcatsToAll(message string)  - отпарвляет сообщение  всем клиентам
*/
func (h *WsHandler) GobroadcastToAll(message string) {
	for wsid, conn := range conectionPool {
		fmt.Printf("broadcatsToAll сообщение для %s:%s\n", wsid, []byte(message))
		conn.Conn.WriteMessage(websocket.TextMessage, []byte(message))
	}
}
