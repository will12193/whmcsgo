package integration

import (
	"fmt"
	"testing"

	"github.com/chrisjoyce911/whmcsgo"
	"github.com/jinzhu/now"
	"github.com/kelseyhightower/envconfig"
	"github.com/stretchr/testify/assert"
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
- Setting -> Products/Services - Atleast one Product Group must be created

Whats left (not cleaned up) after running the tests:
- The test product
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

// Get Authenticated and connected to WHMCS API
func loadWhmcs() (*whmcsgo.Client, *Config) {
	auth := whmcsgo.NewAuth(map[string]string{"identifier": whmcsConfig.Ident, "secret": whmcsConfig.Secret, "accesskey": whmcsConfig.Access})
	whmcs := whmcsgo.NewClient(nil, auth, whmcsConfig.URL)
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
	assert.Greater(t, wc.Numreturned, 0)

	clientExists := false
	productExists := false
	for _, thisCustomer := range wc.Clients.Client {
		t.Logf("\n\n%v\n", thisCustomer)

		// Test GetClientDetails
		wd, _, _ := client.Accounts.GetClientsDetails(
			map[string]string{
				"clientid": fmt.Sprintf("%d", thisCustomer.ID), "limitstart": "0", "limitnum": "500",
			},
		)
		t.Log(wd)
		if wd.Email == "testdudes@divisia.io" {
			clientExists = true
		}

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
		if len(wp.Products.Product) > 0 {
			productExists = true
		}
	}
	assert.True(t, clientExists)
	assert.True(t, productExists)

	err = deleteClient(client, tc.ID)
	if err != nil {
		t.Error(err)
	}
}

func TestClientContactList(t *testing.T) {
	client, _ := loadWhmcs()
	tc, err := createTestClient(client)
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
	t.Logf("\nActive Contacts:\n%v\nInactive Contacts:\n%v", active, inactive)

	contacts := append(active, inactive...)
	assert.Greater(t, len(contacts), 0)

	err = deleteClient(client, tc.ID)
	if err != nil {
		t.Error(err)
	}
}

func TestCreateInvoice(t *testing.T) {
	client, _ := loadWhmcs()
	tc, err := createTestClient(client)
	if err != nil {
		panic(err)
	}

	// Create a new Invoice
	invoice := whmcsgo.CreateInvoiceRequest{}
	invoice.SendInvoice = false
	invoice.Status = "Draft"
	invoice.DueDate = now.EndOfMonth()

	lineitems := []whmcsgo.InvoiceLineItems{}
	lineItem := whmcsgo.InvoiceLineItems{}

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
	supportInvoice, _, err := client.Billing.CreateInvoice(tc.ID, invoice)
	if err != nil {
		t.Errorf("ERROR %s", err)
	}
	t.Logf("\ninvoice ID: %d\n", supportInvoice)

	err = deleteClient(client, tc.ID)
	if err != nil {
		t.Error(err)
	}
}
