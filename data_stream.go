package fyersgosdk

import (
	"fmt"
	fyersws "fyers-go-sdk/websocket"
	"os"
	"os/signal"
	"syscall"
)

func DataSocket(fyModel *FyersModel, webSocketRequest DataSocketRequest) (map[string]interface{}, error) {

	accessTokenStr := fmt.Sprintf("%s:%s", fyModel.appId, fyModel.accessToken)

	var dataSocket *fyersws.FyersDataSocket

	onDataConnect := func() {
		dataSocket.Subscribe(webSocketRequest.Symbols, webSocketRequest.DataType)
	}

	dataSocket = fyersws.NewFyersDataSocket(
		accessTokenStr,            // Access token in the format "appid:accesstoken"
		"",                        // Log path - leave empty to auto-create logs in the current directory
		webSocketRequest.LiteMode, // Lite mode disabled. Set to true if you want a lite response
		false,                     // Save response in a log file instead of printing it
		true,                      // Enable auto-reconnection to WebSocket on disconnection
		50,                        // reconnectRetry: max reconnect attempts (same as Python default; cap 50)
		onDataConnect,             // Callback: subscribe on every connect (first + after reconnect)
		onDataClose,               // Callback function to handle WebSocket connection close events
		onDataError,               // Callback function to handle WebSocket errors
		onDataMessage,             // Callback function to handle incoming messages from the WebSocket
	)

	err := dataSocket.Connect()
	if err != nil {
		return nil, fmt.Errorf("failed to connect to Data Socket: %v", err)
	}

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	<-sigChan
	fmt.Println("\nReceived interrupt signal, closing connection...")

	dataSocket.CloseConnection()
	fmt.Println("Data Socket connection closed")

	return map[string]interface{}{
		"status":       "connected",
		"symbols":      webSocketRequest.Symbols,
		"subscription": webSocketRequest.DataType,
		"message":      "Data Socket connected and subscribed successfully",
	}, nil
}

func onDataMessage(message fyersws.DataResponse) {
	fmt.Printf("Response: %s\n", message)
}

func onDataError(message fyersws.DataError) {
	fmt.Printf("Error: %s\n", message)
}

func onDataClose(message fyersws.DataClose) {
	fmt.Printf("Connection closed: %s\n", message)
}

func OrderSocket(fyModel *FyersModel, orderSocketRequest OrderSocketRequest) (map[string]interface{}, error) {

	accessTokenStr := fmt.Sprintf("%s:%s", fyModel.appId, fyModel.accessToken)

	orderSocket := fyersws.NewFyersOrderSocket(
		accessTokenStr,   // Access token in the format "appid:accesstoken"
		false,            // Write to file - set to true if you want to save responses to a log file
		"",               // Log path - leave empty to auto-create logs in the current directory
		onOrderTrades,    // Callback function to handle trade events
		onOrderPositions, // Callback function to handle position events
		onOrderUpdates,   // Callback function to handle order events
		onOrderGeneral,   // Callback function to handle general events
		onOrderError,     // Callback function to handle WebSocket errors
		nil,              // Callback function called when WebSocket connection is established
		onOrderClose,     // Callback function to handle WebSocket connection close events
		true,             // Enable auto-reconnection to WebSocket on disconnection
		5,                // Maximum number of reconnection attempts
	)

	err := orderSocket.Connect()
	if err != nil {
		return nil, fmt.Errorf("failed to connect to Order Socket: %v", err)
	}

	if len(orderSocketRequest.TradeOperations) > 0 {
		orderSocket.SubscribeMultiple(orderSocketRequest.TradeOperations)
	}

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	<-sigChan
	fmt.Println("\nReceived interrupt signal, closing connection...")

	orderSocket.CloseConnection()
	fmt.Println("Order Socket connection closed")

	return map[string]interface{}{
		"status":        "connected",
		"subscriptions": orderSocketRequest.TradeOperations,
		"message":       "Order Socket connected and subscribed successfully",
	}, nil
}

func onOrderTrades(message fyersws.OrderMessage) {
	fmt.Printf("Trade Response: %s\n", message)
}

func onOrderPositions(message fyersws.OrderMessage) {
	fmt.Printf("Position Response: %s\n", message)
}

func onOrderUpdates(message fyersws.OrderMessage) {
	fmt.Printf("Order Response: %s\n", message)
}

func onOrderGeneral(message fyersws.OrderMessage) {
	fmt.Printf("General: %s\n", message)
}

func onOrderError(message fyersws.OrderError) {
	fmt.Printf("Error: %s\n", message)
}

func onOrderClose(message fyersws.OrderClose) {
	fmt.Printf("Response: %s\n", message)
}

