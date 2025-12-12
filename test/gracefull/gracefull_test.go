//go:build integration
// +build integration

package gracefull_test

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"github.com/docker/docker/api/types/build"
	"github.com/docker/docker/client"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/network"
	"github.com/testcontainers/testcontainers-go/wait"
)

const (
	dbName     = "appdb"
	dbUser     = "app"
	dbPassword = "secret"
	serverPort = "8080/tcp"
)

type stack struct {
	network  *testcontainers.DockerNetwork
	postgres testcontainers.Container
	server   testcontainers.Container
	agent    testcontainers.Container
	hostDSN  string
}

func ensureDockerAvailable(t *testing.T) {
	t.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		t.Skipf("docker not available: %v", err)
	}
	defer cli.Close()

	if _, err := cli.Ping(ctx); err != nil {
		t.Skipf("docker daemon unreachable: %v", err)
	}
}

func TestGracefullDeliveryAndPersistence(t *testing.T) {
	ensureDockerAvailable(t)
	// signals := []string{"SIGINT", "SIGTERM", "SIGQUIT"}
	signals := []string{"SIGINT"}
	for _, sig := range signals {
		t.Run(sig, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 6*time.Minute)
			defer cancel()

			st := startStack(ctx, t)
			t.Cleanup(func() {
				terminate(ctx, t, st.agent)
				terminate(ctx, t, st.server)
				terminate(ctx, t, st.postgres)
				if st.network != nil {
					_ = st.network.Remove(ctx)
				}
			})

			waitForAgentDelivery(ctx, t, st.hostDSN)

			sendSignal(ctx, t, st.agent, sig)
			waitForExit(ctx, t, st.agent)

			beforeServerStop := snapshotMetrics(ctx, t, st.hostDSN)
			require.Greater(t, beforeServerStop.pollCount, int64(0), "poll counter should be persisted before stopping server")
			require.Greater(t, beforeServerStop.gaugeRows, 0, "gauges should be persisted before stopping server")

			sendSignal(ctx, t, st.server, sig)
			waitForExit(ctx, t, st.server)

			afterServerStop := snapshotMetrics(ctx, t, st.hostDSN)
			require.Equal(t, beforeServerStop.pollCount, afterServerStop.pollCount, "counter value should survive server shutdown")
			require.Equal(t, beforeServerStop.gaugeRows, afterServerStop.gaugeRows, "gauge count should survive server shutdown")
		})
	}
}

func startStack(ctx context.Context, t *testing.T) stack {
	t.Helper()

	nw, err := network.New(ctx)
	require.NoError(t, err)

	pg := startPostgres(ctx, t, nw)
	hostDSN := postgresDSN(ctx, t, pg)

	srv := startServer(ctx, t, nw)
	agent := startAgent(ctx, t, nw)

	return stack{
		network:  nw,
		postgres: pg,
		server:   srv,
		agent:    agent,
		hostDSN:  hostDSN,
	}
}

func startPostgres(ctx context.Context, t *testing.T, nw *testcontainers.DockerNetwork) testcontainers.Container {
	t.Helper()

	req := testcontainers.ContainerRequest{
		Image:        "postgres:16-alpine",
		Env:          map[string]string{"POSTGRES_DB": dbName, "POSTGRES_USER": dbUser, "POSTGRES_PASSWORD": dbPassword},
		ExposedPorts: []string{"5432/tcp"},
		Networks:     []string{nw.Name},
		NetworkAliases: map[string][]string{
			nw.Name: {"pg"},
		},
		WaitingFor: wait.ForListeningPort("5432/tcp").WithStartupTimeout(90 * time.Second),
	}

	pg, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{ContainerRequest: req, Started: true})
	require.NoError(t, err)
	return pg
}

func startServer(ctx context.Context, t *testing.T, nw *testcontainers.DockerNetwork) testcontainers.Container {
	t.Helper()

	req := testcontainers.ContainerRequest{
		FromDockerfile: serverDockerfile(t),
		ExposedPorts:   []string{serverPort},
		Env: map[string]string{
			"ADDRESS":        "0.0.0.0:8080",
			"DATABASE_DSN":   fmt.Sprintf("postgres://%s:%s@pg:5432/%s?sslmode=disable", dbUser, dbPassword, dbName),
			"STORE_INTERVAL": "0",
			"RESTORE":        "false",
		},
		Networks: []string{nw.Name},
		NetworkAliases: map[string][]string{
			nw.Name: {"server"},
		},
		WaitingFor: wait.ForHTTP("/").WithPort(serverPort).WithStartupTimeout(2 * time.Minute),
	}

	srv, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{ContainerRequest: req, Started: true})
	require.NoError(t, err)
	return srv
}

func startAgent(ctx context.Context, t *testing.T, nw *testcontainers.DockerNetwork) testcontainers.Container {
	t.Helper()

	req := testcontainers.ContainerRequest{
		FromDockerfile: agentDockerfile(t),
		Env: map[string]string{
			"ADDRESS":         "server:8080",
			"POLL_INTERVAL":   "1",
			"REPORT_INTERVAL": "1",
			"RATE_LIMIT":      "4",
		},
		Networks: []string{nw.Name},
		NetworkAliases: map[string][]string{
			nw.Name: {"agent"},
		},
	}

	agt, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{ContainerRequest: req, Started: true})
	require.NoError(t, err)
	return agt
}

func serverDockerfile(t *testing.T) testcontainers.FromDockerfile {
	t.Helper()
	return targetedDockerfile(t, "server")
}

func agentDockerfile(t *testing.T) testcontainers.FromDockerfile {
	t.Helper()
	return targetedDockerfile(t, "agent")
}

func targetedDockerfile(t *testing.T, target string) testcontainers.FromDockerfile {
	t.Helper()

	return testcontainers.FromDockerfile{
		Context:    repoRoot(t),
		Dockerfile: filepath.Join("test", "gracefull", "Dockerfile"),
		Repo:       fmt.Sprintf("gracefull-%s", target),
		Tag:        "latest",
		KeepImage:  true,
		BuildOptionsModifier: func(opts *build.ImageBuildOptions) {
			opts.Target = target
		},
	}
}

func repoRoot(t *testing.T) string {
	t.Helper()
	root, err := findModuleRoot()
	require.NoError(t, err)
	return root
}

func findModuleRoot() (string, error) {
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		return "", fmt.Errorf("unable to determine caller")
	}

	dir := filepath.Dir(filename)
	for {
		candidate := filepath.Join(dir, "go.mod")
		if _, err := os.Stat(candidate); err == nil {
			return dir, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", fmt.Errorf("go.mod not found from %s", filename)
		}
		dir = parent
	}
}

func waitForAgentDelivery(ctx context.Context, t *testing.T, dsn string) {
	t.Helper()

	require.Eventually(t, func() bool {
		snap := snapshotMetrics(ctx, t, dsn)
		return snap.pollCount > 0 && snap.gaugeRows > 0
	}, 90*time.Second, 1*time.Second)
}

type metricsSnapshot struct {
	pollCount int64
	gaugeRows int
}

func snapshotMetrics(ctx context.Context, t *testing.T, dsn string) metricsSnapshot {
	t.Helper()

	db, err := sql.Open("pgx", dsn)
	require.NoError(t, err)
	defer db.Close()

	var pollCount sql.NullInt64
	err = db.QueryRowContext(ctx, `SELECT value FROM counters WHERE id='PollCount'`).Scan(&pollCount)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		require.NoError(t, err)
	}
	var gaugeRows int
	require.NoError(t, db.QueryRowContext(ctx, `SELECT COUNT(*) FROM gauges`).Scan(&gaugeRows))

	return metricsSnapshot{pollCount: pollCount.Int64, gaugeRows: gaugeRows}
}

func sendSignal(ctx context.Context, t *testing.T, ctr testcontainers.Container, sig string) {
	t.Helper()

	_, _, err := ctr.Exec(ctx, []string{"kill", "-s", sig, "1"})
	require.NoError(t, err)
}

func waitForExit(ctx context.Context, t *testing.T, ctr testcontainers.Container) {
	t.Helper()

	require.Eventually(t, func() bool {
		state, err := ctr.State(ctx)
		require.NoError(t, err)
		return !state.Running
	}, 30*time.Second, 500*time.Millisecond)
}

func terminate(ctx context.Context, t *testing.T, ctr testcontainers.Container) {
	t.Helper()
	if ctr == nil {
		return
	}
	_ = ctr.Terminate(ctx)
}

func postgresDSN(ctx context.Context, t *testing.T, pg testcontainers.Container) string {
	t.Helper()

	host, err := pg.Host(ctx)
	require.NoError(t, err)
	port, err := pg.MappedPort(ctx, "5432/tcp")
	require.NoError(t, err)

	return fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable", dbUser, dbPassword, host, port.Port(), dbName)
}
