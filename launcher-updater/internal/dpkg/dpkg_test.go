package dpkg

import "testing"

func TestParseInstalledUsesCurrentStatus(t *testing.T) {
	output := []byte("normal\tinstalled\nheld\tinstalled\nremoved\tconfig-files\nhalf\thalf-configured\n")

	installed, err := parseInstalled(output)
	if err != nil {
		t.Fatalf("parseInstalled returned an error: %v", err)
	}
	for _, pkgName := range []string{"normal", "held"} {
		if _, ok := installed[pkgName]; !ok {
			t.Errorf("expected %q to be installed", pkgName)
		}
	}
	for _, pkgName := range []string{"removed", "half"} {
		if _, ok := installed[pkgName]; ok {
			t.Errorf("did not expect %q to be installed", pkgName)
		}
	}
}

func TestParseInstalledRejectsMalformedOutput(t *testing.T) {
	if _, err := parseInstalled([]byte("missing-separator\n")); err == nil {
		t.Fatal("expected malformed output to be rejected")
	}
}
