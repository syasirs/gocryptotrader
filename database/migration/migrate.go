package migrations

import (
	"bytes"
	"database/sql"
	"errors"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
)

// LoadMigrations will load all migrations in the ./database/migration/migrations folder
func (m *Migrator) LoadMigrations() error {
	migration, err := filepath.Glob("./database/migration/migrations/*.sql")

	if err != nil {
		return errors.New("failed to load migrations")
	}

	sort.Strings(migration)

	for x := range migration {
		err = m.loadMigration(migration[x])
		if err != nil {
			return err
		}
	}

	return nil
}

func (m *Migrator) loadMigration(migration string) error {
	file, err := os.Open(migration)
	if err != nil {
		m.Log.Println(err)
	}
	fileData := strings.Split(file.Name(), "/")

	fileSeq := strings.Split(fileData[3], "_")
	seq, _ := strconv.Atoi(fileSeq[0])

	b, err := ioutil.ReadAll(file)

	up := bytes.Split(b, []byte("-- up"))
	down := strings.Split(string(up[1]), "-- down")

	temp := Migration{
		Sequence: seq,
		Name:     fileData[3],
		UpSQL:    down[0],
		DownSQL:  down[1],
	}

	m.Migrations = append(m.Migrations, temp)

	return nil
}

// RunMigration attempts to run current migrations against a database
func (m *Migrator) RunMigration() (err error) {
	err = m.checkVersionTableExists()
	if err != nil {
		return
	}

	v, _ := m.getCurrentVersion()
	m.Log.Printf("Current database version: %v\n", v)

	latestSeq := m.Migrations[len(m.Migrations)-1].Sequence

	if v == latestSeq {
		m.Log.Println("no migrations to be run")
		return
	}

	tx, err := m.Conn.SQL.Begin()
	if err != nil {
		return
	}

	for y := v; y < len(m.Migrations); y++ {
		err = m.txBegin(tx, m.checkConvert(m.Migrations[y].UpSQL))
		if err != nil {
			return tx.Rollback()
		}

		_, err = tx.Exec("update version set version=$1", m.Migrations[y].Sequence)
		if err != nil {
			return tx.Rollback()
		}
	}

	err = tx.Commit()
	if err != nil {
		return tx.Rollback()
	}

	m.Log.Println("Migration completed")
	return nil
}

func (m *Migrator) txBegin(tx *sql.Tx, input string) error {
	_, err := tx.Exec(input)
	if err != nil {
		return tx.Rollback()
	}

	return nil
}

func (m *Migrator) getCurrentVersion() (v int, err error) {
	err = m.Conn.SQL.QueryRow("select version from version").Scan(&v)
	return
}

func (m *Migrator) checkVersionTableExists() error {
	query := `
		CREATE TABLE IF NOT EXISTS version(
		    version int not null
		);

	INSERT INTO version SELECT 0 WHERE 0=(SELECT COUNT(*) from version);
`

	_, err := m.Conn.SQL.Exec(m.checkConvert(query))
	if err != nil {
		return err
	}

	return nil
}

func (m *Migrator) checkConvert(input string) string {

	if m.Conn.Config.Driver != "sqlite" {
		return input
	}

	// Common PSQL -> SQLITE conversion
	// TODO: Find a better way to handle this list

	r := strings.NewReplacer(
		"bigserial", "integer",
		"int", "integer",
		"now()", "CURRENT_TIMESTAMP")

	return r.Replace(input)
}
