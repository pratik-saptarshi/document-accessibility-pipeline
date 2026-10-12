package pdfua

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRootPDFUAContractContainsNoProcessOrJavaAdapter(t *testing.T) {
	entries, err := os.ReadDir(".")
	if err != nil {
		t.Fatal(err)
	}
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".go") || strings.HasSuffix(entry.Name(), "_test.go") {
			continue
		}
		data, err := os.ReadFile(filepath.Clean(entry.Name()))
		if err != nil {
			t.Fatal(err)
		}
		for _, forbidden := range []string{"\"os/exec\"", "exec.Command", "java -jar", "--flavour ua2"} {
			if strings.Contains(string(data), forbidden) {
				t.Errorf("root Go report package contains validator execution marker %q in %s", forbidden, entry.Name())
			}
		}
	}
}
