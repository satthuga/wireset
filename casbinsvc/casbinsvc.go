package casbinsvc

import (
	"fmt"
	"github.com/casbin/casbin/v2"
	model "github.com/casbin/casbin/v2/model"
	"github.com/casbin/casbin/v2/persist"
	mongodbadapter "github.com/casbin/mongodb-adapter/v3"
	"github.com/pkg/errors"
	"os"
)

type CasbinMiddleware struct {
	Enforcer *casbin.Enforcer
}

func NewModel() model.Model {
	m := model.NewModel()
	m.AddDef("r", "r", "sub, obj, act")
	m.AddDef("p", "p", "sub, obj, act")
	m.AddDef("e", "e", "some(where (p.eft == allow))")
	m.AddDef("m", "m", "r.sub == p.sub && (keyMatch(r.obj, p.obj) || keyMatch2(r.obj, p.obj) || globMatch(r.obj, p.obj)) && (r.act == p.act || regexMatch(r.act, p.act))")
	return m
}

func NewMongoAdapterFromEnv() (persist.BatchAdapter, error) {
	uri := os.Getenv("CASBIN_MONGO_URI")
	if uri == "" {
		return nil, fmt.Errorf("CASBIN_MONGO_URI is required")
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

func NewEnforcer(model model.Model, adapter persist.BatchAdapter) (*casbin.Enforcer, error) {
	enforcer, err := casbin.NewEnforcer(model, adapter)
	if err != nil {
		return nil, errors.WithMessage(err, "failed to create casbin enforcer")
	}

	enforcer.EnableAcceptJsonRequest(true)
	enforcer.EnableLog(true)
	enforcer.EnableAutoSave(true)

	return enforcer, nil
}

func NewCasbinMiddleware(enforcer *casbin.Enforcer) (*CasbinMiddleware, error) {

	casbinMiddleware := &CasbinMiddleware{
		Enforcer: enforcer,
	}

	return casbinMiddleware, nil
}
