package integration

import (
	"fmt"
	"log"
	"testing"

	"github.com/chrisjoyce911/whmcsgo"
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
- Setting -> Products/Services - Atleast one Product Group must be created
*/

type Config struct {
	URL           string `default:""`
	Access        string `default:""`
	Ident         string `default:""`
	Secret        string `default:""`
	PaymentMethod string `default:""`
}

// Import environment variables
func whmcs() (whmcsConfig *Config, err error) {
	whmcsConfig = &Config{}
	if err = envconfig.Process("WHM", whmcsConfig); err != nil {
		return nil, fmt.Errorf("envconfig.Process %v: %w", "WHM", err)
	}
	return whmcsConfig, err
}

// Get Authenticated and connected to WHMCS API
func loadWhmcs() (*whmcsgo.Client, *Config) {
	whmcsConfig, err := whmcs()
	if err != nil {
		panic(err)
	}

	auth := whmcsgo.NewAuth(map[string]string{"identifier": whmcsConfig.Ident, "secret": whmcsConfig.Secret, "accesskey": whmcsConfig.Access})
	whmcs := whmcsgo.NewClient(nil, auth, whmcsConfig.URL)
	return whmcs, whmcsConfig
}

func TestGetClients(t *testing.T) {
	whmcs, whmcsConfig := loadWhmcs()
	tc, err := createTestClient(whmcs)
	if err != nil {
		panic(err)
	}
	productID, err := createTestProduct(whmcs)
	if err != nil {
		panic(err)
	}
	_, err = createTestOrder(whmcs, tc.ID, *productID, whmcsConfig.PaymentMethod)
	if err != nil {
		panic(err)
	}

	// Test GetClients
	fmt.Println("Load Private Clients")
	wc, _, err := whmcs.Accounts.GetClients(map[string]string{"sorting": "ASC", "limitstart": "0", "limitnum": "2500"})
	if err != nil {
		panic(err)
	}

	for _, thisCustomer := range wc.Clients.Client {
		fmt.Printf("\n\n%v\n", thisCustomer)

		// Test GetClientDetails
		wd, _, _ := whmcs.Accounts.GetClientsDetails(
			map[string]string{
				"clientid": fmt.Sprintf("%d", thisCustomer.ID), "limitstart": "0", "limitnum": "500",
			},
		)
		fmt.Println(wd)

		// Test GetClientProducts
		wp, _, _ := whmcs.Accounts.GetClientsProducts(
			map[string]string{
				"clientid": fmt.Sprintf("%d", thisCustomer.ID), "limitstart": "0", "limitnum": "500",
			},
		)
		fmt.Printf("\nProducts for %s\n", thisCustomer.Email)
		for _, thisProduct := range wp.Products.Product {
			fmt.Println(thisProduct)
		}
	}
}

func TestClientContactList(t *testing.T) {
	whmcs, _ := loadWhmcs()
	_, err := createTestClient(whmcs)
	if err != nil {
		panic(err)
	}

	active, err := whmcs.Accounts.ClientContactList("Active")
	if err != nil {
		panic(err)
	}
	inactive, err := whmcs.Accounts.ClientContactList("Inactive")
	if err != nil {
		panic(err)
	}
	fmt.Printf("Active Contacts:\n%v\nInactive Contacts:\n%v", active, inactive)
}

func TestCreateInvoice(t *testing.T) {
	whmcs, _ := loadWhmcs()
	client, err := createTestClient(whmcs)
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
	supportInvoice, _, err := whmcs.Billing.CreateInvoice(client.ID, invoice)
	if err != nil {
		log.Printf("ERROR %s", err)
	}
	fmt.Printf("invoice ID: %d\n", supportInvoice)
}
