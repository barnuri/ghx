package config

import "testing"

func TestParseOrgs(t *testing.T) {
	cfg, err := Parse([]byte("profiles: {work: {user: a}}\norgs: {\"chkp-*\": work}"))
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Orgs["chkp-*"] != "work" {
		t.Fatalf("unexpected orgs %+v", cfg.Orgs)
	}
}

func TestParseRejectsInvalidOrgs(t *testing.T) {
	cases := map[string]string{
		"unknown org profile": "profiles: {work: {user: a}}\norgs: {chkp-*: nope}",
		"bad org pattern":     "profiles: {work: {user: a}}\norgs: {\"chkp-[\": work}",
	}
	for name, content := range cases {
		t.Run(name, func(t *testing.T) {
			if _, err := Parse([]byte(content)); err == nil {
				t.Fatal("expected error")
			}
		})
	}
}
