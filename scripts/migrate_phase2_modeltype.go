//go:build ignore

// One-off store migration for the Device Class → Model Type rename.
// Run once per CloverDB store:  go run scripts/migrate_phase2_modeltype.go -s <store-dir>
//
// Backs the store up first (<dir>.bak-<ts>), then:
//   1. copies the `devclasses` collection into `modeltypes` (preserving `_id`,
//      models reference it by id) and drops `devclasses`;
//   2. renames the `class_id` field → `model_type_id` on every `models` doc.
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
		log.Fatal("usage: go run scripts/migrate_phase2_modeltype.go -s <store-dir>")
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

	// 1. devclasses -> modeltypes (the doc's name lives under `name`, so only the
	//    collection name changes; preserve _id since models reference it).
	if has, _ := db.HasCollection("devclasses"); has {
		if has2, _ := db.HasCollection("modeltypes"); !has2 {
			if err := db.CreateCollection("modeltypes"); err != nil {
				log.Fatalf("create modeltypes: %v", err)
			}
		}
		docs, err := db.FindAll(q.NewQuery("devclasses"))
		if err != nil {
			log.Fatalf("read devclasses: %v", err)
		}
		for _, doc := range docs {
			if _, err := db.InsertOne("modeltypes", d.NewDocumentOf(doc.AsMap())); err != nil {
				log.Fatalf("insert modeltype: %v", err)
			}
		}
		if err := db.DropCollection("devclasses"); err != nil {
			log.Fatalf("drop devclasses: %v", err)
		}
		fmt.Printf("migrated %d model-type(s); dropped devclasses\n", len(docs))
	} else {
		fmt.Println("no devclasses collection (already migrated?)")
	}

	// 2. models: rename class_id -> model_type_id.
	if has, _ := db.HasCollection("models"); has {
		docs, _ := db.FindAll(q.NewQuery("models"))
		n := 0
		for _, doc := range docs {
			m := doc.AsMap()
			v, ok := m["class_id"]
			if !ok {
				continue
			}
			m["model_type_id"] = v
			delete(m, "class_id")
			id, _ := m["_id"].(string)
			if err := db.Delete(q.NewQuery("models").Where(q.Field("_id").Eq(id))); err != nil {
				log.Fatalf("models delete %s: %v", id, err)
			}
			if _, err := db.InsertOne("models", d.NewDocumentOf(m)); err != nil {
				log.Fatalf("models reinsert %s: %v", id, err)
			}
			n++
		}
		fmt.Printf("models: renamed class_id->model_type_id on %d doc(s)\n", n)
	}
	fmt.Println("done.")
}
