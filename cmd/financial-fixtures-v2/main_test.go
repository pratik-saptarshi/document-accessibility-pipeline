package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestReadObservationRecordsRequiresFourBoundedRows(t *testing.T) {
	const hash = "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
	records := strings.Join([]string{
		"statement-positive-v2\t" + hash + "\t" + hash + "\t" + hash,
		"statement-damaged-v2\t" + hash + "\t" + hash + "\t" + hash,
		"invoice-positive-v2\t" + hash + "\t" + hash + "\t" + hash,
		"invoice-damaged-v2\t" + hash + "\t" + hash + "\t" + hash,
	}, "\n") + "\n"
	path := filepath.Join(t.TempDir(), "records.tsv")
	if err := os.WriteFile(path, []byte(records), 0o600); err != nil {
		t.Fatal(err)
	}
	got, err := readObservationRecords(path)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 4 || got[0].Name != "statement-positive-v2" || got[3].Name != "invoice-damaged-v2" {
		t.Fatalf("observation records = %#v", got)
	}
	if err := os.WriteFile(path, []byte("statement-positive-v2\t"+hash+"\t"+hash+"\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := readObservationRecords(path); err == nil || !strings.Contains(err.Error(), "fields") {
		t.Fatalf("malformed TSV error = %v, want field-count rejection", err)
	}
}
