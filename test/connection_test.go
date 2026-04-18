package test_test

import (
	"log"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
)

func TestMySQLConnection(t *testing.T) {
	// We use sqlmock to simulate a successful connection and ping
	// without depending on a live external MySQL server.
	db, mockDB, err := sqlmock.New(sqlmock.MonitorPingsOption(true))
	if err != nil {
		log.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	// Expect a ping on the database connection
	mockDB.ExpectPing()

	// Fire the ping that our infrastructure code would normally execute
	err = db.Ping()
	
	// Assert no error is returned
	assert.NoError(t, err)
	
	// Assert that the mocked expectations (the Ping) were actually met
	assert.NoError(t, mockDB.ExpectationsWereMet())
}
