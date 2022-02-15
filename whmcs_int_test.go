package whmcsgo

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/jinzhu/now"
	"github.com/kelseyhightower/envconfig"
)

/*
Integration tests to be run with an active development instance of WHMCS.

The following environment variables are required to run the tests:
export WHM_IDENT="put your api identity key here"
export WHM_SECRET="put your api secret key here"
export WHM_ACCESS="put your api access key here"
export WHM_URL="http://localhost:####/"
export WHM_PAYMENTMETHOD="payment method setup in dev env"

Prerequisites to running the tests:
- Working instance of WHMCS
- Settings -> API Credentials - Created API Role with appropriate access and create credentials
- Settings -> Payment Gateways - Atleast one payment gateway must be selected
*/

type Config struct {
	URL           string `default:"http://localhost:3000/"`
	Access        string `default:"RAND0MT3STK3Y"`
	Ident         string `default:"JnbGfwNUq1CIHxhEoqRbMKb084gcvwwz"`
	Secret        string `default:"Wx9Lqeqe0Os0paUUDtbc37k89qfpqdvZ"`
	PaymentMethod string `default:"testPaymentMethod"`
}

var whmcsConfig *Config

// Import environment variables
func init() {
	whmcsConfig = &Config{}
	if err := envconfig.Process("WHM", whmcsConfig); err != nil {
		panic(err)
	}
}

// Creates a test product
func createTestProduct(whmcs *Client) (*int, error) {
	var (
		prod Product
	)
	_, response, err := whmcs.Products.AddProduct(
		map[string]string{
			"name": "TestProduct", "gid": "1",
		},
	)
	if err != nil {
		return nil, fmt.Errorf("AddProduct Failed: %w", err)
	}
	json.Unmarshal([]byte(response.Body), &prod)

	if response.StatusCode == 201 || response.StatusCode == 200 {
		fmt.Printf("Created test product. ProductID: %d\n", prod.Pid)
		return &prod.Pid, err
	} else {
		return nil, fmt.Errorf("error, AddProduct returned status of: %d\n", response.StatusCode)
	}
}

// Creates a test client
func createTestClient(whmcs *Client) (*Account, error) {
	_, response, err := whmcs.Accounts.AddClient(
		map[string]string{
			"firstname": "Test", "lastname": "Dude", "companyname": "test corp", "email": "testdudes@divisia.io",
			"address1": "123 Fake Street", "city": "Brisbane", "state": "Queensland", "postcode": "4000",
			"country": "AU", "phonenumber": "1234123123", "password2": "4me2test",
		},
	)
	if err != nil {
		return nil, fmt.Errorf("AddClient failed: %w", err)
	}

	if response.StatusCode == 201 || response.StatusCode == 200 {
		client, _, err := whmcs.Accounts.GetClientsDetails(map[string]string{"email": "testdudes@divisia.io"})
		if err != nil {
			return nil, fmt.Errorf("GetClientDetails failed: %w", err)
		}
		fmt.Printf("Created test client with email: %s\n", client.Email)
		return client, err
	} else {
		return nil, fmt.Errorf("error, AddClient returned status of: %s\n", response.Status)
	}
}

// Adds and accepts an order
func createTestOrder(whmcs *Client, clientID int, productID int, paymentMethod string) (*Order, error) {
	// Add the order
	order, resp, err := whmcs.Orders.AddOrder(map[string]string{
		"clientid": fmt.Sprintf("%d", clientID), "paymentmethod": paymentMethod,
		"pid": fmt.Sprintf("1, %d", productID),
	})
	if err != nil {
		return nil, err
	} else if resp.StatusCode != 200 && resp.StatusCode != 201 {
		return nil, fmt.Errorf("error, AddOrder returned status of: %s\n", resp.Status)
	}
	json.Unmarshal([]byte(resp.Body), order)

	// Accept the order
	_, resp, err = whmcs.Orders.AcceptOrder(map[string]string{
		"orderid": fmt.Sprintf("%d", order.OrderID),
	})
	if err != nil {
		return nil, err
	} else if resp.StatusCode != 200 && resp.StatusCode != 201 {
		return nil, fmt.Errorf("error, AcceptOrder returned status of: %s\n", resp.Status)
	}
	// t.Logf or pass in logger ?!?
	fmt.Printf("Created test order with ID: %d\n", order.OrderID)

	return order, err
}

func loadWhmcs() (*Client, *Config) {
	auth := NewAuth(map[string]string{"identifier": whmcsConfig.Ident, "secret": whmcsConfig.Secret, "accesskey": whmcsConfig.Access})
	whmcs := NewClient(nil, auth, whmcsConfig.URL)
	return whmcs, whmcsConfig
}

func TestGetClients(t *testing.T) {
	client, whmcsConfig := loadWhmcs()
	tc, err := createTestClient(client)
	if err != nil {
		t.Error(err)
	}
	productID, err := createTestProduct(client)
	if err != nil {
		t.Error(err)
	}
	_, err = createTestOrder(client, tc.ID, *productID, whmcsConfig.PaymentMethod)
	if err != nil {
		t.Error(err)
	}

	// Test GetClients
	t.Log("Load Private Clients")
	wc, _, err := client.Accounts.GetClients(map[string]string{"sorting": "ASC", "limitstart": "0", "limitnum": "2500"})
	if err != nil {
		t.Error(err)
	}

	for _, thisCustomer := range wc.Clients.Client {
		t.Logf("\n\n%v\n", thisCustomer)

		// Test GetClientDetails
		wd, _, _ := client.Accounts.GetClientsDetails(
			map[string]string{
				"clientid": fmt.Sprintf("%d", thisCustomer.ID), "limitstart": "0", "limitnum": "500",
			},
		)
		t.Log(wd)

		// Test GetClientProducts
		wp, _, _ := client.Accounts.GetClientsProducts(
			map[string]string{
				"clientid": fmt.Sprintf("%d", thisCustomer.ID), "limitstart": "0", "limitnum": "500",
			},
		)
		t.Logf("\nProducts for %s\n", thisCustomer.Email)
		for _, thisProduct := range wp.Products.Product {
			t.Log(thisProduct)
		}
	}
}

func TestClientContactList(t *testing.T) {
	client, _ := loadWhmcs()
	_, err := createTestClient(client)
	if err != nil {
		t.Error(err)
	}

	active, err := client.Accounts.ClientContactList("Active")
	if err != nil {
		t.Error(err)
	}
	inactive, err := client.Accounts.ClientContactList("Inactive")
	if err != nil {
		t.Error(err)
	}
	t.Logf("Active Contacts:\n%v\nInactive Contacts:\n%v", active, inactive)
}

func TestCreateInvoice(t *testing.T) {
	whmcs, _ := loadWhmcs()
	client, err := createTestClient(whmcs)
	if err != nil {
		panic(err)
	}

	// Create a new Invoice
	invoice := CreateInvoiceRequest{}
	invoice.SendInvoice = false
	invoice.Status = "Draft"
	invoice.DueDate = now.EndOfMonth()

	lineitems := []InvoiceLineItems{}
	lineItem := InvoiceLineItems{}

	lineItem.ItemOrder = 1
	lineItem.ItemDescription = "This is a really cool test invoice!"
	lineItem.ItemTaxed = false
	lineItem.ItemAmount = 0
	lineitems = append(lineitems, lineItem)

	lineItem.ItemOrder = 2
	lineItem.ItemDescription = "Wow, look at this amazing test invoice ive made"
	lineItem.ItemTaxed = true
	lineItem.ItemAmount = 10
	lineitems = append(lineitems, lineItem)

	invoice.LineItems = lineitems
	supportInvoice, _, err := whmcs.Billing.CreateInvoice(client.ID, invoice)
	if err != nil {
		t.Errorf("ERROR %s", err)
	}
	t.Logf("invoice ID: %d\n", supportInvoice)
}
