package dgraphsvc

import (
	"crypto/tls"
	"crypto/x509"
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

// NewDgraphSvc creates a new Dgraph client.
// It returns an error if the connection fails.
// It returns a cleanup function that should be called when the client is no longer needed.
func NewDgraphSvc(
	logger *zap.Logger,
) (*dgo.Dgraph, func(), error) {
	host := os.Getenv("DGRAPH_ADDRESS")

	var opts []grpc.DialOption

	opts = append(opts, grpc.WithAuthority(host))

	systemRoots, err := x509.SystemCertPool()
	if err != nil {
		return nil, nil, errors.WithMessage(err, "failed to get system cert pool")
	}

	cred := credentials.NewTLS(&tls.Config{
		RootCAs: systemRoots,
	})
	opts = append(opts, grpc.WithTransportCredentials(cred))

	conn, err := grpc.Dial(host, opts...)
	if err != nil {
		return nil, nil, errors.WithMessage(err, "failed to connect to dgraph")
	}
	dgraphClient := dgo.NewDgraphClient(api.NewDgraphClient(conn))

	cleanup := func() {
		if err := conn.Close(); err != nil {
			logger.Error("failed to close connection", zap.Error(err))
		}

	}

	return dgraphClient, cleanup, nil
}
