package model

import (
	"reflect"
	"testing"
)

func TestGetChannelLimits(t *testing.T) {
	tests := []struct {
		name    string
		raw     string
		want    []int
		wantNil bool
	}{
		{"empty", "", nil, true},
		{"single", "3", []int{3}, false},
		{"multiple", "3,7,11", []int{3, 7, 11}, false},
		{"with spaces", " 3 , 7 , 11 ", []int{3, 7, 11}, false},
		{"skip invalid", "1,abc,2,,3", []int{1, 2, 3}, false},
		{"skip zero and negative", "0,-1,5", []int{5}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			token := &Token{ChannelLimits: tt.raw}
			got := token.GetChannelLimits()
			if tt.wantNil {
				if got != nil {
					t.Fatalf("expected nil, got %v", got)
				}
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Fatalf("expected %v, got %v", tt.want, got)
			}
		})
	}
}