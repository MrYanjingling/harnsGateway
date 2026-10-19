package apis

import "time"

// Meta holds the common base fields shared by persistent entities.
type Meta struct {
	// ID is the entity identifier.
	ID string `json:"id"`
	// Version is the entity version, incremented on every update. It can be
	// used for optimistic concurrency control.
	Version int64 `json:"version"`
	// CreatedUser is the creator of the entity.
	CreatedUser string `json:"createdUser"`
	// UpdatedUser is the last updater of the entity.
	UpdatedUser string `json:"updatedUser"`
	// CreatedTime is when the entity was created.
	CreatedTime time.Time `json:"createdTime"`
	// UpdatedTime is when the entity was last updated.
	UpdatedTime time.Time `json:"updatedTime"`
}
