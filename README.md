# Flag - Parsing command line arguments in Go

[![GoDoc](https://godoc.org/github.com/etombini/flag?status.svg)](https://godoc.org/github.com/etombini/flag)

Package flag is dedicated to handle command line flags in a CLI application.
It can handle multivaluated flags when flag usage is repeated or parse given
string to extract values. Flag values can also be set using environment variables.
The following examples are equivalent from the application perspective.

```
 $ ./app --server 10.0.0.1 --server 10.0.0.2       # a slice with 2 values is available to use
 [...]
 $ ./app --server 10.0.0.1,10.0.0.2
 [...]
 $ export SERVERS="10.0.0.1,10.0.0.2"
 $ ./app        # values are set using environment variables or default values
```

Default values can be set within the application.

Flags are categorized as boolean, monovaluated (1 and only 1 value can be set) or
multivaluated (several values can be associated with a flag).

The following snippet declares :
-b and --boolean as boolean flags;
-s and --server as multivaluated flags (slice), settable with an environment variable;
-i and --interval as a monovaluated flag, to stored as a uint64

The sep tag allows the user to set several values at once using a separator.

```go
type config struct {
	Path     string   `names:"-p,--p"`
	Servers  []string `names:"-s,--server" env:"SERVERS_TEST" sep:","`
	Interval uint64   `names:"-i,--interval" env:"INTERVAL_TEST"`
	SomeBool bool     `names:"-b,--boolean" env:"BOOL_TEST"`
}

func main() {
	c := &config{
		Path:     "some path",
		Servers:  []string{"srv01", "srv02"},
		Interval: 10,
	}

	f := NewFlagSet(c)

	if err := f.Parse(); err != nil {
		fmt.Printf("ERROR: %s\n", err)
	}

	fmt.Printf("CONFIG:\npath: %s\nservers: %s\ninterval: %d\nsomeBool: %t\n",
		c.Path,
		strings.Join(c.Servers, "|"),
		c.Interval,
		c.SomeBool,
	)
}

Setting values is done this way for each flag: 
1. Parsing the commande line
2. If nothing is set from 1., parse environment variables
3. If nothing is set from 2., default values already set apply