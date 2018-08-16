/*
Package flag is dedicated to handle command line flags in a CLI application.
It can handle multivaluated flags when flag usage is repeated or parse given
string to extract values. Flag values can also be set using environment variables.
The following examples are equivalent from the application perspective.

 $ ./app --server 10.0.0.1 --server 10.0.0.2       # a slice with 2 values is available to use
 [...]
 $ ./app --server 10.0.0.1,10.0.0.2
 [...]
 $ export SERVERS="10.0.0.1,10.0.0.2"
 $ ./app        # values are set using environment variables or default values

Default values can be set within the application.

Flags are categorized as boolean, monovaluated (1 and only 1 value can be set) or
multivaluated (several values can be associated with a flag).

The following snippet declares :
-b and --boolean as boolean flags;
-s and --server as multivaluated flags (slice), settable with an environment variable;
-i and --interval as a monovaluated flag, to stored as a uint64

The sep tag allows the user to set several values at once using a separator.

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
*/
package flag

import (
	"fmt"
	"os"
	"reflect"
	"strconv"
	"strings"
)

type valuation int

const (
	none valuation = iota
	mono
	multi
)

type flag struct {
	names     []string
	values    []string
	valuation valuation
	env       string
	finalType reflect.Kind
	index     int
	usage     string
	separator string
	isSet     bool
}

func (f *flag) String() string {
	return fmt.Sprintf("Flag.names: %s\nvalues: %s\nvaluation: %d\nenv: %s\ntype: %s\nis set: %t\nindex: %d\n",
		strings.Join(f.names, ";"),
		strings.Join(f.values, ";"),
		int(f.valuation),
		f.env,
		f.finalType.String(),
		f.isSet,
		f.index,
	)

}

//FlagSet is a set of flags holding parameters to populate the final data structure
//provided
type FlagSet struct {
	config interface{}
	fmap   map[string]*flag
	flist  []string
}

//NewFlagSet returns a pointer to a new FlagSet.
//config is a pointer to the struct to be populated with user inputs on command line
//or using environment variables. For example:
// type config struct {
//	 Help bool `names:"-h,--help" usage:"prints this help message"
//	 Targets []string `names:"-s,--server" env:"SERVERS" sep:"," usage:"server to contact"`
// }
//
func NewFlagSet(config interface{}) *FlagSet {
	fs := &FlagSet{
		config: config,
		fmap:   make(map[string]*flag),
		flist:  make([]string, 0),
	}

	if err := fs.setupFlags(); err != nil {
		panic("could not create FlagSet: " + err.Error())
	}
	return fs
}

func (fs *FlagSet) setupFlags() error {
	if reflect.TypeOf(fs.config).Kind() != reflect.Ptr {
		return fmt.Errorf("interface provided to NewFlagSet must be a pointer to a struct")
	}
	t := reflect.TypeOf(fs.config).Elem()

	for i := 0; i < t.NumField(); i++ {
		ft := t.Field(i)

		if ft.Type.Kind() == reflect.Ptr {
			return fmt.Errorf("pointer in config structure is not supported (%s)", ft.Name)
		}
		if ft.Type.Kind() == reflect.Map {
			return fmt.Errorf("map in config structure is not supported (%s)", ft.Name)
		}
		if ft.Type.Kind() == reflect.Chan {
			return fmt.Errorf("chan in config structure is not supported (%s)", ft.Name)
		}

		//valuation for this flag
		ftValuation := mono
		if ft.Type.Kind() == reflect.Slice {
			ftValuation = multi
		}
		if ft.Type.Kind() == reflect.Bool {
			ftValuation = none
		}

		flag := &flag{
			names:     make([]string, 0),
			values:    make([]string, 0),
			valuation: ftValuation,
			env:       "",
			finalType: ft.Type.Kind(),
			index:     i,
			usage:     "",
			separator: "",
			isSet:     false,
		}

		// get names for this flag
		namesTag, ok := ft.Tag.Lookup("names")
		if !ok {
			return fmt.Errorf("improper tag usage for flags: tag \"names\" is required")
		}
		names := strings.Split(namesTag, ",")
		for _, s := range names {
			s = strings.TrimSpace(s)
			if len(s) == 0 {
				continue
			}
			flag.names = append(flag.names, s)
		}
		if len(flag.names) == 0 {
			return fmt.Errorf("could not get any names tag for %s", ft.Name)
		}

		if envTag, ok := ft.Tag.Lookup("env"); ok {
			envTag = strings.TrimSpace(envTag)
			flag.env = envTag
		}

		if sepTag, ok := ft.Tag.Lookup("sep"); ok {
			flag.separator = strings.TrimSpace(sepTag)
		}

		if usageTag, ok := ft.Tag.Lookup("usage"); ok {
			flag.usage = strings.TrimSpace(usageTag)
		}

		for _, name := range flag.names {
			fs.fmap[name] = flag
		}
		fs.flist = append(fs.flist, flag.names[0])
	}
	return nil
}

//Parse parse command line and populate provided configuration structure
func (fs *FlagSet) Parse() error {

	if err := fs.parseCommand(os.Args[1:]); err != nil {
		return fmt.Errorf("could not parse commande line: %s", err)
	}

	if err := fs.parseEnv(); err != nil {
		return fmt.Errorf("could not get values from environment variables: %s", err)
	}

	if err := fs.setConfig(); err != nil {
		return fmt.Errorf("could not populate data structure: %s", err)
	}

	return nil
}

func (fs *FlagSet) parseCommand(args []string) error {
	if len(args) == 0 {
		return nil
	}

	arg := args[0]
	fitem, ok := fs.fmap[arg]
	if !ok {
		return fmt.Errorf("%s is not a valid flag", arg)
	}

	//boolean flag (valuation == none)
	if fs.fmap[arg].finalType == reflect.Bool {
		fs.fmap[arg].isSet = true
		return fs.parseCommand(args[1:])
	}

	if len(args) < 2 {
		return fmt.Errorf("missing value for flag %s", arg)
	}
	values := args[1]

	//mono flag (valuation == mono)
	if fitem.valuation == mono && fitem.isSet {
		return fmt.Errorf("flag %s already set", arg)
	}

	if fitem.valuation == mono {
		fitem.values = append(fitem.values, values)
		fitem.isSet = true
		return fs.parseCommand(args[2:])
	}

	//multi flag (valuation == multi)
	if len(fitem.separator) != 0 {
		splitted := strings.Split(values, fitem.separator)
		found := false
		for _, v := range splitted {
			if len(strings.TrimSpace(v)) != 0 {
				fitem.values = append(fitem.values, v)
				found = true
				fitem.isSet = true
			}
		}
		if !found {
			return fmt.Errorf("missing value for flag %s", arg)
		}
	} else {
		fitem.values = append(fitem.values, values)
		fitem.isSet = true
	}
	return fs.parseCommand(args[2:])
}

func (fs *FlagSet) parseEnv() error {

	for _, fname := range fs.flist {
		fitem := fs.fmap[fname]
		if fitem.isSet || len(fitem.env) == 0 {
			continue
		}

		values := os.Getenv(fitem.env)
		if len(values) == 0 {
			continue
		}

		if fitem.valuation == none {
			fitem.isSet = true
			continue
		}

		if fitem.valuation == mono {
			fitem.values = append(fitem.values, values)
			fitem.isSet = true
			continue
		}

		if len(fitem.separator) != 0 {
			splitted := strings.Split(values, fitem.separator)
			for _, v := range splitted {
				if len(strings.TrimSpace(v)) != 0 {
					fitem.values = append(fitem.values, v)
					fitem.isSet = true
				}
			}
		}

		if len(fitem.values) == 0 {
			fitem.values = append(fitem.values, values)
			fitem.isSet = true
		}
	}

	return nil
}

func (fs *FlagSet) setConfig() error {
	if !reflect.ValueOf(fs.config).Elem().Field(0).CanAddr() {
		fmt.Printf("can not addr fs.config field(0)\n")
	}
	if !reflect.ValueOf(fs.config).Elem().Field(0).IsValid() {
		fmt.Printf("not valid fs.config field(0)\n")
	}
	if !reflect.ValueOf(fs.config).Elem().Field(0).CanSet() {
		fmt.Printf("can not set fs.config field(0)\n")
	}

	for _, fname := range fs.flist {
		fitem := fs.fmap[fname]
		if !fitem.isSet {
			continue
		}

		ith := reflect.ValueOf(fs.config).Elem().Field(fitem.index)
		if fitem.valuation == none {
			ith.SetBool(true)
			continue
		}

		if fitem.valuation == mono {
			switch fitem.finalType {
			case reflect.String:
				ith.SetString(fitem.values[0])
				continue
			case reflect.Uint:
				v, err := strconv.ParseUint(fitem.values[0], 10, 0)
				if err != nil {
					return err
				}
				ith.SetUint(v)
				continue
			case reflect.Uint8:
				v, err := strconv.ParseUint(fitem.values[0], 10, 8)
				if err != nil {
					return err
				}
				ith.SetUint(v)
				continue
			case reflect.Uint16:
				v, err := strconv.ParseUint(fitem.values[0], 10, 16)
				if err != nil {
					return err
				}
				ith.SetUint(v)
				continue
			case reflect.Uint32:
				v, err := strconv.ParseUint(fitem.values[0], 10, 32)
				if err != nil {
					return err
				}
				ith.SetUint(v)
				continue
			case reflect.Uint64:
				v, err := strconv.ParseUint(fitem.values[0], 10, 64)
				if err != nil {
					return err
				}
				ith.SetUint(v)
				continue
			case reflect.Int:
				v, err := strconv.ParseInt(fitem.values[0], 10, 0)
				if err != nil {
					return err
				}
				ith.SetInt(v)
				continue
			case reflect.Int8:
				v, err := strconv.ParseInt(fitem.values[0], 10, 8)
				if err != nil {
					return err
				}
				ith.SetInt(v)
				continue
			case reflect.Int16:
				v, err := strconv.ParseInt(fitem.values[0], 10, 16)
				if err != nil {
					return err
				}
				ith.SetInt(v)
				continue
			case reflect.Int32:
				v, err := strconv.ParseInt(fitem.values[0], 10, 32)
				if err != nil {
					return err
				}
				ith.SetInt(v)
				continue
			case reflect.Int64:
				v, err := strconv.ParseInt(fitem.values[0], 10, 64)
				if err != nil {
					return err
				}
				ith.SetInt(v)
				continue
			case reflect.Float32:
				v, err := strconv.ParseFloat(fitem.values[0], 32)
				if err != nil {
					return err
				}
				ith.SetFloat(v)
				continue
			case reflect.Float64:
				v, err := strconv.ParseFloat(fitem.values[0], 64)
				if err != nil {
					return err
				}
				ith.SetFloat(v)
				continue
			default:
				return fmt.Errorf("can not guess type: %s", fitem.finalType.String())
			}
		}

		if fitem.valuation == multi {
			newSlice := reflect.MakeSlice(ith.Type(), 0, 0)

			switch ith.Type().Elem().Kind() {
			case reflect.String:
				for _, vstr := range fitem.values {
					rv := reflect.ValueOf(vstr)
					newSlice = reflect.Append(newSlice, rv)
				}
				ith.Set(newSlice)
				continue
			case reflect.Uint:
				for _, vstr := range fitem.values {
					v, err := strconv.ParseUint(vstr, 10, 0)
					if err != nil {
						return err
					}
					rv := reflect.ValueOf(uint(v))
					newSlice = reflect.Append(newSlice, rv)
				}
				ith.Set(newSlice)
				continue
			case reflect.Uint8:
				for _, vstr := range fitem.values {
					v, err := strconv.ParseUint(vstr, 10, 8)
					if err != nil {
						return err
					}
					rv := reflect.ValueOf(uint8(v))
					newSlice = reflect.Append(newSlice, rv)
				}
				ith.Set(newSlice)
				continue
			case reflect.Uint16:
				for _, vstr := range fitem.values {
					v, err := strconv.ParseUint(vstr, 10, 16)
					if err != nil {
						return err
					}
					rv := reflect.ValueOf(uint16(v))
					newSlice = reflect.Append(newSlice, rv)
				}
				ith.Set(newSlice)
				continue
			case reflect.Uint32:
				for _, vstr := range fitem.values {
					v, err := strconv.ParseUint(vstr, 10, 32)
					if err != nil {
						return err
					}
					rv := reflect.ValueOf(uint32(v))
					newSlice = reflect.Append(newSlice, rv)
				}
				ith.Set(newSlice)
				continue
			case reflect.Uint64:
				for _, vstr := range fitem.values {
					v, err := strconv.ParseUint(vstr, 10, 64)
					if err != nil {
						return err
					}
					rv := reflect.ValueOf(uint64(v))
					newSlice = reflect.Append(newSlice, rv)
				}
				ith.Set(newSlice)
				continue
			case reflect.Int:
				for _, vstr := range fitem.values {
					v, err := strconv.ParseInt(vstr, 10, 0)
					if err != nil {
						return err
					}
					rv := reflect.ValueOf(int(v))
					newSlice = reflect.Append(newSlice, rv)
				}
				ith.Set(newSlice)
				continue
			case reflect.Int8:
				for _, vstr := range fitem.values {
					v, err := strconv.ParseInt(vstr, 10, 8)
					if err != nil {
						return err
					}
					rv := reflect.ValueOf(int8(v))
					newSlice = reflect.Append(newSlice, rv)
				}
				ith.Set(newSlice)
				continue
			case reflect.Int16:
				for _, vstr := range fitem.values {
					v, err := strconv.ParseInt(vstr, 10, 16)
					if err != nil {
						return err
					}
					rv := reflect.ValueOf(int16(v))
					newSlice = reflect.Append(newSlice, rv)
				}
				ith.Set(newSlice)
				continue
			case reflect.Int32:
				for _, vstr := range fitem.values {
					v, err := strconv.ParseInt(vstr, 10, 32)
					if err != nil {
						return err
					}
					rv := reflect.ValueOf(int32(v))
					newSlice = reflect.Append(newSlice, rv)
				}
				ith.Set(newSlice)
				continue
			case reflect.Int64:
				for _, vstr := range fitem.values {
					v, err := strconv.ParseInt(vstr, 10, 64)
					if err != nil {
						return err
					}
					rv := reflect.ValueOf(int64(v))
					newSlice = reflect.Append(newSlice, rv)
				}
				ith.Set(newSlice)
				continue
			case reflect.Float32:
				for _, vstr := range fitem.values {
					v, err := strconv.ParseFloat(vstr, 32)
					if err != nil {
						return err
					}
					rv := reflect.ValueOf(float32(v))
					newSlice = reflect.Append(newSlice, rv)
				}
				ith.Set(newSlice)
				continue
			case reflect.Float64:
				for _, vstr := range fitem.values {
					v, err := strconv.ParseFloat(vstr, 64)
					if err != nil {
						return err
					}
					rv := reflect.ValueOf(float64(v))
					newSlice = reflect.Append(newSlice, rv)
				}
				ith.Set(newSlice)
				continue
			default:
				return fmt.Errorf("can not guess type: %s", fitem.finalType.String())
			}
		}
	}
	return nil
}
