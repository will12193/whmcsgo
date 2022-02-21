package integration

import (
	"encoding/json"
	"fmt"

	"github.com/chrisjoyce911/whmcsgo"
)

// Creates a test product
func createTestProduct(whmcs *whmcsgo.Client, name string, gid int) (*int, error) {
	var (
		prod whmcsgo.Product
	)
	_, response, err := whmcs.Products.AddProduct(
		map[string]string{
			"name": name, "gid": fmt.Sprintf("%d", gid),
		},
	)
	if err != nil {
		return nil, fmt.Errorf("whmcs.Products.AddProduct failed: %w", err)
	}
	err = json.Unmarshal([]byte(response.Body), &prod)
	if err != nil {
		return nil, fmt.Errorf("json.Unmarshal failed: %w", err)
	}

	return &prod.Pid, err
}

// Creates a test client (if client with same email already exists, no new client will be made)
func createTestClient(whmcs *whmcsgo.Client) (*whmcsgo.Account, error) {
	email := "testdude@divisia.io"

	_, response, err := whmcs.Accounts.AddClient(
		map[string]string{
			"firstname": "Test", "lastname": "Dude", "companyname": "test corp", "email": email,
			"address1": "123 Fake Street", "city": "Brisbane", "state": "Queensland", "postcode": "4000",
			"country": "AU", "phonenumber": "1234123123", "password2": "4me2test",
		},
	)
	if err != nil {
		return nil, fmt.Errorf("AddClient failed: %w", err)
	}

	apiResp := struct {
		Result   string
		Message  string
		ClientID int `json:"client_id"`
		OwnerID  int `json:"owner_id"`
	}{}

	err = json.Unmarshal([]byte(response.Body), &apiResp)

	if err != nil {
		return nil, fmt.Errorf("body Unmarshal failed: %w", err)
	}

	ok := (apiResp.Result == "error" &&
		apiResp.Message == "A client already exists with that email address") ||
		apiResp.Result == "success"

	if ok {
		client, _, err := whmcs.Accounts.GetClientsDetails(map[string]string{"email": email})
		if err != nil {
			return nil, fmt.Errorf("GetClientDetails failed: %w", err)
		}

		if client == nil {
			return nil, fmt.Errorf("bah %+v", apiResp)
		}
		return client, err
	}
	return nil, fmt.Errorf("error, AddClient returned status of: %v", response)
}

// Adds and accepts a test order for the given client
func createTestOrder(whmcs *whmcsgo.Client, clientID, productID int, paymentMethod string) (*whmcsgo.Order, error) {
	// Add the order
	order, resp, err := whmcs.Orders.AddOrder(map[string]string{
		"clientid": fmt.Sprintf("%d", clientID), "paymentmethod": paymentMethod,
		"pid": fmt.Sprintf("%d, 1", productID),
	})
	if err != nil {
		return nil, fmt.Errorf("whmcs.Orders.AddOrder failed: %w", err)
	}

	err = json.Unmarshal([]byte(resp.Body), order)
	if err != nil {
		return nil, fmt.Errorf("json.Unmarshal failed: %w", err)
	}

	if order.Result != "success" {
		return nil, fmt.Errorf("order result invalid : %+v", order)
	}

	// Accept the order
	if _, err = whmcs.Orders.AcceptOrder(map[string]string{
		"orderid": fmt.Sprintf("%d", order.OrderID),
	}); err != nil {
		return nil, fmt.Errorf("whmcs.Orders.AcceptOrder failed: %w", err)
	}

	return order, err
}

func deleteClient(whmcs *whmcsgo.Client, clientID int) error {
	_, err := whmcs.Accounts.DeleteClient(map[string]string{
		"clientid": fmt.Sprintf("%d", clientID), "deleteusers": "true",
		"deletetransactions": "true",
	})
	if err != nil {
		return fmt.Errorf("whmcs.DeleteClient failed: %w", err)
	}

	return nil
}
