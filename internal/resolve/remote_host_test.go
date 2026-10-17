package resolve

import "testing"

func TestRemoteHost(t *testing.T) {
	cases := map[string]string{
		"git@gh1:org/repo.git":            "gh1",
		"gh2:org/repo":                    "gh2",
		"ssh://git@gh3/org/repo.git":      "gh3",
		"ssh://git@gh4:2222/org/repo.git": "gh4",
		"https://github.com/org/repo.git": "github.com",
		"/srv/git/repo.git":               "",
		"./relative/path:with-colon":      "",
		"":                                "",
	}
	for remote, want := range cases {
		if got := remoteHost(remote); got != want {
			t.Errorf("remoteHost(%q) = %q, want %q", remote, got, want)
		}
	}
}

func TestRemoteUser(t *testing.T) {
	cases := []struct {
		remote, wantUser, wantHost string
	}{
		{"https://vbrozik@github.com/vbrozik/cp-py-tools.git", "vbrozik", "github.com"},
		{"http://me:secret@ghe.corp:8443/org/repo", "me", "ghe.corp"},
		{"https://github.com/org/repo.git", "", ""},
		{"ssh://git@github.com/org/repo.git", "", ""},
		{"git@github.com:org/repo.git", "", ""},
		{"", "", ""},
	}
	for _, tc := range cases {
		user, host := remoteUser(tc.remote)
		if user != tc.wantUser || host != tc.wantHost {
			t.Errorf("remoteUser(%q) = %q, %q, want %q, %q", tc.remote, user, host, tc.wantUser, tc.wantHost)
		}
	}
}
