package fyersgosdk

import (
	"encoding/base64"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

type FyersDataSocket struct {
	url                  string
	accessToken          string
	hsmToken             string
	logPath              string
	lite                 bool
	maxRetry             int
	source               string
	channelNum           int
	channels             []int
	runningChannels      map[int]bool
	dataType             string
	OnMessage            func(DataResponse)
	OnError              func(DataError)
	OnOpen               func()
	OnClose              func(DataClose)
	updateTick           bool
	ackCount             int
	wsRun                *websocket.Conn
	writeToFile          bool
	backgroundFlag       bool
	updateCount          int
	liteResp             map[string]interface{}
	channelSymbol        []string
	symbolDict           map[string]string
	scripsCount          map[string]int
	scripsPerChannel     map[int][]string
	restartFlag          bool
	reconnectAttempts    int
	reconnectDelay       int
	maxReconnectAttempts int
	mu                   sync.Mutex
	connected            bool
	messageQueue         chan []byte
	stopChan             chan bool
	stopOnce             sync.Once

	scripsSym     map[uint16]string
	indexSym      map[uint16]string
	dpSym         map[uint16]string
	resp          map[string]map[string]interface{}
	fieldMappings map[string][]string

	lastLtpScrips map[string]int32
	lastLtpIndex  map[string]int32

	cachedSubscribeSymbols  []string
	cachedSubscribeDataType string
	cachedSubscribeDataDict map[string]string
}

const readDeadlineDuration = 30 * time.Second

func NewFyersDataSocket(
	accessToken string,
	logPath string,
	liteMode bool,
	writeToFile bool,
	reconnect bool,
	reconnectRetry int,
	onConnect func(),
	onClose func(DataClose),
	onError func(DataError),
	onMessage func(DataResponse),
) *FyersDataSocket {
	fieldMappings, err := loadFieldMappingsOnce()
	if err != nil {
		fmt.Printf("Failed to load field mappings: %v\n", err)
		return nil
	}
	if reconnectRetry <= 0 {
		reconnectRetry = 5
	}
	maxReconnectAttempts := 50
	if reconnectRetry < maxReconnectAttempts {
		maxReconnectAttempts = reconnectRetry
	}
	return &FyersDataSocket{
		url:                  "wss://socket.fyers.in/hsm/v1-5/prod",
		accessToken:          accessToken,
		hsmToken:             "",
		logPath:              logPath,
		lite:                 liteMode,
		maxRetry:             reconnectRetry,
		source:               "GoSDK-1.0.0",
		channelNum:           11,
		channels:             []int{},
		runningChannels:      make(map[int]bool),
		dataType:             "",
		OnMessage:            onMessage,
		OnError:              onError,
		OnOpen:               onConnect,
		OnClose:              onClose,
		updateTick:           false,
		ackCount:             0,
		wsRun:                nil,
		writeToFile:          writeToFile,
		backgroundFlag:       false,
		updateCount:          0,
		liteResp:             make(map[string]interface{}),
		channelSymbol:        []string{},
		symbolDict:           make(map[string]string),
		scripsCount:          make(map[string]int),
		scripsPerChannel:     make(map[int][]string),
		restartFlag:          reconnect,
		reconnectAttempts:    0,
		reconnectDelay:       0,
		maxReconnectAttempts: maxReconnectAttempts,
		connected:            false,
		messageQueue:         make(chan []byte, 1000),
		stopChan:             make(chan bool),
		scripsSym:            make(map[uint16]string),
		indexSym:             make(map[uint16]string),
		dpSym:                make(map[uint16]string),
		resp:                 make(map[string]map[string]interface{}),
		fieldMappings:        fieldMappings,
		lastLtpScrips:        make(map[string]int32),
		lastLtpIndex:         make(map[string]int32),
	}
}

func loadFieldMappingsOnce() (map[string][]string, error) {
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		return nil, fmt.Errorf("failed to get current file path")
	}
	dir := filepath.Dir(filename)
	mapPath := filepath.Join(dir, "map.json")
	data, err := os.ReadFile(mapPath)
	if err != nil {
		return nil, err
	}
	var mapData map[string]interface{}
	err = json.Unmarshal(data, &mapData)
	if err != nil {
		return nil, err
	}
	fieldMappings := make(map[string][]string)
	if dataVal, ok := mapData["data_val"].([]interface{}); ok {
		fieldMappings["data_val"] = make([]string, len(dataVal))
		for i, v := range dataVal {
			fieldMappings["data_val"][i] = v.(string)
		}
	}
	if indexVal, ok := mapData["index_val"].([]interface{}); ok {
		fieldMappings["index_val"] = make([]string, len(indexVal))
		for i, v := range indexVal {
			fieldMappings["index_val"][i] = v.(string)
		}
	}
	if depthVal, ok := mapData["depthvalue"].([]interface{}); ok {
		fieldMappings["depthvalue"] = make([]string, len(depthVal))
		for i, v := range depthVal {
			fieldMappings["depthvalue"][i] = v.(string)
		}
	}
	return fieldMappings, nil
}

var (
	marshalMappings     map[string][]string
	marshalMappingsOnce sync.Once
)

func getMappingsForMarshal() map[string][]string {
	marshalMappingsOnce.Do(func() {
		marshalMappings, _ = loadFieldMappingsOnce()
	})
	return marshalMappings
}

func buildOrderedKeys(resp map[string]interface{}, mappings map[string][]string) []string {
	if mappings == nil {
		keys := make([]string, 0, len(resp))
		for k := range resp {
			keys = append(keys, k)
		}
		return keys
	}
	seen := make(map[string]bool)
	ordered := make([]string, 0, len(resp))

	dataType, _ := resp["type"].(string)
	var fieldOrder []string
	switch dataType {
	case "sf", "scrips":
		fieldOrder = mappings["data_val"]
	case "if":
		fieldOrder = mappings["index_val"]
	case "dp":
		fieldOrder = mappings["depthvalue"]
	}

	for _, key := range fieldOrder {
		if _, exists := resp[key]; exists {
			ordered = append(ordered, key)
			seen[key] = true
		}
	}
	for _, key := range []string{"ch", "chp"} {
		if _, exists := resp[key]; exists && !seen[key] {
			ordered = append(ordered, key)
			seen[key] = true
		}
	}
	for _, key := range []string{"symbol", "type"} {
		if _, exists := resp[key]; exists && !seen[key] {
			ordered = append(ordered, key)
			seen[key] = true
		}
	}
	var remainder []string
	for key := range resp {
		if !seen[key] {
			remainder = append(remainder, key)
		}
	}
	sort.Strings(remainder)
	ordered = append(ordered, remainder...)
	return ordered
}

func MarshalDataResponseInOrder(resp map[string]interface{}) ([]byte, error) {
	if resp == nil {
		return []byte("null"), nil
	}
	mappings := getMappingsForMarshal()
	orderedKeys := buildOrderedKeys(resp, mappings)
	var buf strings.Builder
	buf.Grow(512)
	buf.WriteByte('{')
	for i, key := range orderedKeys {
		if i > 0 {
			buf.WriteByte(',')
		}
		keyBytes, _ := json.Marshal(key)
		buf.Write(keyBytes)
		buf.WriteByte(':')
		valBytes, err := json.Marshal(resp[key])
		if err != nil {
			return nil, err
		}
		buf.Write(valBytes)
	}
	buf.WriteByte('}')
	return []byte(buf.String()), nil
}

func FormatDataResponseInOrder(resp map[string]interface{}) string {
	b, err := MarshalDataResponseInOrder(resp)
	if err != nil {
		return fmt.Sprintf("%v", resp)
	}
	return string(b)
}

type DataResponse map[string]interface{}

func (d DataResponse) String() string {
	return FormatDataResponseInOrder(map[string]interface{}(d))
}

type DataError map[string]interface{}

func (d DataError) String() string {
	b, err := json.Marshal(map[string]interface{}(d))
	if err != nil {
		return fmt.Sprintf("%v", map[string]interface{}(d))
	}
	return string(b)
}

type DataClose map[string]interface{}

func (d DataClose) String() string {
	b, err := json.Marshal(map[string]interface{}(d))
	if err != nil {
		return fmt.Sprintf("%v", map[string]interface{}(d))
	}
	return string(b)
}

func (f *FyersDataSocket) AccessTokenToHSMToken() bool {
	if !strings.Contains(f.accessToken, ":") {
		fmt.Printf("Access token format error: expected format 'APPID:TOKEN', got: %s\n", f.accessToken)
		if f.OnError != nil {
			f.OnError(DataError{
				"type":    AUTH_TYPE,
				"code":    AUTH_ERROR_CODE,
				"message": INVALID_TOKEN,
				"s":       ERROR,
			})
		}
		return false
	}

	parts := strings.Split(f.accessToken, ":")
	if len(parts) != 2 {
		fmt.Printf("Access token format error: expected exactly one colon, got: %s\n", f.accessToken)
		if f.OnError != nil {
			f.OnError(DataError{
				"type":    AUTH_TYPE,
				"code":    AUTH_ERROR_CODE,
				"message": INVALID_TOKEN,
				"s":       ERROR,
			})
		}
		return false
	}

	tokenPart := parts[1]

	tokenParts := strings.Split(tokenPart, ".")
	if len(tokenParts) != 3 {
		fmt.Printf("JWT token format error: expected 3 parts, got %d\n", len(tokenParts))
		if f.OnError != nil {
			f.OnError(DataError{
				"type":    AUTH_TYPE,
				"code":    AUTH_ERROR_CODE,
				"message": INVALID_TOKEN,
				"s":       ERROR,
			})
		}
		return false
	}

	payloadB64 := tokenParts[1]

	if len(payloadB64)%4 != 0 {
		payloadB64 += strings.Repeat("=", 4-len(payloadB64)%4)
	}

	payloadBytes, err := base64.URLEncoding.DecodeString(payloadB64)
	if err != nil {
		fmt.Printf("Base64 decode error: %v\n", err)
		if f.OnError != nil {
			f.OnError(DataError{
				"type":    AUTH_TYPE,
				"code":    AUTH_ERROR_CODE,
				"message": INVALID_TOKEN,
				"s":       ERROR,
			})
		}
		return false
	}

	var payload map[string]interface{}
	err = json.Unmarshal(payloadBytes, &payload)
	if err != nil {
		fmt.Printf("JSON unmarshal error: %v\n", err)
		if f.OnError != nil {
			f.OnError(DataError{
				"type":    AUTH_TYPE,
				"code":    AUTH_ERROR_CODE,
				"message": INVALID_TOKEN,
				"s":       ERROR,
			})
		}
		return false
	}

	if exp, exists := payload["exp"]; exists {
		expTime := int64(exp.(float64))
		currentTime := time.Now().Unix()
		if expTime-currentTime < 0 {
			if f.OnError != nil {
				f.OnError(DataError{
					"type":    AUTH_TYPE,
					"code":    TOKEN_EXPIRED,
					"message": TOKEN_EXPIRED_MSG,
					"s":       ERROR,
				})
			}
			fmt.Printf("Token expired: exp=%d, current=%d\n", expTime, currentTime)
			return false
		}
	}

	if hsmKey, exists := payload["hsm_key"]; exists {
		f.hsmToken = hsmKey.(string)
		return true
	}

	fmt.Printf("hsm_key not found in token payload\n")
	if f.OnError != nil {
		f.OnError(DataError{
			"type":    AUTH_TYPE,
			"code":    AUTH_ERROR_CODE,
			"message": INVALID_TOKEN,
			"s":       ERROR,
		})
	}
	return false
}

func (f *FyersDataSocket) Connect() error {
	if !f.AccessTokenToHSMToken() {
		return fmt.Errorf("failed to get HSM token")
	}

	dialer := websocket.Dialer{}
	conn, _, err := dialer.Dial(f.url, nil)
	if err != nil {
		return err
	}

	f.mu.Lock()
	_ = f.reconnectAttempts > 0
	f.wsRun = conn
	f.connected = true
	f.reconnectAttempts = 0
	f.reconnectDelay = 0
	f.mu.Unlock()

	go f.processMessageQueue()

	authMsg := f.createAuthMessage()
	err = f.wsRun.WriteMessage(websocket.BinaryMessage, authMsg)
	if err != nil {
		return err
	}

	var modeMsg []byte
	if f.lite {
		modeMsg = f.createLiteModeMessage()
	} else {
		modeMsg = f.createFullModeMessage()
	}
	err = f.wsRun.WriteMessage(websocket.BinaryMessage, modeMsg)
	if err != nil {
		return err
	}

	go f.readMessages()

	if f.OnOpen != nil {
		f.OnOpen()
	}

	return nil
}

func (f *FyersDataSocket) createAuthMessage() []byte {
	bufferSize := 18 + len(f.hsmToken) + len(f.source)

	buffer := make([]byte, 0, bufferSize)

	lengthBytes := make([]byte, 2)
	binary.BigEndian.PutUint16(lengthBytes, uint16(bufferSize-2))
	buffer = append(buffer, lengthBytes...)

	buffer = append(buffer, 1)

	buffer = append(buffer, 4)

	buffer = append(buffer, 1)
	tokenLenBytes := make([]byte, 2)
	binary.BigEndian.PutUint16(tokenLenBytes, uint16(len(f.hsmToken)))
	buffer = append(buffer, tokenLenBytes...)
	buffer = append(buffer, []byte(f.hsmToken)...)

	buffer = append(buffer, 2)
	buffer = append(buffer, 0, 1)
	buffer = append(buffer, '1')

	buffer = append(buffer, 3)
	buffer = append(buffer, 0, 1)
	buffer = append(buffer, 1)

	buffer = append(buffer, 4)
	sourceLenBytes := make([]byte, 2)
	binary.BigEndian.PutUint16(sourceLenBytes, uint16(len(f.source)))
	buffer = append(buffer, sourceLenBytes...)
	buffer = append(buffer, []byte(f.source)...)

	return buffer
}

func (f *FyersDataSocket) readMessages() {
	for {
		select {
		case <-f.stopChan:
			return
		default:
		}

		f.mu.Lock()
		conn := f.wsRun
		f.mu.Unlock()
		if conn == nil {
			return
		}

		conn.SetReadDeadline(time.Now().Add(readDeadlineDuration))
		_, message, err := conn.ReadMessage()
		if err != nil {
			f.mu.Lock()
			f.wsRun = nil
			f.connected = false
			restartFlag := f.restartFlag
			maxAttempts := f.maxReconnectAttempts
			_ = f.reconnectAttempts
			f.mu.Unlock()

			if f.OnError != nil {
				errMsg := err.Error()
				if strings.Contains(errMsg, "timeout") || strings.Contains(errMsg, "deadline") {
					errMsg = "Connection timed out"
				}
				f.OnError(DataError{"error": errMsg})
			}

			if !restartFlag {
				if f.OnClose != nil {
					f.OnClose(DataClose{
						"code":    SUCCESS_CODE,
						"message": CONNECTION_CLOSED,
						"s":       SUCCESS,
					})
				}
				return
			}

			f.runReconnectLoop(maxAttempts)
			return
		}

		f.handleMessage(message)
	}
}

func (f *FyersDataSocket) runReconnectLoop(maxAttempts int) {
	for {
		select {
		case <-f.stopChan:
			return
		default:
		}

		f.mu.Lock()
		attempts := f.reconnectAttempts
		f.mu.Unlock()

		if attempts >= maxAttempts {
			if f.OnClose != nil {
				f.OnClose(DataClose{
					"code":    SUCCESS_CODE,
					"message": MAX_RECONNECT_ATTEMPTS_REACHED,
					"s":       ERROR,
				})
			}
			return
		}

		attemptNum := attempts + 1
		_ = attemptNum

		f.mu.Lock()
		delay := f.reconnectDelay
		if attempts%5 == 0 {
			f.reconnectDelay += 5
			delay = f.reconnectDelay
		}
		f.mu.Unlock()

		if delay > 0 {
			select {
			case <-f.stopChan:

				return
			case <-time.After(time.Duration(delay) * time.Second):
			}
		}

		f.mu.Lock()
		f.reconnectAttempts++
		f.scripsPerChannel[f.channelNum] = nil

		f.mu.Unlock()

		err := f.Connect()
		if err == nil {

			return
		}

	}
}

func (f *FyersDataSocket) handleMessage(message []byte) {

	if len(message) < 3 {
		if f.OnError != nil {
			f.OnError(DataError{"error": "Message too short"})
		}
		return
	}

	_, respType := binary.BigEndian.Uint16(message[:2]), message[2]

	switch respType {
	case 1:
		f.handleAuthResponse(message)
	case 4:
		f.handleSubscribeResponse(message)
	case 5:
		f.handleUnsubscribeResponse(message)
	case 6:
		f.handleDataFeedResponse(message)
	case 7, 8:
		f.handleResumePauseResponse(message, int(respType))
	case 12:
		f.handleLiteFullModeResponse(message)
	default:

	}
}

func (f *FyersDataSocket) handleAuthResponse(data []byte) {
	if len(data) < 4 {
		return
	}

	offset := 4
	if offset+1 >= len(data) {
		return
	}
	offset += 1

	if offset+2 >= len(data) {
		return
	}
	fieldLength := binary.BigEndian.Uint16(data[offset : offset+2])
	offset += 2

	if offset+int(fieldLength) > len(data) {
		return
	}
	stringVal := string(data[offset : offset+int(fieldLength)])
	offset += int(fieldLength)

	if stringVal == "K" {
		if f.OnMessage != nil {
			f.OnMessage(DataResponse{
				"type":    AUTH_TYPE,
				"code":    SUCCESS_CODE,
				"message": AUTH_SUCCESS,
				"s":       SUCCESS,
			})
		}
	} else {
		if f.OnError != nil {
			f.OnError(DataError{
				"type":    AUTH_TYPE,
				"code":    AUTH_ERROR_CODE,
				"message": AUTH_FAIL,
				"s":       ERROR,
			})
		}
	}

	if offset+1+2+4 <= len(data) {
		offset += 1
		offset += 2
		f.ackCount = int(binary.BigEndian.Uint32(data[offset : offset+4]))
	}
}

func (f *FyersDataSocket) handleSubscribeResponse(data []byte) {
	if len(data) < 5 {
		return
	}

	offset := 5
	if offset+2 >= len(data) {
		return
	}
	fieldLength := binary.BigEndian.Uint16(data[offset : offset+2])
	offset += 2

	if offset+int(fieldLength) > len(data) {
		return
	}
	stringVal := string(data[offset : offset+int(fieldLength)])

	if stringVal == "K" {
		if f.OnMessage != nil {
			f.OnMessage(DataResponse{
				"type":    SUBS_TYPE,
				"code":    SUCCESS_CODE,
				"message": SUBSCRIBE_SUCCESS,
				"s":       SUCCESS,
			})
		}
	} else {
		if f.OnError != nil {
			f.OnError(DataError{
				"type":    SUBS_TYPE,
				"code":    SUBS_ERROR_CODE,
				"message": SUBSCRIBE_FAIL,
				"s":       ERROR,
			})
		}
	}
}

func (f *FyersDataSocket) handleUnsubscribeResponse(data []byte) {
	if len(data) < 5 {
		return
	}

	offset := 5
	if offset+2 >= len(data) {
		return
	}
	fieldLength := binary.BigEndian.Uint16(data[offset : offset+2])
	offset += 2

	if offset+int(fieldLength) > len(data) {
		return
	}
	stringVal := string(data[offset : offset+int(fieldLength)])

	if stringVal == "K" {
		if f.OnMessage != nil {
			f.OnMessage(DataResponse{
				"type":    UNSUBS_TYPE,
				"code":    SUCCESS_CODE,
				"message": UNSUBSCRIBE_SUCCESS,
				"s":       SUCCESS,
			})
		}
	} else {
		if f.OnError != nil {
			f.OnError(DataError{
				"type":    UNSUBS_TYPE,
				"code":    UNSUBS_ERROR_CODE,
				"message": UNSUBSCRIBE_FAIL,
				"s":       ERROR,
			})
		}
	}
}

func (f *FyersDataSocket) handleDataFeedResponse(data []byte) {
	fieldMappings := f.fieldMappings
	if fieldMappings == nil {
		if f.OnError != nil {
			f.OnError(DataError{"error": "Field mappings not loaded"})
		}
		return
	}

	if len(data) < 9 {
		return
	}

	if f.ackCount > 0 {
		f.updateCount++
		messageNum := binary.BigEndian.Uint32(data[3:7])
		if f.updateCount == f.ackCount {
			ackMsg := f.createAcknowledgmentMessage(int(messageNum))
			f.messageQueue <- ackMsg
			f.updateCount = 0
		}
	}

	scripCount := binary.BigEndian.Uint16(data[7:9])
	offset := 9

	for i := 0; i < int(scripCount); i++ {
		if offset >= len(data) {
			break
		}

		dataType := data[offset]
		offset++

		if dataType == 83 {
			offset = f.handleSnapshotData(data, offset, fieldMappings)
		} else if dataType == 85 {
			offset = f.handleFullModeData(data, offset, fieldMappings)
		} else if dataType == 76 {
			offset = f.handleLiteModeData(data, offset, fieldMappings)
		}
	}
}

func (f *FyersDataSocket) handleSnapshotData(data []byte, offset int, fieldMappings map[string][]string) int {
	if offset+2 >= len(data) {
		return offset
	}

	topicID := binary.BigEndian.Uint16(data[offset : offset+2])
	offset += 2

	if offset+1 >= len(data) {
		return offset
	}
	topicNameLen := data[offset]
	offset++

	if offset+int(topicNameLen) > len(data) {
		return offset
	}
	topicName := string(data[offset : offset+int(topicNameLen)])
	offset += int(topicNameLen)

	var dataType string
	var fieldNames []string

	if strings.HasPrefix(topicName, "dp") {
		dataType = "depth"
		fieldNames = fieldMappings["depthvalue"]
		f.dpSym[topicID] = topicName
		f.resp[topicName] = make(map[string]interface{})
	} else if strings.HasPrefix(topicName, "if") {
		dataType = "index"
		fieldNames = fieldMappings["index_val"]
		f.indexSym[topicID] = topicName
		f.resp[topicName] = make(map[string]interface{})
	} else if strings.HasPrefix(topicName, "sf") {
		dataType = "scrips"
		fieldNames = fieldMappings["data_val"]
		f.scripsSym[topicID] = topicName
		f.resp[topicName] = make(map[string]interface{})
	} else {
		return offset
	}

	if offset+1 >= len(data) {
		return offset
	}
	fieldCount := data[offset]
	offset++

	for i := 0; i < int(fieldCount); i++ {
		if offset+4 > len(data) {
			break
		}
		value := int32(binary.BigEndian.Uint32(data[offset : offset+4]))
		offset += 4

		if value != -2147483648 && i < len(fieldNames) {
			f.resp[topicName][fieldNames[i]] = value
		}
	}

	offset += 2

	if offset+3 <= len(data) {
		multiplier := binary.BigEndian.Uint16(data[offset : offset+2])
		offset += 2
		precision := data[offset]
		offset++

		f.resp[topicName]["multiplier"] = multiplier
		f.resp[topicName]["precision"] = precision
	}

	valNames := []string{"exchange", "exchange_token", "symbol"}
	for _, valName := range valNames {
		if offset+1 <= len(data) {
			stringLen := data[offset]
			offset++
			if offset+int(stringLen) <= len(data) {
				stringData := string(data[offset : offset+int(stringLen)])
				f.resp[topicName][valName] = stringData
				offset += int(stringLen)
			}
		}
	}

	f.resp[topicName]["type"] = dataType
	if symbol, exists := f.symbolDict[topicName]; exists {
		f.resp[topicName]["symbol"] = symbol
	}

	processedResponse := f.applyPrecisionAndMultiplier(f.resp[topicName], dataType)

	if f.OnMessage != nil {
		f.OnMessage(DataResponse(processedResponse))
	}

	if dataType == "scrips" {
		if ltp, ok := f.resp[topicName]["ltp"].(int32); ok {
			f.lastLtpScrips[topicName] = ltp
		}
	} else if dataType == "index" {
		if ltp, ok := f.resp[topicName]["ltp"].(int32); ok {
			f.lastLtpIndex[topicName] = ltp
		}
	}

	return offset
}

var fullModeTickRelevantScrips = map[string]bool{
	"ltp": true, "vol_traded_today": true, "bid_size": true, "ask_size": true,
	"bid_price": true, "ask_price": true, "last_traded_qty": true, "avg_trade_price": true,
	"low_price": true, "high_price": true, "open_price": true, "prev_close_price": true,
	"OI": true, "lower_ckt": true, "upper_ckt": true,
}

var fullModeTickRelevantIndex = map[string]bool{
	"ltp": true, "prev_close_price": true, "high_price": true, "low_price": true, "open_price": true,
}

func (f *FyersDataSocket) handleFullModeData(data []byte, offset int, fieldMappings map[string][]string) int {
	if offset+2 >= len(data) {
		return offset
	}

	topicID := binary.BigEndian.Uint16(data[offset : offset+2])
	offset += 2

	if offset+1 >= len(data) {
		return offset
	}
	fieldCount := data[offset]
	offset++

	sfFlag, idxFlag, dpFlag := false, false, false
	f.updateTick = false

	for i := 0; i < int(fieldCount); i++ {
		if offset+4 > len(data) {
			break
		}
		value := int32(binary.BigEndian.Uint32(data[offset : offset+4]))
		offset += 4

		if topicName, exists := f.scripsSym[topicID]; exists {
			if i < len(fieldMappings["data_val"]) {
				fieldName := fieldMappings["data_val"][i]
				if existingValue, hasValue := f.resp[topicName][fieldName]; hasValue {
					if existingValue != value && value != -2147483648 {
						f.resp[topicName][fieldName] = value
						if fullModeTickRelevantScrips[fieldName] {
							f.updateTick = true
						}
					}
				} else if value != -2147483648 {
					f.resp[topicName][fieldName] = value
					if fullModeTickRelevantScrips[fieldName] {
						f.updateTick = true
					}
				}
			}
			sfFlag = true
		} else if topicName, exists := f.indexSym[topicID]; exists {
			if i < len(fieldMappings["index_val"]) {
				fieldName := fieldMappings["index_val"][i]
				if existingValue, hasValue := f.resp[topicName][fieldName]; hasValue {
					if existingValue != value && value != -2147483648 {
						f.resp[topicName][fieldName] = value
						if fullModeTickRelevantIndex[fieldName] {
							f.updateTick = true
						}
					}
				} else if value != -2147483648 {
					f.resp[topicName][fieldName] = value
					if fullModeTickRelevantIndex[fieldName] {
						f.updateTick = true
					}
				}
			}
			idxFlag = true
		} else if topicName, exists := f.dpSym[topicID]; exists {
			if i < len(fieldMappings["depthvalue"]) {
				fieldName := fieldMappings["depthvalue"][i]
				if existingValue, hasValue := f.resp[topicName][fieldName]; hasValue {
					if existingValue != value && value != -2147483648 {
						f.resp[topicName][fieldName] = value
						f.updateTick = true
					}
				} else if value != -2147483648 {
					f.resp[topicName][fieldName] = value
					f.updateTick = true
				}
			}
			dpFlag = true
		}
	}

	if f.updateTick {
		if sfFlag {
			if topicName, exists := f.scripsSym[topicID]; exists {

				if _, hasPrecision := f.resp[topicName]["precision"]; !hasPrecision {

					f.resp[topicName]["precision"] = uint8(2)
					f.resp[topicName]["multiplier"] = uint16(100)
				}
				processedResponse := f.applyPrecisionAndMultiplier(f.resp[topicName], "scrips")
				if f.OnMessage != nil {
					f.OnMessage(DataResponse(processedResponse))
				}
			}
		} else if idxFlag {
			if topicName, exists := f.indexSym[topicID]; exists {

				if _, hasPrecision := f.resp[topicName]["precision"]; !hasPrecision {

					f.resp[topicName]["precision"] = uint8(2)
					f.resp[topicName]["multiplier"] = uint16(100)
				}
				processedResponse := f.applyPrecisionAndMultiplier(f.resp[topicName], "index")
				if f.OnMessage != nil {
					f.OnMessage(DataResponse(processedResponse))
				}
			}
		} else if dpFlag {
			if topicName, exists := f.dpSym[topicID]; exists {

				if _, hasPrecision := f.resp[topicName]["precision"]; !hasPrecision {

					f.resp[topicName]["precision"] = uint8(2)
					f.resp[topicName]["multiplier"] = uint16(100)
				}
				processedResponse := f.applyPrecisionAndMultiplier(f.resp[topicName], "depth")
				if f.OnMessage != nil {
					f.OnMessage(DataResponse(processedResponse))
				}
			}
		}
	}

	return offset
}

type FloatSDK float64

func (f FloatSDK) MarshalJSON() ([]byte, error) {
	v := float64(f)
	if math.IsNaN(v) || math.IsInf(v, 0) {
		return json.Marshal(v)
	}
	if v == math.Trunc(v) {
		return []byte(fmt.Sprintf("%.1f", v)), nil
	}
	return json.Marshal(v)
}

func asFloat64(v interface{}) (float64, bool) {
	switch n := v.(type) {
	case int:
		return float64(n), true
	case int32:
		return float64(n), true
	case float64:
		return n, true
	case FloatSDK:
		return float64(n), true
	default:
		return 0, false
	}
}

func (f *FyersDataSocket) applyPrecisionAndMultiplier(response map[string]interface{}, dataType string) map[string]interface{} {
	precision, precisionOk := response["precision"].(uint8)
	multiplier, multiplierOk := response["multiplier"].(uint16)

	if !precisionOk || !multiplierOk {
		return response
	}

	precisionCalcValue := []string{
		"ltp", "bid_price", "ask_price", "avg_trade_price", "low_price",
		"high_price", "open_price", "prev_close_price",
	}

	newResponse := make(map[string]interface{})

	if f.lite {
		if ltp, exists := response["ltp"]; exists {
			if intValue, ok := ltp.(int32); ok {
				newResponse["ltp"] = FloatSDK(float64(intValue) / (math.Pow(10, float64(precision)) * float64(multiplier)))
			}
		}

		if symbol, exists := response["symbol"]; exists {
			symbolStr, _ := symbol.(string)
			if symbolStr != "" && !strings.Contains(symbolStr, ":") {
				if exchange, hasExchange := response["exchange"]; hasExchange {
					if exStr, ok := exchange.(string); ok && exStr != "" {
						symbolStr = exStr + ":" + symbolStr
					} else {
						symbolStr = "NSE:" + symbolStr
					}
				} else {
					symbolStr = "NSE:" + symbolStr
				}
			}
			newResponse["symbol"] = symbolStr
		}
		newResponse["type"] = "sf"
	} else {
		switch dataType {
		case "scrips":
			fieldMappings := f.fieldMappings
			if fieldMappings != nil {
				for _, fieldName := range fieldMappings["data_val"] {
					if value, exists := response[fieldName]; exists {
						if intValue, ok := value.(int32); ok {
							needsPrecision := false
							for _, precisionField := range precisionCalcValue {
								if fieldName == precisionField {
									needsPrecision = true
									break
								}
							}
							if needsPrecision && fieldName != "upper_ckt" && fieldName != "lower_ckt" {
								newResponse[fieldName] = FloatSDK(float64(intValue) / (math.Pow(10, float64(precision)) * float64(multiplier)))
							} else {
								newResponse[fieldName] = value
							}
						} else {
							newResponse[fieldName] = value
						}
					}
				}
			}
			newResponse["lower_ckt"] = 0
			newResponse["upper_ckt"] = 0
			delete(newResponse, "OI")
			delete(newResponse, "Yhigh")
			delete(newResponse, "Ylow")
			newResponse["type"] = "sf"
			if symbol, exists := newResponse["symbol"]; exists {
				symbolStr := symbol.(string)
				if !strings.Contains(symbolStr, ":") {
					symbolStr = "NSE:" + symbolStr
				}
				newResponse["symbol"] = symbolStr
			}
			if ltp, ltpOk := asFloat64(newResponse["ltp"]); ltpOk {
				if prevClose, prevCloseOk := asFloat64(newResponse["prev_close_price"]); prevCloseOk && prevClose != 0 {
					change := ltp - prevClose
					changePercent := (change / prevClose) * 100
					newResponse["ch"] = FloatSDK(math.Round(change*10000) / 10000)
					newResponse["chp"] = FloatSDK(math.Round(changePercent*10000) / 10000)
				}
			}
		case "index":
			fieldMappings := f.fieldMappings
			if fieldMappings != nil {
				for _, fieldName := range fieldMappings["index_val"] {
					if value, exists := response[fieldName]; exists {
						if intValue, ok := value.(int32); ok {
							needsPrecision := false
							for i, indexField := range fieldMappings["index_val"] {
								if fieldName == indexField && (i == 0 || i == 1 || i == 3 || i == 4 || i == 5) {
									needsPrecision = true
									break
								}
							}
							if needsPrecision {
								newResponse[fieldName] = FloatSDK(float64(intValue) / (math.Pow(10, float64(precision)) * float64(multiplier)))
							} else {
								newResponse[fieldName] = value
							}
						} else {
							newResponse[fieldName] = value
						}
					}
				}
			}
			newResponse["type"] = "if"
			if ltp, ltpOk := asFloat64(newResponse["ltp"]); ltpOk {
				if prevClose, prevCloseOk := asFloat64(newResponse["prev_close_price"]); prevCloseOk && prevClose != 0 {
					change := ltp - prevClose
					changePercent := (change / prevClose) * 100
					newResponse["ch"] = FloatSDK(math.Round(change*100) / 100)
					newResponse["chp"] = FloatSDK(math.Round(changePercent*100) / 100)
				}
			}
		case "depth":
			fieldMappings := f.fieldMappings
			if fieldMappings != nil {
				for i, fieldName := range fieldMappings["depthvalue"] {
					if value, exists := response[fieldName]; exists {
						if intValue, ok := value.(int32); ok {
							if i < 10 {
								newResponse[fieldName] = FloatSDK(float64(intValue) / (math.Pow(10, float64(precision)) * float64(multiplier)))
							} else {
								newResponse[fieldName] = value
							}
						} else {
							newResponse[fieldName] = value
						}
					}
				}
			}
			newResponse["type"] = "dp"
		}
	}
	return newResponse
}

func (f *FyersDataSocket) createAcknowledgmentMessage(messageNumber int) []byte {
	buffer := make([]byte, 0, 10)

	buffer = append(buffer, 0, 0)

	buffer = append(buffer, 6)

	msgNumBytes := make([]byte, 4)
	binary.BigEndian.PutUint32(msgNumBytes, uint32(messageNumber))
	buffer = append(buffer, msgNumBytes...)

	length := len(buffer) - 2
	binary.BigEndian.PutUint16(buffer[:2], uint16(length))

	return buffer
}

func (f *FyersDataSocket) processMessageQueue() {
	for {
		select {
		case msg := <-f.messageQueue:
			if f.wsRun != nil && f.connected {
				err := f.wsRun.WriteMessage(websocket.BinaryMessage, msg)
				if err != nil {
					if f.OnError != nil {
						f.OnError(DataError{"error": err.Error()})
					}
				}
			}
		case <-f.stopChan:
			return
		}
	}
}

func (f *FyersDataSocket) createLiteModeMessage() []byte {
	buffer := make([]byte, 0, 20)

	buffer = append(buffer, 0, 0)

	buffer = append(buffer, 12)

	buffer = append(buffer, 2)

	buffer = append(buffer, 1)
	buffer = append(buffer, 0, 8)
	channelBits := uint64(1) << f.channelNum
	channelBytes := make([]byte, 8)
	binary.BigEndian.PutUint64(channelBytes, channelBits)
	buffer = append(buffer, channelBytes...)

	buffer = append(buffer, 2)
	buffer = append(buffer, 0, 1)
	buffer = append(buffer, 76)

	length := len(buffer) - 2
	binary.BigEndian.PutUint16(buffer[:2], uint16(length))

	return buffer
}

func (f *FyersDataSocket) createFullModeMessage() []byte {
	buffer := make([]byte, 0, 20)

	buffer = append(buffer, 0, 0)

	buffer = append(buffer, 12)

	buffer = append(buffer, 2)

	buffer = append(buffer, 1)
	buffer = append(buffer, 0, 8)
	channelBits := uint64(1) << f.channelNum
	channelBytes := make([]byte, 8)
	binary.BigEndian.PutUint64(channelBytes, channelBits)
	buffer = append(buffer, channelBytes...)

	buffer = append(buffer, 2)
	buffer = append(buffer, 0, 1)
	buffer = append(buffer, 70)

	length := len(buffer) - 2
	binary.BigEndian.PutUint16(buffer[:2], uint16(length))

	return buffer
}

func (f *FyersDataSocket) createSubscriptionMessage(symbols map[string]string) []byte {

	symbolsData := make([]byte, 0)
	symbolsData = append(symbolsData, 0, 0)

	symbolCount := 0
	for hsmSymbol := range symbols {
		symbolBytes := []byte(hsmSymbol)
		symbolsData = append(symbolsData, byte(len(symbolBytes)))
		symbolsData = append(symbolsData, symbolBytes...)
		symbolCount++
	}

	binary.BigEndian.PutUint16(symbolsData[:2], uint16(symbolCount))

	dataLen := 18 + len(symbolsData) + len(f.accessToken) + len(f.source)

	buffer := make([]byte, 0, dataLen)

	lengthBytes := make([]byte, 2)
	binary.BigEndian.PutUint16(lengthBytes, uint16(dataLen))
	buffer = append(buffer, lengthBytes...)

	buffer = append(buffer, 4)

	buffer = append(buffer, 2)

	buffer = append(buffer, 1)
	symbolsLenBytes := make([]byte, 2)
	binary.BigEndian.PutUint16(symbolsLenBytes, uint16(len(symbolsData)))
	buffer = append(buffer, symbolsLenBytes...)
	buffer = append(buffer, symbolsData...)

	buffer = append(buffer, 2)
	buffer = append(buffer, 0, 1)
	buffer = append(buffer, byte(f.channelNum))

	return buffer
}

func (f *FyersDataSocket) createUnsubscriptionMessage(symbols map[string]string) []byte {

	symbolsData := make([]byte, 0)
	symbolsData = append(symbolsData, 0, 0)

	symbolCount := 0
	for hsmSymbol := range symbols {
		symbolBytes := []byte(hsmSymbol)
		symbolsData = append(symbolsData, byte(len(symbolBytes)))
		symbolsData = append(symbolsData, symbolBytes...)
		symbolCount++
	}

	binary.BigEndian.PutUint16(symbolsData[:2], uint16(symbolCount))

	dataLen := 18 + len(symbolsData) + len(f.accessToken) + len(f.source)

	buffer := make([]byte, 0, dataLen)

	lengthBytes := make([]byte, 2)
	binary.BigEndian.PutUint16(lengthBytes, uint16(dataLen))
	buffer = append(buffer, lengthBytes...)

	buffer = append(buffer, 5)

	buffer = append(buffer, 2)

	buffer = append(buffer, 1)
	symbolsLenBytes := make([]byte, 2)
	binary.BigEndian.PutUint16(symbolsLenBytes, uint16(len(symbolsData)))
	buffer = append(buffer, symbolsLenBytes...)
	buffer = append(buffer, symbolsData...)

	buffer = append(buffer, 2)
	buffer = append(buffer, 0, 1)
	buffer = append(buffer, byte(f.channelNum))

	return buffer
}

func slicesEqual(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func (f *FyersDataSocket) Subscribe(symbols []string, dataType string) {
	f.dataType = dataType

	sc := newSymbolConversion(f.accessToken, dataType, f.logPath)
	dataDict, wrongSymbols, dpIndexFlag, errorMsg := sc.symbolToHSMToken(symbols)

	if errorMsg != "" || dataDict == nil {

		f.mu.Lock()
		cached := f.cachedSubscribeDataDict != nil && f.cachedSubscribeDataType == dataType && slicesEqual(f.cachedSubscribeSymbols, symbols)
		dataDictToUse := f.cachedSubscribeDataDict
		f.mu.Unlock()
		if cached && dataDictToUse != nil {
			f.mu.Lock()
			for hsm, userSym := range dataDictToUse {
				f.symbolDict[hsm] = userSym
			}
			f.mu.Unlock()
			msg := f.createSubscriptionMessage(dataDictToUse)
			f.messageQueue <- msg
			return
		}
		if f.OnError != nil {
			f.OnError(DataError{"error": errorMsg})
		}
		return
	}

	f.mu.Lock()
	f.cachedSubscribeSymbols = make([]string, len(symbols))
	copy(f.cachedSubscribeSymbols, symbols)
	f.cachedSubscribeDataType = dataType
	f.cachedSubscribeDataDict = make(map[string]string, len(dataDict))
	for k, v := range dataDict {
		f.cachedSubscribeDataDict[k] = v
	}
	for hsm, userSym := range dataDict {
		f.symbolDict[hsm] = userSym
	}
	f.mu.Unlock()

	if len(wrongSymbols) > 0 {
		if f.OnError != nil {
			f.OnError(DataError{"invalid_symbols": wrongSymbols})
		}
	}

	if dpIndexFlag {
		if f.OnError != nil {
			f.OnError(DataError{"error": INDEX_DEPTH_ERROR_MESSAGE})
		}
	}

	msg := f.createSubscriptionMessage(dataDict)
	f.messageQueue <- msg
}

func (f *FyersDataSocket) Unsubscribe(symbols []string, dataType string) {

	sc := newSymbolConversion(f.accessToken, dataType, f.logPath)
	dataDict, _, _, _ := sc.symbolToHSMToken(symbols)

	msg := f.createUnsubscriptionMessage(dataDict)
	f.messageQueue <- msg
}

func (f *FyersDataSocket) KeepRunning() {

	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			if f.wsRun != nil && f.connected {
				err := f.wsRun.WriteMessage(websocket.TextMessage, []byte("ping"))
				if err != nil {
					if f.OnError != nil {
						f.OnError(DataError{"error": err.Error()})
					}
				}
			}
		case <-f.stopChan:
			return
		}
	}
}

func (f *FyersDataSocket) CloseConnection() {

	f.mu.Lock()
	f.restartFlag = false
	if f.wsRun != nil {
		f.wsRun.Close()
		f.wsRun = nil
	}
	f.connected = false
	f.mu.Unlock()

	f.stopOnce.Do(func() { close(f.stopChan) })

	if f.OnClose != nil {
		f.OnClose(DataClose{
			"code":    SUCCESS_CODE,
			"message": CONNECTION_CLOSED,
			"s":       SUCCESS,
		})
	}
}

func (f *FyersDataSocket) IsConnected() bool {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.connected && f.wsRun != nil
}

func (f *FyersDataSocket) handleResumePauseResponse(data []byte, respType int) {

	if f.OnMessage != nil {
		messageType := "resume"
		if respType == 7 {
			messageType = "pause"
		}
		f.OnMessage(DataResponse{
			"type": messageType,
			"data": data,
		})
	}
}

func (f *FyersDataSocket) handleLiteFullModeResponse(data []byte) {
	if len(data) < 4 {
		return
	}

	offset := 3
	if offset+1 >= len(data) {
		return
	}
	fieldCount := data[offset]
	offset += 1

	if fieldCount >= 1 && offset+1+2 < len(data) {
		offset += 1
		fieldLength := binary.BigEndian.Uint16(data[offset : offset+2])
		offset += 2

		if offset+int(fieldLength) <= len(data) {
			stringVal := string(data[offset : offset+int(fieldLength)])

			if stringVal == "K" {
				if f.OnMessage != nil {
					messageType := FULL_MODE_TYPE
					message := FULL_MODE
					if f.lite {
						messageType = LITE_MODE_TYPE
						message = LITE_MODE
					}
					f.OnMessage(DataResponse{
						"type":    messageType,
						"code":    SUCCESS_CODE,
						"message": message,
						"s":       SUCCESS,
					})
				}
			} else {
				if f.OnError != nil {
					f.OnError(DataError{
						"code":    MODE_ERROR_CODE,
						"message": MODE_CHANGE_ERROR,
						"s":       ERROR,
					})
				}
			}
		}
	}
}

func (f *FyersDataSocket) handleLiteModeData(data []byte, offset int, fieldMappings map[string][]string) int {
	if offset+2 >= len(data) {
		return offset
	}

	topicID := binary.BigEndian.Uint16(data[offset : offset+2])
	offset += 2

	if topicName, exists := f.scripsSym[topicID]; exists {
		if offset+4 <= len(data) {
			value := int32(binary.BigEndian.Uint32(data[offset : offset+4]))
			offset += 4

			if value != -2147483648 && len(fieldMappings["data_val"]) > 0 {
				last := f.lastLtpScrips[topicName]
				if last != value {
					f.lastLtpScrips[topicName] = value
					fieldName := fieldMappings["data_val"][0]
					if f.resp[topicName] == nil {
						f.resp[topicName] = make(map[string]interface{})
					}
					f.resp[topicName][fieldName] = value
					f.resp[topicName]["type"] = "sf"
					processedResponse := f.applyPrecisionAndMultiplier(f.resp[topicName], "scrips")
					if f.OnMessage != nil {
						f.OnMessage(DataResponse(processedResponse))
					}
				}
			}
		}
	} else if topicName, exists := f.indexSym[topicID]; exists {
		if offset+4 <= len(data) {
			value := int32(binary.BigEndian.Uint32(data[offset : offset+4]))
			offset += 4

			if value != -2147483648 && len(fieldMappings["index_val"]) > 0 {
				last := f.lastLtpIndex[topicName]
				if last != value {
					f.lastLtpIndex[topicName] = value
					fieldName := fieldMappings["index_val"][0]
					if f.resp[topicName] == nil {
						f.resp[topicName] = make(map[string]interface{})
					}
					f.resp[topicName][fieldName] = value
					f.resp[topicName]["type"] = "if"
					processedResponse := f.applyPrecisionAndMultiplier(f.resp[topicName], "index")
					if f.OnMessage != nil {
						f.OnMessage(DataResponse(processedResponse))
					}
				}
			}
		}
	}

	return offset
}
