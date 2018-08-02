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
 $ ./app        # values are set using environment variable

Default values can be set within the application.

Flags are categorized as boolean, monovaluated (1 and only 1 value can be set) or
multivaluated (several values can be associated with a flag).

The following snippet declares -b and --boolean as boolean flags; -l and --long as
multivaluated flags, settable with an environment variable, with default values;
-w and --without-env in the same way, except for the environment variable

    f := flag.NewFlag()
    if f == nil {
        fmt.Printf("can not create flag")
        os.Exit(1)
    }

	if err := f.AddBoolFlags([]string{"-b", "--boolean"}, "a boolean flag"); err != nil {
        fmt.Printf("can not create boolean flag: %s", err)
        os.Exit(1)
    }
	if err := f.AddMultiFlagsWithEnv([]string{"-l", "--long"}, "LONG_FLAG_ENV", "1,2", ",", "-l and --long set the long things"); err != nil {
        fmt.Printf("can not create multivaluated flag: %s", err)
        os.Exit(1)
    }
    if err := f.AddMultiFlags([]string{"-w", "--without-env"}, "value01,value02", ",", "without environment variable"); err != nil {
        fmt.Printf("can not create multivaluated flag: %s", err)
        os.Exit(1)
    }

Values can be retrieved using specific method per expected type.

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

*/
package flag

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

//Flag is the main data structure holding application flags
type Flag struct {
	f map[string]*flagItem
}

//Valuation tells if a Flag is multivaluated, monovaluated, not valuated (boolean)
type Valuation int

const (
	//Multi allow an Argument to hold several values
	Multi Valuation = 0
	//Mono allows an Argument to hold only one value
	Mono Valuation = 1
	//None does not allow an Argument to hold a value
	None Valuation = 2

	spaces string = " \n\r\t\v\f"
)

func checkEnvFormat(env string) error {
	if len(env) == 0 {
		return nil
	}
	if strings.ContainsAny(env, spaces) {
		return fmt.Errorf("environment variable %s contains space caracters", env)
	}
	return nil
}

func checkFlagFormat(flag string) error {
	if len(flag) == 0 {
		return fmt.Errorf("flag name is empty")
	}
	if strings.ContainsAny(flag, spaces) {
		return fmt.Errorf("flag %s contains space caracters", flag)
	}
	return nil
}

//Flag struct represent a specific flag defined by a user
type flagItem struct {
	flags       []string
	envFlag     string
	isSet       bool
	values      []string
	defaults    []string
	valuation   Valuation
	separator   string
	description string
}

//NewFlag returns a new pointer to Flag
func NewFlag() *Flag {
	return &Flag{
		f: make(map[string]*flagItem),
	}
}

//AddBoolFlagsWithEnv add several boolean flags for the same behavior. For example
//-f and --force
func (f *Flag) AddBoolFlagsWithEnv(flags []string, env string, description string) error {
	return f.add(flags, env, []string{}, None, "", description)
}

//AddBoolFlagWithEnv add a boolean flag, possibly set using an environment variable
func (f *Flag) AddBoolFlagWithEnv(flag string, env string, description string) error {
	return f.AddBoolFlagsWithEnv([]string{flag}, env, description)
}

//AddBoolFlags add several boolean flags for the same behavior. For example
//-f and --force
func (f *Flag) AddBoolFlags(flags []string, description string) error {
	return f.AddBoolFlagsWithEnv(flags, "", description)
}

//AddBoolFlag add a boolean flag
func (f *Flag) AddBoolFlag(flag string, description string) error {
	return f.AddBoolFlagWithEnv(flag, "", description)
}

//AddMonoFlagsWithEnv add several flags that can handle only one value and can not be used multiple times
func (f *Flag) AddMonoFlagsWithEnv(flags []string, env string, defaults string, description string) error {
	return f.add(flags, env, []string{defaults}, Mono, "", description)
}

//AddMonoFlagWithEnv add a flag that can handle only one value and can not be used multiple times
func (f *Flag) AddMonoFlagWithEnv(flag string, env string, defaults string, description string) error {
	return f.AddMonoFlagsWithEnv([]string{flag}, env, defaults, description)
}

//AddMonoFlags add several flags that can handle only one value and can not be used multiple times
func (f *Flag) AddMonoFlags(flags []string, defaults string, description string) error {
	return f.AddMonoFlagsWithEnv(flags, "", defaults, description)
}

//AddMonoFlag add a flag that can handle only one value and can not be used multiple times
func (f *Flag) AddMonoFlag(flag string, defaults string, description string) error {
	return f.AddMonoFlagWithEnv(flag, "", defaults, description)
}

//AddMultiFlagsWithEnv add a flag that can handle several values and can be used multiple times.
//For example --server 10.1.2.3 --server 10.2.3.4 or --server 10.1.2.3,10.2.3.4 where "," is defined as a separator
func (f *Flag) AddMultiFlagsWithEnv(flags []string, env string, defaults string, separator string, description string) error {
	if len(separator) == 0 {
		return f.add(flags, env, []string{defaults}, Multi, separator, description)
	}
	d := make([]string, 0)
	for _, v := range strings.Split(defaults, separator) {
		if len(v) == 0 {
			continue
		}
		d = append(d, v)
	}
	return f.add(flags, env, d, Multi, separator, description)
}

//AddMultiFlagWithEnv add a flag that can handle several values and can be used multiple times.
//For example --server 10.1.2.3 --server 10.2.3.4 or --server 10.1.2.3,10.2.3.4 where "," is defined as a separator
func (f *Flag) AddMultiFlagWithEnv(flag string, env string, defaults string, separator string, description string) error {
	return f.AddMultiFlagsWithEnv([]string{flag}, env, defaults, separator, description)
}

//AddMultiFlags add a flag that can handle several values and can be used multiple times.
//For example --server 10.1.2.3 --server 10.2.3.4 or --server 10.1.2.3,10.2.3.4 where "," is defined as a separator
func (f *Flag) AddMultiFlags(flags []string, defaults string, separator string, description string) error {
	return f.AddMultiFlagsWithEnv(flags, "", defaults, separator, description)
}

//AddMultiFlag add a flag that can handle several values and can be used multiple times.
//For example --server 10.1.2.3 --server 10.2.3.4 or --server 10.1.2.3,10.2.3.4 where "," is defined as a separator
func (f *Flag) AddMultiFlag(flag string, defaults string, separator string, description string) error {
	return f.AddMultiFlagWithEnv(flag, "", defaults, separator, description)
}

//AddMulti adds several flags to handle a value, for example allowing both usage of "-c" and "--config"
func (f *Flag) add(flags []string, env string, defaults []string, valuation Valuation, separator string, description string) error {
	if err := checkEnvFormat(env); err != nil {
		return err
	}

	for _, flag := range flags {
		if err := checkFlagFormat(flag); err != nil {
			return err
		}
	}

	for _, flag := range flags {
		if _, ok := f.f[flag]; ok {
			return fmt.Errorf("flag [%s] is already defined", flag)
		}
	}

	if len(defaults) > 0 && valuation == None {
		return fmt.Errorf("default value(s) defined for boolean flag(s) %v", flags)
	}
	if len(defaults) > 1 && valuation == Mono {
		return fmt.Errorf("default multivaluation defined for monovaluated flag(s) %s", flags)
	}
	ff := &flagItem{
		flags:       make([]string, 0),
		envFlag:     env,
		isSet:       false,
		values:      make([]string, 0),
		defaults:    make([]string, 0),
		valuation:   valuation,
		separator:   separator,
		description: description,
	}
	for _, flag := range flags {
		ff.flags = append(ff.flags, flag)
	}
	for _, d := range defaults {
		if len(d) == 0 {
			continue
		}
		ff.defaults = append(ff.defaults, d)
	}
	for _, flag := range flags {
		f.f[flag] = ff
	}
	return nil
}

//Get returns values (as strings) set for this flag
func (f *Flag) Get(flag string) ([]string, error) {
	res, ok := f.f[flag]
	if !ok {
		return nil, fmt.Errorf("flag %s not defined", flag)
	}
	return res.values, nil
}

//Parse parse command line instructions and environment variables to set up the Flag struct
func (f *Flag) Parse() error {
	if err := f.parse(os.Args[1:]); err != nil {
		return err
	}
	if err := f.parseEnv(); err != nil {
		return err
	}
	if err := f.parseDefaults(); err != nil {
		return err
	}
	return nil
}

func (f *Flag) parse(args []string) error {
	if len(args) == 0 {
		return nil
	}
	fv, ok := f.f[args[0]]
	if !ok {
		return fmt.Errorf("unknown flag: %s", args[0])
	}

	if (fv.valuation == Mono || fv.valuation == Multi) && len(args) < 2 {
		return fmt.Errorf("flag %s requires a value", args[0])
	}

	//Valuation None
	if fv.valuation == None {
		return f.parseNone(fv, args[1:])
	}

	//Valuation Mono
	if fv.valuation == Mono {
		return f.parseMono(fv, args[1:])
	}

	//Valuation Multi
	if fv.valuation == Multi {
		return f.parseMulti(fv, args[1:])
	}

	return fmt.Errorf("error with %s flag, it must be a boolean, mono-valuated or multi-valuated", args[0])
}

func (f *Flag) parseNone(fItem *flagItem, args []string) error {
	fItem.values = append(fItem.values, "true")
	fItem.isSet = true
	return f.parse(args)
}

func (f *Flag) parseMono(fItem *flagItem, args []string) error {
	if fItem.isSet {
		return fmt.Errorf("flag %s defined several times", args[0])
	}
	fItem.values = append(fItem.values, args[0])
	fItem.isSet = true
	return f.parse(args[1:])
}

func (f *Flag) parseMulti(fItem *flagItem, args []string) error {
	if len(fItem.separator) == 0 || !strings.Contains(args[0], fItem.separator) {
		fItem.values = append(fItem.values, args[0])
		fItem.isSet = true
		return f.parse(args[1:])
	}
	splitted := strings.Split(args[0], fItem.separator)
	for _, v := range splitted {
		if len(v) == 0 {
			continue
		}
		fItem.values = append(fItem.values, v)
		fItem.isSet = true
	}
	return f.parse(args[1:])
}

func (f *Flag) parseEnv() error {
	for _, fv := range f.f {
		if fv.isSet {
			continue
		}
		if len(fv.envFlag) == 0 {
			continue
		}
		env := os.Getenv(fv.envFlag)
		if len(env) == 0 {
			continue
		}
		if fv.valuation == None {
			fv.values = append(fv.values, "true")
			fv.isSet = true
			continue
		}
		if fv.valuation == Mono {
			fv.values = append(fv.values, env)
			fv.isSet = true
			continue
		}
		if fv.valuation == Multi {
			if len(fv.separator) == 0 {
				fv.values = append(fv.values, env)
				fv.isSet = true
				continue
			}
			splitted := strings.Split(env, fv.separator)
			for _, v := range splitted {
				if len(v) == 0 {
					continue
				}
				fv.values = append(fv.values, v)
				fv.isSet = true
			}
			continue
		}
		return fmt.Errorf("error with %s environment flag, it must be a boolean, mono-valuated or multi-valuated", fv.envFlag)
	}
	return nil
}

func (f *Flag) parseDefaults() error {
	for _, fv := range f.f {
		if fv.isSet {
			continue
		}
		if len(fv.defaults) == 0 {
			continue
		}
		for _, v := range fv.defaults {
			if len(v) == 0 {
				continue
			}
			fv.values = append(fv.values, v)
			fv.isSet = true
		}
	}
	return nil
}

func (f *Flag) isOK(key string, v Valuation) error {
	fv, ok := f.f[key]
	if !ok {
		return fmt.Errorf("%s is not defined", key)
	}
	if fv.valuation != v {
		return fmt.Errorf("%s is not of type %v", key, v)
	}
	return nil
}

func (f *Flag) isNone(key string) bool {
	if err := f.isOK(key, None); err != nil {
		return false
	}
	return true
}

func (f *Flag) isMono(key string) bool {
	if err := f.isOK(key, Mono); err != nil {
		return false
	}
	return true
}

func (f *Flag) isMulti(key string) bool {
	if err := f.isOK(key, Multi); err != nil {
		return false
	}
	return true
}

//GetBool return the flag value for the corresponding boolean key
func (f *Flag) GetBool(key string) (bool, error) {
	if !f.isNone(key) {
		return false, fmt.Errorf("%s key is not a boolean flag", key)
	}

	return f.f[key].isSet, nil
}

//GetString returns a slice of string for each value assiciated with the flag
func (f *Flag) GetString(key string) ([]string, error) {
	if !f.isMono(key) && !f.isMulti(key) {
		return nil, fmt.Errorf("%s key is a boolean flag", key)
	}
	return f.f[key].values, nil
}

//GetInt returns a slice of Int for each value associated with the flag
func (f *Flag) GetInt(key string) ([]int, error) {
	if !f.isMono(key) && !f.isMulti(key) {
		return nil, fmt.Errorf("%s key is a boolean flag", key)
	}
	res := make([]int, 0)
	for _, v := range f.f[key].values {
		i, err := strconv.Atoi(v)
		if err != nil {
			return res, err
		}
		res = append(res, i)
	}
	return res, nil
}

//GetInt8 returns a slice of Int for each value associated with the flag
func (f *Flag) GetInt8(key string) ([]int8, error) {
	if !f.isMono(key) && !f.isMulti(key) {
		return nil, fmt.Errorf("%s key is a boolean flag", key)
	}
	res := make([]int8, 0)
	for _, v := range f.f[key].values {
		i, err := strconv.ParseInt(v, 10, 8)
		if err != nil {
			return res, err
		}
		res = append(res, int8(i))
	}
	return res, nil
}

//GetInt16 returns a slice of Int for each value associated with the flag
func (f *Flag) GetInt16(key string) ([]int16, error) {
	if !f.isMono(key) && !f.isMulti(key) {
		return nil, fmt.Errorf("%s key is a boolean flag", key)
	}
	res := make([]int16, 0)
	for _, v := range f.f[key].values {
		i, err := strconv.ParseInt(v, 10, 16)
		if err != nil {
			return res, err
		}
		res = append(res, int16(i))
	}
	return res, nil
}

//GetInt32 returns a slice of Int for each value associated with the flag
func (f *Flag) GetInt32(key string) ([]int32, error) {
	if !f.isMono(key) && !f.isMulti(key) {
		return nil, fmt.Errorf("%s key is a boolean flag", key)
	}
	res := make([]int32, 0)
	for _, v := range f.f[key].values {
		i, err := strconv.ParseInt(v, 10, 32)
		if err != nil {
			return res, err
		}
		res = append(res, int32(i))
	}
	return res, nil
}

//GetInt64 returns a slice of Int for each value associated with the flag
func (f *Flag) GetInt64(key string) ([]int64, error) {
	if !f.isMono(key) && !f.isMulti(key) {
		return nil, fmt.Errorf("%s key is a boolean flag", key)
	}
	res := make([]int64, 0)
	for _, v := range f.f[key].values {
		i, err := strconv.ParseInt(v, 10, 64)
		if err != nil {
			return res, err
		}
		res = append(res, int64(i))
	}
	return res, nil
}

//GetUint returns a slice of Int for each value associated with the flag
func (f *Flag) GetUint(key string) ([]uint, error) {
	if !f.isMono(key) && !f.isMulti(key) {
		return nil, fmt.Errorf("%s key is a boolean flag", key)
	}
	res := make([]uint, 0)
	for _, v := range f.f[key].values {
		i, err := strconv.ParseUint(v, 10, 0)
		if err != nil {
			return res, err
		}
		res = append(res, uint(i))
	}
	return res, nil
}

//GetUint8 returns a slice of Int for each value associated with the flag
func (f *Flag) GetUint8(key string) ([]uint8, error) {
	if !f.isMono(key) && !f.isMulti(key) {
		return nil, fmt.Errorf("%s key is a boolean flag", key)
	}
	res := make([]uint8, 0)
	for _, v := range f.f[key].values {
		i, err := strconv.ParseUint(v, 10, 8)
		if err != nil {
			return res, err
		}
		res = append(res, uint8(i))
	}
	return res, nil
}

//GetUint16 returns a slice of Int for each value associated with the flag
func (f *Flag) GetUint16(key string) ([]uint16, error) {
	if !f.isMono(key) && !f.isMulti(key) {
		return nil, fmt.Errorf("%s key is a boolean flag", key)
	}
	res := make([]uint16, 0)
	for _, v := range f.f[key].values {
		i, err := strconv.ParseUint(v, 10, 16)
		if err != nil {
			return res, err
		}
		res = append(res, uint16(i))
	}
	return res, nil
}

//GetUint32 returns a slice of Int for each value associated with the flag
func (f *Flag) GetUint32(key string) ([]uint32, error) {
	if !f.isMono(key) && !f.isMulti(key) {
		return nil, fmt.Errorf("%s key is a boolean flag", key)
	}
	res := make([]uint32, 0)
	for _, v := range f.f[key].values {
		i, err := strconv.ParseUint(v, 10, 32)
		if err != nil {
			return res, err
		}
		res = append(res, uint32(i))
	}
	return res, nil
}

//GetUint64 returns a slice of Int for each value associated with the flag
func (f *Flag) GetUint64(key string) ([]uint64, error) {
	if !f.isMono(key) && !f.isMulti(key) {
		return nil, fmt.Errorf("%s key is a boolean flag", key)
	}
	res := make([]uint64, 0)
	for _, v := range f.f[key].values {
		i, err := strconv.ParseUint(v, 10, 64)
		if err != nil {
			return res, err
		}
		res = append(res, uint64(i))
	}
	return res, nil
}

//GetFloat32 returns a slice of Int for each value associated with the flag
func (f *Flag) GetFloat32(key string) ([]float32, error) {
	if !f.isMono(key) && !f.isMulti(key) {
		return nil, fmt.Errorf("%s key is a boolean flag", key)
	}
	res := make([]float32, 0)
	for _, v := range f.f[key].values {
		i, err := strconv.ParseFloat(v, 32)
		if err != nil {
			return res, err
		}
		res = append(res, float32(i))
	}
	return res, nil
}

//GetFloat64 returns a slice of Int for each value associated with the flag
func (f *Flag) GetFloat64(key string) ([]float64, error) {
	if !f.isMono(key) && !f.isMulti(key) {
		return nil, fmt.Errorf("%s key is a boolean flag", key)
	}
	res := make([]float64, 0)
	for _, v := range f.f[key].values {
		i, err := strconv.ParseFloat(v, 64)
		if err != nil {
			return res, err
		}
		res = append(res, float64(i))
	}
	return res, nil
}

//Usage prints command usage on stdout
func (f *Flag) Usage() {
	fmt.Printf("Usage: %s [OPTIONS]\nOptions:\n", os.Args[0])

	fItem := make(map[*flagItem]bool)

	for _, fi := range f.f {
		fItem[fi] = true
	}

	for fi := range fItem {
		if fi.valuation == Multi {
			fmt.Printf("  %s", strings.Join(fi.flags, ", and/or "))
		} else {
			fmt.Printf("  %s", strings.Join(fi.flags, ", or "))
		}

		if len(fi.envFlag) > 0 {
			fmt.Printf(", or set $%s", fi.envFlag)
		}
		if len(fi.separator) > 0 {
			fmt.Printf(" (set multiple values at once using separator '%s')", fi.separator)
		}
		description := strings.Replace(fi.description, "\n", "\n\t\t", -1)
		fmt.Printf("\n\t\t%s\n\n", description)

	}
}
