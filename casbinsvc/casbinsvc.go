package casbinsvc

import (
	"fmt"
	"github.com/casbin/casbin/v2"
	model "github.com/casbin/casbin/v2/model"
	mongodbadapter "github.com/casbin/mongodb-adapter/v3"
	"github.com/google/wire"
	"os"
)

var DefaultWireset = wire.NewSet(
	NewCasbinSvcConfigFromEnv,
	NewEnforcer,
	NewCasbinSvc,
)

type Config struct {
	MongoURI string
	Model    model.Model
}

func NewCasbinSvcConfigFromEnv() (*Config, error) {
	uri := os.Getenv("MONGO_URI")
	if uri == "" {
		return nil, fmt.Errorf("MONGO_URI is required")
	}

	m := model.NewModel()
	m.AddDef("r", "r", "sub, obj, act")
	m.AddDef("p", "p", "sub, obj, act")
	m.AddDef("e", "e", "some(where (p.eft == allow))")
	m.AddDef("m", "m", "r.sub == p.sub && (keyMatch(r.obj, p.obj) || keyMatch2(r.obj, p.obj) || globMatch(r.obj, p.obj)) && (r.act == p.act || regexMatch(r.act, p.act))")

	return &Config{
		MongoURI: uri,
		Model:    m,
	}, nil
}

type CasbinMiddleware struct {
	Enforcer *casbin.Enforcer
}

func NewEnforcer(config *Config) (*casbin.Enforcer, error) {
	adapter, err := mongodbadapter.NewAdapter(config.MongoURI)
	if err != nil {
		return nil, err
	}

	enforcer, err := casbin.NewEnforcer(config.Model, adapter)
	if err != nil {
		return nil, err
	}

	enforcer.EnableAcceptJsonRequest(true)
	enforcer.EnableLog(true)
	enforcer.EnableAutoSave(true)

	return enforcer, nil
}

func NewCasbinSvc(config *Config, enforcer *casbin.Enforcer) (*CasbinMiddleware, error) {

	casbinMiddleware := &CasbinMiddleware{
		Enforcer: enforcer,
	}

	return casbinMiddleware, nil
}
