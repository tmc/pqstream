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

	event := &pqs.RawEvent{
		Schema: "public",
		Table:  "users",
		Payload: &google_protobuf.Struct{
			Fields: map[string]*google_protobuf.Value{
				"first_name": {
					Kind: &google_protobuf.Value_StringValue{StringValue: "first_name"},
				},
				"last_name": {
					Kind: &google_protobuf.Value_StringValue{StringValue: "last_name"},
				},
				"password": {
					Kind: &google_protobuf.Value_StringValue{StringValue: "_insecure_"},
				},
				"email": {
					Kind: &google_protobuf.Value_StringValue{StringValue: "someone@corp.com"},
				},
			},
		},
	}

	type args struct {
		redactions FieldRedactions
		incoming   *pqs.RawEvent
		expected   *pqs.RawEvent
	}
	tests := []struct {
		name string
		args args
	}{
		{"nil", args{redactions: rfields, incoming: nil}},
		{"nil_payload", args{redactions: rfields, incoming: &pqs.RawEvent{}}},
		{"nil_payload_matching", args{redactions: rfields, incoming: &pqs.RawEvent{
			Schema: "public",
			Table:  "users",
		}}},
		{"nil_payload_nonnil_previous", args{redactions: rfields, incoming: &pqs.RawEvent{
			Schema: "public",
			Table:  "users",
			Previous: &google_protobuf.Struct{
				Fields: map[string]*google_protobuf.Value{
					"password": {
						Kind: &google_protobuf.Value_StringValue{StringValue: "password"},
					},
				},
			},
		}}},
		{
			name: "found",
			args: args{
				redactions: rfields,
				incoming:   event,
				expected: &pqs.RawEvent{
					Schema: "public",
					Table:  "users",
					Payload: &google_protobuf.Struct{
						Fields: map[string]*google_protobuf.Value{
							"first_name": {
								Kind: &google_protobuf.Value_StringValue{StringValue: "first_name"},
							},
							"last_name": {
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

			if got := tt.args.incoming; tt.args.expected != nil && !cmp.Equal(got, tt.args.expected) {
				t.Errorf("s.redactFields()= %v, want %v", got, tt.args.expected)
			}
		})
	}
}

func TestDecodeRedactions(t *testing.T) {
	type args struct {
		r string
	}
	tests := []struct {
		name    string
		args    args
		want    FieldRedactions
		wantErr bool
	}{
		{
			name: "basic",
			args: args{r: `{"public":{"users":["first_name","last_name","email"]}}`},
			want: FieldRedactions{
				"public": {"users": []string{
					"first_name",
					"last_name",
					"email",
				},
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := DecodeRedactions(tt.args.r)
			if (err != nil) != tt.wantErr {
				t.Errorf("DecodeRedactions() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !cmp.Equal(got, tt.want) {
				t.Errorf("DecodeRedactions() = %v, want %v", got, tt.want)
			}
		})
	}
}
