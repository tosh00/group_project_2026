package dotenv

import (
	"os"
	"path/filepath"
	"testing"
)

func TestWczytaj(t *testing.T) {
	plik := filepath.Join(t.TempDir(), ".env")
	tresc := `# komentarz
DOTENV_A=1
export DOTENV_B = "z cudzysłowem"
DOTENV_C='pojedynczy'
DOTENV_D=postgres://u:p@h:1/db?sslmode=disable

DOTENV_JUZ=z_pliku
`
	if err := os.WriteFile(plik, []byte(tresc), 0o600); err != nil {
		t.Fatal(err)
	}
	t.Setenv("DOTENV_JUZ", "ze_srodowiska")
	for _, k := range []string{"DOTENV_A", "DOTENV_B", "DOTENV_C", "DOTENV_D"} {
		t.Setenv(k, "") // przywrócenie po teście
		os.Unsetenv(k)
	}

	if err := Wczytaj(plik); err != nil {
		t.Fatal(err)
	}
	chce := map[string]string{
		"DOTENV_A":   "1",
		"DOTENV_B":   "z cudzysłowem",
		"DOTENV_C":   "pojedynczy",
		"DOTENV_D":   "postgres://u:p@h:1/db?sslmode=disable",
		"DOTENV_JUZ": "ze_srodowiska", // środowisko wygrywa
	}
	for k, v := range chce {
		if got := os.Getenv(k); got != v {
			t.Errorf("%s = %q, chcę %q", k, got, v)
		}
	}
}

func TestWczytajBledy(t *testing.T) {
	if err := Wczytaj(filepath.Join(t.TempDir(), "brak")); err != nil {
		t.Errorf("brak pliku: %v", err)
	}
	plik := filepath.Join(t.TempDir(), ".env")
	if err := os.WriteFile(plik, []byte("BEZ_ROWNA_SIE\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := Wczytaj(plik); err == nil {
		t.Error("linia bez = powinna być błędem")
	}
}

func TestWczytajZRepoSzukaWGore(t *testing.T) {
	repo := t.TempDir()
	pakiet := filepath.Join(repo, "internal", "x")
	if err := os.MkdirAll(pakiet, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(repo, "go.mod"), []byte("module x\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(repo, ".env"), []byte("DOTENV_REPO=tak\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	t.Setenv("DOTENV_REPO", "")
	os.Unsetenv("DOTENV_REPO")
	stary, _ := os.Getwd()
	if err := os.Chdir(pakiet); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chdir(stary) })

	if err := WczytajZRepo(".env"); err != nil {
		t.Fatal(err)
	}
	if got := os.Getenv("DOTENV_REPO"); got != "tak" {
		t.Errorf("DOTENV_REPO = %q", got)
	}
}
