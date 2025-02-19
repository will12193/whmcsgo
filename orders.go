package whmcsgo

import (
	"encoding/json"
	"fmt"
)

// OrdersService provides access to the orders related functions
// in the WHMCS API.
//
// WHMCS API docs: https://developers.whmcs.com/api/api-index/
type OrdersService struct {
	client *Client
}

// Order represents a WHCMS Order for an account
type Order struct {
	Result        string  `json:"result"`
	OrderID       int     `json:"orderid"`
	ClientID      *string `json:"clientid"`
	PID           *string `json:"pid"`
	Domain        *string `json:"domain"`
	BillingCycle  *string `json:"billingcycle"`
	DomainType    *string `json:"domaintype"`
	RegPeriod     *string `json:"regperiod"`
	EppCode       *int    `json:"eppcode"`
	NameServer1   *string `json:"nameserver1"`
	PaymentMethod *string `json:"paymentmethod"`
	HostName      *string `json:"hostname"`
}

type AcceptOrder struct {
	Result string `json:"result"`
}

func (o Order) String() string {
	return Stringify(o)
}

// AddOrder adds an  new order
//
// WHMCs API docs: https://developers.whmcs.com/api-reference/addorder/
func (s *OrdersService) AddOrder(parms map[string]string) (*Order, *Response, error) {
	order := new(Order)
	resp, err := apiRequest(s.client, Params{parms: parms, u: "AddOrder"}, order)
	if err != nil {
		return nil, resp, err
	}
	return order, resp, err
}

// AcceptOrder accepts an existing order
//
// WHMCs API docs: https://developers.whmcs.com/api-reference/acceptorder/
func (s *OrdersService) AcceptOrder(parms map[string]string) (*AcceptOrder, error) {
	order := AcceptOrder{}

	resp, err := apiRequest(s.client, Params{parms: parms, u: "AcceptOrder"}, nil)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal([]byte(resp.Body), &order)

	if err != nil {
		return nil, fmt.Errorf("json.Unmarshal failed : %w", err)
	}

	if order.Result != "success" {
		return &order, fmt.Errorf("accept order did not succeed")
	}

	return &order, err
}

// GetOrders the orders for a user.  Passing the empty string will list
// orders for the authenticated user.
//
// WHMCS API docs: https://developers.whmcs.com/api-reference/getorders/
func (s *OrdersService) GetOrders(parms map[string]string) (*[]Order, *Response, error) {
	orders := new([]Order)
	resp, err := apiRequest(s.client, Params{parms: parms, u: "GetOrders"}, orders)
	if err != nil {
		return nil, resp, err
	}
	return orders, resp, err
}

// GetOrderStatuses the status of the order
//
// WHMCS API docs: https://developers.whmcs.com/api-reference/getorderstatuses/
// TO-DO this shall return *[]OrderStatus
func (s *OrdersService) GetOrderStatuses(parms map[string]string) (*Order, *Response, error) {
	order := new(Order)
	resp, err := apiRequest(s.client, Params{parms: parms, u: "GetOrderStatuses"}, order)
	if err != nil {
		return nil, resp, err
	}
	return order, resp, err
}

// CancelOrder Cancel a Pending Order
//
// WHMCS API docs: https://developers.whmcs.com/api-reference/cancelorder/
func (s *OrdersService) CancelOrder(parms map[string]string) (*Order, *Response, error) {
	order := new(Order)
	resp, err := apiRequest(s.client, Params{parms: parms, u: "CancelOrder"}, order)
	if err != nil {
		return nil, resp, err
	}
	return order, resp, err
}
