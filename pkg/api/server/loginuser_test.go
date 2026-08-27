package server

import "testing"

func TestLoginUserFor(t *testing.T) {
	t.Parallel()
	cases := []struct {
		productType string
		distro      string
		want        string
		ok          bool
	}{
		// Verified live against provisioned servers.
		{ProductTypeHPVS, "ubuntu-noble", LoginUserUbuntu, true},
		{ProductTypeHPVS, "debian-trixie", LoginUserDebian, true},
		{ProductTypeHPVS, "almalinux-10", LoginUserAlmaLinux, true},
		{ProductTypeLINVPS, "ubuntu-noble-pvh", LoginUserRoot, true},

		// The same distro differs by product family — the point of the
		// function.
		{ProductTypeHPVS, "ubuntu-jammy", LoginUserUbuntu, true},
		{ProductTypeLINVPS, "ubuntu-jammy-pvh", LoginUserRoot, true},

		// Other releases in a known family resolve by prefix.
		{ProductTypeHPVS, "ubuntu-resolute", LoginUserUbuntu, true},
		{ProductTypeHPVS, "debian-bookworm", LoginUserDebian, true},
		{ProductTypeHPVS, "almalinux-8", LoginUserAlmaLinux, true},

		// Case and whitespace should not matter.
		{"hpvs", "  Ubuntu-Noble  ", LoginUserUbuntu, true},

		// Windows is RDP, not SSH: the lookup must not hand back an
		// account an SSH client would try.
		{ProductTypeWINVPS, "win2019", "", false},
		{ProductTypeHPVS, "win2022", "", false},
		{ProductTypeHPVS, "windows-server-2025", "", false},

		// Unrecognised families and products are not guessed.
		{ProductTypeHPVS, "gentoo", "", false},
		{"DVS", "ubuntu-noble", "", false},
		{"VDSERV", "ubuntu-noble", "", false},
		{"", "ubuntu-noble", "", false},
		{ProductTypeHPVS, "", "", false},
	}
	for _, c := range cases {
		got, ok := LoginUserFor(c.productType, c.distro)
		if ok != c.ok {
			t.Errorf("LoginUserFor(%q, %q) ok = %v, want %v", c.productType, c.distro, ok, c.ok)
			continue
		}
		if got != c.want {
			t.Errorf("LoginUserFor(%q, %q) = %q, want %q", c.productType, c.distro, got, c.want)
		}
	}
}
