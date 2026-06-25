//go:build ignore

// One-off store migration for the scan-profile device_type → os_type rename.
// Run once per CloverDB store:  go run scripts/migrate_phase3a_ostype.go -s <store-dir>
//
// Backs the store up first (<dir>.bak-<ts>), then renames the `device_type`
// field → `os_type` on every `scanprofiles` doc. Idempotent.
package main

import (
	"flag"
	"fmt"
	"log"
	"os/exec"
	"time"

	c "github.com/ostafen/clover/v2"
	d "github.com/ostafen/clover/v2/document"
	q "github.com/ostafen/clover/v2/query"
)

func main() {
	dir := flag.String("s", "", "CloverDB store directory")
	flag.Parse()
	if *dir == "" {
		log.Fatal("usage: go run scripts/migrate_phase3a_ostype.go -s <store-dir>")
	}

	bak := fmt.Sprintf("%s.bak-%s", *dir, time.Now().Format("20060102-150405"))
	if out, err := exec.Command("cp", "-r", *dir, bak).CombinedOutput(); err != nil {
		log.Fatalf("backup failed: %v: %s", err, out)
	}
	fmt.Printf("backed up %s -> %s\n", *dir, bak)

	db, err := c.Open(*dir)
	if err != nil {
		log.Fatalf("open: %v", err)
	}
	defer db.Close()

	if has, _ := db.HasCollection("scanprofiles"); !has {
		fmt.Println("no scanprofiles collection; nothing to do")
		return
	}
	docs, _ := db.FindAll(q.NewQuery("scanprofiles"))
	n := 0
	for _, doc := range docs {
		m := doc.AsMap()
		v, ok := m["device_type"]
		if !ok {
			continue
		}
		m["os_type"] = v
		delete(m, "device_type")
		id, _ := m["_id"].(string)
		if err := db.Delete(q.NewQuery("scanprofiles").Where(q.Field("_id").Eq(id))); err != nil {
			log.Fatalf("delete %s: %v", id, err)
		}
		if _, err := db.InsertOne("scanprofiles", d.NewDocumentOf(m)); err != nil {
			log.Fatalf("reinsert %s: %v", id, err)
		}
		n++
	}
	fmt.Printf("scanprofiles: renamed device_type->os_type on %d doc(s)\ndone.\n", n)
}
