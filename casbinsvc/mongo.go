package casbinsvc

import (
	"fmt"
	"github.com/casbin/casbin/v2"
	"github.com/casbin/casbin/v2/persist"
	mongodbadapter "github.com/casbin/mongodb-adapter/v3"
	"github.com/pkg/errors"
	"os"
)

func NewMongoAdapterFromEnv() (persist.BatchAdapter, error) {
	uri := os.Getenv("CASBIN_MONGODB_URI")
	if uri == "" {
		return nil, fmt.Errorf("CASBIN_MONGODB_URI is required")
	}

	adapter, err := mongodbadapter.NewAdapter(uri)
	if err != nil {
		return nil, err
	}

	return adapter, nil
}

func NewMongoEnforcerFromEnv() (*casbin.Enforcer, error) {
	adapter, err := NewMongoAdapterFromEnv()
	if err != nil {
		return nil, errors.WithMessage(err, "failed to create mongo adapter")
	}

	enforcer, err := NewEnforcer(NewModel(), adapter)
	if err != nil {
		return nil, errors.WithMessage(err, "failed to create casbin enforcer")
	}

	return enforcer, nil
}
