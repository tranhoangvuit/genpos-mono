package datastore

import (
	"time"

	"github.com/jackc/pgx/v5/pgtype"
)

// textOrNull converts a Go string to pgtype.Text, treating empty as NULL.
func textOrNull(s string) pgtype.Text {
	if s == "" {
		return pgtype.Text{Valid: false}
	}
	return pgtype.Text{String: s, Valid: true}
}

func textString(t pgtype.Text) string {
	if !t.Valid {
		return ""
	}
	return t.String
}

// uuidOrNull converts a Go string to pgtype.UUID, treating empty as NULL.
func uuidOrNull(s string) (pgtype.UUID, error) {
	if s == "" {
		return pgtype.UUID{Valid: false}, nil
	}
	return parseUUID(s)
}

// uuidStrings renders a slice of pgtype.UUID into Go strings, skipping NULLs.
func uuidStrings(ids []pgtype.UUID) []string {
	out := make([]string, 0, len(ids))
	for _, id := range ids {
		if !id.Valid {
			continue
		}
		out = append(out, uuidString(id))
	}
	return out
}

// numericFromString parses a decimal string into pgtype.Numeric.
func numericFromString(s string) (pgtype.Numeric, error) {
	var n pgtype.Numeric
	if s == "" {
		s = "0"
	}
	if err := n.Scan(s); err != nil {
		return n, err
	}
	return n, nil
}

// numericOrNull parses a decimal string into pgtype.Numeric, treating empty as NULL.
func numericOrNull(s string) (pgtype.Numeric, error) {
	if s == "" {
		return pgtype.Numeric{Valid: false}, nil
	}
	return numericFromString(s)
}

// timestampOrNull converts a Go time to pgtype.Timestamptz, treating zero as NULL.
func timestampOrNull(t time.Time) pgtype.Timestamptz {
	if t.IsZero() {
		return pgtype.Timestamptz{Valid: false}
	}
	return pgtype.Timestamptz{Time: t, Valid: true}
}

// timestampTime extracts a time.Time from pgtype.Timestamptz (zero if NULL).
func timestampTime(t pgtype.Timestamptz) time.Time {
	if !t.Valid {
		return time.Time{}
	}
	return t.Time
}

// dateOrNull converts a Go time to pgtype.Date (date-only), zero time = NULL.
func dateOrNull(t time.Time) pgtype.Date {
	if t.IsZero() {
		return pgtype.Date{Valid: false}
	}
	return pgtype.Date{Time: t, Valid: true}
}

// dateTime extracts a time.Time from pgtype.Date (zero if NULL).
func dateTime(d pgtype.Date) time.Time {
	if !d.Valid {
		return time.Time{}
	}
	return d.Time
}

// numericToString renders a pgtype.Numeric as a decimal string.
func numericToString(n pgtype.Numeric) string {
	if !n.Valid {
		return "0"
	}
	b, err := n.MarshalJSON()
	if err != nil {
		return "0"
	}
	// MarshalJSON wraps in quotes; strip them.
	if len(b) >= 2 && b[0] == '"' && b[len(b)-1] == '"' {
		return string(b[1 : len(b)-1])
	}
	return string(b)
}
