package langfuse_test

import (
	"context"
	"io"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	langfuse "github.com/xops-infra/GoLangfuse"
	"github.com/xops-infra/GoLangfuse/config"
	"github.com/xops-infra/GoLangfuse/mock"
	"github.com/xops-infra/GoLangfuse/types"
)

func Test_AddEvent_ShouldCallClientToSendEvents(t *testing.T) {
	os.Setenv("LANGFUSE_URL", "http://localhost:3000")
	os.Setenv("LANGFUSE_PUBLIC_KEY", "LangfusePublicKey")
	os.Setenv("LANGFUSE_SECRET_KEY", "LangfuseSecretKey")
	cfg, err := config.LoadLangfuseConfig()
	require.NoError(t, err, "Failed to load configuration")

	httpClient := &http.Client{}
	eventID := uuid.MustParse("f8359e80-1ecd-471b-bf2a-49d2009a9179")
	mockTransport := mock.AddMockTransport(t, httpClient)

	resp := &http.Response{Body: io.NopCloser(strings.NewReader("{}"))}
	mockTransport.ExpectWith("POST", "http://localhost:3000/api/public/ingestion").Return(resp, nil)
	subject, err := langfuse.NewWithClient(cfg, httpClient)

	subject.AddEvent(context.TODO(), &types.TraceEvent{ID: &eventID, Name: "LLM"})

	assert.Eventually(t, func() bool {
		return mockTransport.AllExpectationMet()
	}, time.Second*10, time.Millisecond*100)

	body, err := io.ReadAll(mockTransport.RecordedRequests()[0].Body)
	require.NoError(t, err)

	assert.Contains(t, string(body), `"id":"f8359e80-1ecd-471b-bf2a-49d2009a9179"`)
	assert.Contains(t, string(body), `"type":"trace-create"`)
	assert.Contains(t, string(body), `"name":"LLM"`)
	assert.Contains(t, string(body), `"public":false`)
}
