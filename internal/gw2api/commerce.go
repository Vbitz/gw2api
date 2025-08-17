package gw2api

import "time"

// Price represents trading post price information for an item
type Price struct {
	ID          int       `json:"id"`
	Whitelisted bool      `json:"whitelisted"`
	Buys        PriceInfo `json:"buys"`
	Sells       PriceInfo `json:"sells"`
}

// PriceInfo represents buy/sell price and quantity information
type PriceInfo struct {
	Quantity  int `json:"quantity"`
	UnitPrice int `json:"unit_price"`
}

// Listing represents detailed trading post listings for an item
type Listing struct {
	ID    int           `json:"id"`
	Buys  []ListingInfo `json:"buys"`
	Sells []ListingInfo `json:"sells"`
}

// ListingInfo represents individual buy/sell listing information
type ListingInfo struct {
	Listings  int `json:"listings"`
	UnitPrice int `json:"unit_price"`
	Quantity  int `json:"quantity"`
}

// ExchangeRate represents gem/gold exchange rate information
type ExchangeRate struct {
	CoinsPerGem int `json:"coins_per_gem"`
	Quantity    int `json:"quantity"`
}

// Transaction represents a trading post transaction
type Transaction struct {
	ID        int       `json:"id"`
	ItemID    int       `json:"item_id"`
	Price     int       `json:"price"`
	Quantity  int       `json:"quantity"`
	Created   time.Time `json:"created"`
	Purchased time.Time `json:"purchased,omitempty"`
}

// DeliveryItem represents an item available for pickup
type DeliveryItem struct {
	ID       int `json:"id"`
	Count    int `json:"count"`
	ItemID   int `json:"item_id"`
	Quantity int `json:"quantity"`
}
