//go:build ignore

// One-off store migration for the proprietary → owner rename.
// Run once per CloverDB store:  go run scripts/migrate_phase1_owner.go -s <store-dir>
//
// It backs the store up first (<dir>.bak-<ts>), then:
//   1. copies the `proprietaries` collection into `owners`, renaming each doc's
//      `proprietary` field → `owner` and PRESERVING `_id` (devices/zones
//      reference owners by id), and drops `proprietaries`;
//   2. renames the `proprietary` field → `owner` on every `devices`/`zones` doc.
// Idempotent: re-running when already migrated is a no-op.
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
		log.Fatal("usage: go run scripts/migrate_phase1_owner.go -s <store-dir>")
	}

	// 1. Backup.
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

	// 2. proprietaries -> owners (preserve _id, rename field within the doc).
	if has, _ := db.HasCollection("proprietaries"); has {
		if has2, _ := db.HasCollection("owners"); !has2 {
			if err := db.CreateCollection("owners"); err != nil {
				log.Fatalf("create owners: %v", err)
			}
		}
		docs, err := db.FindAll(q.NewQuery("proprietaries"))
		if err != nil {
			log.Fatalf("read proprietaries: %v", err)
		}
		for _, doc := range docs {
			m := doc.AsMap()
			if v, ok := m["proprietary"]; ok {
				m["owner"] = v
				delete(m, "proprietary")
			}
			if _, err := db.InsertOne("owners", d.NewDocumentOf(m)); err != nil {
				log.Fatalf("insert owner: %v", err)
			}
		}
		if err := db.DropCollection("proprietaries"); err != nil {
			log.Fatalf("drop proprietaries: %v", err)
		}
		fmt.Printf("migrated %d owner(s); dropped proprietaries\n", len(docs))
	} else {
		fmt.Println("no proprietaries collection (already migrated?)")
	}

	// 3. devices/zones: rename the proprietary field -> owner in place.
	for _, coll := range []string{"devices", "zones"} {
		if has, _ := db.HasCollection(coll); !has {
			continue
		}
		docs, _ := db.FindAll(q.NewQuery(coll))
		n := 0
		for _, doc := range docs {
			m := doc.AsMap()
			v, ok := m["proprietary"]
			if !ok {
				continue
			}
			m["owner"] = v
			delete(m, "proprietary")
			id, _ := m["_id"].(string)
			// Replace the doc so the stale `proprietary` key is removed.
			if err := db.Delete(q.NewQuery(coll).Where(q.Field("_id").Eq(id))); err != nil {
				log.Fatalf("%s delete %s: %v", coll, id, err)
			}
			if _, err := db.InsertOne(coll, d.NewDocumentOf(m)); err != nil {
				log.Fatalf("%s reinsert %s: %v", coll, id, err)
			}
			n++
		}
		fmt.Printf("%s: renamed proprietary->owner on %d doc(s)\n", coll, n)
	}
	fmt.Println("done.")
}
