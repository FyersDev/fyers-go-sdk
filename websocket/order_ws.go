package fyersgosdk

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"path/filepath"
	"runtime"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

type OrderMessage map[string]interface{}

func (o OrderMessage) String() string {
	b, err := json.Marshal(map[string]interface{}(o))
	if err != nil {
		return fmt.Sprintf("%v", map[string]interface{}(o))
	}
	return string(b)
}

type OrderError map[string]interface{}

func (o OrderError) String() string {
	b, err := json.Marshal(map[string]interface{}(o))
	if err != nil {
		return fmt.Sprintf("%v", map[string]interface{}(o))
	}
	return string(b)
}

type OrderClose map[string]interface{}

func (o OrderClose) String() string {
	b, err := json.Marshal(map[string]interface{}(o))
	if err != nil {
		return fmt.Sprintf("%v", map[string]interface{}(o))
	}
	return string(b)
}

type FyersOrderSocket struct {
	accessToken          string
	logPath              string
	wsObject             *websocket.Conn
	wsRun                bool
	pingThread           *time.Ticker
	writeToFile          bool
	backgroundFlag       bool
	reconnectDelay       int
	onTrades             func(OrderMessage)
	onPosition           func(OrderMessage)
	restartFlag          bool
	onOrder              func(OrderMessage)
	onGeneral            func(OrderMessage)
	onError              func(OrderError)
	onOpen               func()
	maxReconnectAttempts int
	reconnectAttempts    int
	onClose              func(OrderClose)
	runningThread        bool
	url                  string
	positionMapper       map[string]interface{}
	orderMapper          map[string]interface{}
	tradeMapper          map[string]interface{}
	orderLogger          *FyersLogger
	websocketTask        bool
	socketType           map[string]interface{}
	mu                   sync.Mutex
	connected            bool
	stopChan             chan bool
}

func NewFyersOrderSocket(
	accessToken string,
	writeToFile bool,
	logPath string,
	onTrades func(OrderMessage),
	onPositions func(OrderMessage),
	onOrders func(OrderMessage),
	onGeneral func(OrderMessage),
	onError func(OrderError),
	onConnect func(),
	onClose func(OrderClose),
	reconnect bool,
	reconnectRetry int,
) *FyersOrderSocket {

	maxReconnectAttempts := 50
	if reconnectRetry < maxReconnectAttempts {
		maxReconnectAttempts = reconnectRetry
	}

	mapData, err := loadMapJSON()
	if err != nil {
		fmt.Printf("Failed to load map.json: %v\n", err)
	}

	positionMapper := make(map[string]interface{})
	orderMapper := make(map[string]interface{})
	tradeMapper := make(map[string]interface{})

	if mapData != nil {
		if posMap, exists := mapData["position_mapper"]; exists {
			positionMapper = posMap.(map[string]interface{})
		}
		if ordMap, exists := mapData["order_mapper"]; exists {
			orderMapper = ordMap.(map[string]interface{})
		}
		if trdMap, exists := mapData["trade_mapper"]; exists {
			tradeMapper = trdMap.(map[string]interface{})
		}
	}

	var loggerPath string
	if logPath != "" {
		loggerPath = logPath
	} else {
		loggerPath = ""
	}

	orderLogger := NewFyersLogger("FyersOrderSocket", "DEBUG", 2, loggerPath)

	socketType := map[string]interface{}{
		"OnOrders":    "orders",
		"OnTrades":    "trades",
		"OnPositions": "positions",
		"OnGeneral":   []string{"edis", "pricealerts", "login"},
	}

	return &FyersOrderSocket{
		accessToken:          accessToken,
		logPath:              logPath,
		wsObject:             nil,
		wsRun:                false,
		pingThread:           nil,
		writeToFile:          writeToFile,
		backgroundFlag:       false,
		reconnectDelay:       0,
		onTrades:             onTrades,
		onPosition:           onPositions,
		restartFlag:          reconnect,
		onOrder:              onOrders,
		onGeneral:            onGeneral,
		onError:              onError,
		onOpen:               onConnect,
		maxReconnectAttempts: maxReconnectAttempts,
		reconnectAttempts:    0,
		onClose:              onClose,
		runningThread:        false,
		url:                  "wss://socket.fyers.in/trade/v3",
		positionMapper:       positionMapper,
		orderMapper:          orderMapper,
		tradeMapper:          tradeMapper,
		orderLogger:          orderLogger,
		websocketTask:        false,
		socketType:           socketType,
		connected:            false,
		stopChan:             make(chan bool),
	}
}

func loadMapJSON() (map[string]interface{}, error) {
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		return nil, fmt.Errorf("failed to get current file path")
	}

	dir := filepath.Dir(filename)
	mapPath := filepath.Join(dir, "map.json")

	data, err := ioutil.ReadFile(mapPath)
	if err != nil {
		return nil, err
	}

	var mapData map[string]interface{}
	err = json.Unmarshal(data, &mapData)
	if err != nil {
		return nil, err
	}

	return mapData, nil
}

func (f *FyersOrderSocket) parsePositionData(msg map[string]interface{}) map[string]interface{} {
	positionData := make(map[string]interface{})

	if positions, exists := msg["positions"]; exists {
		if posMap, ok := positions.(map[string]interface{}); ok {
			for rawKey, mappedKey := range f.positionMapper {
				if val, exists := posMap[rawKey]; exists {
					positionData[mappedKey.(string)] = val
				}
			}
		}
	}

	return map[string]interface{}{
		"s":         msg["s"],
		"positions": positionData,
	}
}

func (f *FyersOrderSocket) parseTradeData(msg map[string]interface{}) map[string]interface{} {
	tradeData := make(map[string]interface{})

	if trades, exists := msg["trades"]; exists {
		if trdMap, ok := trades.(map[string]interface{}); ok {
			for rawKey, mappedKey := range f.tradeMapper {
				if val, exists := trdMap[rawKey]; exists {
					tradeData[mappedKey.(string)] = val
				}
			}
		}
	}

	return map[string]interface{}{
		"s":      msg["s"],
		"trades": tradeData,
	}
}

func (f *FyersOrderSocket) parseOrderData(msg map[string]interface{}) map[string]interface{} {
	orderData := make(map[string]interface{})

	if orders, exists := msg["orders"]; exists {
		if ordMap, ok := orders.(map[string]interface{}); ok {
			for rawKey, mappedKey := range f.orderMapper {
				if val, exists := ordMap[rawKey]; exists {
					orderData[mappedKey.(string)] = val
				}
			}

			if id, exists := ordMap["id"]; exists {
				if orgOrdStatus, exists := ordMap["org_ord_status"]; exists {
					orderData["orderNumStatus"] = fmt.Sprintf("%v:%v", id, orgOrdStatus)
				}
			}
		}
	}

	return map[string]interface{}{
		"s":      msg["s"],
		"orders": orderData,
	}
}

func (f *FyersOrderSocket) OnTrades(message map[string]interface{}) {
	if f.onTrades != nil {
		f.onTrades(OrderMessage(message))
	} else {
		fmt.Printf("Trade : %s\n", OrderMessage(message))
	}
}

func (f *FyersOrderSocket) OnPositions(message map[string]interface{}) {
	if f.onPosition != nil {
		f.onPosition(OrderMessage(message))
	} else {
		fmt.Printf("Position : %s\n", OrderMessage(message))
	}
}

func (f *FyersOrderSocket) OnOrder(message map[string]interface{}) {
	if f.onOrder != nil {
		f.onOrder(OrderMessage(message))
	} else {
		fmt.Printf("Order : %s\n", OrderMessage(message))
	}
}

func (f *FyersOrderSocket) OnGeneral(message map[string]interface{}) {
	if f.onGeneral != nil {
		f.onGeneral(OrderMessage(message))
	} else {
		fmt.Printf("General : %s\n", OrderMessage(message))
	}
}

func (f *FyersOrderSocket) OnError(message interface{}) {
	if f.onError != nil {
		f.onError(OrderError{"error": message})
	} else {
		fmt.Printf("Error : %v\n", message)
	}
}

func (f *FyersOrderSocket) handleMessage(message []byte) {
	messageStr := string(message)

	if messageStr == "pong" {
		return
	}

	var data map[string]interface{}
	err := json.Unmarshal(message, &data)
	if err != nil {
		f.OnError("Failed to parse message")
		return
	}

	if _, exists := data["orders"]; exists {
		parsedData := f.parseOrderData(data)
		f.OnOrder(parsedData)
	} else if _, exists := data["positions"]; exists {
		parsedData := f.parsePositionData(data)
		f.OnPositions(parsedData)
	} else if _, exists := data["trades"]; exists {
		parsedData := f.parseTradeData(data)
		f.OnTrades(parsedData)
	} else {
		f.OnGeneral(data)
	}
}

func (f *FyersOrderSocket) readMessages() {
	for {
		select {
		case <-f.stopChan:
			return
		default:
			if f.wsObject == nil {
				return
			}

			_, message, err := f.wsObject.ReadMessage()
			if err != nil {
				if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
					f.OnError(err.Error())
				} else {
					f.orderLogger.Info("WebSocket connection closed normally")
				}
				return
			}

			f.handleMessage(message)
		}
	}
}

func (f *FyersOrderSocket) ping() {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			f.mu.Lock()
			if f.wsObject != nil && f.connected {
				err := f.wsObject.WriteMessage(websocket.TextMessage, []byte("ping"))
				if err != nil {
					f.orderLogger.Error("Failed to send ping: " + err.Error())
				}
			}
			f.mu.Unlock()
		case <-f.stopChan:
			return
		}
	}
}

func (f *FyersOrderSocket) Connect() error {
	f.mu.Lock()
	defer f.mu.Unlock()

	headers := http.Header{}
	headers.Set("authorization", f.accessToken)

	dialer := websocket.Dialer{}
	conn, _, err := dialer.Dial(f.url, headers)
	if err != nil {
		return err
	}

	f.wsObject = conn
	f.connected = true
	f.wsRun = true

	authResponse := map[string]interface{}{
		"type":    AUTH_TYPE,
		"code":    200,
		"message": AUTH_SUCCESS,
		"s":       "ok",
	}
	jsonData, _ := json.Marshal(authResponse)
	fmt.Printf("Response: %s\n", string(jsonData))

	go f.readMessages()

	go f.ping()

	if f.onOpen != nil {
		f.onOpen()
	}

	return nil
}

func (f *FyersOrderSocket) socketTypeToSlist(dataTypes []string) []string {
	var dataTypeList []string
	for _, dataType := range dataTypes {
		if socketType, exists := f.socketType[dataType]; exists {
			if list, ok := socketType.([]string); ok {
				dataTypeList = append(dataTypeList, list...)
			} else if str, ok := socketType.(string); ok {
				dataTypeList = append(dataTypeList, str)
			}
		}
	}
	return dataTypeList
}

func (f *FyersOrderSocket) Subscribe(dataType string) {
	f.SubscribeMultiple([]string{dataType})
}

func (f *FyersOrderSocket) SubscribeMultiple(dataTypes []string) {
	dataTypeList := f.socketTypeToSlist(dataTypes)
	if len(dataTypeList) == 0 {
		return
	}

	msg := map[string]interface{}{
		"T":     "SUB_ORD",
		"SLIST": dataTypeList,
		"SUB_T": 1,
	}

	jsonData, _ := json.Marshal(msg)
	if f.wsObject != nil && f.connected {
		err := f.wsObject.WriteMessage(websocket.TextMessage, jsonData)
		if err != nil {
			f.OnError(err.Error())
		}
	}
}

func (f *FyersOrderSocket) Unsubscribe(dataType string) {
	f.UnsubscribeMultiple([]string{dataType})
}

func (f *FyersOrderSocket) UnsubscribeMultiple(dataTypes []string) {
	dataTypeList := f.socketTypeToSlist(dataTypes)
	if len(dataTypeList) == 0 {
		return
	}

	msg := map[string]interface{}{
		"T":     "SUB_ORD",
		"SLIST": dataTypeList,
		"SUB_T": -1,
	}

	jsonData, _ := json.Marshal(msg)
	if f.wsObject != nil && f.connected {
		err := f.wsObject.WriteMessage(websocket.TextMessage, jsonData)
		if err != nil {
			f.OnError(err.Error())
		}
	}
}

func (f *FyersOrderSocket) KeepRunning() {
	select {}
}

func (f *FyersOrderSocket) StopRunning() {
	f.CloseConnection()
}

func (f *FyersOrderSocket) CloseConnection() {
	f.mu.Lock()
	defer f.mu.Unlock()

	if f.wsObject != nil {
		f.wsObject.Close()
		f.wsObject = nil
	}

	f.connected = false
	f.wsRun = false

	select {
	case <-f.stopChan:
	default:
		close(f.stopChan)
	}

	if f.onClose != nil {
		f.onClose(OrderClose{"message": "Connection closed"})
	}
}

func (f *FyersOrderSocket) IsConnected() bool {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.connected && f.wsObject != nil
}
