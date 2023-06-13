package flagsvc

import (
	"context"

	firebase "firebase.google.com/go"
	"github.com/google/wire"
	"github.com/pkg/errors"
	"go.uber.org/zap"
)

var DefaultWireSet = wire.NewSet(
	NewFlagSvc,
)

// FlagSvc is feature flag service
type FlagSvc struct {
	firebaseApp *firebase.App
	logger      *zap.Logger
}

func NewFlagSvc(ctx context.Context, firebaseApp *firebase.App, logger *zap.Logger) (*FlagSvc, func(), error) {

	return nil, nil, errors.New("not implemented")
}

func (s *FlagSvc) BoolVariation(ctx context.Context, flagName string, defaultVal bool) (bool, error) {
	return defaultVal, nil
}

func (s *FlagSvc) IntVariation(ctx context.Context, flagName string, defaultVal int64) (int64, error) {
	// TODO: Implement the logic for the IntVariation function
	// For now, we'll just return the default value
	return defaultVal, nil
}

func (s *FlagSvc) Float64Variation(ctx context.Context, flagName string, defaultVal float64) (float64, error) {
	// TODO: Implement the logic for the Float64Variation function
	// For now, we'll just return the default value
	return defaultVal, nil
}

func (s *FlagSvc) StringVariation(ctx context.Context, flagName string, defaultVal string) (string, error) {
	// TODO: Implement the logic for the StringVariation function
	// For now, we'll just return the default value
	return defaultVal, nil
}

func (s *FlagSvc) JSONVariation(ctx context.Context, flagName string, defaultVal interface{}) (interface{}, error) {
	// TODO: Implement the logic for the JSONVariation function
	// For now, we'll just return the default value
	return defaultVal, nil
}
