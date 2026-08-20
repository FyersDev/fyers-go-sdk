package fyersgosdk

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
)

type Protocol interface {
	Encode() []byte
	Decode()
}

type HsmProtocol struct {
	Mode     string `json:"mode"`
	Protocol string `json:"protocol"`
	Version  string `json:"version"`
	Type     string `json:"type"`
}

type HsmProtocolV2 struct {
	Type string `json:"type"`
}

type symbolConversion struct {
	dataType        string
	accessToken     string
	logPath         string
	symbolsTokenAPI string
	dataLogger      *FyersLogger
}

var (
	indexHSMMappingCache map[string]interface{}
	indexHSMMappingMutex sync.Mutex
)

func newSymbolConversion(accessToken, dataType, logPath string) *symbolConversion {
	if strings.Contains(accessToken, ":") {
		parts := strings.Split(accessToken, ":")
		if len(parts) > 1 {
			accessToken = parts[1]
		}
	}

	if logPath != "" {
		logPath = logPath + "/"
	}

	logger := NewFyersLogger("FyersDataSocket", "DEBUG", 3, logPath)

	return &symbolConversion{
		dataType:        dataType,
		accessToken:     accessToken,
		logPath:         logPath,
		symbolsTokenAPI: "https://api-t1.fyers.in/data/symbol-token",
		dataLogger:      logger,
	}
}

// Get index mapping.
// First call get from public API, falling back to map.json on error.
// Result is cached in memory for subsequent calls.
func (sc *symbolConversion) getIndexHSMMapping(
	fallbackIndexDict map[string]interface{},
) map[string]interface{} {

	indexHSMMappingMutex.Lock()
	defer indexHSMMappingMutex.Unlock()

	// Already loaded in memory.
	if indexHSMMappingCache != nil {
		sc.dataLogger.Debug(
			"Index mapping already loaded in memory. Returning cached mapping.",
		)
		return indexHSMMappingCache
	}

	indexDict, err := sc.getIndexDictMapJSON()
	if err != nil {
		sc.dataLogger.Exception(err)
		sc.dataLogger.Debug(
			"S3 index mapping load failed. Falling back to map.json index_dict.",
		)
		// API call failed -> use local map.json.
		indexHSMMappingCache = fallbackIndexDict
	} else {
		sc.dataLogger.Debug(
			"API index mapping load succeeded. Caching API mapping.",
		)
		// API call succeeded -> cache API mapping.
		indexHSMMappingCache = indexDict
	}

	return indexHSMMappingCache
}

func (sc *symbolConversion) symbolToHSMToken(symbols []string) (map[string]string, []string, bool, string) {
	data := map[string]interface{}{
		"symbols": symbols,
	}

	jsonData, err := json.Marshal(data)
	if err != nil {
		sc.dataLogger.Exception(err)
		return nil, nil, false, ""
	}

	req, err := http.NewRequest("POST", sc.symbolsTokenAPI, bytes.NewBuffer(jsonData))
	if err != nil {
		sc.dataLogger.Exception(err)
		return nil, nil, false, ""
	}

	req.Header.Set("Authorization", sc.accessToken)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		sc.dataLogger.Exception(err)
		return nil, nil, false, ""
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		sc.dataLogger.Exception(err)
		return nil, nil, false, ""
	}

	var responseData map[string]interface{}
	err = json.Unmarshal(body, &responseData)
	if err != nil {
		sc.dataLogger.Exception(err)
		return nil, nil, false, ""
	}

	mapData, err := sc.loadMapJSON()
	if err != nil {
		sc.dataLogger.Exception(err)
		return nil, nil, false, ""
	}

	indexDict := sc.getIndexHSMMapping(mapData["index_dict"].(map[string]interface{}))
	exchSegDict := mapData["exch_seg_dict"].(map[string]interface{})

	dataDict := make(map[string]string)
	wrongSymbol := []string{}
	dpIndexFlag := false

	if responseData["s"] == "ok" {
		validSymbols := responseData["validSymbol"].(map[string]interface{})
		for symbol, fytoken := range validSymbols {
			fytokenStr := fytoken.(string)
			exSg := fytokenStr[:4]

			if _, exists := exchSegDict[exSg]; !exists {
				continue
			}

			segment := exchSegDict[exSg].(string)
			symbolSplit := strings.Split(symbol, "-")
			updateDict := true

			var hsmSymbol string
			if len(symbolSplit) > 1 && symbolSplit[len(symbolSplit)-1] == "INDEX" && sc.dataType != "DepthUpdate" {
				var exchToken string
				if indexToken, exists := indexDict[symbol]; exists {
					exchToken = indexToken.(string)
				} else {
					parts := strings.Split(symbol, ":")
					if len(parts) > 1 {
						subParts := strings.Split(parts[1], "-")
						if len(subParts) > 0 {
							exchToken = subParts[0]
						}
					}
				}
				hsmSymbol = "if|" + segment + "|" + exchToken
			} else if sc.dataType == "DepthUpdate" && symbolSplit[len(symbolSplit)-1] != "INDEX" {
				exchToken := fytokenStr[10:]
				hsmSymbol = "dp|" + segment + "|" + exchToken
			} else if sc.dataType == "SymbolUpdate" {
				exchToken := fytokenStr[10:]
				hsmSymbol = "sf|" + segment + "|" + exchToken
			} else if sc.dataType == "DepthUpdate" && symbolSplit[len(symbolSplit)-1] == "INDEX" {
				updateDict = false
				dpIndexFlag = true
			}

			if updateDict {
				dataDict[hsmSymbol] = symbol
			}
		}

		if invalidSymbols, exists := responseData["invalidSymbol"]; exists {
			if invalidList, ok := invalidSymbols.([]interface{}); ok {
				for _, symbol := range invalidList {
					wrongSymbol = append(wrongSymbol, symbol.(string))
				}
			}
		}

		return dataDict, wrongSymbol, dpIndexFlag, ""
	} else if responseData["s"] == "error" {
		message := ""
		if msg, exists := responseData["message"]; exists {
			message = msg.(string)
		}
		return nil, nil, dpIndexFlag, message
	}

	return nil, nil, false, ""
}

func (sc *symbolConversion) loadMapJSON() (map[string]interface{}, error) {
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

	return mapData, nil
}

func (sc *symbolConversion) getIndexDictMapJSON() (map[string]interface{}, error) {
	resp, err := http.Get(INDEX_HSM_MAPPING_URL)
	if err != nil {
		return nil, fmt.Errorf("failed to get index HSM mapping JSON: %w", err)
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	var mapData map[string]interface{}
	if err := json.Unmarshal(data, &mapData); err != nil {
		return nil, fmt.Errorf("failed to parse index HSM mapping JSON: %w", err)
	}

	return mapData, nil
}
