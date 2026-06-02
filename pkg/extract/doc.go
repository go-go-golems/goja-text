// Package extract provides deterministic helpers for locating structured-data
// candidates inside larger text documents.
//
// Extraction is intentionally separate from parsing and sanitizing. Extractors
// return source spans and wrapper metadata; callers can then validate, sanitize,
// or parse candidate payloads explicitly.
package extract
