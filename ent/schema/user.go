package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"github.com/google/uuid"
)

// User holds the schema definition for the User entity.
//
// A user is created (or identified, if already known) purely from the
// Bearer token presented on the first authenticated request — see
// internal/middleware/auth.go. This is deliberately the "basic auth" the
// assignment asks for; a production system would swap the token lookup
// for verified Firebase/JWT claims without touching this schema.
type User struct {
	ent.Schema
}

func (User) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).
			Default(uuid.New).
			Unique().
			Immutable(),
		field.String("token").
			NotEmpty().
			Unique().
			Sensitive().
			Comment("Bearer token used to identify/create this user"),
		field.Time("created_at").
			Default(time.Now).
			Immutable(),
	}
}

func (User) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("subscription", Subscription.Type).
			Unique(),
	}
}
