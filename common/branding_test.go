package common

import "testing"

func TestNormalizeSystemName(t *testing.T) {
	for _, legacyName := range []string{
		"",
		"New API",
		"new-api",
		"newapi",
		"robocoding",
		"RoboCodingAI",
		"  NEW API  ",
	} {
		if got := NormalizeSystemName(legacyName); got != DefaultSystemName {
			t.Fatalf("NormalizeSystemName(%q) = %q, want %q", legacyName, got, DefaultSystemName)
		}
	}
}

func TestNormalizeSystemNamePreservesCustomBrand(t *testing.T) {
	if got := NormalizeSystemName("  Acme Robotics  "); got != "Acme Robotics" {
		t.Fatalf("NormalizeSystemName() = %q, want custom brand", got)
	}
}

func TestNormalizeSystemLogo(t *testing.T) {
	for _, legacyLogo := range []string{"", "/logo.png", "  /logo.png  "} {
		if got := NormalizeSystemLogo(legacyLogo); got != DefaultLogo {
			t.Fatalf("NormalizeSystemLogo(%q) = %q, want %q", legacyLogo, got, DefaultLogo)
		}
	}

	const customLogo = "https://cdn.example.com/acme.svg"
	if got := NormalizeSystemLogo("  " + customLogo + "  "); got != customLogo {
		t.Fatalf("NormalizeSystemLogo() = %q, want custom logo", got)
	}
}
