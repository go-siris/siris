package qs

import (
	"reflect"
	"strings"
	"testing"
)

type defaultPresenceTestCase struct {
	tagStr    reflect.StructTag
	defaultMP MarshalPresence
	mp        MarshalPresence
	defaultUP UnmarshalPresence
	up        UnmarshalPresence
}

func TestParseTag_DefaultPresence(t *testing.T) {
	testCases := []defaultPresenceTestCase{
		{`qs:"name"`, KeepEmpty, KeepEmpty, Nil, Nil},
		{`qs:"name"`, OmitEmpty, OmitEmpty, Opt, Opt},
		{`qs:"name"`, OmitEmpty, OmitEmpty, Req, Req},
		{`qs:"name,omitempty"`, KeepEmpty, OmitEmpty, Nil, Nil},
		{`qs:"name,keepempty"`, KeepEmpty, KeepEmpty, Nil, Nil},
		{`qs:"name,omitempty"`, OmitEmpty, OmitEmpty, Nil, Nil},
		{`qs:"name,keepempty"`, OmitEmpty, KeepEmpty, Nil, Nil},
		{`qs:"name,nil"`, KeepEmpty, KeepEmpty, Opt, Nil},
		{`qs:"name,opt"`, KeepEmpty, KeepEmpty, Nil, Opt},
		{`qs:"name,req"`, KeepEmpty, KeepEmpty, Nil, Req},
		{`qs:"name,keepempty,opt"`, OmitEmpty, KeepEmpty, Nil, Opt},
	}

	for _, tc := range testCases {
		tag, err := parseFieldTag(tc.tagStr, tc.defaultMP, tc.defaultUP)
		if err != nil {
			t.Errorf("unexpected error - tag: %q :: %v", tc.tagStr, err)
			continue
		}
		if tag.Name != "name" {
			t.Errorf("tag.Name == %q, want %q", tag.Name, "name")
		}
		if tag.MarshalPresence != tc.mp {
			t.Errorf("tag=%q, DefaultMarshalPresence=%v, MarshalPresence=%v, want %v",
				tc.tagStr, tc.defaultMP, tag.MarshalPresence, tc.mp)
		}
		if tag.UnmarshalPresence != tc.up {
			t.Errorf("tag=%q, DefaultUnmarshalPresence=%v, UnmarshalPresence=%v, want %v",
				tc.tagStr, tc.defaultUP, tag.UnmarshalPresence, tc.up)
		}
	}
}

func TestParseTag_SurplusComma(t *testing.T) {
	tagStrList := []reflect.StructTag{
		`qs:","`,
		`qs:"-,"`,
		`qs:"name,"`,
		`qs:",opt,"`,
		`qs:"-,opt,"`,
		`qs:"name,opt,"`,
		`qs:",,opt"`,
		`qs:"-,,opt"`,
		`qs:"name,,opt"`,
	}
	for _, tagStr := range tagStrList {
		_, err := parseFieldTag(tagStr, KeepEmpty, Opt)
		if err == nil {
			t.Errorf("unexpected success - tag: %q", tagStr)
			continue
		}
		if !strings.Contains(err.Error(), "tag string contains a surplus comma") {
			t.Errorf("expected a different error :: %v", err)
		}
	}
}

func TestParseTag_IncompatibleOptions(t *testing.T) {
	tagStrList := []reflect.StructTag{
		`qs:",nil,opt,req"`,
		`qs:",nil,req,opt"`,
		`qs:",opt,req,nil"`,
		`qs:",opt,nil,req"`,
		`qs:",req,nil,opt"`,
		`qs:",req,opt,nil"`,
		`qs:",req,opt"`,
		`qs:",opt,req"`,
		`qs:",req,nil"`,
		`qs:",nil,req"`,
		`qs:",nil,opt"`,
		`qs:",opt,nil"`,
		`qs:",keepempty,omitempty"`,
		`qs:",omitempty,keepempty"`,
	}
	for _, tagStr := range tagStrList {
		_, err := parseFieldTag(tagStr, KeepEmpty, Opt)
		if err == nil {
			t.Errorf("unexpected success - tag: %q", tagStr)
			continue
		}
		if !strings.Contains(err.Error(), "option is allwed - you've specified at least two") {
			t.Errorf("expected a different error :: %v", err)
		}
	}
}

var snakeTestCases = map[string]string{
	"woof_woof":                     "woof_woof",
	"_woof_woof":                    "_woof_woof",
	"woof_woof_":                    "woof_woof_",
	"WOOF":                          "woof",
	"Woof":                          "woof",
	"woof":                          "woof",
	"woof0_woof1":                   "woof0_woof1",
	"_woof0_woof1_2":                "_woof0_woof1_2",
	"woof0_WOOF1_2":                 "woof0_woof1_2",
	"WOOF0":                         "woof0",
	"Woof1":                         "woof1",
	"woof2":                         "woof2",
	"woofWoof":                      "woof_woof",
	"woofWOOF":                      "woof_woof",
	"woof_WOOF":                     "woof_woof",
	"Woof_WOOF":                     "woof_woof",
	"WOOFWoofWoofWOOFWoofWoof":      "woof_woof_woof_woof_woof_woof",
	"WOOF_Woof_woof_WOOF_Woof_woof": "woof_woof_woof_woof_woof_woof",
	"Woof_W":                        "woof_w",
	"Woof_w":                        "woof_w",
	"WoofW":                         "woof_w",
	"Woof_W_":                       "woof_w_",
	"Woof_w_":                       "woof_w_",
	"WoofW_":                        "woof_w_",
	"WOOF_":                         "woof_",
	"W_Woof":                        "w_woof",
	"w_Woof":                        "w_woof",
	"WWoof":                         "w_woof",
	"_W_Woof":                       "_w_woof",
	"_w_Woof":                       "_w_woof",
	"_WWoof":                        "_w_woof",
	"_WOOF":                         "_woof",
	"_woof":                         "_woof",
}

func TestSnakeCase(t *testing.T) {
	for input, output := range snakeTestCases {
		if snakeCase(input) != output {
			t.Errorf("snakeCase(%q) -> %q, want %q", input, snakeCase(input), output)
		}
	}
}
