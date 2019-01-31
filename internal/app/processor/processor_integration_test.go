// +build integration

package processor_test

import (
	"fmt"
	"testing"

	"github.com/eggsbenjamin/square_enix/internal/app/db"
	"github.com/eggsbenjamin/square_enix/internal/app/models"
	"github.com/eggsbenjamin/square_enix/internal/app/processor"
	"github.com/eggsbenjamin/square_enix/internal/app/repository"
	"github.com/eggsbenjamin/square_enix/pkg/env"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/require"
)

func TestProcessor(t *testing.T) {
	dsn := fmt.Sprintf(
		"%s@tcp(%s:3306)/%s?parseTime=true",
		env.MustGetEnv("MYSQL_USER"),
		env.MustGetEnv("MYSQL_HOST"),
		env.MustGetEnv("MYSQL_DB"),
	)
	conn, err := sqlx.Connect("mysql", dsn)
	require.NoError(t, err)

	db := db.NewDB(conn)

	t.Run("ProcessBatch", func(t *testing.T) {
		t.Run("No Running Process", func(t *testing.T) {
			defer func() {
				if err := ResetDB(conn); err != nil {
					t.Logf("error resetting Process table: %q\n", err)
				}
			}()

			require.NoError(t, ResetDB(conn))

			_, err = conn.Exec("INSERT INTO Process (status) VALUES ('COMPLETE')")
			require.NoError(t, err)

			proc := processor.NewProcessor(
				db,
				repository.NewProcessRepositoryFactory(),
				repository.NewElementRepositoryFactory(),
			)

			require.Equal(t, processor.ErrNoRunningProcessExists, proc.ProcessBatch(1))
		})

		t.Run("Elements to Process", func(t *testing.T) {
			defer func() {
				if err := ResetDB(conn); err != nil {
					t.Logf("error resetting Process table: %q\n", err)
				}
			}()

			require.NoError(t, ResetDB(conn))

			_, err = conn.Exec("INSERT INTO Element (id, data) VALUES (1, 'test')")
			require.NoError(t, err)

			_, err = conn.Exec("INSERT INTO Element (id, data) VALUES (2, 'test')")
			require.NoError(t, err)

			_, err = conn.Exec("INSERT INTO Process (id, status, created_at) VALUES (1, 'RUNNING', NOW() + INTERVAL 1 DAY)")
			require.NoError(t, err)

			proc := processor.NewProcessor(
				db,
				repository.NewProcessRepositoryFactory(),
				repository.NewElementRepositoryFactory(),
			)

			require.NoError(t, proc.ProcessBatch(2))

			processedElements, err := repository.NewElementRepositoryFactory().CreateElementRepository(db).GetElementsByProcessID(1)
			require.NoError(t, err)
			require.Equal(t, 2, len(processedElements))

			processes, err := repository.NewProcessRepositoryFactory().CreateProcessRepository(db).GetByStatus(models.PROCESS_STATUS_RUNNING)
			require.NoError(t, err)
			require.Equal(t, 1, len(processes))
			require.Equal(t, 1, processes[0].ID)
		})

		t.Run("No Elements to Process", func(t *testing.T) {
			defer func() {
				if err := ResetDB(conn); err != nil {
					t.Logf("error resetting Process table: %q\n", err)
				}
			}()

			require.NoError(t, ResetDB(conn))

			_, err = conn.Exec("INSERT INTO Element (id, data) VALUES (1, 'test')")
			require.NoError(t, err)

			_, err = conn.Exec("INSERT INTO Element (id, data) VALUES (2, 'test')")
			require.NoError(t, err)

			_, err = conn.Exec("INSERT INTO Process (id, status, created_at) VALUES (1, 'RUNNING', NOW() + INTERVAL 1 DAY)")
			require.NoError(t, err)

			_, err = conn.Exec("INSERT INTO ProcessElement (process_id, element_id) VALUES (1, 1)")
			require.NoError(t, err)

			_, err = conn.Exec("INSERT INTO ProcessElement (process_id, element_id) VALUES (1, 2)")
			require.NoError(t, err)

			proc := processor.NewProcessor(
				db,
				repository.NewProcessRepositoryFactory(),
				repository.NewElementRepositoryFactory(),
			)

			require.NoError(t, proc.ProcessBatch(2))

			processes, err := repository.NewProcessRepositoryFactory().CreateProcessRepository(db).GetByStatus(models.PROCESS_STATUS_COMPLETE)
			require.NoError(t, err)
			require.Equal(t, 1, len(processes))
			require.Equal(t, 1, processes[0].ID)
		})
	})
}

func ResetDB(conn *sqlx.DB) error {
	if _, err := conn.Exec("DELETE FROM ProcessElement"); err != nil {
		return err
	}

	if _, err := conn.Exec("DELETE FROM Process"); err != nil {
		return err
	}

	if _, err := conn.Exec("DELETE FROM Element"); err != nil {
		return err
	}

	return nil
}
