package sqlcsqlite

import (
	_ "embed"
)

// This gets the schema so it can be used in other submodules
//
//go:embed schema.sql
var Ddl string
