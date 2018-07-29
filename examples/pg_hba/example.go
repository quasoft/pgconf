package main

import (
	"fmt"

	"github.com/quasoft/pgconf/hba"
)

func main() {
	c, err := hba.Open("../../hba/testdata/sample.conf") // pg_hba.conf
	if err != nil {
		panic(fmt.Errorf("Failed opening file pg_hba.conf: %s", err))
	}

	// Find all rows for replication host-based authentication
	rows, err := c.LookupAll(hba.Database, "replication")
	if err != nil {
		panic(fmt.Errorf("Failed looking up for replication rows: %s", err))
	}

	fmt.Printf("Found %d replication rows with addresses as follows:\n", len(rows))
	for _, row := range rows {
		// Get and print value for ADDRESS column
		address, err := c.String(row, hba.Address)
		if err != nil {
			panic(fmt.Errorf("Could not read address value: %s", err))
		}
		fmt.Println(" - " + address)
	}
}
