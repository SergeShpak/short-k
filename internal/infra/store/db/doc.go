/*
Package db handles access to the SQL data store.
*/
package db

//go:generate go run go.uber.org/mock/mockgen -source=db.go -destination=internal/mocks/repo_mock.gen.go -package=mocks
