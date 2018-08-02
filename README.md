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
 $ ./app        # values are set using environment variable
```

Default values can be set within the application.

Flags are categorized as boolean, monovaluated (1 and only 1 value can be set) or
multivaluated (several values can be associated with a flag).

The following snippet declares `-b` and `--boolean` as boolean flags; `-l` and `--long` as
multivaluated flags, settable with an environment variable, with default values;
`-w` and `--without-env ` in the same way, except for the environment variable

```go
f := flag.NewFlag()
if f == nil {
    fmt.Printf("can not create flag")
    os.Exit(1)
}
if err := f.AddBoolFlags([]string{"-b", "--boolean"}, "a boolean flag"); err != nil {
    fmt.Printf("can not create boolean flag: %s", err)
    os.Exit(1)
}
if err := f.AddMultiFlagsWithEnv([]string{"-l", "--long"}, "LONG_FLAG_ENV", "1,2", ",", "-l and --long set the long things"); err != il {
    fmt.Printf("can not create multivaluated flag: %s", err)
    os.Exit(1)
}
if err := f.AddMultiFlags([]string{"-w", "--without-env"}, "value01,value02", ",", "without environment variable"); err != nil {
    fmt.Printf("can not create multivaluated flag: %s", err)
    os.Exit(1)
}
```

Values can be retrieved using specific method per expected type.

```go
firstFlag, err := f.GetBool("-b")
if err != nil {
	return err
}

secondFlag, err := f.GetInt("-l")
if err != nil {
	return err
}

for _, values := range secondFlag {
	// do related stuff
}
```

Setting values is done this way for each flag: 
1. Parsing the commande line
2. If nothing is set from 1., parse environment variables
3. If nothing is set from 2., parse default values