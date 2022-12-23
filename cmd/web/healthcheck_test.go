package main

import (
	"fmt"
	"net/http"
	"testing"

	"trolly.hunterwilkins.dev/internal/assert"
)

func TestHealthCheck(t *testing.T) {
	app := newTestApplication(t)

	ts := newTestServer(t, app.routes())
	defer ts.Close()

	code, _, body := ts.get(t, "/api/healthcheck")

	assert.Equal(t, code, http.StatusOK)
	assert.Equal(t, body, fmt.Sprintf(`{"environment":%q,"status":%q,"version":%q}`, "testing", "available", version))
}
