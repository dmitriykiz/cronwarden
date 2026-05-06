package api_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestJobCheckpoint_PutAndGet(t *testing.T) {
	db := tempDB(t)
	srv := httptest.NewServer(testRouter(db))
	defer srv.Close()

	body := `{"payload":"{\"offset\":42}"}`
	resp, err := http.NewRequest(http.MethodPut, srv.URL+"/api/jobs/import-job/checkpoint", bytes.NewBufferString(body))
	require.NoError(t, err)
	resp.Header.Set("Content-Type", "application/json")
	res, err := http.DefaultClient.Do(resp)
	require.NoError(t, err)
	assert.Equal(t, http.StatusNoContent, res.StatusCode)

	get, err := http.Get(srv.URL + "/api/jobs/import-job/checkpoint")
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, get.StatusCode)

	var out map[string]interface{}
	require.NoError(t, json.NewDecoder(get.Body).Decode(&out))
	assert.Equal(t, "import-job", out["job_name"])
	assert.NotEmpty(t, out["payload"])
}

func TestJobCheckpoint_GetNotFound(t *testing.T) {
	db := tempDB(t)
	srv := httptest.NewServer(testRouter(db))
	defer srv.Close()

	res, err := http.Get(srv.URL + "/api/jobs/ghost-job/checkpoint")
	require.NoError(t, err)
	assert.Equal(t, http.StatusNotFound, res.StatusCode)
}

func TestJobCheckpoint_Delete(t *testing.T) {
	db := tempDB(t)
	srv := httptest.NewServer(testRouter(db))
	defer srv.Close()

	body := `{"payload":"{\"step\":1}"}`
	req, _ := http.NewRequest(http.MethodPut, srv.URL+"/api/jobs/batch-job/checkpoint", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	_, err := http.DefaultClient.Do(req)
	require.NoError(t, err)

	del, _ := http.NewRequest(http.MethodDelete, srv.URL+"/api/jobs/batch-job/checkpoint", nil)
	res, err := http.DefaultClient.Do(del)
	require.NoError(t, err)
	assert.Equal(t, http.StatusNoContent, res.StatusCode)

	get, err := http.Get(srv.URL + "/api/jobs/batch-job/checkpoint")
	require.NoError(t, err)
	assert.Equal(t, http.StatusNotFound, get.StatusCode)
}

func TestJobCheckpoint_MethodNotAllowed(t *testing.T) {
	db := tempDB(t)
	srv := httptest.NewServer(testRouter(db))
	defer srv.Close()

	req, _ := http.NewRequest(http.MethodPost, srv.URL+"/api/jobs/any-job/checkpoint", nil)
	res, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	assert.Equal(t, http.StatusMethodNotAllowed, res.StatusCode)
}
