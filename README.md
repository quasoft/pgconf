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

    "github.com/quasoft/pgconf"
)

conf, err := pgconf.Open("/data/postgresql.conf")
if err != nil {
    panic("Could not open conf file: " + err.Error())
}

// .AsString automatically dequotes values
dest, err := conf.AsString("log_destination")
if err == pgconf.ErrKeyNotFound || dest != "syslog" {
    // If key was not set or has the wrong value
    fmt.Println("log_destination value is not what we want, changing it now")
    conf.SetString("log_destination", "syslog") // .SetString automatically quotes values
}

err = pgconf.Save("/data/postgresql.conf")
if err != nil {
    panic("Could not save file: " + err.Error())
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

    "github.com/quasoft/pgconf"
)

conf, err := pgconf.Open("/data/postgresql.conf")
if err != nil {
    panic("Could not open conf file: " + err.Error())
}

dest, err := conf.AsString("log_destination")
if err == pgconf.ErrKeyNotFound || dest != "syslog" {
    fmt.Println("log_destination value is not what we want, changing it now")
    conf.SetString("log_destination", "syslog")
}

err = pgconf.Save("/data/postgresql.conf.tmp")
if err != nil {
    panic("Could not save to tmp file: " + err.Error())
}

err = os.Rename("/data/postgresql.conf.tmp", "/data/postgresql.conf")
if err != nil {
    panic("Could not rename tmp file to conf: " + err.Error())
}
```