package whmcsgo

// ProductsService handles .... related
// methods of the WHMCS API.
//
// WHMCS API docs: https://developers.whmcs.com/api/api-index/
type ProductsService struct {
	client *Client
}
type Product struct {
	Result string `json:"result"`
	Pid    int    `json:"pid"`
}
