package whmcsgo

/*
DeleteClient - Delete client

WHMCS API docs

https://developers.whmcs.com/api-reference/deleteclient/

Request Parameters

see WHMCS API docs
*/
func (s *AccountsService) DeleteClient(parms map[string]string) (*Response, error) {
	a := new(Account)
	resp, err := do(s.client, Params{parms: parms, u: "DeleteClient"}, a)
	if err != nil {
		return resp, err
	}

	return resp, err
}
