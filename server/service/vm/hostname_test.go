package vm

import "testing"

func TestValidHostname(t *testing.T) {
	for _, name := range []string{
		"nanokvm", "kvm-1", "a", "KVM", "rack1.lab", "1kvm",
		"a23456789012345678901234567890123456789012345678901234567890123",
	} {
		if !validHostname(name) {
			t.Errorf("refused %q", name)
		}
	}

	for _, name := range []string{
		"", "-kvm", "kvm-", "kvm_1", "kvm 1", "kvm\n", "kvm/1", "kvm.", ".kvm", "a..b",
		"a234567890123456789012345678901234567890123456789012345678901234",
		"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa.aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
	} {
		if validHostname(name) {
			t.Errorf("accepted %q", name)
		}
	}
}

func TestRenameHostChangesOnlyWholeNames(t *testing.T) {
	hosts := "127.0.0.1\tlocalhost\n127.0.1.1 local local.lan # local\n::1 ip6-localhost\n"
	got := renameHost(hosts, "local", "kvm")
	want := "127.0.0.1\tlocalhost\n127.0.1.1 kvm local.lan # local\n::1 ip6-localhost\n"
	if got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}

func TestRenameHostLeavesTheAddressAlone(t *testing.T) {
	hosts := "kvm kvm\n"
	if got := renameHost(hosts, "kvm", "rack"); got != "kvm rack\n" {
		t.Fatalf("got %q", got)
	}
}

func TestRenameHostWithNoOldNameChangesNothing(t *testing.T) {
	hosts := "127.0.0.1 localhost\n"
	if got := renameHost(hosts, "", "kvm"); got != hosts {
		t.Fatalf("got %q", got)
	}
}
