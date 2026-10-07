package roommic

import (
	"os"
	"path/filepath"
	"testing"
)

// The card list of slot B, the mainline kernel (ironkvm-dist trial 57).
const slotBCards = ` 0 [sg2002onboard  ]: simple-card - sg2002-onboard
                      sg2002-onboard
 1 [UAC1Gadget     ]: UAC1_Gadget - UAC1_Gadget
                      UAC1_Gadget 0
`

// The card list of slot A, the vendor kernel. Recording from its codec hangs
// the board, so this must never count as having the microphone.
const slotACards = ` 0 [cv182xaadc     ]: cv182xa_adc - cv182xa_adc
                      cv182xa_adc
 1 [cv182xadac     ]: cv182xa_dac - cv182xa_dac
                      cv182xa_dac
 2 [UAC1Gadget     ]: UAC1_Gadget - UAC1_Gadget
                      UAC1_Gadget 0
`

func withCards(t *testing.T, cards string) {
	t.Helper()

	path := filepath.Join(t.TempDir(), "cards")
	if err := os.WriteFile(path, []byte(cards), 0o644); err != nil {
		t.Fatal(err)
	}

	original := cardsPath
	cardsPath = path
	t.Cleanup(func() { cardsPath = original })
}

func TestTheMainlineKernelHasTheMicrophone(t *testing.T) {
	withCards(t, slotBCards)

	if !Available() {
		t.Fatal("slot B's card list should offer the microphone")
	}
}

func TestTheVendorKernelDoesNot(t *testing.T) {
	withCards(t, slotACards)

	if Available() {
		t.Fatal("slot A's codec must never be opened: recording from it hangs the board")
	}
}

func TestNoCardListMeansNoMicrophone(t *testing.T) {
	original := cardsPath
	cardsPath = filepath.Join(t.TempDir(), "absent")
	t.Cleanup(func() { cardsPath = original })

	if Available() {
		t.Fatal("a missing card list should mean no microphone")
	}
}

// Only the card id counts. A long name, a driver name or a gadget that carries
// the word somewhere else is not the card.
func TestOnlyTheCardIdCounts(t *testing.T) {
	cases := map[string]string{
		"in the long name":   " 0 [other          ]: simple-card - sg2002onboard\n",
		"a longer id":        " 0 [sg2002onboard2 ]: simple-card - x\n",
		"on a second line":   " 0 [other          ]: x - y\n                      [sg2002onboard]\n",
		"empty list":         "--- no soundcards ---\n",
		"no index before it": "[sg2002onboard  ]: simple-card - sg2002-onboard\n",
	}

	for name, cards := range cases {
		if hasCard([]byte(cards), CardName) {
			t.Errorf("%s: matched %q", name, cards)
		}
	}

	if !hasCard([]byte("10 [sg2002onboard]: simple-card - sg2002-onboard\n"), CardName) {
		t.Error("a two-digit index or an unpadded id should still match")
	}
}

func TestTheDeviceIsOpenedByName(t *testing.T) {
	if Device != "hw:sg2002onboard,1" {
		t.Fatalf("device %q: the microphone is device 1 of the card, opened by name", Device)
	}
	if Format.Channels != 1 {
		t.Fatalf("the codec captures one channel, the format asks for %d", Format.Channels)
	}
}
