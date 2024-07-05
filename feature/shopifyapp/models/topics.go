package models

import "github.com/aiocean/wireset/feature/realtime/models"

const TopicFetchActivateSubscription models.WebsocketTopic = "fetchActiveSubscription"

const TopicSetActivateSubscription models.WebsocketTopic = "setActiveSubscription"

type SetActivateSubscriptionPayload struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	TrialDays int    `json:"trialDays"`
	Status    string `json:"status"`
}

const TopicCreateSubscription models.WebsocketTopic = "createSubscription"

type CreateSubscriptionPayload struct {
	Plan string `json:"plan"`
}

const TopicNavigateTo models.WebsocketTopic = "navigateTo"

type NavigateToPayload struct {
	URL string `json:"url"`
}
