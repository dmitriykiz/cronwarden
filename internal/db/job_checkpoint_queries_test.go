package db_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUpsertAndGetJobCheckpoint(t *testing.T) {
	db := tempDB(t)

	err := db.UpsertJobCheckpoint("backup-job", `{"last_file": "data.tar.gz", "offset": 1024}`)
	require.NoError(t, err)

	cp, err := db.GetJobCheckpoint("backup-job")
	require.NoError(t, err)
	assert.Equal(t, "backup-job", cp.JobName)
	assert.JSONEq(t, `{"last_file": "data.tar.gz", "offset": 1024}`, cp.Payload)
	assert.False(t, cp.UpdatedAt.IsZero())
}

func TestUpsertJobCheckpoint_Replaces(t *testing.T) {
	db := tempDB(t)

	require.NoError(t, db.UpsertJobCheckpoint("sync-job", `{"page": 1}`))
	require.NoError(t, db.UpsertJobCheckpoint("sync-job", `{"page": 5}`))

	cp, err := db.GetJobCheckpoint("sync-job")
	require.NoError(t, err)
	assert.JSONEq(t, `{"page": 5}`, cp.Payload)
}

func TestGetJobCheckpoint_NotFound(t *testing.T) {
	db := tempDB(t)

	_, err := db.GetJobCheckpoint("nonexistent-job")
	require.Error(t, err)
	assert.True(t, isNoRows(err))
}

func TestDeleteJobCheckpoint(t *testing.T) {
	db := tempDB(t)

	require.NoError(t, db.UpsertJobCheckpoint("cleanup-job", `{"done": true}`))

	err := db.DeleteJobCheckpoint("cleanup-job")
	require.NoError(t, err)

	_, err = db.GetJobCheckpoint("cleanup-job")
	require.Error(t, err)
	assert.True(t, isNoRows(err))
}
