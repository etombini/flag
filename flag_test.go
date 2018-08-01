package flag

import (
	"os"
	"reflect"
	"strings"
	"testing"
)

func TestNewFlag(t *testing.T) {
	funcName := "TestNewFflag"
	f := NewFlag()
	if f == nil {
		t.Errorf("%s error: returned value is nil", funcName)
	}

}

func helperIsErr(e error) bool {
	if e != nil {
		return true
	}
	return false
}

func TestCheckEnvFormat(t *testing.T) {
	funcName := "TestCheckEnvFormat"

	expected := make(map[string]bool)
	expected[""] = false
	expected["noSpace"] = false
	expected[" leadingSpace"] = true
	expected["trailingSpace "] = true
	expected["inside Space"] = true
	expected["\tleadingSpace"] = true
	expected["\rleadingSpace"] = true
	expected["\nleadingSpace"] = true

	for k, v := range expected {
		if err := checkEnvFormat(k); helperIsErr(err) != v {
			t.Errorf("%s error: expected checkEnvFormat(%s) to return %t, got %v", funcName, k, v, err)
		}
	}
}

func TestCheckFlagFormat(t *testing.T) {
	funcName := "TestCheckFlagFormat"

	expected := make(map[string]bool)
	expected[""] = true
	expected["noSpace"] = false
	expected[" leadingSpace"] = true
	expected["trailingSpace "] = true
	expected["inside Space"] = true
	expected["\tleadingSpace"] = true
	expected["\rleadingSpace"] = true
	expected["\nleadingSpace"] = true

	for k, v := range expected {
		if err := checkFlagFormat(k); helperIsErr(err) != v {
			t.Errorf("%s error: expected checkFlagFormat(%s) to return %t, got %v", funcName, k, v, err)
		}
	}
}

func TestIsValuationNone(t *testing.T) {
	funcName := "TestIsValuation"

	f := NewFlag()
	if f == nil {
		t.Errorf("%s error: can not get new Flag", funcName)
	}
	if err := f.AddBoolFlag("-bool", "boolean flag"); err != nil {
		t.Errorf("%s error: failed to declare a mono valuated flag: %s", funcName, err)
	}
	if !f.isNone("-bool") {
		t.Errorf("%s error: expecting flag to be a boolean flag", funcName)
	}
	if f.isMono("-bool") {
		t.Errorf("%s error: expecting flag to be a boolean flag, not mono valuated", funcName)
	}
	if f.isMulti("-bool") {
		t.Errorf("%s error: expecting flag to be a boolean flag, not multi valuated", funcName)
	}
}

func TestIsValuationMono(t *testing.T) {
	funcName := "TestIsValuationMono"

	f := NewFlag()
	if f == nil {
		t.Errorf("%s error: can not get new Flag", funcName)
	}
	if err := f.AddMonoFlag("-mono", "oneValue", "mono valuated flag"); err != nil {
		t.Errorf("%s error: failed to declare a mono valuated flag: %s", funcName, err)
	}
	if !f.isMono("-mono") {
		t.Errorf("%s error: expecting flag to be mono valuated", funcName)
	}
	if f.isMulti("-mono") {
		t.Errorf("%s error: expecting flag to be mono valuated, not multi valuated", funcName)
	}
	if f.isNone("-mono") {
		t.Errorf("%s error: expecting flag to be mono valuated, not a boolean flag", funcName)
	}
}

func TestIsValuationMulti(t *testing.T) {
	funcName := "TestIsValuationMono"

	f := NewFlag()
	if f == nil {
		t.Errorf("%s error: can not get new Flag", funcName)
	}
	if err := f.AddMultiFlag("-multi", "several|default|values", "|", "multi valuated flag"); err != nil {
		t.Errorf("%s error: failed to declare a multi valuated flag: %s", funcName, err)
	}
	if !f.isMulti("-multi") {
		t.Errorf("%s error: expecting flag to be multi valuated", funcName)
	}
	if f.isMono("-multi") {
		t.Errorf("%s error: expecting flag to be multi valuated, not mono valuated", funcName)
	}
	if f.isNone("-multi") {
		t.Errorf("%s error: expecting flag to be multi valuated, not a boolean flag", funcName)
	}
}

type input struct {
	flags         []string
	envName       string
	valuation     Valuation
	defaultValues []string
	separator     string
	description   string
	setEnv        map[string]string
}

type output struct {
	errInstantiation   bool
	errParsing         bool
	errParsingDefaults bool
	errParsingEnv      bool
	values             map[string][]string
}

type testItem struct {
	testName string
	input    input
	command  []string
	output   output
}

func TestAddMultisWithEnv(t *testing.T) {
	funcName := "TestAddMultisWithEnv"

	testTable := make([]testItem, 0)

	//MULTI
	testTable = append(testTable, testItem{
		testName: "multi-from-command-line",
		input: input{
			flags:         []string{"-l", "--long", "--long-is-long"},
			valuation:     Multi,
			envName:       "TEST_LONG",
			defaultValues: []string{"default_value"},
			separator:     ",",
			description:   "some description",
			setEnv:        make(map[string]string),
		},
		command: []string{"-l", "value1", "--long", "value2", "--long-is-long", "value3"},
		output: output{
			errInstantiation:   false,
			errParsing:         false,
			errParsingDefaults: false,
			errParsingEnv:      false,
			values: map[string][]string{
				"-l":             []string{"value1", "value2", "value3"},
				"--long":         []string{"value1", "value2", "value3"},
				"--long-is-long": []string{"value1", "value2", "value3"},
			},
		},
	})

	testTable = append(testTable, testItem{
		testName: "multi-from-default",
		input: input{
			flags:         []string{"-l", "--long", "--long-is-long"},
			valuation:     Multi,
			envName:       "",
			defaultValues: []string{"default_value01", "default_value02"},
			separator:     ",",
			description:   "some description",
			setEnv:        make(map[string]string),
		},
		command: []string{},
		output: output{
			errInstantiation:   false,
			errParsing:         false,
			errParsingDefaults: false,
			errParsingEnv:      false,
			values: map[string][]string{
				"-l":             []string{"default_value01", "default_value02"},
				"--long":         []string{"default_value01", "default_value02"},
				"--long-is-long": []string{"default_value01", "default_value02"},
			},
		},
	})

	testTable = append(testTable, testItem{
		testName: "multi-from-env-using-separator",
		input: input{
			flags:         []string{"-l", "--long", "--long-is-long"},
			valuation:     Multi,
			envName:       "TEST_LONG",
			defaultValues: []string{"default_value01", "default_value02", "default_value03"},
			separator:     ",",
			description:   "some description",
			setEnv: map[string]string{
				"TEST_LONG": "from_env01,from_env02,from_env03",
			},
		},
		command: []string{},
		output: output{
			errInstantiation:   false,
			errParsing:         false,
			errParsingDefaults: false,
			errParsingEnv:      false,
			values: map[string][]string{
				"-l":             []string{"from_env01", "from_env02", "from_env03"},
				"--long":         []string{"from_env01", "from_env02", "from_env03"},
				"--long-is-long": []string{"from_env01", "from_env02", "from_env03"},
			},
		},
	})

	testTable = append(testTable, testItem{
		testName: "multi-using-separator",
		input: input{
			flags:         []string{"-l", "--long", "--long-is-long"},
			valuation:     Multi,
			envName:       "",
			defaultValues: []string{},
			separator:     ",",
			description:   "some description",
			setEnv:        make(map[string]string),
		},
		command: []string{"-l", "value1,value2", "--long-is-long", "value3"},
		output: output{
			errInstantiation:   false,
			errParsing:         false,
			errParsingDefaults: false,
			errParsingEnv:      false,
			values: map[string][]string{
				"-l":             []string{"value1", "value2", "value3"},
				"--long":         []string{"value1", "value2", "value3"},
				"--long-is-long": []string{"value1", "value2", "value3"},
			},
		},
	})

	testTable = append(testTable, testItem{
		testName: "multi-with-error-because-space-in-flag",
		input: input{
			flags:         []string{"--flag-with space"},
			valuation:     Multi,
			envName:       "",
			defaultValues: []string{"default_value01"},
			separator:     ",",
			description:   "some description",
			setEnv:        make(map[string]string),
		},
		command: []string{"-l", "whatever"},
		output: output{
			errInstantiation:   true,
			errParsing:         false,
			errParsingDefaults: false,
			errParsingEnv:      false,
			values:             map[string][]string{},
		},
	})

	testTable = append(testTable, testItem{
		testName: "multi-with-error-because-space-in-env",
		input: input{
			flags:         []string{"--flag"},
			valuation:     Multi,
			envName:       "ENV_WITH SPACE",
			defaultValues: []string{"default_value01"},
			separator:     ",",
			description:   "some description",
			setEnv:        make(map[string]string),
		},
		command: []string{},
		output: output{
			errInstantiation:   true,
			errParsing:         false,
			errParsingDefaults: false,
			errParsingEnv:      false,
			values:             map[string][]string{},
		},
	})

	//MONO
	testTable = append(testTable, testItem{
		testName: "mono-from-command-line",
		input: input{
			flags:         []string{"-l", "--long", "--long-is-long"},
			valuation:     Mono,
			envName:       "",
			defaultValues: []string{"default_value"},
			separator:     ",",
			description:   "some description",
			setEnv:        make(map[string]string),
		},
		command: []string{"-l", "value1"},
		output: output{
			errInstantiation:   false,
			errParsing:         false,
			errParsingDefaults: false,
			errParsingEnv:      false,
			values: map[string][]string{
				"-l":             []string{"value1"},
				"--long":         []string{"value1"},
				"--long-is-long": []string{"value1"},
			},
		},
	})

	testTable = append(testTable, testItem{
		testName: "mono-from-command-env",
		input: input{
			flags:         []string{"-l", "--long", "--long-is-long"},
			valuation:     Mono,
			envName:       "TEST_MONO_FROM_ENV",
			defaultValues: []string{"default_value"},
			separator:     ",",
			description:   "some description",
			setEnv: map[string]string{
				"TEST_MONO_FROM_ENV": "value1",
			},
		},
		command: []string{},
		output: output{
			errInstantiation:   false,
			errParsing:         false,
			errParsingDefaults: false,
			errParsingEnv:      false,
			values: map[string][]string{
				"-l":             []string{"value1"},
				"--long":         []string{"value1"},
				"--long-is-long": []string{"value1"},
			},
		},
	})

	testTable = append(testTable, testItem{
		testName: "mono-from-command-line-err-when-multivaluated",
		input: input{
			flags:         []string{"-l", "--long", "--long-is-long"},
			valuation:     Mono,
			envName:       "",
			defaultValues: []string{"default_value"},
			separator:     ",",
			description:   "some description",
			setEnv:        make(map[string]string),
		},
		command: []string{"-l", "value1", "--long", "value2"},
		output: output{
			errInstantiation:   false,
			errParsing:         true,
			errParsingDefaults: false,
			errParsingEnv:      false,
			values: map[string][]string{
				"-l":             []string{"value1"},
				"--long":         []string{"value1"},
				"--long-is-long": []string{"value1"},
			},
		},
	})

	testTable = append(testTable, testItem{
		testName: "mono-from-command-line-err-when-multidefaults",
		input: input{
			flags:         []string{"-l", "--long", "--long-is-long"},
			valuation:     Mono,
			envName:       "",
			defaultValues: []string{"default_value01", "default_value02"},
			separator:     ",",
			description:   "some description",
			setEnv:        make(map[string]string),
		},
		command: []string{"-l", "value1", "--long", "value2"},
		output: output{
			errInstantiation:   true,
			errParsing:         false,
			errParsingDefaults: false,
			errParsingEnv:      false,
			values: map[string][]string{
				"-l":             []string{"value1"},
				"--long":         []string{"value1"},
				"--long-is-long": []string{"value1"},
			},
		},
	})

	testTable = append(testTable, testItem{
		testName: "mono-from-command-line-err-when-unknown-flag",
		input: input{
			flags:         []string{"-l", "--long", "--long-is-long"},
			valuation:     Mono,
			envName:       "",
			defaultValues: []string{},
			separator:     ",",
			description:   "some description",
			setEnv:        make(map[string]string),
		},
		command: []string{"-l", "value1", "--unknown-flag", "unknown"},
		output: output{
			errInstantiation:   false,
			errParsing:         true,
			errParsingDefaults: false,
			errParsingEnv:      false,
			values: map[string][]string{
				"-l":             []string{"value1"},
				"--long":         []string{"value1"},
				"--long-is-long": []string{"value1"},
			},
		},
	})

	testTable = append(testTable, testItem{
		testName: "mono-from-command-line-err-when-value-is-missing",
		input: input{
			flags:         []string{"-l", "--long", "--long-is-long"},
			valuation:     Mono,
			envName:       "",
			defaultValues: []string{},
			separator:     ",",
			description:   "some description",
			setEnv:        make(map[string]string),
		},
		command: []string{"-l"},
		output: output{
			errInstantiation:   false,
			errParsing:         true,
			errParsingDefaults: false,
			errParsingEnv:      false,
			values: map[string][]string{
				"-l":             []string{"value1"},
				"--long":         []string{"value1"},
				"--long-is-long": []string{"value1"},
			},
		},
	})

	//BOOL (NONE)
	testTable = append(testTable, testItem{
		testName: "bool",
		input: input{
			flags:         []string{"-f", "--flag"},
			valuation:     None,
			envName:       "",
			defaultValues: []string{},
			separator:     "",
			description:   "some description",
			setEnv:        make(map[string]string),
		},
		command: []string{"-f"},
		output: output{
			errInstantiation:   false,
			errParsing:         false,
			errParsingDefaults: false,
			errParsingEnv:      false,
			values: map[string][]string{
				"-f":     []string{"true"},
				"--flag": []string{"true"},
			},
		},
	})

	testTable = append(testTable, testItem{
		testName: "bool-using-env",
		input: input{
			flags:         []string{"-f", "--flag"},
			valuation:     None,
			envName:       "FLAG_TEST_ENV",
			defaultValues: []string{},
			separator:     "",
			description:   "some description",
			setEnv: map[string]string{
				"FLAG_TEST_ENV": "true",
			},
		},
		command: []string{},
		output: output{
			errInstantiation:   false,
			errParsing:         false,
			errParsingDefaults: false,
			errParsingEnv:      false,
			values: map[string][]string{
				"-f":     []string{"true"},
				"--flag": []string{"true"},
			},
		},
	})

	testTable = append(testTable, testItem{
		testName: "bool-err-with-default-value",
		input: input{
			flags:         []string{"-f", "--flag"},
			valuation:     None,
			envName:       "FLAG_TEST_ENV",
			defaultValues: []string{"true"},
			separator:     "",
			description:   "some description",
			setEnv: map[string]string{
				"FLAG_TEST_ENV": "true",
			},
		},
		command: []string{},
		output: output{
			errInstantiation:   true,
			errParsing:         false,
			errParsingDefaults: false,
			errParsingEnv:      false,
			values: map[string][]string{
				"-f":     []string{"true"},
				"--flag": []string{"true"},
			},
		},
	})

	for _, ti := range testTable {
		//setting up environment variables
		for k, v := range ti.input.setEnv {
			if len(k) == 0 || strings.Contains(k, spaces) {
				t.Errorf("%s error [%s]: incompatible environment variable name [%s]", funcName, ti.testName, k)
			}
			if err := os.Setenv(k, v); err != nil {
				t.Errorf("%s error [%s]: can not set environment variable %s: %s", funcName, ti.testName, k, err)
			}
		}

		//testing instanciation
		f := NewFlag()
		err := f.add(ti.input.flags, ti.input.envName, ti.input.defaultValues, ti.input.valuation, ti.input.separator, ti.input.description)
		if err != nil {
			if !ti.output.errInstantiation {
				t.Errorf("%s error [%s]: %s", funcName, ti.testName, err)
			}
			continue
		}

		for _, flagName := range ti.input.flags {
			fi, ok := f.f[flagName]
			if !ok {
				t.Errorf("%s error [%s]: flag %s is not declared", funcName, ti.testName, flagName)
			}
			if fi.isSet {
				t.Errorf("%s error [%s]: %s is set and is supposed not to be", funcName, ti.testName, flagName)
			}
			if fi.valuation != ti.input.valuation {
				t.Errorf("%s error [%s]: expected valuation %v, returned %v", funcName, ti.testName, ti.input.valuation, fi.valuation)
			}
			if len(ti.input.defaultValues) != len(fi.defaults) {
				t.Errorf("%s error [%s]: expecting %d default values, returned %d (%v)", funcName, ti.testName, len(ti.input.defaultValues), len(fi.defaults), fi.defaults)
			}
			for _, inputDef := range ti.input.defaultValues {
				found := false
				for _, setDef := range fi.defaults {
					if inputDef == setDef {
						found = true
						continue
					}
				}
				if !found {
					t.Errorf("%s error [%s]: default value %s not set", funcName, ti.testName, inputDef)
				}
			}
			if ti.input.separator != fi.separator {
				t.Errorf("%s error [%s]: expecting separator [%s], got [%s]", funcName, ti.testName, ti.input.separator, fi.separator)
			}
			if ti.input.description != fi.description {
				t.Errorf("%s error [%s]: expecting description [%s], got [%s]", funcName, ti.testName, ti.input.description, fi.description)
			}
		}

		//testing parsing
		if err := f.parse(ti.command); err != nil {
			if !ti.output.errParsing {
				t.Errorf("%s error [%s]: parsing failed: %s", funcName, ti.testName, err)
			}
			continue
		}

		if err := f.parseEnv(); err != nil {
			if !ti.output.errParsingEnv {
				t.Errorf("%s error [%s]: parsing environment variables failed: %s", funcName, ti.testName, err)
			}
			continue
		}

		if err := f.parseDefaults(); err != nil {
			if !ti.output.errParsingDefaults {
				t.Errorf("%s error [%s]: parsing default values failed: %s", funcName, ti.testName, err)
			}
			continue
		}

		for flagName, expectedValues := range ti.output.values {
			setValues, err := f.Get(flagName)
			if err != nil {
				t.Errorf("%s error [%s]: can not get flag %s values: %s", funcName, ti.testName, flagName, err)
			}
			for _, expectedValue := range expectedValues {
				found := false
				for _, setValue := range setValues {
					if expectedValue == setValue {
						found = true
					}
				}
				if !found {
					t.Errorf("%s error [%s]: expecting value %s for flag %s to be set", funcName, ti.testName, expectedValue, flagName)
				}
			}
		}

		//cleaning environment variables
		for k := range ti.input.setEnv {
			if err := os.Unsetenv(k); err != nil {
				t.Errorf("%s error [%s]: can not unset environment variable %s: %s", funcName, ti.testName, k, err)
			}
		}

	}

}

func TestGetBool(t *testing.T) {
	funcName := "TestGetBool"
	f := NewFlag()
	if f == nil {
		t.Errorf("%s error: can not create flag", funcName)
	}

	f.AddBoolFlag("-f1", "some flag")
	f.AddBoolFlag("-f2", "some other flag")
	f.parse([]string{"-f1"})

	res1, err := f.GetBool("-f1")
	if err != nil {
		t.Errorf("%s error: can not get boolean flag -f1 status", funcName)
	}
	if !res1 {
		t.Errorf("%s error: -f1 is expected to be true", funcName)
	}

	res2, err := f.GetBool("-f2")
	if err != nil {
		t.Errorf("%s error: can not get boolean flag -f2 status", funcName)
	}
	if res2 {
		t.Errorf("%s error: -f2 is expected to be false", funcName)
	}
	_, err1 := f.GetBool("-f3")
	if err1 == nil {
		t.Errorf("%s error: -f3 flag does not exist, expecting an error", funcName)
	}
}

func TestGetMono(t *testing.T) {
	funcName := "TestGetMono"
	f := NewFlag()
	if f == nil {
		t.Errorf("%s error: can not create flag", funcName)
	}

	values := make(map[string]string)
	values["string"] = "some string"
	values["int"] = "-10"
	values["int8"] = "-10"
	values["int16"] = "-1024"
	values["int32"] = "-65537"
	values["int64"] = "-4294967297"
	values["uint"] = "10"
	values["uint8"] = "10"
	values["uint16"] = "1024"
	values["uint32"] = "65537"
	values["uint64"] = "4294967297"
	values["float32"] = "65537.65537"
	values["float64"] = "4294967297.4294967297"

	for k, v := range values {
		f.AddMonoFlag(k, v, k+v)
	}
	if err := f.parseDefaults(); err != nil {
		t.Errorf("%s error: can not parse default values: %s", funcName, err)
	}

	if _, err := f.Get("-unknown-flag"); err == nil {
		t.Errorf("%s error: expecting an error when getting values for a nonexistent flag", funcName)
	}

	for k := range values {
		switch k {
		case "string":
			slice, err := f.GetString(k)
			if err != nil {
				t.Errorf("%s error: get type %s error %s", funcName, k, err)
			}
			if reflect.TypeOf(slice[0]).String() != k {
				t.Errorf("%s error: wrong type returned", funcName)
			}
		case "int":
			slice, err := f.GetInt(k)
			if err != nil {
				t.Errorf("%s error: get type %s error %s", funcName, k, err)
			}
			if reflect.TypeOf(slice[0]).String() != k {
				t.Errorf("%s error: wrong type returned", funcName)
			}
		case "int8":
			slice, err := f.GetInt8(k)
			if err != nil {
				t.Errorf("%s error: get type %s error %s", funcName, k, err)
			}
			if reflect.TypeOf(slice[0]).String() != k {
				t.Errorf("%s error: wrong type returned", funcName)
			}
		case "int16":
			slice, err := f.GetInt16(k)
			if err != nil {
				t.Errorf("%s error: get type %s error %s", funcName, k, err)
			}
			if reflect.TypeOf(slice[0]).String() != k {
				t.Errorf("%s error: wrong type returned", funcName)
			}
		case "int32":
			slice, err := f.GetInt32(k)
			if err != nil {
				t.Errorf("%s error: get type %s error %s", funcName, k, err)
			}
			if reflect.TypeOf(slice[0]).String() != k {
				t.Errorf("%s error: wrong type returned", funcName)
			}
		case "int64":
			slice, err := f.GetInt64(k)
			if err != nil {
				t.Errorf("%s error: get type %s error %s", funcName, k, err)
			}
			if reflect.TypeOf(slice[0]).String() != k {
				t.Errorf("%s error: wrong type returned", funcName)
			}
		case "uint":
			slice, err := f.GetUint(k)
			if err != nil {
				t.Errorf("%s error: get type %s error %s", funcName, k, err)
			}
			if reflect.TypeOf(slice[0]).String() != k {
				t.Errorf("%s error: wrong type returned", funcName)
			}
		case "uint8":
			slice, err := f.GetUint8(k)
			if err != nil {
				t.Errorf("%s error: get type %s error %s", funcName, k, err)
			}
			if reflect.TypeOf(slice[0]).String() != k {
				t.Errorf("%s error: wrong type returned", funcName)
			}
		case "uint16":
			slice, err := f.GetUint16(k)
			if err != nil {
				t.Errorf("%s error: get type %s error %s", funcName, k, err)
			}
			if reflect.TypeOf(slice[0]).String() != k {
				t.Errorf("%s error: wrong type returned", funcName)
			}
		case "uint32":
			slice, err := f.GetUint32(k)
			if err != nil {
				t.Errorf("%s error: get type %s error %s", funcName, k, err)
			}
			if reflect.TypeOf(slice[0]).String() != k {
				t.Errorf("%s error: wrong type returned", funcName)
			}
		case "uint64":
			slice, err := f.GetUint64(k)
			if err != nil {
				t.Errorf("%s error: get type %s error %s", funcName, k, err)
			}
			if reflect.TypeOf(slice[0]).String() != k {
				t.Errorf("%s error: wrong type returned", funcName)
			}
		case "float32":
			slice, err := f.GetFloat32(k)
			if err != nil {
				t.Errorf("%s error: get type %s error %s", funcName, k, err)
			}
			if reflect.TypeOf(slice[0]).String() != k {
				t.Errorf("%s error: wrong type returned", funcName)
			}
		case "float64":
			slice, err := f.GetFloat64(k)
			if err != nil {
				t.Errorf("%s error: get type %s error %s", funcName, k, err)
			}
			if reflect.TypeOf(slice[0]).String() != k {
				t.Errorf("%s error: wrong type returned", funcName)
			}
		default:
			t.Errorf("%s error: unknown type to convert %s", funcName, k)
		}
	}

}
