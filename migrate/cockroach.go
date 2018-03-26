package migrate

import (
	"database/sql"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"path/filepath"

	_ "github.com/lib/pq"

	"github.com/mattes/migrate"
	"github.com/mattes/migrate/database/cockroachdb"
	"strconv"
	"strings"
)

type Cockroach struct {
	db *sql.DB

	MigrationsPath string
}

type CockroachOptions struct {
	*ConnectionOptions

	MigrationTable string
	Secure         bool
	SSL            *CockroachSSL
}

type CockroachSSL struct {
	CertPath, KeyPath, Mode, RootCert string
}

const (
	DefaultHost           = "localhost"
	DefaultUser           = "cockroach"
	DefaultPort           = "26257"
	DefaultDatabase       = "service-db"
	DefaultMigrationTable = "schema_migrations"
)

// Connection string concatenates the CockroachOptions down to a string,
// applying defaults to options that were not set ready to be used in a
// connection to the database
func (co *CockroachOptions) ConnectionString() (string, error) {
	var conn strings.Builder

	conn.WriteString("cockroach://")

	// Prevent panics and just return the exact default upon nil options
	if co.ConnectionOptions == nil {
		return fmt.Sprintf(
			"cockroach://%s@%s:%s/%s?sslmode=disable&x-migrations-table=%s",
			DefaultUser, DefaultHost, DefaultPort, DefaultDatabase, DefaultMigrationTable,
		), nil
	}

	// User
	if len(co.User) != 0 {
		conn.WriteString(co.User)
	} else {
		conn.WriteString(DefaultUser)
	}

	conn.WriteString("@")

	// Host
	if len(co.Host) != 0 {
		conn.WriteString(co.Host)
	} else {
		conn.WriteString(DefaultHost)
	}

	conn.WriteString(":")

	// Port
	if co.Port != 0 {
		conn.WriteString(strconv.Itoa(co.Port))
	} else {
		conn.WriteString(DefaultPort)
	}

	conn.WriteString("/")

	if len(co.Database) != 0 {
		conn.WriteString(co.Database)
	} else {
		conn.WriteString(DefaultDatabase)
	}

	// Set connection security
	if co.Secure {
		conn.WriteString(fmt.Sprintf("?sslcert=%s", co.SSL.CertPath))
		conn.WriteString(fmt.Sprintf("?sslkey=%s", co.SSL.KeyPath))
		conn.WriteString(fmt.Sprintf("?sslmode=%s", co.SSL.Mode))
		conn.WriteString(fmt.Sprintf("?sslrootcert=%s", co.SSL.RootCert))
	} else {
		conn.WriteString("?sslmode=disable")
	}

	if len(co.MigrationTable) != 0 {
		conn.WriteString(fmt.Sprintf("&x-migrations-table=%s",
			co.MigrationTable))
	} else {
		conn.WriteString(fmt.Sprintf("&x-migrations-table=%s"))
	}

	return conn.String(), nil
}

func NewCockroach(opts *CockroachOptions) (*Cockroach, error) {
	conn, err := opts.ConnectionString()
	if err != nil {
		return nil, err
	}

	db, err := sql.Open("postgres", conn)
	if err != nil {
		return nil, err
	}

	return &Cockroach{db: db}, nil
}

func (c *Cockroach) Migrate() error {
	driver, err := cockroachdb.WithInstance(c.db, &cockroachdb.Config{})
	if err != nil {
		log.Fatalf("could not get migrations driver: %s", err)
	}

	if !c.migrationsInPath() {
		return errors.New("migrate: no migration files exist in given path")
	}

	m, err := migrate.NewWithDatabaseInstance(
		fmt.Sprintf("file://%s", c.MigrationsPath),
		"sql",
		driver,
	)

	if err != nil {
		return fmt.Errorf("migrate: could not initialise migrations: %s", err)
	}

	return m.Up()
}

// migrationsInPath will assert whether the configured migrations path has
// migrations ready to be ran.
func (c *Cockroach) migrationsInPath() bool {
	// Ensure we have a migration path
	if len(c.MigrationsPath) == 0 {
		return false
	}

	// Pull all files up from the migrations path
	files, err := ioutil.ReadDir(c.MigrationsPath)
	if err != nil {
		return false
	}

	// If one sql file exists we can consider this a valid migration directory.
	for _, file := range files {
		if file.Mode().IsRegular() &&
			filepath.Ext(file.Name()) == ".sql" {
			return true
		}
	}

	// If no SQL files were found whilst traversing a path - then none exists!
	return false
}