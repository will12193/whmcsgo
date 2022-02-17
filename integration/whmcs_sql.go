package integration

import (
	"database/sql"
	"fmt"

	_ "github.com/go-sql-driver/mysql"
)

func doSQLQuery(query string) (*sql.Rows, error) {
	db, err := sql.Open("mysql", "root:sUper3R4nd0m@tcp(localhost:3306)/whmcs")
	if err != nil {
		return nil, fmt.Errorf("sql.Open failed: %w", err)
	}
	defer db.Close()

	results, err := db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("db.Query failed: %w", err)
	}

	return results, err
}

// Checks for exisiting payment gateway, and if none, create one
func createPaymentGW(name string) error {
	var (
		id *int
	)

	selectQuery := `
		SELECT id FROM whmcs.tblpaymentgateways 
			WHERE tblpaymentgateways.gateway = "` + name + `"
	`

	results, err := doSQLQuery(selectQuery)
	if err != nil {
		return fmt.Errorf("doSQLQuery failed: %w", err)
	}
	for results.Next() {
		err = results.Scan(&id)
		if err != nil {
			return fmt.Errorf("Scan failed: %w", err)
		}
	}
	if id != nil {
		return nil
	}

	insertQuery := `
	INSERT INTO whmcs.tblpaymentgateways (
		tblpaymentgateways.gateway,
		tblpaymentgateways.setting,
		tblpaymentgateways.value,
		tblpaymentgateways.order
	)
	VALUES
		(
			"` + name + `",
			"name",
			"Test Payment Gateway",
			0
		),
		(
			"` + name + `",
			"type",
			"Bank",
			0
		),
		(
			"` + name + `",
			"visible",
			"on",
			0
		),
		(
			"` + name + `",
			"merchantID",
			"015632c85d903a1a918f7674eed7e5a882626ab7757ec382a138de6f659afdc9439e609bfc1dc6ee229345d0bd7ccec4",
			0
		),
		(
			"` + name + `",
			"password",
			"5ae131360e757725e5a869a0de9b5c5d63de03a8121f58fa29de3766a4067865457786a993f455858221a9351f08d019",
			0
		),
		(
			"` + name + `",
			"testMode",
			"58b945070cf4c02a9840a3aca595f660fcb00f7af1d7e736d62cdee3c86a285041348b19b7b3f4ac26dce53918d73bb5",
			0
		)
	`
	_, err = doSQLQuery(insertQuery)
	return err
}

// creates a new product group and returns the GID for it
func createProductGroup(name string) (*int, error) {
	var (
		id *int
	)

	insertQuery := `
		INSERT INTO whmcs.tblproductgroups (
			name,
			slug,
			headline,
			tagline,
			orderfrmtpl,
			disabledgateways,
			hidden,
			tblproductgroups.order,
			created_at,
			updated_at
		)
		VALUES
			(
				"` + name + `",
				"` + name + `",
				"Epic headline",
				"Sick tagline",
				'',
				'',
				'0',
				'1',
				'2022-02-15 03:37:11',
				'2022-02-15 03:37:11'
		);
	`
	_, err := doSQLQuery(insertQuery)
	if err != nil {
		return nil, fmt.Errorf("doSQLQuery failed: %w", err)
	}

	selectQuery := `
	SELECT id FROM whmcs.tblproductgroups
		WHERE tblproductgroups.name = "` + name + `"
	`

	results, err := doSQLQuery(selectQuery)
	if err != nil {
		return nil, fmt.Errorf("doSQLQuery failed: %w", err)
	}

	for results.Next() {
		err = results.Scan(&id)
		if err != nil {
			return nil, fmt.Errorf("Scan failed: %w", err)
		}
	}
	if id == nil {
		return nil, fmt.Errorf("error: Product group %s not found", name)
	}

	return id, err
}

func deleteProductGroup(name string) error {
	query := `
		DELETE FROM whmcs.tblproductgroups WHERE (tblproductgroups.name = "` + name + `");
	`
	_, err := doSQLQuery(query)
	return err
}

func deleteProduct(name string) error {
	query := `
		DELETE FROM whmcs.tblproducts WHERE (tblproducts.name = "` + name + `");
	`
	_, err := doSQLQuery(query)
	return err
}
