package integration_test

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/joho/godotenv"
	"github.com/kelseyhightower/envconfig"
	"github.com/stretchr/testify/suite"

	langfuse "github.com/bdpiprava/GoLangfuse"
	"github.com/bdpiprava/GoLangfuse/config"
	"github.com/bdpiprava/GoLangfuse/types"
	"github.com/bdpiprava/easy-http/pkg/httpx"
)

// LangfuseIntegrationTestSuite is a test suite for Langfuse integration tests.
type LangfuseIntegrationTestSuite struct {
	suite.Suite
	subject langfuse.Langfuse
	cfg     *config.Langfuse
}

func TestIntegrationTestSuite(t *testing.T) {
	suite.Run(t, new(LangfuseIntegrationTestSuite))
}

func (s *LangfuseIntegrationTestSuite) SetupSuite() {
	err := godotenv.Load(".integration.env")
	s.Require().NoError(err, "Failed to load .integration.env file")

	s.cfg, err = loadEnv()
	s.Require().NoError(err, "Failed to load configuration from environment variables")

	s.subject = langfuse.New(s.cfg)
}

func (s *LangfuseIntegrationTestSuite) Test_SendTrace() {
	traceEvent := NewTestTraceEvent()
	expectedSession := BuildSession(traceEvent)

	s.subject.AddEvent(context.TODO(), traceEvent)

	s.Eventually(func() bool {
		tracePath := fmt.Sprintf("/api/public/traces/%s", traceEvent.ID.String())
		got, err := getEventType[types.TraceEvent](s.cfg, tracePath)
		if err != nil {
			s.T().Logf("Failed to get trace event: %v", err)
			return false
		}

		return s.Equal(traceEvent, got)
	}, 10*time.Second, 100*time.Millisecond, "Trace event should be sent successfully")

	s.Eventually(func() bool {
		sessionPath := fmt.Sprintf("/api/public/sessions/%s", traceEvent.SessionID)
		got, err := getEventType[types.Session](s.cfg, sessionPath)
		if err != nil {
			s.T().Logf("Failed to get session: %v", err)
			return false
		}

		return s.Equal(expectedSession, got,
			cmpopts.IgnoreFields(types.Trace{}, "Timestamp", "CreatedAt", "UpdatedAt"),
			cmpopts.IgnoreFields(types.Session{}, "CreatedAt"),
		)
	}, 10*time.Second, 100*time.Millisecond, "Trace event should be sent successfully")
}

func (s *LangfuseIntegrationTestSuite) Equal(expected, got any, opts ...cmp.Option) bool {
	less := func(a, b string) bool { return a < b }
	opts = append(opts, cmpopts.SortSlices(less))

	diff := cmp.Diff(expected, got, opts...)
	if diff != "" {
		s.T().Logf("Trace event mismatch (-want +got):\n%s", diff)
		return false
	}

	return true
}

func getEventType[T any](cfg *config.Langfuse, path string) (*T, error) {
	r, err := httpx.GET[string](
		httpx.WithPath(path),
		httpx.WithBaseURL(cfg.URL),
		httpx.WithHeader("Authorization", "Basic "+basicAuth(cfg.PublicKey, cfg.SecretKey)),
		httpx.WithHeader("Content-Type", "application/json"),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get event type: %w", err)
	}

	var t T
	err = json.Unmarshal(r.RawBody, &t)
	return &t, err
}

func basicAuth(username, password string) string {
	auth := username + ":" + password
	return base64.StdEncoding.EncodeToString([]byte(auth))
}

// loadEnv load config variables into config.Langfuse.
func loadEnv() (*config.Langfuse, error) {
	var cfg config.Langfuse
	err := envconfig.Process("", &cfg)
	if err != nil {
		return nil, err
	}
	return &cfg, err
}
