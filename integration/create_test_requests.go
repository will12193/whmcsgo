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
		fmt.Printf("Created test product. ProductID: %d\n", prod.Pid)
		return &prod.Pid, err
	} else {
		return nil, fmt.Errorf("error, AddProduct returned status of: %d\n", response.StatusCode)
	}
}

// Creates a test client
func createTestClient(whmcs *whmcsgo.Client) (*whmcsgo.Account, error) {
	_, response, err := whmcs.Accounts.AddClient(
		map[string]string{
			"firstname": "Test", "lastname": "Dude", "companyname": "test corp", "email": "testdudes@divisia.io",
			"address1": "123 Fake Street", "city": "Brisbane", "state": "Queensland", "postcode": "4000",
			"country": "AU", "phonenumber": "1234123123", "password2": "4me2test",
		},
	)
	if err != nil {
		return nil, fmt.Errorf("whmcs.Accounts.AddClient failed: %w", err)
	}

	if response.StatusCode == 201 || response.StatusCode == 200 {
		client, _, err := whmcs.Accounts.GetClientsDetails(map[string]string{"email": "testdudes@divisia.io"})
		if err != nil {
			return nil, fmt.Errorf("whmcs.Accounts.GetClientsDetails failed: %w", err)
		}
		fmt.Printf("Created test client with email: %s\n", client.Email)
		return client, err
	} else {
		return nil, fmt.Errorf("error, AddClient returned status of: %s\n", response.Status)
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
	fmt.Printf("Created test order with ID: %d\n", order.OrderID)

	return order, err
}
