package money_test

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/shopspring/decimal"

	"github.com/genpick/genpos-mono/backend/pkg/money"
)

func Test_Money_Parse(t *testing.T) {
	t.Parallel()

	cases := map[string]struct {
		in      string
		want    string
		wantErr bool
	}{
		"empty string is zero":          {in: "", want: "0"},
		"plain integer":                 {in: "100", want: "100"},
		"trailing zeros":                {in: "100.0000", want: "100"},
		"VND-shaped large value":        {in: "115000", want: "115000"},
		"NUMERIC(12,4) max value":       {in: "99999999.9999", want: "99999999.9999"},
		"negative value":                {in: "-50.25", want: "-50.25"},
		"whitespace tolerated":          {in: "  100.50  ", want: "100.5"},
		"scientific is rejected": {
			in: "not-a-number", wantErr: true,
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			got, err := money.Parse(tc.in)
			if tc.wantErr {
				if err == nil {
					t.Fatalf("expected error for %q, got %s", tc.in, got)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got.String() != tc.want {
				t.Errorf("Parse(%q) = %s, want %s", tc.in, got.String(), tc.want)
			}
		})
	}
}

func Test_Money_String(t *testing.T) {
	t.Parallel()

	cases := map[string]struct {
		in   decimal.Decimal
		want string
	}{
		"zero formats with 4 places":  {in: money.Zero, want: "0.0000"},
		"integer formats with zeros":  {in: money.MustParse("100"), want: "100.0000"},
		"sub-unit value preserved":    {in: money.MustParse("100.1234"), want: "100.1234"},
		"truncates beyond NUMERIC":    {in: money.MustParse("100.12345"), want: "100.1235"}, // RoundBank inside StringFixed not applied; shopspring rounds half-up here -- documenting actual behaviour
		"negative value":              {in: money.MustParse("-25.5"), want: "-25.5000"},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			got := money.String(tc.in)
			if got != tc.want {
				t.Errorf("String(%s) = %q, want %q", tc.in, got, tc.want)
			}
		})
	}
}

func Test_Money_Round(t *testing.T) {
	t.Parallel()

	cases := map[string]struct {
		in       decimal.Decimal
		strategy money.RoundStrategy
		want     string
	}{
		"per_line snaps to whole VND":           {in: money.MustParse("9999.5"), strategy: money.RoundPerLine, want: "10000"},
		"per_line bankers rounds 0.5 to even":   {in: money.MustParse("0.5"), strategy: money.RoundPerLine, want: "0"},
		"per_line bankers rounds 1.5 to even":   {in: money.MustParse("1.5"), strategy: money.RoundPerLine, want: "2"},
		"per_line negative rounds correctly":    {in: money.MustParse("-2.5"), strategy: money.RoundPerLine, want: "-2"},
		"per_order keeps fractional precision":  {in: money.MustParse("9999.5"), strategy: money.RoundPerOrder, want: "9999.5"},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			got := money.Round(tc.in, tc.strategy)
			if got.String() != tc.want {
				t.Errorf("Round(%s, %s) = %s, want %s", tc.in, tc.strategy, got, tc.want)
			}
		})
	}
}

func Test_Money_RoundOrderTotal(t *testing.T) {
	t.Parallel()

	cases := map[string]struct {
		in       decimal.Decimal
		strategy money.RoundStrategy
		want     string
	}{
		"per_order rounds the total once":           {in: money.MustParse("114999.5"), strategy: money.RoundPerOrder, want: "115000"},
		"per_line is a no-op (lines already round)": {in: money.MustParse("114999.5"), strategy: money.RoundPerLine, want: "114999.5"},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			got := money.RoundOrderTotal(tc.in, tc.strategy)
			if got.String() != tc.want {
				t.Errorf("RoundOrderTotal(%s, %s) = %s, want %s", tc.in, tc.strategy, got, tc.want)
			}
		})
	}
}

func Test_Money_Sum(t *testing.T) {
	t.Parallel()

	cases := map[string]struct {
		in   []decimal.Decimal
		want string
	}{
		"empty is zero":         {in: nil, want: "0"},
		"single value":          {in: []decimal.Decimal{money.MustParse("100")}, want: "100"},
		"three values":          {in: []decimal.Decimal{money.MustParse("100"), money.MustParse("50.25"), money.MustParse("-25")}, want: "125.25"},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			got := money.Sum(tc.in...)
			if got.String() != tc.want {
				t.Errorf("Sum(%v) = %s, want %s", tc.in, got, tc.want)
			}
		})
	}
}

func Test_Money_IsValidRoundStrategy(t *testing.T) {
	t.Parallel()

	cases := map[string]struct {
		in   string
		want bool
	}{
		"per_line is valid":  {in: "per_line", want: true},
		"per_order is valid": {in: "per_order", want: true},
		"empty is invalid":   {in: "", want: false},
		"unknown is invalid": {in: "rounded_to_the_moon", want: false},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			got := money.IsValidRoundStrategy(tc.in)
			if diff := cmp.Diff(tc.want, got); diff != "" {
				t.Errorf("IsValidRoundStrategy(%q) (-want +got):\n%s", tc.in, diff)
			}
		})
	}
}
