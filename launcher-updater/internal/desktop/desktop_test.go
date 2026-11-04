package desktop

import (
	"strings"
	"testing"
)

func TestRenderTemplateLauncherOnlyChangesMainGroup(t *testing.T) {
	source := []byte(`[Desktop Entry]
Type=Application
Name=Example Tool - description
Name[it]=Strumento di esempio
Icon=example
Exec=example --help

[Desktop Action Alternate]
Name=Alternate
Icon=alternate
Exec=example --alternate
`)

	generated, err := renderTemplateLauncher(source, "example-package")
	if err != nil {
		t.Fatalf("renderTemplateLauncher returned an error: %v", err)
	}
	text := string(generated)
	for _, expected := range []string{
		"Name=[not installed] Example Tool\n",
		"Name[it]=[not installed] Strumento di esempio\n",
		"Icon=software-manager\n",
		"Exec=parrot-exec --install example-package\n",
		"Terminal=true\n",
		"[Desktop Action Alternate]\nName=Alternate\nIcon=alternate\nExec=example --alternate\n",
	} {
		if !strings.Contains(text, expected) {
			t.Errorf("generated launcher does not contain %q:\n%s", expected, text)
		}
	}
}

func TestRenderTemplateLauncherRequiresDesktopEntry(t *testing.T) {
	if _, err := renderTemplateLauncher([]byte("Name=invalid\n"), "example"); err == nil {
		t.Fatal("expected missing Desktop Entry group to be rejected")
	}
}
