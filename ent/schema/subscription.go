package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"github.com/google/uuid"
)

// Subscription holds the schema definition for the Subscription entity.
//
// One subscription per user (see the "user" edge below). Postgres is the
// source of truth for this table; Redis only ever holds a short-lived,
// invalidate-on-write cache of it — see internal/handlers/subscription.go.
type Subscription struct {
	ent.Schema
}

func (Subscription) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).
			Default(uuid.New).
			Unique().
			Immutable(),
		field.Enum("status").
			Values("free", "active", "grace_period", "expired", "cancelled").
			Default("free"),
		field.String("plan").
			Optional().
			Comment("e.g. pro_monthly_99, pro_monthly_299"),
		field.Time("current_period_end").
			Optional().
			Nillable(),
		field.Time("updated_at").
			Default(time.Now).
			UpdateDefault(time.Now),
	}
}

func (Subscription) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("user", User.Type).
			Ref("subscription").
			Unique().
			Required(),
	}
}
