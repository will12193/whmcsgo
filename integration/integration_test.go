package integration

import (
	"encoding/json"
	"fmt"
	"strconv"
	"testing"

	"github.com/chrisjoyce911/whmcsgo"
	"github.com/jinzhu/now"
	"github.com/kelseyhightower/envconfig"
	"github.com/stretchr/testify/assert"
)

type tHelper interface {
	Helper()
}

/*
Integration tests to be run with an active development instance of WHMCS.

The following environment variables are required to run the tests:
export WHM_IDENT="put your api identity key here"
export WHM_SECRET="put your api secret key here"
export WHM_ACCESS="put your api access key here"
export WHM_URL="http://localhost:####/"
export WHM_PAYMENTMETHOD="payment method setup in dev env"
export WHM_DB_PASSWORD="password for the root user in DB"

Prerequisites to running the tests:
- Working instance of WHMCS
- Settings -> API Credentials - Created API Role with appropriate access and create credentials

Whats left (not cleaned up) after running the tests:
- A Payment Gateway is setup
*/

type Config struct {
	URL           string `default:"http://localhost:3000/"`
	Access        string `default:"RAND0MT3STK3Y"`
	Ident         string `default:"JnbGfwNUq1CIHxhEoqRbMKb084gcvwwz"`
	Secret        string `default:"Wx9Lqeqe0Os0paUUDtbc37k89qfpqdvZ"`
	PaymentMethod string `default:"testPaymentMethod"`
	DBPassword    string `default:"sUper3R4nd0m"`
}

var whmcsConfig *Config
var testUser = "testdude@divisia.io"
var testProductGroup = "Test Product Group"
var testProduct = "Test Product"

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
	assert.NoError(t, err)
	assert.NotNil(t, tc)
	gid, err := createProductGroup(whmcsConfig.DBPassword, testProductGroup)
	assert.NoError(t, err)
	assert.GreaterOrEqual(t, *gid, 0)
	productID, err := createTestProduct(client, testProduct, *gid)
	assert.NoError(t, err)
	assert.GreaterOrEqual(t, *productID, 0)
	err = createPaymentGW(whmcsConfig.DBPassword, whmcsConfig.PaymentMethod)
	assert.NoError(t, err)
	_, err = createTestOrder(client, tc.ID, *productID, whmcsConfig.PaymentMethod)
	assert.NoError(t, err)

	// Test GetClients
	t.Log("Load Private Clients")
	wc, _, err := client.Accounts.GetClients(map[string]string{"sorting": "ASC", "limitstart": "0", "limitnum": "2500"})
	assert.NoError(t, err)
	assert.NotNil(t, wc)
	assert.Greater(t, wc.Numreturned, 0)

	clientExists := false
	productExists := false
	for _, thisCustomer := range wc.Clients.Client {
		t.Logf("\n\n%v\n\n", thisCustomer)

		// Test GetClientDetails
		wd, _, _ := client.Accounts.GetClientsDetails(
			map[string]string{
				"clientid": fmt.Sprintf("%d", thisCustomer.ID), "limitstart": "0", "limitnum": "500",
			},
		)
		t.Log(wd)
		if wd.Email == testUser {
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
	err = deleteProduct(whmcsConfig.DBPassword, testProduct)
	if err != nil {
		t.Error(err)
	}
	err = deleteProductGroup(whmcsConfig.DBPassword, testProductGroup)
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

	var ContactContains assert.ComparisonAssertionFunc = func(
		t assert.TestingT, a interface{}, b interface{}, msgAndArgs ...interface{},
	) bool {
		if h, ok := t.(tHelper); ok {
			h.Helper()
		}
		contact := a.([]whmcsgo.ContactList)
		expected := b.(whmcsgo.ContactList)

		for _, c := range contact {
			if c.CompanyName == expected.CompanyName &&
				c.FullName == expected.FullName &&
				c.Phone == expected.Phone &&
				c.Status == expected.Status &&
				c.Email == expected.Email &&
				c.State == expected.State {
				return true
			}
		}

		assert.Fail(t, "Doesn't contain matching company", msgAndArgs)
		return false
	}

	active, err := client.Accounts.ClientContactList("Active")
	if err != nil {
		t.Error(err)
	}

	assert.Greater(t, len(active), 0, active)

	ContactContains(t, active, whmcsgo.ContactList{
		CompanyName: "test corp",
		FullName:    "Test Dude",
		Phone:       "01234123123",
		Status:      "Active",
		State:       "Queensland",
		Email:       testUser,
	}, active)

	inactive, err := client.Accounts.ClientContactList("Inactive")
	assert.NoError(t, err)

	t.Logf("\nActive Contacts:\n%v\nInactive Contacts:\n%v", active, inactive)
	assert.Greater(t, len(active)+len(inactive), 0)

	err = deleteClient(client, tc.ID)
	if err != nil {
		t.Error(err)
	}
}

func TestCreateInvoice(t *testing.T) {
	client, _ := loadWhmcs()
	tc, err := createTestClient(client)
	assert.NoError(t, err)

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

	var AssertInvoice assert.ComparisonAssertionFunc = func(
		t assert.TestingT, expected interface{}, b interface{}, msgAndArgs ...interface{},
	) bool {
		if h, ok := t.(tHelper); ok {
			h.Helper()
		}
		expectedObj := expected.(*whmcsgo.InvoiceResponse)
		actual := b.(*whmcsgo.InvoiceResponse)
		assert.Equal(t, expectedObj.Status, actual.Status)
		assert.Equal(t, expectedObj.Result, actual.Result)
		assert.Greater(t, expectedObj.InvoiceID, 0)
		return true
	}

	invoiceid, supportInvoice, err := client.Billing.CreateInvoice(tc.ID, invoice)

	assert.NoError(t, err)

	var apiResp whmcsgo.InvoiceResponse
	err = json.Unmarshal([]byte(supportInvoice.Body), &apiResp)
	assert.NoError(t, err)

	AssertInvoice(t, &apiResp, &whmcsgo.InvoiceResponse{
		Status: "Draft",
		Result: "success",
	})

	// Get Invoice and check line items
	inv, err := client.Billing.GetLastInvoice(tc.ID, "")
	assert.NoError(t, err)

	assert.Equal(t, inv.ID, invoiceid)
	subtotal, err := strconv.ParseFloat(inv.Subtotal, 64)
	assert.NoError(t, err)
	assert.Equal(t, subtotal, float64(lineItem.ItemAmount))

	t.Logf("invoice ID: %d\n", invoiceid)

	err = deleteClient(client, tc.ID)
	if err != nil {
		t.Error(err)
	}
}
