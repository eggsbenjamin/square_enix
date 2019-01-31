// +build integration

package repository_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/eggsbenjamin/square_enix/internal/app/db"
	"github.com/eggsbenjamin/square_enix/internal/app/repository"
	"github.com/eggsbenjamin/square_enix/pkg/env"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/require"
)

func TestElementRepository(t *testing.T) {
	dsn := fmt.Sprintf(
		"%s@tcp(%s:3306)/%s?parseTime=true",
		env.MustGetEnv("MYSQL_USER"),
		env.MustGetEnv("MYSQL_HOST"),
		env.MustGetEnv("MYSQL_DB"),
	)
	conn, err := sqlx.Connect("mysql", dsn)
	require.NoError(t, err)

	db := db.NewQuerier(conn)

	t.Run("GetElementsCreatedBefore", func(t *testing.T) {
		defer func() {
			_, err := conn.Exec("DELETE FROM ProcessElement")
			_, err = conn.Exec("DELETE FROM Process")
			_, err = conn.Exec("DELETE FROM Element")
			if err != nil {
				t.Logf("error resetting Process table: %q\n", err)
			}
		}()

		_, err := conn.Exec("DELETE FROM ProcessElement")
		require.NoError(t, err)

		_, err = conn.Exec("DELETE FROM Process")
		require.NoError(t, err)

		_, err = conn.Exec("DELETE FROM Element")
		require.NoError(t, err)

		_, err = conn.Exec("INSERT INTO Element (id, data ) VALUES (1, 'test')")
		require.NoError(t, err)

		_, err = conn.Exec("INSERT INTO Element (id, data, created_at) VALUES (2, 'test', NOW() + INTERVAL 1 DAY)")
		require.NoError(t, err)

		repo := repository.NewElementRepository(db)
		require.NoError(t, err)

		elements, err := repo.GetElementsCreatedBefore(time.Now())
		require.NoError(t, err)
		require.Equal(t, 1, len(elements))
		require.Equal(t, 1, elements[0].ID)
	})
}
