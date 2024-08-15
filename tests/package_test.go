package tests

import (
	"database/sql"
	"fmt"
	"log"
	"net/url"
	"testing"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/pgx"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/stretchr/testify/suite"
)

type IntegrationTestSuite struct {
	suite.Suite
}

// docker run --rm -p 5432:5432 -e POSTGRES_PASSWORD=pgpassword -e POSTGRES_USER=pguser -e POSTGRES_DB=testdb postgres:16
func (s *IntegrationTestSuite) SetupSuite() {
	v := make(url.Values, 1)
	v.Set("sslmode", "disable")
	connString := url.URL{
		Scheme:   "postgres",
		User:     url.UserPassword("pguser", "pgpassword"),
		Host:     fmt.Sprintf("%s:%d", "localhost", 5432),
		Path:     "testdb",
		RawQuery: v.Encode(),
	}
	db, err := sql.Open("pgx", connString.String())
	if err != nil {
		log.Fatal(err)
	}

	err = db.Ping()
	if err != nil {
		log.Fatalf("db.Ping: %v", err)
	}

	driver, err := pgx.WithInstance(db, &pgx.Config{
		MigrationsTable: "migrations",
	})
	if err != nil {
		log.Fatalf("pgx.WithInstance: %v", err)
	}

	migrator, err := migrate.NewWithDatabaseInstance("file://../migrations", "pgx", driver)
	if err != nil {
		log.Fatalf("migrate.NewWithDatabaseInstance: %v", err)
	}

	err = migrator.Up()
	if err != nil {
		log.Fatalf("migrator.Up: %v", err)
	}
}

func TestIntegrationTestSuite(t *testing.T) {
	suite.Run(t, new(IntegrationTestSuite))
}
