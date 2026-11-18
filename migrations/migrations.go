// Package migrations osadza migracje goose w binarce serwera (ADR 0008).
//
// Pliki NNNNN_opis.sql to kontrakt schematu bazy. Zmiana tylko przez PR z review (docs/contracts/).
package migrations

import "embed"

// FS zawiera pliki migracji *.sql.
//
//go:embed *.sql
var FS embed.FS
