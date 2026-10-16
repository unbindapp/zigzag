//go:build ignore

package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"time"

	atlasMigrate "ariga.io/atlas/sql/migrate"
	"ariga.io/atlas/sql/sqltool"
	"entgo.io/ent/dialect"
	"entgo.io/ent/dialect/sql/schema"
	embeddedpostgres "github.com/fergusstrange/embedded-postgres"
	"github.com/zigzagapp-org/zigzag-api/ent/migrate"
)

func regenerateChecksum(directory *sqltool.GooseDir) error {
	hashFile, err := directory.Checksum()
	if err != nil {
		return fmt.Errorf("failed to generate checksum: %w", err)
	}

	if err := atlasMigrate.WriteSumFile(directory, hashFile); err != nil {
		return fmt.Errorf("failed to write hash file: %w", err)
	}

	return nil
}

func main() {
	ctx := context.Background()

	_, thisFile, _, _ := runtime.Caller(0)

	migrationsDir := filepath.Join(thisFile, "../migrations")

	gooseDir, err := sqltool.NewGooseDir(migrationsDir)
	if err != nil {
		fmt.Printf("Error creating atlas directory: %v\n", err)
		os.Exit(1)
	}

	name := flag.String("name", "", "Name of the migration to create")
	checksum := flag.Bool("checksum", false, "Regenerate atlas checksum")
	databaseURL := flag.String("database-url", "", "Diff against this empty scratch database instead of an embedded postgres")
	flag.Parse()

	if err := os.MkdirAll(migrationsDir, 0755); err != nil {
		fmt.Printf("Error creating migrations directory: %v\n", err)
		os.Exit(1)
	}

	if *checksum {
		if err := regenerateChecksum(gooseDir); err != nil {
			fmt.Printf("Error regenerating checksum: %v\n", err)
			os.Exit(1)
		}

		fmt.Println("Checksum regenerated successfully")
		return
	}

	if *name == "" {
		fmt.Println("Error: migration name is required")
		flag.Usage()
		os.Exit(1)
	}

	startingFiles, err := gooseDir.Files()
	if err != nil {
		fmt.Printf("Error getting files: %v\n", err)
		return
	}

	opts := []schema.MigrateOption{
		schema.WithDir(gooseDir),
		schema.WithMigrationMode(schema.ModeReplay),
		schema.WithDialect(dialect.Postgres),
		schema.WithDropColumn(true),
		schema.WithDropIndex(true),
		schema.WithIndent("  "),
		schema.WithFormatter(sqltool.GooseFormatter),
	}

	dbConnString := *databaseURL
	if dbConnString == "" {
		stop, url, err := startEmbeddedPostgres()
		if err != nil {
			fmt.Printf("Error starting embedded postgres: %v\n", err)
			os.Exit(1)
		}
		defer stop()
		dbConnString = url
	}

	if err = migrate.NamedDiff(
		ctx,
		dbConnString,
		*name,
		opts...,
	); err != nil {
		fmt.Printf("Error creating migration: %v\n", err)
		return
	}

	afterFiles, err := gooseDir.Files()
	if err != nil {
		fmt.Printf("Error getting files: %v\n", err)
		return
	}

	diff := diffFiles(startingFiles, afterFiles)
	if len(diff) == 0 {
		fmt.Println("No new migration files created")
		return
	}

	fmt.Println("New migration files created:")
	for _, file := range diff {
		if strings.HasPrefix(file, "U") {
			fmt.Printf("- %s (reversal)\n", file)
			continue
		}
		fmt.Printf("- %s\n", file)
	}
}

func startEmbeddedPostgres() (func(), string, error) {
	port := 34529
	ep := embeddedpostgres.NewDatabase(
		embeddedpostgres.DefaultConfig().
			Version(embeddedpostgres.V16).
			Port(uint32(port)).
			StartTimeout(30 * time.Second).
			Logger(io.Discard),
	)
	if err := ep.Start(); err != nil {
		return nil, "", err
	}
	stop := func() {
		if err := ep.Stop(); err != nil {
			fmt.Printf("Error stopping embedded postgres: %v\n", err)
		}
	}
	return stop, fmt.Sprintf("postgres://postgres:postgres@localhost:%d/postgres?sslmode=disable", port), nil
}

func diffFiles(startingFiles, afterFiles []atlasMigrate.File) []string {
	beforeMap := make(map[string]atlasMigrate.File)
	for _, file := range startingFiles {
		beforeMap[file.Name()] = file
	}

	var diff []string
	for _, file := range afterFiles {
		if _, exists := beforeMap[file.Name()]; !exists {
			diff = append(diff, file.Name())
		}
	}
	sort.Strings(diff)
	return diff
}
