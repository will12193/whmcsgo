package integration

import (
	"encoding/json"
	"fmt"

	"github.com/chrisjoyce911/whmcsgo"
)

// Creates a test product
func createTestProduct(whmcs *whmcsgo.Client) (*int, error) {
	var (
		prod whmcsgo.Product
	)
	_, response, err := whmcs.Products.AddProduct(
		map[string]string{
			"name": "TestProduct", "gid": "1",
		},
	)
	if err != nil {
		return nil, fmt.Errorf("whmcs.Products.AddProduct failed: %w", err)
	}
	err = json.Unmarshal([]byte(response.Body), &prod)
	if err != nil {
		return nil, fmt.Errorf("json.Unmarshal failed: %w", err)
	}

	if response.StatusCode == 201 || response.StatusCode == 200 {
		return &prod.Pid, err
	} else {
		return nil, fmt.Errorf("error, AddProduct returned status of: %d\n", response.StatusCode)
	}
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
		return nil, fmt.Errorf("Body Unmarshal failed: %w", err)
	}

	if response.StatusCode == 200 ||
		(apiResp.Result == "error" && apiResp.Message == "A user already exists with that email address") {
		client, _, err := whmcs.Accounts.GetClientsDetails(map[string]string{"email": email})
		if err != nil {
			return nil, fmt.Errorf("GetClientDetails failed: %w", err)
		}
		fmt.Printf("Created test client with email: %s\n", client.Email)
		return client, err
	} else {
		return nil, fmt.Errorf("error, AddClient returned status of: %+v\n", response)
	}
}

// Adds and accepts a test order for the given client
func createTestOrder(whmcs *whmcsgo.Client, clientID int, productID int, paymentMethod string) (*whmcsgo.Order, error) {
	// Add the order
	order, resp, err := whmcs.Orders.AddOrder(map[string]string{
		"clientid": fmt.Sprintf("%d", clientID), "paymentmethod": paymentMethod,
		"pid": fmt.Sprintf("1, %d", productID),
	})
	if err != nil {
		return nil, fmt.Errorf("whmcs.Orders.AddOrder failed: %w", err)
	} else if resp.StatusCode != 200 && resp.StatusCode != 201 {
		return nil, fmt.Errorf("error, AddOrder returned status of: %s\n", resp.Status)
	}
	err = json.Unmarshal([]byte(resp.Body), order)
	if err != nil {
		return nil, fmt.Errorf("json.Unmarshal failed: %w", err)
	}

	// Accept the order
	_, resp, err = whmcs.Orders.AcceptOrder(map[string]string{
		"orderid": fmt.Sprintf("%d", order.OrderID),
	})
	if err != nil {
		return nil, fmt.Errorf("whmcs.Orders.AcceptOrder failed: %w", err)
	} else if resp.StatusCode != 200 && resp.StatusCode != 201 {
		return nil, fmt.Errorf("error, AcceptOrder returned status of: %s\n", resp.Status)
	}

	return order, err
}

func deleteClient(whmcs *whmcsgo.Client, clientID int) error {
	resp, err := whmcsgo.DeleteClient(map[string]string{
		"clientid": fmt.Sprintf("%d", clientID), "deleteusers": "true",
		"deletetransactions": "true",
	})
	if err != nil {
		return fmt.Errorf("whmcs.DeleteClient failed: %w", err)
	} else if resp.StatusCode != 200 {
		return fmt.Errorf("error, DeleteClient returned status of: %s\n", resp.Status)
	}

	return nil
}
