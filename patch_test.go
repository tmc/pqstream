package pqstream

import (
	"testing"

	"github.com/golang/protobuf/jsonpb"
	ptypes_struct "github.com/golang/protobuf/ptypes/struct"
	"github.com/google/go-cmp/cmp"
)

func Test_generatePatch(t *testing.T) {
	type args struct {
		a *ptypes_struct.Struct
		b *ptypes_struct.Struct
	}
	tests := []struct {
		name     string
		args     args
		wantJson string
		wantErr  bool
	}{
		{"nils", args{nil, nil}, "{}", false},
		{"empties", args{&ptypes_struct.Struct{}, &ptypes_struct.Struct{}}, "{}", false},
		{"basic", args{&ptypes_struct.Struct{}, &ptypes_struct.Struct{
			map[string]*ptypes_struct.Value{
				"foo": &ptypes_struct.Value{
					Kind: &ptypes_struct.Value_StringValue{
						StringValue: "bar",
					},
				},
			},
		}}, `{"foo":"bar"}`, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := generatePatch(tt.args.a, tt.args.b)
			if (err != nil) != tt.wantErr {
				t.Errorf("generatePatch() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			gotJson, err := (&jsonpb.Marshaler{}).MarshalToString(got)
			if err != nil {
				t.Error(err)
			}
			if !cmp.Equal(gotJson, tt.wantJson) {
				t.Errorf("generatePatch() = %v, want %v\n%s", gotJson, tt.wantJson, cmp.Diff(gotJson, tt.wantJson))
			}
		})
	}
}
