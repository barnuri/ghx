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
