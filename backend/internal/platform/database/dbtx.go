package database

import "alongtu/backend/internal/platform/database/dbgen"

// DBTX is the common sqlc executor contract implemented by pgx pools and
// transactions. Business repositories depend on generated queries rather
// than embedding SQL strings.
type DBTX = dbgen.DBTX
