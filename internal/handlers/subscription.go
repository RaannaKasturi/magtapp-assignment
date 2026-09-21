package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/redis/go-redis/v9"

	"magtapp-backend/ent"
	"magtapp-backend/ent/subscription"
	"magtapp-backend/internal/dto"
	"magtapp-backend/internal/middleware"
)

// SubscriptionHandler holds the two stores this service touches:
// Postgres (source of truth, via Ent) and Redis (read cache +
// idempotency locks). No other package talks to either store directly.
type SubscriptionHandler struct {
	DB    *ent.Client
	Cache *redis.Client
}

func NewSubscriptionHandler(db *ent.Client, cache *redis.Client) *SubscriptionHandler {
	return &SubscriptionHandler{DB: db, Cache: cache}
}

const cacheTTL = 5 * time.Minute

func entitlementKey(userID string) string {
	return "entitlement:" + userID
}

// Get returns the caller's subscription status.
//
// Redis is the fast path; Postgres is always the fallback of record on a
// cache miss, and a miss is repopulated before responding. If the user
// has never had a subscription row created, one is created now, defaulted
// to "free" (matches Ent's schema default) — this is the "creates or
// identifies" behavior applied to subscriptions, mirroring what Auth
// middleware does for users.
func (h *SubscriptionHandler) Get(c echo.Context) error {
	u := middleware.FromContext(c)
	ctx := c.Request().Context()
	key := entitlementKey(u.ID.String())

	if cached, err := h.Cache.Get(ctx, key).Result(); err == nil {
		var resp dto.SubscriptionResponse
		if json.Unmarshal([]byte(cached), &resp) == nil {
			return c.JSON(http.StatusOK, resp)
		}
	}

	sub, err := u.QuerySubscription().Only(ctx)
	if ent.IsNotFound(err) {
		sub, err = h.DB.Subscription.Create().SetUser(u).Save(ctx)
	}
	if err != nil {
		return c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Error: "failed to load subscription"})
	}

	resp := toResponse(u.ID.String(), sub)
	h.writeCache(ctx, key, resp)
	return c.JSON(http.StatusOK, resp)
}

// Update changes the caller's subscription status.
//
// An optional Idempotency-Key header (mirroring Apple/Google's
// transaction/notification IDs in the real design) makes retried calls
// safe: the first request with a given key processes normally; any
// retry with the same key short-circuits to a plain read instead of
// re-applying the write. Keys are held for 24h, comfortably longer than
// any realistic client/webhook retry window.
func (h *SubscriptionHandler) Update(c echo.Context) error {
	u := middleware.FromContext(c)
	ctx := c.Request().Context()

	var req dto.UpdateSubscriptionRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "invalid request body"})
	}
	if err := subscription.StatusValidator(subscription.Status(req.Status)); err != nil {
		return c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "invalid status value"})
	}

	if idemKey := c.Request().Header.Get("Idempotency-Key"); idemKey != "" {
		lockKey := "idempotency:" + idemKey
		acquired, err := h.Cache.SetNX(ctx, lockKey, "1", 24*time.Hour).Result()
		if err == nil && !acquired {
			// Already processed once — return current state, don't reapply.
			return h.Get(c)
		}
	}

	sub, err := u.QuerySubscription().Only(ctx)
	if ent.IsNotFound(err) {
		sub, err = h.DB.Subscription.Create().SetUser(u).Save(ctx)
	}
	if err != nil {
		return c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Error: "failed to load subscription"})
	}

	update := sub.Update().SetStatus(subscription.Status(req.Status))
	if req.Plan != "" {
		update = update.SetPlan(req.Plan)
	}
	if req.CurrentPeriodEnd != nil {
		update = update.SetCurrentPeriodEnd(*req.CurrentPeriodEnd)
	}
	sub, err = update.Save(ctx)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Error: "failed to update subscription"})
	}

	// Postgres (source of truth) just changed — invalidate rather than
	// patch the cache in place, then repopulate with the fresh value.
	key := entitlementKey(u.ID.String())
	h.Cache.Del(ctx, key)

	resp := toResponse(u.ID.String(), sub)
	h.writeCache(ctx, key, resp)
	return c.JSON(http.StatusOK, resp)
}

func toResponse(userID string, sub *ent.Subscription) dto.SubscriptionResponse {
	return dto.SubscriptionResponse{
		UserID:           userID,
		Status:           string(sub.Status),
		Plan:             sub.Plan,
		CurrentPeriodEnd: sub.CurrentPeriodEnd,
		UpdatedAt:        sub.UpdatedAt,
	}
}

func (h *SubscriptionHandler) writeCache(ctx context.Context, key string, resp dto.SubscriptionResponse) {
	b, err := json.Marshal(resp)
	if err != nil {
		return
	}
	h.Cache.Set(ctx, key, b, cacheTTL)
}
