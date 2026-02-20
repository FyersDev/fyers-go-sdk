package fyersgosdk

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

type SubscriptionModes string

const (
	DEPTH SubscriptionModes = "depth"
)

type Depth struct {
	Tbq       int       `json:"tbq"`
	Tsq       int       `json:"tsq"`
	BidPrice  []float64 `json:"bidprice"`
	AskPrice  []float64 `json:"askprice"`
	BidQty    []float64 `json:"bidqty"`
	AskQty    []float64 `json:"askqty"`
	BidOrdn   []float64 `json:"bidordn"`
	AskOrdn   []float64 `json:"askordn"`
	Snapshot  bool      `json:"snapshot"`
	Timestamp int64     `json:"timestamp"`
	SendTime  int64     `json:"sendtime"`
	SeqNo     int64     `json:"seqNo"`
}

func NewDepth() *Depth {
	return &Depth{
		Tbq:       0,
		Tsq:       0,
		BidPrice:  make([]float64, 50),
		AskPrice:  make([]float64, 50),
		BidQty:    make([]float64, 50),
		AskQty:    make([]float64, 50),
		BidOrdn:   make([]float64, 50),
		AskOrdn:   make([]float64, 50),
		Snapshot:  false,
		Timestamp: 0,
		SendTime:  0,
		SeqNo:     0,
	}
}

func (d *Depth) String() string {
	return fmt.Sprintf("Depth{ts: %d, send_ts: %d, tbq: %d, tsq: %d, bidprice: %v, askprice: %v, bidqty: %v, askqty: %v, bidordn: %v, askordn: %v, snapshot: %t, sNo: %d}",
		d.Timestamp, d.SendTime, d.Tbq, d.Tsq, d.BidPrice, d.AskPrice, d.BidQty, d.AskQty, d.BidOrdn, d.AskOrdn, d.Snapshot, d.SeqNo)
}

type SubscriptionInfo struct {
	symbols        map[string]map[string]bool
	modeInfo       map[string]SubscriptionModes
	activeChannels map[string]bool
	mu             sync.RWMutex
}

func NewSubscriptionInfo() *SubscriptionInfo {
	return &SubscriptionInfo{
		symbols:        make(map[string]map[string]bool),
		modeInfo:       make(map[string]SubscriptionModes),
		activeChannels: make(map[string]bool),
	}
}

func (si *SubscriptionInfo) Subscribe(symbols map[string]bool, channelNo string, mode SubscriptionModes) {
	si.mu.Lock()
	defer si.mu.Unlock()

	if existingSymbols, exists := si.symbols[channelNo]; exists {
		for symbol := range symbols {
			existingSymbols[symbol] = true
		}
	} else {
		si.symbols[channelNo] = make(map[string]bool)
		for symbol := range symbols {
			si.symbols[channelNo][symbol] = true
		}
	}
	si.modeInfo[channelNo] = mode
}

func (si *SubscriptionInfo) Unsubscribe(symbols map[string]bool, channelNo string) {
	si.mu.Lock()
	defer si.mu.Unlock()

	if existingSymbols, exists := si.symbols[channelNo]; exists {
		for symbol := range symbols {
			delete(existingSymbols, symbol)
		}
		if len(existingSymbols) == 0 {
			delete(si.symbols, channelNo)
		}
	}
}

func (si *SubscriptionInfo) UpdateChannels(pauseChannels, resumeChannels map[string]bool) {
	si.mu.Lock()
	defer si.mu.Unlock()

	for channel := range pauseChannels {
		delete(si.activeChannels, channel)
	}
	for channel := range resumeChannels {
		si.activeChannels[channel] = true
	}
}

func (si *SubscriptionInfo) GetSymbolsInfo(chanNo string) map[string]bool {
	si.mu.RLock()
	defer si.mu.RUnlock()

	if symbols, exists := si.symbols[chanNo]; exists {
		result := make(map[string]bool)
		for symbol := range symbols {
			result[symbol] = true
		}
		return result
	}
	return make(map[string]bool)
}

func (si *SubscriptionInfo) GetModeInfo(chanNo string) SubscriptionModes {
	si.mu.RLock()
	defer si.mu.RUnlock()

	if mode, exists := si.modeInfo[chanNo]; exists {
		return mode
	}
	return DEPTH
}

func (si *SubscriptionInfo) GetChannelInfo() map[string]bool {
	si.mu.RLock()
	defer si.mu.RUnlock()

	result := make(map[string]bool)
	for channel := range si.activeChannels {
		result[channel] = true
	}
	return result
}

type DataStore struct {
	depth map[string]*Depth
	mu    sync.RWMutex
}

func NewDataStore() *DataStore {
	return &DataStore{
		depth: make(map[string]*Depth),
	}
}

func (ds *DataStore) UpdateDepth(packet map[string]interface{}, cb func(string, *Depth), diffOnly bool) {
	if feeds, exists := packet["feeds"]; exists {
		if feedsMap, ok := feeds.(map[string]interface{}); ok {
			for _, value := range feedsMap {
				if feed, ok := value.(map[string]interface{}); ok {
					if ticker, exists := feed["ticker"]; exists {
						symbol := ticker.(string)

						ds.mu.Lock()
						if _, exists := ds.depth[symbol]; !exists {
							ds.depth[symbol] = NewDepth()
						}
						ds.mu.Unlock()

						if !diffOnly {

							ds.mu.Lock()
							ds.updateDepthFromFeed(ds.depth[symbol], feed, packet["snapshot"].(bool))
							ds.mu.Unlock()
							cb(symbol, ds.depth[symbol])
						} else {

							depth := NewDepth()
							ds.updateDepthFromFeed(depth, feed, packet["snapshot"].(bool))
							cb(symbol, depth)
						}
					}
				}
			}
		}
	}
}

func (ds *DataStore) updateDepthFromFeed(depth *Depth, feed map[string]interface{}, isSnapshot bool) {
	depth.Snapshot = isSnapshot

	if depthData, exists := feed["depth"]; exists {
		if depthMap, ok := depthData.(map[string]interface{}); ok {
			if tbq, exists := depthMap["tbq"]; exists {
				if tbqMap, ok := tbq.(map[string]interface{}); ok {
					if value, exists := tbqMap["value"]; exists {
						depth.Tbq = int(value.(float64))
					}
				}
			}

			if tsq, exists := depthMap["tsq"]; exists {
				if tsqMap, ok := tsq.(map[string]interface{}); ok {
					if value, exists := tsqMap["value"]; exists {
						depth.Tsq = int(value.(float64))
					}
				}
			}

			if asks, exists := depthMap["asks"]; exists {
				if asksArray, ok := asks.([]interface{}); ok {
					for i, ask := range asksArray {
						if i >= 50 {
							break
						}
						if askMap, ok := ask.(map[string]interface{}); ok {
							if price, exists := askMap["price"]; exists {
								if priceMap, ok := price.(map[string]interface{}); ok {
									if value, exists := priceMap["value"]; exists {
										depth.AskPrice[i] = value.(float64) / 100
									}
								}
							}
							if qty, exists := askMap["qty"]; exists {
								if qtyMap, ok := qty.(map[string]interface{}); ok {
									if value, exists := qtyMap["value"]; exists {
										depth.AskQty[i] = value.(float64)
									}
								}
							}
							if nord, exists := askMap["nord"]; exists {
								if nordMap, ok := nord.(map[string]interface{}); ok {
									if value, exists := nordMap["value"]; exists {
										depth.AskOrdn[i] = value.(float64)
									}
								}
							}
						}
					}
				}
			}

			if bids, exists := depthMap["bids"]; exists {
				if bidsArray, ok := bids.([]interface{}); ok {
					for i, bid := range bidsArray {
						if i >= 50 {
							break
						}
						if bidMap, ok := bid.(map[string]interface{}); ok {
							if price, exists := bidMap["price"]; exists {
								if priceMap, ok := price.(map[string]interface{}); ok {
									if value, exists := priceMap["value"]; exists {
										depth.BidPrice[i] = value.(float64) / 100
									}
								}
							}
							if qty, exists := bidMap["qty"]; exists {
								if qtyMap, ok := qty.(map[string]interface{}); ok {
									if value, exists := qtyMap["value"]; exists {
										depth.BidQty[i] = value.(float64)
									}
								}
							}
							if nord, exists := bidMap["nord"]; exists {
								if nordMap, ok := nord.(map[string]interface{}); ok {
									if value, exists := nordMap["value"]; exists {
										depth.BidOrdn[i] = value.(float64)
									}
								}
							}
						}
					}
				}
			}
		}
	}

	if feedTime, exists := feed["feed_time"]; exists {
		if feedTimeMap, ok := feedTime.(map[string]interface{}); ok {
			if value, exists := feedTimeMap["value"]; exists {
				depth.Timestamp = int64(value.(float64))
			}
		}
	}

	if sendTime, exists := feed["send_time"]; exists {
		if sendTimeMap, ok := sendTime.(map[string]interface{}); ok {
			if value, exists := sendTimeMap["value"]; exists {
				depth.SendTime = int64(value.(float64))
			}
		}
	}

	if seqNo, exists := feed["sequence_no"]; exists {
		depth.SeqNo = int64(seqNo.(float64))
	}
}

func getURL(accessToken string) string {
	req, err := http.NewRequest("GET", "https://api-t1.fyers.in/indus/home/tbtws", nil)
	if err != nil {
		return "wss://rtsocket-api.fyers.in/versova"
	}

	req.Header.Set("Authorization", accessToken)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "wss://rtsocket-api.fyers.in/versova"
	}
	defer resp.Body.Close()

	if resp.StatusCode == 200 {
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return "wss://rtsocket-api.fyers.in/versova"
		}

		var data map[string]interface{}
		err = json.Unmarshal(body, &data)
		if err != nil {
			return "wss://rtsocket-api.fyers.in/versova"
		}

		if dataMap, ok := data["data"].(map[string]interface{}); ok {
			if socketURL, exists := dataMap["socket_url"]; exists {
				return socketURL.(string)
			}
		}
	}

	return "wss://rtsocket-api.fyers.in/versova"
}

type FyersTbtSocket struct {
	datastore            *DataStore
	subsInfo             *SubscriptionInfo
	accessToken          string
	logPath              string
	wsObject             *websocket.Conn
	wsRun                bool
	pingThread           *time.Ticker
	writeToFile          bool
	backgroundFlag       bool
	reconnectDelay       int
	onDepthUpdate        func(string, *Depth)
	onErrorMsg           func(string)
	onError              func(map[string]interface{})
	onConnect            func()
	onClose              func(map[string]interface{})
	onOpen               func()
	restartFlag          bool
	diffOnly             bool
	maxReconnectAttempts int
	reconnectAttempts    int
	tbtLogger            *FyersLogger
	mu                   sync.Mutex
	connected            bool
	stopChan             chan bool
}

func NewFyersTbtSocket(
	accessToken string,
	writeToFile bool,
	logPath string,
	onDepthUpdate func(string, *Depth),
	onErrorMessage func(string),
	onError func(map[string]interface{}),
	onConnect func(),
	onClose func(map[string]interface{}),
	onOpen func(),
	reconnect bool,
	diffOnly bool,
	reconnectRetry int,
) *FyersTbtSocket {

	maxReconnectAttempts := 50
	if reconnectRetry < maxReconnectAttempts {
		maxReconnectAttempts = reconnectRetry
	}

	var loggerPath string
	if logPath != "" {
		loggerPath = logPath + "/fyersTbtSocket.log"
	} else {
		loggerPath = "fyersTbtSocket.log"
	}

	tbtLogger := NewFyersLogger("FyersTbtSocket", "DEBUG", 2, loggerPath)

	return &FyersTbtSocket{
		datastore:            NewDataStore(),
		subsInfo:             NewSubscriptionInfo(),
		accessToken:          accessToken,
		logPath:              logPath,
		wsObject:             nil,
		wsRun:                false,
		pingThread:           nil,
		writeToFile:          writeToFile,
		backgroundFlag:       false,
		reconnectDelay:       0,
		onDepthUpdate:        onDepthUpdate,
		onErrorMsg:           onErrorMessage,
		onError:              onError,
		onConnect:            onConnect,
		onClose:              onClose,
		onOpen:               onOpen,
		restartFlag:          reconnect,
		diffOnly:             diffOnly,
		maxReconnectAttempts: maxReconnectAttempts,
		reconnectAttempts:    0,
		tbtLogger:            tbtLogger,
		connected:            false,
		stopChan:             make(chan bool),
	}
}

func (f *FyersTbtSocket) Subscribe(symbolTickers map[string]bool, channelNo string, mode SubscriptionModes) {
	f.subsInfo.Subscribe(symbolTickers, channelNo, mode)

	msg := map[string]interface{}{
		"type":    "subscribe",
		"symbols": symbolTickers,
		"channel": channelNo,
		"mode":    mode,
	}

	jsonData, _ := json.Marshal(msg)
	if f.wsObject != nil && f.connected {
		err := f.wsObject.WriteMessage(websocket.TextMessage, jsonData)
		if err != nil {
			f.OnError(err.Error())
		}
	}
}

func (f *FyersTbtSocket) Unsubscribe(symbolTickers map[string]bool, channelNo string, mode SubscriptionModes) {
	f.subsInfo.Unsubscribe(symbolTickers, channelNo)

	msg := map[string]interface{}{
		"type":    "unsubscribe",
		"symbols": symbolTickers,
		"channel": channelNo,
		"mode":    mode,
	}

	jsonData, _ := json.Marshal(msg)
	if f.wsObject != nil && f.connected {
		err := f.wsObject.WriteMessage(websocket.TextMessage, jsonData)
		if err != nil {
			f.OnError(err.Error())
		}
	}
}

func (f *FyersTbtSocket) SwitchChannel(resumeChannels, pauseChannels map[string]bool) {
	f.subsInfo.UpdateChannels(pauseChannels, resumeChannels)

	msg := map[string]interface{}{
		"type":            "switch_channel",
		"resume_channels": resumeChannels,
		"pause_channels":  pauseChannels,
	}

	jsonData, _ := json.Marshal(msg)
	if f.wsObject != nil && f.connected {
		err := f.wsObject.WriteMessage(websocket.TextMessage, jsonData)
		if err != nil {
			f.OnError(err.Error())
		}
	}
}

func (f *FyersTbtSocket) OnDepthUpdate(ticker string, message *Depth) {
	if f.onDepthUpdate != nil {
		f.onDepthUpdate(ticker, message)
	} else {
		fmt.Printf("Depth Update for %s: %s\n", ticker, message.String())
	}
}

func (f *FyersTbtSocket) OnErrorMessage(message string) {
	if f.onErrorMsg != nil {
		f.onErrorMsg(message)
	} else {
		fmt.Printf("Error Message: %s\n", message)
	}
}

func (f *FyersTbtSocket) OnError(message interface{}) {
	if f.onError != nil {
		f.onError(map[string]interface{}{"error": message})
	} else {
		fmt.Printf("Error: %v\n", message)
	}
}

func (f *FyersTbtSocket) handleMessage(message []byte) {
	var data map[string]interface{}
	err := json.Unmarshal(message, &data)
	if err != nil {
		f.OnError("Failed to parse message")
		return
	}

	if msgType, exists := data["type"]; exists {
		switch msgType {
		case "depth":
			f.datastore.UpdateDepth(data, f.OnDepthUpdate, f.diffOnly)
		case "error":
			if errorMsg, exists := data["message"]; exists {
				f.OnErrorMessage(errorMsg.(string))
			}
		default:
			f.OnError(fmt.Sprintf("Unknown message type: %v", msgType))
		}
	}
}

func (f *FyersTbtSocket) readMessages() {
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
				f.OnError(err.Error())
				return
			}

			f.handleMessage(message)
		}
	}
}

func (f *FyersTbtSocket) ping() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			if f.wsObject != nil && f.connected {
				err := f.wsObject.WriteMessage(websocket.TextMessage, []byte("ping"))
				if err != nil {
					f.OnError(err.Error())
				}
			}
		case <-f.stopChan:
			return
		}
	}
}

func (f *FyersTbtSocket) Connect() error {
	f.mu.Lock()
	defer f.mu.Unlock()

	url := getURL(f.accessToken)
	dialer := websocket.Dialer{}
	conn, _, err := dialer.Dial(url, nil)
	if err != nil {
		return err
	}

	f.wsObject = conn
	f.connected = true
	f.wsRun = true

	go f.readMessages()

	go f.ping()

	if f.onOpen != nil {
		f.onOpen()
	}

	return nil
}

func (f *FyersTbtSocket) KeepRunning() {
	select {}
}

func (f *FyersTbtSocket) StopRunning() {
	f.CloseConnection()
}

func (f *FyersTbtSocket) CloseConnection() {
	f.mu.Lock()
	defer f.mu.Unlock()

	if f.wsObject != nil {
		f.wsObject.Close()
		f.wsObject = nil
	}

	f.connected = false
	f.wsRun = false
	close(f.stopChan)

	if f.onClose != nil {
		f.onClose(map[string]interface{}{"message": "Connection closed"})
	}
}

func (f *FyersTbtSocket) IsConnected() bool {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.connected && f.wsObject != nil
}
