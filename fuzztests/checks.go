package main

import (
	"database/sql"
	"fmt"
	"testing"
)

func CheckWKTParse(t *testing.T, db *sql.DB, candidates []string) {
	for i, candidate := range candidates {
		t.Run(fmt.Sprintf("CheckWKTParse_%d", i), func(t *testing.T) {
			_ = candidate
		})
	}
}
