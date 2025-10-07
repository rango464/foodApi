package structs

type (
	ClientComands struct { // структура команд от клиента на сервер
		Command  string   `json:"command"`
		Symbols  []string `json:"symbols,omitempty"`
		Fps      float64  `json:"fps"`
		DateFrom uint64   `json:"dfrom"`
		DateTo   uint64   `json:"dto"`
		Msg      string   `json:"msg"`
	}

	SymbolPrice struct {
		Symbol string `json:"symbol"`
		Price  string `json:"price"`
	}
	//команды
	/*
		{ "command": "show", "symbols":["BTCUSDT", "ETHUSDT","BTCUSDT", "BNBUSDT", "BNBBTC", "LTCBTC"], "datefrom": ,"dateto"} - показат данные о котировках за период
		{ "command": "stream", "symbols":["BTCUSDT", "ETHUSDT","BTCUSDT", "BNBUSDT", "BNBBTC", "LTCBTC"], "fps":0.5 }- стримить текущее значение тикеров
		{ "command": "unstream", "symbols":["BTCUSDT", "ETHUSDT","BTCUSDT", "BNBUSDT", "BNBBTC", "LTCBTC"] }- прекратить стримить текущее значение тикеров
		{"command": "broadcast" "msg":""}
	*/

	ARSymbolPrise struct {
		Symbol string `json:"symbol"`
		Price  string `json:"price"`
	}

	ClientRequest struct { // сообщения от клиента на сервер имеют одинаковый формат
		Command string   `json:"command"`
		Symbols []string `json:"symbols,omitempty"`
		Fps     float64  `json:"fps"`
	}
	CommandRequest struct { // сообщения от клиента на сервер имеют одинаковый формат
		Command string                 `json:"command"`
		Data    RqData                 `json:"data,omitempty"`
		Message string                 `json:"message,omitempty"`
		Meta    map[string]interface{} `json:"meta,omitempty"`
	}

	RqData struct {
		Symbol string
		Fps    float64
	}
	RsData struct {
		Symbol string
		Val    float64
	}

	CommandResponse struct { /// ответы сервера
		Ok      bool        `json:"ok"`
		Reply   string      `json:"reply,omitempty"`
		Data    interface{} `json:"data,omitempty"`
		Error   string      `json:"error,omitempty"`
		Command string      `json:"command,omitempty"`
	}
)
