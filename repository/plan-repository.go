package repository

import (
	"errors"
)

var (
	ErrPlanNotFound = errors.New("plan not found")
	ErrNoPlanFound  = errors.New("no plan found for shop")
)

// Plan is a model for a pricing plan.
type Plan struct {
	ID          string     `json:"id"`
	Name        string     `json:"name"`
	Description string     `json:"description"`
	Price       float64    `json:"price"`
	Features    []*Feature `json:"features"`
}

// Feature is a model for a feature.
type Feature struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

// PlanRepository is an interface for a pricing plan repository.
type PlanRepository interface {
	GetPlan(ID string) (*Plan, error)
	ListPlans() ([]*Plan, error)
	CreatePlan(plan *Plan) error
	UpdatePlan(plan *Plan) error
	DeletePlan(ID string) error
	GetFeaturesForPlan(ID string) ([]*Feature, error)
	CanPlanFeature(planID, featureID string) (bool, error)
	GetPlansOfShop(shopID string) ([]*Plan, error)
	CanShopFeature(shopID, featureID string) (bool, error)
}

// MemoryPlanRepository is an in-memory implementation of PlanRepository.
type MemoryPlanRepository struct {
	plans    []*Plan
	shopPlan map[string]string
}

// NewMemoryPlanRepository creates a new instance of MemoryPlanRepository.
func NewMemoryPlanRepository(plans []*Plan) *MemoryPlanRepository {
	return &MemoryPlanRepository{
		plans: plans,
	}
}

// GetPlan returns the pricing plan with the given ID.
func (r *MemoryPlanRepository) GetPlan(ID string) (*Plan, error) {
	for _, plan := range r.plans {
		if plan.ID == ID {
			return plan, nil
		}
	}
	return nil, ErrPlanNotFound
}

// ListPlans returns a list of all pricing plans.
func (r *MemoryPlanRepository) ListPlans() ([]*Plan, error) {
	return r.plans, nil
}

// CreatePlan creates a new pricing plan.
func (r *MemoryPlanRepository) CreatePlan(plan *Plan) error {
	r.plans = append(r.plans, plan)
	return nil
}

// UpdatePlan updates an existing pricing plan.
func (r *MemoryPlanRepository) UpdatePlan(plan *Plan) error {
	for i, p := range r.plans {
		if p.ID == plan.ID {
			r.plans[i] = plan
			return nil
		}
	}
	return ErrPlanNotFound
}

// DeletePlan deletes a pricing plan.
func (r *MemoryPlanRepository) DeletePlan(ID string) error {
	for i, plan := range r.plans {
		if plan.ID == ID {
			r.plans = append(r.plans[:i], r.plans[i+1:]...)
			return nil
		}
	}
	return ErrPlanNotFound
}

// GetFeaturesForPlan returns a list of features that are included in the given pricing plan.
func (r *MemoryPlanRepository) GetFeaturesForPlan(ID string) ([]*Feature, error) {
	for _, plan := range r.plans {
		if plan.ID == ID {
			return plan.Features, nil
		}
	}
	return nil, ErrPlanNotFound
}

// CanPlanFeature checks if the given plan ID has the given feature ID.
func (r *MemoryPlanRepository) CanPlanFeature(planID, featureID string) (bool, error) {
	for _, plan := range r.plans {
		if plan.ID == planID {
			for _, feature := range plan.Features {
				if feature.ID == featureID {
					return true, nil
				}
			}
			return false, nil
		}
	}
	return false, ErrPlanNotFound
}

// GetPlansOfShop returns a list of pricing plans for the given shop ID.
func (r *MemoryPlanRepository) GetPlansOfShop(shopID string) ([]*Plan, error) {
	var plans []*Plan
	for _, plan := range r.plans {
		if r.shopPlan[shopID] == plan.ID {
			plans = append(plans, plan)
		}
	}
	if len(plans) == 0 {
		return nil, ErrNoPlanFound
	}
	return plans, nil
}

// CanShopFeature checks if the given shop ID has the given feature ID.
func (r *MemoryPlanRepository) CanShopFeature(shopID, featureID string) (bool, error) {
	planID, ok := r.shopPlan[shopID]
	if !ok {
		return false, ErrNoPlanFound
	}
	return r.CanPlanFeature(planID, featureID)
}
