package pqstream

import (
	"bytes"

	jsonpatch "github.com/evanphx/json-patch"
	"github.com/golang/protobuf/jsonpb"
	ptypes_struct "github.com/golang/protobuf/ptypes/struct"
)

func generatePatch(a, b *ptypes_struct.Struct) (*ptypes_struct.Struct, error) {
	abytes := &bytes.Buffer{}
	bbytes := &bytes.Buffer{}
	m := &jsonpb.Marshaler{}

	if a != nil {
		if err := m.Marshal(abytes, a); err != nil {
			return nil, err
		}
	}
	if b != nil {
		if err := m.Marshal(bbytes, b); err != nil {
			return nil, err
		}
	}
	if abytes.Len() == 0 {
		abytes.Write([]byte("{}"))
	}
	if bbytes.Len() == 0 {
		bbytes.Write([]byte("{}"))
	}
	p, err := jsonpatch.CreateMergePatch(abytes.Bytes(), bbytes.Bytes())
	if err != nil {
		return nil, err
	}
	r := &ptypes_struct.Struct{}
	rbytes := bytes.NewReader(p)
	err = (&jsonpb.Unmarshaler{}).Unmarshal(rbytes, r)
	return r, err
}
