package dto

import "time"

// MeResponse is returned by GET /v1/me.
type MeResponse struct {
	ID        string    `json:"id" required:"true"`
	CreatedAt time.Time `json:"created_at" required:"true"`
}

// SubscriptionResponse is returned by GET and PUT /v1/subscription.
type SubscriptionResponse struct {
	UserID           string     `json:"user_id" required:"true"`
	Status           string     `json:"status" required:"true" enum:"free,active,grace_period,expired,cancelled"`
	Plan             string     `json:"plan,omitempty" description:"e.g. pro_monthly_99, pro_monthly_299"`
	CurrentPeriodEnd *time.Time `json:"current_period_end,omitempty"`
	UpdatedAt        time.Time  `json:"updated_at" required:"true"`
}

// UpdateSubscriptionRequest is the body for PUT /v1/subscription.
//
// In a real deployment this endpoint is what an Apple/Google webhook (or a
// verified receipt-check call) would invoke internally — this handler is
// the single place entitlement is ever written, whether the caller is the
// client, an internal job, or a webhook receiver.
type UpdateSubscriptionRequest struct {
	Status           string     `json:"status" required:"true" enum:"free,active,grace_period,expired,cancelled" description:"New subscription status"`
	Plan             string     `json:"plan,omitempty" description:"Plan identifier, e.g. pro_monthly_99"`
	CurrentPeriodEnd *time.Time `json:"current_period_end,omitempty"`
}

// ErrorResponse is the shape of every non-2xx JSON body.
type ErrorResponse struct {
	Error string `json:"error" required:"true"`
}
