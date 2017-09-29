package pqstream

import "github.com/tmc/pqstream/pqs"

// FieldRedactions describes how redaction fields are specified.
// Top level map key is the schema, inner map key is the table and slice is the fields to redact.
type FieldRedactions map[string]map[string][]string

// WithFieldRedactions controls which fields are redacted from the feed.
func WithFieldRedactions(r FieldRedactions) ServerOption {
	return func(s *Server) {
		s.redactions = r
	}
}

// redactFields search through redactionMap if there's any redacted fields
// specified that match the fields of the current event.
func (s *Server) redactFields(e *pqs.Event) {
	if tables, ok := s.redactions[e.GetSchema()]; ok {
		if fields, ok := tables[e.GetTable()]; ok {
			for _, rf := range fields {
				if _, ok := e.Payload.Fields[rf]; ok {
					//remove field from payload
					delete(e.Payload.Fields, rf)
				}
			}
		}
	}
}
