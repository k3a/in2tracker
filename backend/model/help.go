// Package model holds all models, mostly representing database structure.
// Please note that time.Time values are always stored in UTC in the database.
// That means `meddler:"localtime"` should be used for all time.Time struct fields.
package model
