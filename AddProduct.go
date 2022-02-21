package whmcsgo

/*
AddProduct Adds a product to a product group

WHMCs API docs

https://developers.whmcs.com/api-reference/addproduct/

*/
func (s *ProductsService) AddProduct(parms map[string]string) (*Product, *Response, error) {
	a := new(Product)
	resp, err := apiRequest(s.client, Params{parms: parms, u: "AddProduct"}, a)
	if err != nil {
		return nil, resp, err
	}

	return a, resp, err
}
