// +build integration

package repository_test

import (
	"fmt"
	"testing"

	"github.com/eggsbenjamin/square_enix/internal/app/db"
	"github.com/eggsbenjamin/square_enix/internal/app/models"
	"github.com/eggsbenjamin/square_enix/internal/app/repository"
	"github.com/eggsbenjamin/square_enix/pkg/env"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/require"
)

func TestProcessRepository(t *testing.T) {
	dsn := fmt.Sprintf(
		"%s@tcp(%s:3306)/%s?parseTime=true",
		env.MustGetEnv("MYSQL_USER"),
		env.MustGetEnv("MYSQL_HOST"),
		env.MustGetEnv("MYSQL_DB"),
	)
	conn, err := sqlx.Connect("mysql", dsn)
	require.NoError(t, err)

	db := db.NewQuerier(conn)

	t.Run("GetByStatus", func(t *testing.T) {
		defer func() {
			_, err := conn.Exec("DELETE FROM ProcessElement")
			_, err = conn.Exec("DELETE FROM Process")
			if err != nil {
				t.Logf("error resetting Process table: %q\n", err)
			}
		}()

		_, err := conn.Exec("DELETE FROM ProcessElement")
		require.NoError(t, err)

		_, err = conn.Exec("DELETE FROM Process")
		require.NoError(t, err)

		_, err = conn.Exec("INSERT INTO Process (status) VALUES ('COMPLETE')")
		require.NoError(t, err)

		_, err = conn.Exec("INSERT INTO Process (status) VALUES ('RUNNING')")
		require.NoError(t, err)

		repo := repository.NewProcessRepository(db)

		processes, err := repo.GetByStatus("RUNNING")
		require.NoError(t, err)
		require.Equal(t, 1, len(processes))
		require.Equal(t, "RUNNING", processes[0].Status)
	})

	t.Run("CreateNewProcess", func(t *testing.T) {
		defer func() {
			_, err := conn.Exec("DELETE FROM ProcessElement")
			_, err = conn.Exec("DELETE FROM Process")
			if err != nil {
				t.Logf("error resetting Process table: %q\n", err)
			}
		}()

		_, err := conn.Exec("DELETE FROM ProcessElement")
		require.NoError(t, err)

		_, err = conn.Exec("DELETE FROM Process")
		require.NoError(t, err)

		repo := repository.NewProcessRepository(db)

		process, err := repo.CreateNewProcess()
		require.NoError(t, err)

		require.NotZero(t, process.ID)
		require.NotZero(t, process.CreatedAt)
		require.Equal(t, models.PROCESS_STATUS_RUNNING, process.Status)
	})

	t.Run("UpdateProcess", func(t *testing.T) {
		defer func() {
			_, err := conn.Exec("DELETE FROM ProcessElement")
			_, err = conn.Exec("DELETE FROM Process")
			if err != nil {
				t.Logf("error resetting Process table: %q\n", err)
			}
		}()

		_, err := conn.Exec("DELETE FROM ProcessElement")
		require.NoError(t, err)

		_, err = conn.Exec("DELETE FROM Process")
		require.NoError(t, err)

		_, err = conn.Exec("INSERT INTO Process (status) VALUES ('RUNNING')")
		require.NoError(t, err)

		repo := repository.NewProcessRepository(db)

		existingProcesses, err := repo.GetByStatus(models.PROCESS_STATUS_RUNNING)
		require.NoError(t, err)
		require.Equal(t, 1, len(existingProcesses))

		existingProcess := existingProcesses[0]
		existingProcess.Status = models.PROCESS_STATUS_COMPLETE

		require.NoError(t, repo.UpdateProcess(existingProcess))

		updatedProcesses, err := repo.GetByStatus(models.PROCESS_STATUS_COMPLETE)
		require.NoError(t, err)
		updatedProcess := updatedProcesses[0]

		require.Equal(t, existingProcess, updatedProcess)
	})
}
