package httptransport

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/TuneLab/go-truss/gengokit/gentesthelper"
	"github.com/TuneLab/go-truss/svcdef"
	"github.com/davecgh/go-spew/spew"
)

var (
	_ = spew.Sdump
)

var gopath []string

func init() {
	gopath = filepath.SplitList(os.Getenv("GOPATH"))
}

func TestNewMethod(t *testing.T) {
	defStr := `
		syntax = "proto3";

		// General package
		package general;

		import "github.com/TuneLab/go-truss/deftree/googlethirdparty/annotations.proto";

		message SumRequest {
			int64 a = 1;
			int64 b = 2;
		}

		message SumReply {
			int64 v = 1;
			string err = 2;
		}

		service SumSvc {
			rpc Sum(SumRequest) returns (SumReply) {
				option (google.api.http) = {
					get: "/sum/{a}"
				};
			}
		}
	`
	sd, err := svcdef.NewFromString(defStr, gopath)
	if err != nil {
		t.Fatal(err, "Failed to create a service from the definition string")
	}
	binding := &Binding{
		Label:        "SumZero",
		PathTemplate: "/sum/{a}",
		BasePath:     "/sum/",
		Verb:         "get",
		Fields: []*Field{
			&Field{
				Name:                       "A",
				QueryParamName:             "a",
				CamelName:                  "A",
				LowCamelName:               "a",
				LocalName:                  "ASum",
				Location:                   "path",
				GoType:                     "int64",
				ConvertFunc:                "ASum, err := strconv.ParseInt(ASumStr, 10, 64)",
				ConvertFuncNeedsErrorCheck: true,
				TypeConversion:             "ASum",
				IsBaseType:                 true,
			},
			&Field{
				Name:                       "B",
				QueryParamName:             "b",
				CamelName:                  "B",
				LowCamelName:               "b",
				LocalName:                  "BSum",
				Location:                   "query",
				GoType:                     "int64",
				ConvertFunc:                "BSum, err := strconv.ParseInt(BSumStr, 10, 64)",
				ConvertFuncNeedsErrorCheck: true,
				TypeConversion:             "BSum",
				IsBaseType:                 true,
			},
		},
	}
	meth := &Method{
		Name:         "Sum",
		RequestType:  "SumRequest",
		ResponseType: "SumReply",
		Bindings: []*Binding{
			binding,
		},
	}
	binding.Parent = meth

	newMeth := NewMethod(sd.Service.Methods[0])
	if got, want := newMeth, meth; !reflect.DeepEqual(got, want) {
		diff := gentesthelper.DiffStrings(spew.Sdump(got), spew.Sdump(want))
		t.Errorf("got != want; methods differ: %v\n", diff)
	}
}

func TestPathParams(t *testing.T) {
	var cases = []struct {
		url, tmpl, field, want string
	}{
		{"/1234", "/{a}", "a", "1234"},
		{"/v1/1234", "/v1/{a}", "a", "1234"},
		{"/v1/user/5/home", "/v1/user/{userid}/home", "userid", "5"},
		{"/blah/", "/{a}", "a", "blah"},
	}

	for _, test := range cases {
		ret, err := PathParams(test.url, test.tmpl)
		if err != nil {
			t.Errorf("PathParams returned error '%v' on case '%+v'\n", err, test)
		}
		if got, ok := ret[test.field]; ok {
			if got != test.want {
				t.Errorf("PathParams got '%v', want '%v'\n", got, test.want)
			}
		} else {
			t.Errorf("PathParams didn't return map containing field '%v'\n", test.field)
		}
	}
}

// Test that the PathParams function will correctly fail
func TestPathParamsFailure(t *testing.T) {
	var cases = []struct {
		url, tmpl string
	}{
		{"/too/few/params", "/{a}/{b}/{c}/{d}"},
		{"/way/too/many/params", "/{a}"},
	}

	for _, test := range cases {
		_, err := PathParams(test.url, test.tmpl)
		if err == nil {
			t.Errorf("PathParams returned no error when it should have returned an error on case '%+v'\n", test)
		}
	}
}

func TestFuncSourceCode(t *testing.T) {
	_, err := FuncSourceCode(PathParams)
	if err != nil {
		t.Fatalf("Failed to get source code: %s\n", err)
	}
}

func TestAllFuncSourceCode(t *testing.T) {
	_, err := AllFuncSourceCode(PathParams)
	if err != nil {
		t.Fatalf("Failed to get source code: %s\n", err)
	}
}

func TestEnglishNumber(t *testing.T) {
	var cases = []struct {
		i    int
		want string
	}{
		{0, "Zero"},
		{1, "One"},
		{2, "Two"},
		{3, "Three"},
		{4, "Four"},
		{5, "Five"},
		{6, "Six"},
		{7, "Seven"},
		{8, "Eight"},
		{9, "Nine"},

		{11, "OneOne"},
		{22, "TwoTwo"},
		{23, "TwoThree"},
	}

	for _, test := range cases {
		got := EnglishNumber(test.i)
		if got != test.want {
			t.Errorf("Got %v, want %v\n", got, test.want)
		}
	}
}

func TestLowCamelName(t *testing.T) {
	var cases = []struct {
		name, want string
	}{
		{"what", "what"},
		{"example_one", "exampleOne"},
		{"another_example_case", "anotherExampleCase"},
		{"_leading_camel", "xLeadingCamel"},
		{"_a", "xA"},
		{"a", "a"},
	}

	for _, test := range cases {
		got := LowCamelName(test.name)
		if got != test.want {
			t.Errorf("Got %v, want %v\n", got, test.want)
		}
	}
}
