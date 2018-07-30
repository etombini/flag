package flag

import (
	"os"
	"strings"
	"testing"
)

func TestNewFflag(t *testing.T) {
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

func TestParseMulti(t *testing.T) {
	funcName := "TestGlobal"

	cmd1 := []string{
		"-server", "10.0.0.1",
		"-server", "10.0.0.2",
		"-client", "10.0.0.3",
		"-client", "10.0.0.4",
		"-client", "10.0.0.5",
		"-server", "10.0.0.6,10.0.0.7,10.0.0.8",
	}
	f := NewFlag()
	if err := f.AddMultiFlag("-server", "192.168.0.1", ",", "servers IP addresses"); err != nil {
		t.Errorf("%s error: %s", funcName, err)
	}
	if err := f.AddMultiFlag("-client", "192.168.0.10", ",", "servers IP addresses"); err != nil {
		t.Errorf("%s error: %s", funcName, err)
	}
	if err := f.AddMultiFlag("-default", "default1,default2,default3", ",", "some default value"); err != nil {
		t.Errorf("%s error: %s", funcName, err)
	}

	if err := f.parse(cmd1); err != nil {
		t.Errorf("%s error: %s", funcName, err)
	}
	if err := f.parseEnv(); err != nil {
		t.Errorf("%s error: %s", funcName, err)
	}
	if err := f.parseDefaults(); err != nil {
		t.Errorf("%s error: %s", funcName, err)
	}

	s, err := f.GetString("-server")
	if err != nil {
		t.Errorf("%s error : %s", funcName, err)
	}
	if len(s) != 5 {
		t.Errorf("%s error: exepecting 2 items, returned %d (%v)", funcName, len(s), s)
	}
	for _, v := range s {
		found := false
		if v == "10.0.0.1" || v == "10.0.0.2" || v == "10.0.0.6" || v == "10.0.0.7" || v == "10.0.0.8" {
			found = true
		}
		if !found {
			t.Errorf("%s error: flag values for \"-server\" do not match (%v)", funcName, s)
		}
	}

	c, err := f.GetString("-client")
	if err != nil {
		t.Errorf("%s error : %s", funcName, err)
	}
	if len(c) != 3 {
		t.Errorf("%s error: exepecting 3 items, returned %d (%v)", funcName, len(c), c)
	}
	for _, v := range c {
		found := false
		if v == "10.0.0.3" || v == "10.0.0.4" || v == "10.0.0.5" {
			found = true
		}
		if !found {
			t.Errorf("%s error: flag values for \"-client\" do not match (%v)", funcName, c)
		}
	}

	d, err := f.GetString("-default")
	if err != nil {
		t.Errorf("%s error : %s", funcName, err)
	}
	if len(d) != 3 {
		t.Errorf("%s error: exepecting 3 items, returned %d (%v)", funcName, len(d), d)
	}
	for _, v := range d {
		found := false
		if v == "default1" || v == "default2" || v == "default3" {
			found = true
		}
		if !found {
			t.Errorf("%s error: flag values for \"-default\" do not match (%v)", funcName, d)
		}
	}

	if _, err := f.GetString("-doesnotexist"); err == nil {
		t.Errorf("%s error: requesting values for a flag that does not exist must return an error", funcName)
	}

}

func TestAddBoolFlagsWithEnv(t *testing.T) {
	funcName := "TestAddBoolFlagsWithEnv"
	f := NewFlag()
	if f == nil {
		t.Errorf("%s error: returned value is nil", funcName)
	}

	flagNames := []string{"-a", "--aaaaaa-aaaa", "-aa-aa"}
	envName := "FLAG_TEST_BOOL"
	envValue := "some_value"
	description := "some description"

	os.Setenv(envName, envValue)

	f.AddBoolFlagsWithEnv(flagNames, envName, description)
	if err := f.parseEnv(); err != nil {
		t.Errorf("%s error: parse error: %s", funcName, err)
	}

	for _, flagName := range flagNames {
		ff, ok := f.f[flagName]
		if !ok {
			t.Errorf("%s error: can not get Flag information for %s", funcName, flagName)
		}

		if len(ff.values) != 1 && ff.values[0] != "true" {
			t.Errorf("%s error: boolean flag must return 1 value (true), got %d", funcName, len(ff.values))
		}
		if ff.description != description {
			t.Errorf("%s error: description is not set (got %s)", funcName, ff.description)
		}
		for _, v := range ff.values {
			if v != "true" {
				t.Errorf("%s error: expecting flag set from environment variable %s to be %s, got %s ", funcName, envName, "true", v)
			}
		}
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

	//ti1 := testItem{
	testTable = append(testTable, testItem{
		testName: "multi01",
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
		testName: "multiFromEnv01",
		input: input{
			flags:         []string{"-l", "--long", "--long-is-long"},
			valuation:     Multi,
			envName:       "TEST_LONG",
			defaultValues: []string{"default_value"},
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
				"-l":             []string{"default_value"},
				"--long":         []string{"default_value"},
				"--long-is-long": []string{"default_value"},
			},
		},
	})

	testTable = append(testTable, testItem{
		testName: "multiFromEnvWithSeparator",
		input: input{
			flags:         []string{"-l", "--long", "--long-is-long"},
			valuation:     Multi,
			envName:       "TEST_LONG",
			defaultValues: []string{"default_value01", "default_value02", "default_value03"},
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
				"-l":             []string{"default_value01", "default_value02", "default_value03"},
				"--long":         []string{"default_value01", "default_value02", "default_value03"},
				"--long-is-long": []string{"default_value01", "default_value02", "default_value03"},
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
		}

		if err := f.parseEnv(); err != nil {
			if !ti.output.errParsingEnv {
				t.Errorf("%s error [%s]: parsing environment variables failed: %s", funcName, ti.testName, err)
			}
		}

		if err := f.parseDefaults(); err != nil {
			if !ti.output.errParsingDefaults {
				t.Errorf("%s error [%s]: parsing default values failed: %s", funcName, ti.testName, err)
			}
		}

		for flagName, values := range ti.output.values {
			v, err := f.Get(flagName)
			if err != nil {
				t.Errorf("%s error [%s]: can not get flag %s values: %s", funcName, ti.testName, flagName, err)
			}
			for _, expectedValue := range values {
				found := false
				for _, setValue := range v {
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
