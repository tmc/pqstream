package pqstream

import (
	"testing"

	google_protobuf "github.com/golang/protobuf/ptypes/struct"
	"github.com/google/go-cmp/cmp"
	"github.com/tmc/pqstream/pqs"
)

func TestServer_redactFields(t *testing.T) {

	rfields := FieldRedactions{
		"public": {"users": []string{
			"password",
			"email",
		},
		},
	}

	s, err := NewServer(testConnectionString, WithFieldRedactions(rfields))
	if err != nil {
		t.Fatal(err)
	}

	event := &pqs.Event{
		Schema: "public",
		Table:  "users",
		Payload: &google_protobuf.Struct{
			Fields: map[string]*google_protobuf.Value{
				"first_name": &google_protobuf.Value{
					Kind: &google_protobuf.Value_StringValue{StringValue: "first_name"},
				},
				"last_name": &google_protobuf.Value{
					Kind: &google_protobuf.Value_StringValue{StringValue: "last_name"},
				},
				"password": &google_protobuf.Value{
					Kind: &google_protobuf.Value_StringValue{StringValue: "_insecure_"},
				},
				"email": &google_protobuf.Value{
					Kind: &google_protobuf.Value_StringValue{StringValue: "someone@corp.com"},
				},
			},
		},
	}

	type args struct {
		redactions FieldRedactions
		incoming   *pqs.Event
		expected   *pqs.Event
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "found",
			args: args{
				redactions: rfields,
				incoming:   event,
				expected: &pqs.Event{
					Schema: "public",
					Table:  "users",
					Payload: &google_protobuf.Struct{
						Fields: map[string]*google_protobuf.Value{
							"first_name": &google_protobuf.Value{
								Kind: &google_protobuf.Value_StringValue{StringValue: "first_name"},
							},
							"last_name": &google_protobuf.Value{
								Kind: &google_protobuf.Value_StringValue{StringValue: "last_name"},
							},
						},
					},
				},
			},
		},
		{
			name: "not_found",
			args: args{
				redactions: rfields,
				incoming:   event,
				expected:   event,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s.redactions = tt.args.redactions
			s.redactFields(tt.args.incoming)

			if got := tt.args.incoming; !cmp.Equal(got, tt.args.expected) {
				t.Errorf("s.redactFields()= %v, want %v", got, tt.args.expected)
			}
		})
	}
}
