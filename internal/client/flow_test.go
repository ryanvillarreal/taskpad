package client

import "testing"

func TestNormalizeID(t *testing.T) {
	cases := []struct {
		in, want string
	}{
		{"04.16.2026", "04.16.2026"},
		{"4.16.2026", "04.16.2026"},
		{"4.6.2026", "04.06.2026"},
		{"4.16.26", "04.16.2026"},
		{"4.6.26", "04.06.2026"},
		{"12.31.99", "12.31.1999"},
		{"01.01.00", "01.01.2000"},
	}
	for _, c := range cases {
		got, err := NormalizeID(c.in)
		if err != nil {
			t.Errorf("%q: unexpected error: %v", c.in, err)
			continue
		}
		if got != c.want {
			t.Errorf("%q: got %q, want %q", c.in, got, c.want)
		}
	}
}

func TestNormalizeID_Invalid(t *testing.T) {
	bad := []string{
		"banana",
		"13.01.2026",
		"04.32.2026",
		"04-16-2026",
		"2026.04.16",
		"",
	}
	for _, in := range bad {
		if _, err := NormalizeID(in); err == nil {
			t.Errorf("%q: expected error, got nil", in)
		}
	}
}
