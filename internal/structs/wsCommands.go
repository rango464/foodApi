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
)
