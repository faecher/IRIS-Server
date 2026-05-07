// SPDX-License-Identifier: EUPL-1.2

package models

// TraccarTracker represents a tracker device using Traccar
type TraccarTracker struct {
	BaseTracker

	TraccarID string `db:"traccar_id" json:"traccarID"`
}
