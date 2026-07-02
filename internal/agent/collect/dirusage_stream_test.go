package collect

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestDirUsageStreamEmitsProgress(t *testing.T) {
	root := t.TempDir()
	os.WriteFile(filepath.Join(root, "a.txt"), make([]byte, 100), 0644)
	sub := filepath.Join(root, "sub")
	os.Mkdir(sub, 0755)
	os.WriteFile(filepath.Join(sub, "b.bin"), make([]byte, 9000), 0644)

	var emits []DUResult
	DirUsageStream(context.Background(), root, func(r DUResult) { emits = append(emits, r) })

	if len(emits) < 2 {
		t.Fatalf("erwartete mindestens 2 Emits (sofort + final), bekam %d", len(emits))
	}
	// Erster Emit: Ordner als counting, Namen sichtbar.
	first := emits[0]
	var sawCounting bool
	for _, e := range first.Entries {
		if e.Dir && e.Counting {
			sawCounting = true
		}
	}
	if !sawCounting {
		t.Fatalf("erster Emit sollte Ordner als counting markieren: %+v", first.Entries)
	}
	// Letzter Emit: sortiert (sub zuerst), counting aus.
	last := emits[len(emits)-1]
	if last.Entries[0].Name != "sub" || last.Entries[0].Counting || last.Entries[0].Size != 9000 {
		t.Fatalf("finaler Emit falsch: %+v", last.Entries[0])
	}
}
