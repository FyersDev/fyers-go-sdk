<a href="https://fyers.in/"><img src="https://assets.fyers.in/images/logo.svg" align="right" /></a>

# Fyers Go SDK : fyers-api-v3 - v1.0.0

The official Fyers Go SDK for API-V3 Users [FYERS API](https://fyers.in/products/api/).

Fyers API is a set of REST-like APIs that provide integration with our in-house trading platform with which you can build your own customized trading applications.

## Documentation

- [Fyers API documentation](https://myapi.fyers.in/docsv3)

## 🚀 Features

- **Complete API Coverage**: Full implementation of Fyers API v3 endpoints
- **Authentication**: OAuth2-based authentication with automatic token management
- **Real-time Data**: WebSocket streaming for live market data and order updates
- **Trading Operations**: Place, modify, and cancel orders with support for various order types
- **Portfolio Management**: Access holdings, positions, and fund information
- **Market Data**: Historical data, real-time quotes, market depth, and options chain
- **Error Handling**: Comprehensive error handling with detailed error messages
- **Type Safety**: Strongly typed Go structs for all API responses
- **Examples**: Extensive examples for all major functionality
- **Testing**: Comprehensive test suite with mock responses

## 📋 Table of Contents

- [Compatible Go Versions](#compatible-go-versions)
- [Installation](#installation)
- [Quick Start](#quick-start)
- [Authentication](#authentication)
- [API Reference](#api-reference)
- [Examples](#examples)
- [WebSocket Streaming](#websocket-streaming)
- [Error Handling](#error-handling)
- [Contributing](#contributing)
- [License](#license)

## Compatible Go Versions

- **Minimum:** Go **1.18** (required by `go.mod`).
- **Compatible:** Go 1.18 and any later release (1.19, 1.20, 1.21, 1.22, 1.23, etc.).

Using Go 1.19 or newer is recommended for security and tooling support.

## 📦 Installation

```bash
git clone https://github.com/kishore-fyers/fyers-go-sdk.git
```

## 🚀 Quick Start

### Basic Setup

```go
package main

import (
    "fmt"
    "log"
    fyersgosdk "github.com/your-username/fyersgosdk"
)

func main() {
    // 1. Initialize client (auth only)
    fyClient := fyersgosdk.SetClientData("YOUR_APP_ID", "YOUR_APP_SECRET", "YOUR_REDIRECT_URL")
    fmt.Println("Login URL:", fyClient.GetLoginURL())

    // 2. After user authorizes, exchange auth code for access token
    authCode := "AUTH_CODE_FROM_REDIRECT"
    response, err := fyClient.GenerateAccessToken(authCode, fyClient)
    if err != nil {
        log.Fatal("Error generating access token:", err)
    }
    // Parse response JSON to get access_token; then create model for API calls

    // 3. Use FyersModel for all API calls (profile, orders, data, etc.)
    fyModel := fyersgosdk.NewFyersModel("YOUR_APP_ID", "ACCESS_TOKEN")
    profile, err := fyModel.GetProfile()
    if err != nil {
        log.Fatal("Error getting profile:", err)
    }
    fmt.Println("Profile:", profile)
}
```

### Get Market Data

```go
fyModel := fyersgosdk.NewFyersModel(appId, accessToken)
// Quotes (up to 50 symbols)
quotes, err := fyModel.GetStockQuotes([]string{"NSE:SBIN-EQ", "NSE:NIFTY50-INDEX"})
// History
history, err := fyModel.GetHistory(fyersgosdk.HistoryRequest{
    Symbol: "NSE:SBIN-EQ", Resolution: "30", DateFormat: "1",
    RangeFrom: "2021-01-01", RangeTo: "2021-01-02", ContFlag: "",
})
```

### Place an Order

```go
fyModel := fyersgosdk.NewFyersModel(appId, accessToken)
response, err := fyModel.SingleOrderAction(fyersgosdk.OrderRequest{
    Symbol: "NSE:IDEA-EQ", Qty: 1, Type: 1, Side: 1, ProductType: "CNC",
    LimitPrice: 100, Validity: "DAY", DisclosedQty: 0, OfflineOrder: false,
})
```

## 🔐 Authentication

The Fyers Go SDK uses OAuth2 authentication. Here's the complete flow:

### 1. App Registration
First, register your application on the [Fyers Developer Portal](https://myapi.fyers.in/):

1. Create a new app
2. Get your `App ID` and `App Secret`
3. Set your redirect URL
4. Note your `Client ID`

### 2. Authentication Flow

```go
// Step 1: Initialize client (appId, appSecret, redirectUrl)
fyClient := fyersgosdk.SetClientData(appId, appSecret, redirectUrl)

// Step 2: Generate login URL
loginURL := fyClient.GetLoginURL()
fmt.Println("Please visit:", loginURL)

// Step 3: User authorizes; you get auth code from redirect URL
// Step 4: Exchange auth code for access token (returns JSON string)
response, err := fyClient.GenerateAccessToken(authCode, fyClient)

// Step 5: Exchange refresh token for access token
response, err := fyClient.GenerateAccessTokenFromRefreshToken(refreshToken, pin, fyClient)

// Step 6: Create FyersModel for all API calls
fyModel := fyersgosdk.NewFyersModel(appId, accessToken)
```

## 📚 API Reference

All API methods return `(string, error)`; the string is the raw JSON response. Use **Client** only for auth; use **FyersModel** for everything after login.

### Authentication (Client)

| Method | Description |
|--------|-------------|
| `SetClientData(appId, appSecret, redirectUrl string) *Client` | Create client for login/refresh. |
| `GetLoginURL() string` | URL for user to authorize and get auth code. |
| `GenerateAccessToken(authToken string, fyClient *Client) (string, error)` | Exchange auth code for access token. |
| `GenerateAccessTokenFromRefreshToken(refreshToken, pin string, fyClient *Client) (string, error)` | Get new access token from refresh token. |
| `SetAccessToken(accessToken string) *Client` | Set token on client (optional). |

### User & Profile (FyersModel)

| Method | Description |
|--------|-------------|
| `GetProfile() (string, error)` | User profile. |
| `GetFunds() (string, error)` | Account funds. |
| `GetHoldings() (string, error)` | Portfolio holdings. |
| `Logout() (string, error)` | Logout session. |

### Transaction Info (FyersModel)

| Method | Description |
|--------|-------------|
| `GetOrderBook() (string, error)` | All orders. |
| `GetOrderBookByTag(tag string) (string, error)` | Orders by tag. |
| `GetOrderById(id string) (string, error)` | Single order by ID. |
| `GetPositions() (string, error)` | Open positions. |
| `GetTradeBook() (string, error)` | Trade book. |
| `GetTradeBookByTag(tag string) (string, error)` | Trades by tag. |

### Orders (FyersModel)

| Method | Description |
|--------|-------------|
| `SingleOrderAction(orderRequest OrderRequest) (string, error)` | Place single order. |
| `MultiOrderAction(orderRequests []OrderRequest) (string, error)` | Place multiple orders. |
| `MultiLegOrderAction(orderRequests []MultiLegOrderRequest) (string, error)` | Place multi-leg orders. |
| `ModifyOrder(orderRequest ModifyOrderRequest) (string, error)` | Modify single order (PATCH). |
| `ModifyMutliOrder(requests []ModifyMultiOrderItem) (string, error)` | Modify multiple orders (PATCH). |
| `CancelOrder(Id string) (string, error)` | Cancel single order. |
| `CancelMutliOrder(orderIds []string) (string, error)` | Cancel multiple orders (DELETE). |

### GTT Orders (FyersModel)

| Method | Description |
|--------|-------------|
| `GTTSingleOrderAction(orderRequest GTTOrderRequest) (string, error)` | Place GTT order. |
| `GTTMultiOrderAction(orderRequests []GTTOrderRequest) (string, error)` | Place GTT (sends first only). |
| `ModifyGTTOrder(orderRequests []ModifyGTTOrderRequest) (string, error)` | Modify GTT order. |
| `CancelGTTOrder(orderId string) (string, error)` | Cancel GTT order. |
| `GetGTTOrderBook() (string, error)` | GTT order book. |

### Trade Operations / Positions (FyersModel)

| Method | Description |
|--------|-------------|
| `ExitPosition() (string, error)` | Exit all positions (DELETE with exit_all). |
| `ExitPositionById(orderId []string) (string, error)` | Exit by order IDs. |
| `ExitPositionByProductType(req ExitPositionByProductTypeRequest) (string, error)` | Exit by segment/side/productType. |
| `CancelPendingOrders(req CancelPendingOrdersRequest) (string, error)` | Cancel pending orders (optional Id for single symbol). |
| `ConvertPosition(req ConvertPositionRequest) (string, error)` | Convert position (e.g. INTRADAY to CNC). |

### Market Data (FyersModel)

| Method | Description |
|--------|-------------|
| `GetHistory(req HistoryRequest) (string, error)` | OHLCV history (symbol, resolution, range). |
| `GetStockQuotes(symbols []string) (string, error)` | Quotes for up to 50 symbols. |
| `GetMarketDepth(req MarketDepthRequest) (string, error)` | Market depth (symbol, ohlcv_flag). |
| `GetOptionChain(req OptionChainRequest) (string, error)` | Options chain. |

### Broker (FyersModel)

| Method | Description |
|--------|-------------|
| `GetMarketStatus() (string, error)` | Market status. |

### Alerts (FyersModel)

| Method | Description |
|--------|-------------|
| `GetAlerts() (string, error)` | List price alerts. |
| `CreateAlert(req AlertRequest) (string, error)` | Create alert. |
| `UpdateAlert(alertId string, req AlertRequest) (string, error)` | Update alert. |
| `DeleteAlert(alertId string) (string, error)` | Delete alert. |
| `ToggleAlert(alertId string) (string, error)` | Toggle alert. |

### WebSocket (package-level)

| Function | Description |
|----------|-------------|
| `DataSocket(fyModel *FyersModel, req DataSocketRequest) (map[string]interface{}, error)` | Real-time data stream (quotes, depth). |
| `OrderSocket(fyClient *Client, req OrderSocketRequest) (map[string]interface{}, error)` | Real-time orders, trades, positions. |

## 🔌 WebSocket Streaming

Use **DataSocket** for live market data (quotes, depth) and **OrderSocket** for orders, trades, and positions. Both run until interrupt (e.g. Ctrl+C).

```go
// Data WebSocket (symbols, lite/full mode)
var dataSocket *fyersws.FyersDataSocket
onConnect := func() {
    dataSocket.Subscribe(symbols, datatype)
}

dataSocket = fyersws.NewFyersDataSocket(
    accessToken, // Access token in the format "appid:accesstoken"
    "",          // Log path - leave empty to auto-create logs in the current directory
    true,        // Lite mode disabled. Set to true if you want a lite response
    false,       // Save response in a log file instead of printing it
    true,        // Enable auto-reconnection to WebSocket on disconnection
    50,          // reconnectRetry: max reconnect attempts (same as Python default; cap 50)
    onConnect,   // Callback: subscribe on every connect (first + after reconnect)
    onClose,     // Callback function to handle WebSocket connection close events
    onError,     // Callback function to handle WebSocket errors
    onMessage,   // Callback function to handle incoming messages from the WebSocket
)

err := dataSocket.Connect()
if err != nil {
    fmt.Printf("failed to connect to Data Socket: %v", err)
    return
}

sigChan := make(chan os.Signal, 1)
signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

<-sigChan
fmt.Println("\nReceived interrupt signal, closing connection...")

dataSocket.CloseConnection()

// Order WebSocket (trades, positions, orders)
orderSocket := fyersws.NewFyersOrderSocket(
    accessToken,      // Access token in the format "appid:accesstoken"
    false,            // Write to file - set to true if you want to save responses to a log file
    "",               // Log path - leave empty to auto-create logs in the current directory
    onOrderTrades,    // Callback function to handle trade events
    onOrderPositions, // Callback function to handle position events
    onOrderUpdates,   // Callback function to handle order events
    onOrderGeneral,   // Callback function to handle general events
    onOrderError,     // Callback function to handle WebSocket errors
    nil,   // Callback function called when WebSocket connection is established
    onOrderClose,     // Callback function to handle WebSocket connection close events
    true,             // Enable auto-reconnection to WebSocket on disconnection
    5,                // Maximum number of reconnection attempts
)

// Establish a connection to the Fyers Order WebSocket
err := orderSocket.Connect()
if err != nil {
    fmt.Printf("failed to connect to Order Socket: %v", err)
    return
}

if len(tradeOperations) > 0 {
    orderSocket.SubscribeMultiple(tradeOperations)
}

sigChan := make(chan os.Signal, 1)
signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

<-sigChan
fmt.Println("\nReceived interrupt signal, closing connection...")

orderSocket.CloseConnection()
```

## 📁 Examples

All runnable examples live in **[examples/fyers/fyers.go](examples/fyers/fyers.go)**. Uncomment the `main()` you need and run from that directory. The file includes:

### Authentication
- **Get Auth Code URL** – `SetClientData`, `GetLoginURL`
- **Generate Access Token** – `GenerateAccessToken`
- **Generate Access Token From Refresh Token** – `GenerateAccessTokenFromRefreshToken`

### User & Profile
- **Get Profile** – `GetProfile`
- **Get Funds** – `GetFunds`
- **Get Holdings** – `GetHoldings`
- **Logout** – `Logout`

### Transaction Info
- **All Trades** – `GetTradeBook`
- **Trade Book by Tag** – `GetTradeBookByTag`
- **Get Order Book** – `GetOrderBook`
- **Get Order Book by Tag** – `GetOrderBookByTag`
- **Get Positions** – `GetPositions`

### Orders
- **Single Order Placement** – `SingleOrderAction`
- **Multi Order Placement** – `MultiOrderAction`
- **MultiLeg Order** – `MultiLegOrderAction`
- **Modify Orders** – `ModifyOrder`
- **Modify Multi Orders** – `ModifyMutliOrder`
- **Cancel Order** – `CancelOrder`
- **Multi Cancel Order** – `CancelMutliOrder`

### GTT
- **GTT Single** – `GTTSingleOrderAction`
- **GTT OCO** – `GTTMultiOrderAction`
- **GTT Modify** – `ModifyGTTOrder`
- **GTT Cancel / CancelGTT** – `CancelGTTOrder`
- **GTT Get Order Book** – `GetGTTOrderBook`

### Smart Order
- **Smart Limit** – `CreateSmartOrderLimit`
- **Smart Trail** – `CreateSmartOrderTrail`
- **Smart Step** – `CreateSmartOrderStep`
- **Modify Smart Order** – `ModifySmartOrder`
- **Cancel Smart Order** – `CancelSmartOrder`
- **Pause Smart Order** – `PauseSmartOrder`
- **Resume Smart Order** – `ResumeSmartOrder`
- **Smart Order Book** – `GetSmartOrderBookWithFilter`

### Smart Exit
- **Create Smart Exit** – `CreateSmartExitTrigger`
- **Get Smart Exit** – `GetSmartExitTrigger`
- **Update Smart Exit** – `UpdateSmartExitTrigger`
- **Activate/Deactivate Smart Exit** – `ActivateDeactivateSmartExitTrigger`

### Trade Operations / Positions
- **Exit Order** – `ExitPosition`
- **Exit Position By Id** – `ExitPositionById`
- **Exit Position by Tag** – `ExitPositionByProductType`
- **Pending Order Cancel** – `CancelPendingOrders`
- **Convert Position** – `ConvertPosition`

### Alerts
- **Create Price Alert** – `CreateAlert`
- **Get Price Alerts** – `GetAlerts`
- **Modify Price Alert** – `UpdateAlert`
- **Delete Price Alert** – `DeleteAlert`
- **Enable/Disable Price Alert** – `ToggleAlert`

### Market Data & Broker
- **Market Status** – `GetMarketStatus`
- **Quotes** – `GetStockQuotes`
- **Market depth** – `GetMarketDepth`
- **Option Chain** – `GetOptionChain`
- **Get History** – `GetHistory`