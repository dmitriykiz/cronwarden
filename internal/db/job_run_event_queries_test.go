package db_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func seedEventRun(t *testing.T, db *DB) int64 {
	t.Helper()
	id, err := db.InsertJobRun("event-job", true, 0, time.Now())
	require.NoError(t, err)
	return id
}

func TestAddAndListJobRunEvents(t *testing.T) {
	db := tempDB(t)
	runID := seedEventRun(t, db)

	err := db.AddJobRunEvent(runID, "started", "job began execution")
	require.NoError(t, err)

	err = db.AddJobRunEvent(runID, "retry", "attempt 2")
	require.NoError(t, err)

	events, err := db.ListJobRunEvents(runID)
	require.NoError(t, err)
	assert.Len(t, events, 2)
	assert.Equal(t, "started", events[0].EventType)
	assert.Equal(t, "job began execution", events[0].Message)
	assert.Equal(t, "retry", events[1].EventType)
}

func TestListJobRunEvents_Empty(t *testing.T) {
	db := tempDB(t)
	runID := seedEventRun(t, db)

	events, err := db.ListJobRunEvents(runID)
	require.NoError(t, err)
	assert.Empty(t, events)
}

func TestDeleteJobRunEvent_Exists(t *testing.T) {
	db := tempDB(t)
	runID := seedEventRun(t, db)

	err := db.AddJobRunEvent(runID, "warning", "disk usage high")
	require.NoError(t, err)

	events, err := db.ListJobRunEvents(runID)
	require.NoError(t, err)
	require.Len(t, events, 1)

	err = db.DeleteJobRunEvent(events[0].ID)
	require.NoError(t, err)

	events, err = db.ListJobRunEvents(runID)
	require.NoError(t, err)
	assert.Empty(t, events)
}

func TestDeleteJobRunEvent_NotExists(t *testing.T) {
	db := tempDB(t)
	err := db.DeleteJobRunEvent(99999)
	assert.NoError(t, err)
}

func TestDeleteAllJobRunEvents(t *testing.T) {
	db := tempDB(t)
	runID := seedEventRun(t, db)

	_ = db.AddJobRunEvent(runID, "started", "")
	_ = db.AddJobRunEvent(runID, "finished", "")

	err := db.DeleteAllJobRunEvents(runID)
	require.NoError(t, err)

	events, err := db.ListJobRunEvents(runID)
	require.NoError(t, err)
	assert.Empty(t, events)
}
