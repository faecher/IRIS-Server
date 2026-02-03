// SPDX-License-Identifier: EUPL-1.2

package models

// TetraTracker represents a tracker device using TETRA radio
type TetraTracker struct {
	BaseTracker

	ISSI string `db:"issi" json:"issi"`
}
