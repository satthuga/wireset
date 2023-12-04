package casbinsvc

import (
	"github.com/casbin/casbin/v2"
	model "github.com/casbin/casbin/v2/model"
	"github.com/casbin/casbin/v2/persist"
	"github.com/pkg/errors"
)

func NewModel() model.Model {
	m := model.NewModel()
	m.AddDef("r", "r", "sub, obj, act")
	m.AddDef("p", "p", "sub, obj, act")
	m.AddDef("e", "e", "some(where (p.eft == allow))")
	m.AddDef("m", "m", "r.sub == p.sub && (keyMatch(r.obj, p.obj) || keyMatch2(r.obj, p.obj) || globMatch(r.obj, p.obj)) && (r.act == p.act || regexMatch(r.act, p.act))")
	return m
}

func NewEnforcer(model model.Model, adapter persist.Adapter) (*casbin.Enforcer, error) {
	enforcer, err := casbin.NewEnforcer(model, adapter)
	if err != nil {
		return nil, errors.WithMessage(err, "failed to create casbin enforcer")
	}

	enforcer.EnableAcceptJsonRequest(true)
	enforcer.EnableLog(true)
	enforcer.EnableAutoSave(true)

	return enforcer, nil
}
