package ghuser

import "testing"

func TestPickEmail(t *testing.T) {
	user := account{Login: "me", ID: 42, Email: "public@x.com"}
	noPublic := account{Login: "me", ID: 42}
	primary := []emailAddress{{Email: "other@x.com"}, {Email: "primary@x.com", Primary: true}}
	cases := []struct {
		name      string
		addresses []emailAddress
		user      account
		host      string
		want      string
	}{
		{"primary wins", primary, user, "github.com", "primary@x.com"},
		{"public without emails scope", nil, user, "github.com", "public@x.com"},
		{"noreply on github.com", nil, noPublic, "github.com", "42+me@users.noreply.github.com"},
		{"no noreply on enterprise", nil, noPublic, "github.example.com", ""},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := pickEmail(tc.addresses, tc.user, tc.host); got != tc.want {
				t.Fatalf("got %q, want %q", got, tc.want)
			}
		})
	}
}
