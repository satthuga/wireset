package dgraphsvc

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"os"
	"time"

	"github.com/dgraph-io/dgo/v2"
	"github.com/dgraph-io/dgo/v2/protos/api"
	"github.com/google/wire"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

var DefaultWireSet = wire.NewSet(
	NewDgraphSvc,
	ProvideConfig,
)

type Config struct {
	Address     string
	PoolSize    int
	DialTimeout time.Duration
	RetryCount  int
}

func ProvideConfig() (*Config, error) {
	address := os.Getenv("DGRAPH_ADDRESS")
	if address == "" {
		return nil, fmt.Errorf("DGRAPH_ADDRESS environment variable is not provided or is empty")
	}

	return &Config{
		Address:     address,
		PoolSize:    10,
		DialTimeout: 30 * time.Second,
		RetryCount:  3,
	}, nil
}

func createDialOptions(cfg *Config) ([]grpc.DialOption, error) {
	systemRoots, err := x509.SystemCertPool()
	if err != nil {
		return nil, fmt.Errorf("failed to get system certificate pool: %w", err)
	}

	cred := credentials.NewTLS(&tls.Config{
		RootCAs: systemRoots,
	})

	return []grpc.DialOption{
		grpc.WithAuthority(cfg.Address),
		grpc.WithTransportCredentials(cred),
		grpc.WithDefaultCallOptions(grpc.MaxCallRecvMsgSize(100 * 1024 * 1024)),
	}, nil
}

func dialDgraphWithRetry(ctx context.Context, cfg *Config, opts []grpc.DialOption, logger *zap.Logger) (*grpc.ClientConn, error) {
	var conn *grpc.ClientConn
	var err error

	for i := 0; i < cfg.RetryCount; i++ {
		dialCtx, cancel := context.WithTimeout(ctx, cfg.DialTimeout)
		conn, err = grpc.DialContext(dialCtx, cfg.Address, opts...)
		cancel()

		if err == nil {
			return conn, nil
		}

		logger.Warn("Failed to connect to Dgraph, retrying...", zap.Error(err), zap.Int("attempt", i+1))
		time.Sleep(time.Second * time.Duration(i+1))
	}

	return nil, fmt.Errorf("failed to connect to Dgraph after %d attempts: %w", cfg.RetryCount, err)
}

func NewDgraphSvc(cfg *Config, logger *zap.Logger) (*dgo.Dgraph, func(), error) {
	opts, err := createDialOptions(cfg)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create dial options: %w", err)
	}

	ctx := context.Background()
	conn, err := dialDgraphWithRetry(ctx, cfg, opts, logger)
	if err != nil {
		return nil, nil, err
	}

	dgraphClient := dgo.NewDgraphClient(api.NewDgraphClient(conn))

	cleanup := func() {
		if err := conn.Close(); err != nil {
			logger.Error("Failed to close connection", zap.Error(err))
		}
	}

	logger.Info("Successfully connected to Dgraph", zap.String("address", cfg.Address))

	return dgraphClient, cleanup, nil
}
