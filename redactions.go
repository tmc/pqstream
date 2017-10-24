package pqstream

import (
	"encoding/json"
	"strings"

	"github.com/tmc/pqstream/pqs"
)

// FieldRedactions describes how redaction fields are specified.
// Top level map key is the schema, inner map key is the table and slice is the fields to redact.
type FieldRedactions map[string]map[string][]string

// DecodeRedactions returns a FieldRedactions map decoded from redactions specified in json format.
func DecodeRedactions(r string) (FieldRedactions, error) {
	rfields := make(FieldRedactions)
	if err := json.NewDecoder(strings.NewReader(r)).Decode(&rfields); err != nil {
		return nil, err
	}

	return rfields, nil
}

// WithFieldRedactions controls which fields are redacted from the feed.
func WithFieldRedactions(r FieldRedactions) ServerOption {
	return func(s *Server) {
		s.redactions = r
	}
}

// redactFields search through redactionMap if there's any redacted fields
// specified that match the fields of the current event.
func (s *Server) redactFields(e *pqs.RawEvent) {
	if tables, ok := s.redactions[e.GetSchema()]; ok {
		if fields, ok := tables[e.GetTable()]; ok {
			for _, rf := range fields {
				if e.Payload != nil {
					if _, ok := e.Payload.Fields[rf]; ok {
						//remove field from payload
						delete(e.Payload.Fields, rf)
					}
				}
				if e.Previous != nil {
					if _, ok := e.Previous.Fields[rf]; ok {
						//remove field from previous payload
						delete(e.Previous.Fields, rf)
					}
				}
			}
		}
	}
}
