package unis

import (
	"reflect"
	"strings"
	"testing"
)

func TestJoinerFunc_Join(t *testing.T) {
	type args struct {
		part1 string
		part2 string
	}
	tests := []struct {
		name string
		j    JoinerFunc
		args args
		want string
	}{
		{
			"join1",
			NewJoiner("/"),
			args{
				"file",
				"path",
			},
			"file/path",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.j.Join(tt.args.part1, tt.args.part2); got != tt.want {
				t.Errorf("JoinerFunc.Join() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNewJoinerChain(t *testing.T) {
	type args struct {
		joiner     Joiner
		processors []Processor
		part1      string
		part2      string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			"joinerchain1",
			args{
				NewJoiner("/"),
				[]Processor{ProcessorFunc(strings.ToLower), NewPrepender("http://")},
				"the",
				"path",
			},
			"http://the/path",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			joinerchain := NewJoinerChain(tt.args.joiner, tt.args.processors...)
			if got := joinerchain.Join(tt.args.part1, tt.args.part2); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewJoinerChain().Join(%s, %s) = %v, want %v", tt.args.part1, tt.args.part2, got, tt.want)
			}
		})
	}
}
