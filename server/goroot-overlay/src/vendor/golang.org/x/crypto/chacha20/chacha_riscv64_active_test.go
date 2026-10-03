//go:build riscv64

package chacha20

import "testing"

func TestAsmActive(t *testing.T) {
	t.Logf("useXTheadBb = %v", useXTheadBb)
	if !useXTheadBb {
		t.Fatal("assembly path not selected on this core")
	}
}
