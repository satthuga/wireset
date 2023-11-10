package dgraphsvc

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"os"

	"github.com/dgraph-io/dgo/v2"
	"github.com/dgraph-io/dgo/v2/protos/api"
	"github.com/google/wire"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

var DefaultWireSet = wire.NewSet(
	NewDgraphSvc,
)

func FetchDgraphAddress() (string, error) {
	host := os.Getenv("DGRAPH_ADDRESS")
	if host == "" {
		return "", errors.New("DGRAPH_ADDRESS environment variable is not provided or is empty")
	}
	return host, nil
}

// CreateDialOptions creates the gRPC dial options.
func CreateDialOptions(host string) ([]grpc.DialOption, error) {
	var opts []grpc.DialOption
	opts = append(opts, grpc.WithAuthority(host))

	systemRoots, err := x509.SystemCertPool()
	if err != nil {
		return nil, errors.WithMessage(err, "Failed to get system certificate pool")
	}

	cred := credentials.NewTLS(&tls.Config{
		RootCAs: systemRoots,
	})
	opts = append(opts, grpc.WithTransportCredentials(cred))

	return opts, nil
}

// DialDgraph dials the Dgraph host with the specified gRPC options.
func DialDgraph(host string, opts []grpc.DialOption) (*grpc.ClientConn, error) {
	conn, err := grpc.Dial(host, opts...)
	if err != nil {
		return nil, errors.WithMessagef(err, "Failed to connect to Dgraph at %s", host)
	}
	return conn, nil
}

// NewDgraphSvc creates a new Dgraph client.
func NewDgraphSvc(logger *zap.Logger) (*dgo.Dgraph, func(), error) {
	host, err := FetchDgraphAddress()
	if err != nil {
		return nil, nil, fmt.Errorf("failed to fetch Dgraph address: %w", err)
	}

	opts, err := CreateDialOptions(host)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create dial options: %w", err)
	}

	conn, err := DialDgraph(host, opts)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to dial Dgraph: %w", err)
	}

	dgraphClient := dgo.NewDgraphClient(api.NewDgraphClient(conn))

	cleanup := func() {
		if err := conn.Close(); err != nil {
			logger.Error("Failed to close connection", zap.Error(err))
		}
	}

	return dgraphClient, cleanup, nil
}
