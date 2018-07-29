# pgconf

`pgconf` is a simple Go package for reading/writing to PostgreSQL config file (`postgresql.conf`) that:

* Preseves existing whitespace and comments
* Supports single quoted and backslash-quoted values (eg. `search_path = '''$user'', \'public\''`)
* Supports values with optional equal sign (eg. `logconnections yes`)
* Works with ASCII and UTF-8 `.conf` files

## How to use

### `postgresql.conf`:

To read or update `postgresql.conf` files use the `pgconf/conf` package:

```go
import (
	"fmt"

	"github.com/quasoft/pgconf/conf"
)

func main() {
	c, err := conf.Open("/data/postgresql.conf")
	if err != nil {
		panic("Could not open conf file: " + err.Error())
	}

	// StringK automatically dequotes values
	dest, err := c.StringK("log_destination")
	if err != nil || dest != "syslog" {
		// If key was not set or has the wrong value
		fmt.Println("log_destination value is not what we want, changing it now")
		c.SetStringK("log_destination", "syslog") // SetStringK automatically quotes values if necessary
	}

	err = c.WriteFile("/data/postgresql.conf", 0644)
	if err != nil {
		panic("Could not save file: " + err.Error())
	}
}
```

### `pg_hba.conf`

To read or update `pg_hba.conf` files use the `pgconf/hba` package:

```go
package main

import (
	"fmt"

	"github.com/quasoft/pgconf/hba"
)

func main() {
	conf, err := hba.Open("../../hba/testdata/sample.conf") // pg_hba.conf
	if err != nil {
		panic(fmt.Errorf("Failed opening file pg_hba.conf: %s", err))
	}

	// Find all rows for replication host-based authentication
	rows, err := conf.LookupAll(hba.Database, "replication")
	if err != nil {
		panic(fmt.Errorf("Failed looking up for replication rows: %s", err))
	}

	fmt.Printf("Found %d replication rows with addresses as follows:\n", len(rows))
	for _, row := range rows {
		// Get and print value for ADDRESS column
		address, err := conf.String(row, hba.Address)
		if err != nil {
			panic(fmt.Errorf("Could not read address value: %s", err))
		}
		fmt.Println(" - " + address)
	}
}
```

## Hint

Usually it's safer to write changes to a temp file and once that writing is over to rename
the temp file to the actual configuration file:

```go
	...
	err = c.WriteFile("/data/postgresql.conf.tmp", 0644)
	if err != nil {
		panic("Could not save file: " + err.Error())
	}

	err = os.Rename("/data/postgresql.conf.tmp", "/data/postgresql.conf")
	if err != nil {
		panic("Could not rename tmp file to conf: " + err.Error())
	}
	...
}
```

If you use this approach make sure to store the temp file in a secure location (eg. the data
dir) with restricted permissions and not inside the `/tmp` directory.