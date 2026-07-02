package collect

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestDirUsageBasic(t *testing.T) {
	root := t.TempDir()
	os.WriteFile(filepath.Join(root, "a.txt"), make([]byte, 100), 0644)
	sub := filepath.Join(root, "sub")
	os.Mkdir(sub, 0755)
	os.WriteFile(filepath.Join(sub, "b.bin"), make([]byte, 5000), 0644)
	os.WriteFile(filepath.Join(sub, "c.bin"), make([]byte, 3000), 0644)

	var res DUResult
	if err := json.Unmarshal([]byte(DirUsage(context.Background(), root)), &res); err != nil {
		t.Fatal(err)
	}
	if len(res.Entries) != 2 {
		t.Fatalf("erwartete 2 Einträge, bekam %d: %+v", len(res.Entries), res.Entries)
	}
	// größter zuerst: sub (8000) vor a.txt (100)
	if !res.Entries[0].Dir || res.Entries[0].Size != 8000 || res.Entries[0].Items != 2 {
		t.Fatalf("sub falsch: %+v", res.Entries[0])
	}
	if res.Entries[1].Dir || res.Entries[1].Size != 100 {
		t.Fatalf("a.txt falsch: %+v", res.Entries[1])
	}
}
