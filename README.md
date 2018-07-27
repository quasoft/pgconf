# pgconf

`pgconf` is a simple Go package for reading/writing to PostgreSQL config file (`postgresql.conf`) that:

* Preseves existing whitespace and comments
* Supports single quoted and backslash-quoted values (eg. `search_path = '''$user'', \'public\''`)
* Supports values with optional equal sign (eg. `logconnections yes`)
* Works with ASCII and UTF-8 `.conf` files

## How to use

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

## Hint

Usually it's safer to write changes to a temp file and once that writing is over to rename
the temp file to postgresql.conf (or whatever you've named yours).

What follows is an example of this use case. If you use this make sure to store the temp
file in a secure location (eg. the data dir) with restricted permissions and not inside
the /tmp directory.

```go
import (
	"fmt"
	"os"

	"github.com/quasoft/pgconf/conf"
)

func main() {
	c, err := conf.Open("/data/postgresql.conf")
	if err != nil {
		panic("Could not open conf file: " + err.Error())
	}

	dest, err := c.StringK("log_destination")
	if err != nil || dest != "syslog" {
		fmt.Println("log_destination value is not what we want, changing it now")
		c.SetStringK("log_destination", "syslog")
	}

	err = c.WriteFile("/data/postgresql.conf.tmp", 0644)
	if err != nil {
		panic("Could not save file: " + err.Error())
	}

	err = os.Rename("/data/postgresql.conf.tmp", "/data/postgresql.conf")
	if err != nil {
		panic("Could not rename tmp file to conf: " + err.Error())
	}
}

```