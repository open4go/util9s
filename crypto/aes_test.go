package crypto

import "testing"

func TestEncryptDecryptRoundTripAndLegacyFallback(t *testing.T) {
	if err := SetPrimary("0123456789abcdef0123456789abcdef"); err != nil {
		t.Fatal(err)
	}
	ct, err := Encrypt("13800138000")
	if err != nil {
		t.Fatal(err)
	}
	got, err := Decrypt(ct)
	if err != nil || got != "13800138000" {
		t.Fatalf("new key decrypt: %q %v", got, err)
	}

	legacyCT, err := encryptWith([]byte(LegacyHardcodedKey), "13900139000")
	if err != nil {
		t.Fatal(err)
	}
	got, err = Decrypt(legacyCT)
	if err != nil || got != "13900139000" {
		t.Fatalf("legacy fallback: %q %v", got, err)
	}
}

func TestSetPrimaryRejectsDefault(t *testing.T) {
	if err := SetPrimary(LegacyHardcodedKey); err == nil {
		t.Fatal("expected reject default key")
	}
	if err := SetPrimary("short"); err == nil {
		t.Fatal("expected reject short key")
	}
}
