package models

import (
	"errors"
	"time"
)

var ErrNoRecord = errors.New("models: подходящей записи не найдено")

type Snippet struct {
	ID      int
	Title   string
	Content string
	Created time.Time
	Expires time.Time
}


type SessionStartResult struct {
	Success int `json:"success"`
	Meta    struct {
		Source string `json:"_source"`
	} `json:"meta"`
	Data struct {
		SessionID string `json:"sessionId"`
		UserData  struct {
			UserID     int    `json:"userId"`
			Login      string `json:"login"`
			FirstName  string `json:"firstName"`
			LastName   string `json:"lastName"`
			MiddleName string `json:"middleName"`
			IsActive   bool   `json:"isActive"`
		} `json:"userData"`
		UserPermissions []struct {
			UserID       int    `json:"userId"`
			FunctionID   int    `json:"functionId"`
			FunctionCode string `json:"functionCode"`
			FunctionName string `json:"functionName"`
			MerchantID   int    `json:"merchantId"`
		} `json:"userPermissions"`
	} `json:"data"`
}

type KfkOrderStatus struct {
	Subject struct {
		Type string      `json:"type"`
		ID   interface{} `json:"id"`
		Name interface{} `json:"name"`
	} `json:"subject"`
	Predicate struct {
		Type string      `json:"type"`
		ID   interface{} `json:"id"`
		Name interface{} `json:"name"`
	} `json:"predicate"`
	Object struct {
		Type string      `json:"type"`
		ID   interface{} `json:"id"`
		Name interface{} `json:"name"`
	} `json:"object"`
	Attributes struct {
		Item     int64 `json:"item"`
		Order    int   `json:"order"`
		Delivery int   `json:"delivery"`
		Shipment int   `json:"shipment"`
	} `json:"attributes"`
}

type MongoDoc struct {
	ID      int
	Test   string
}
