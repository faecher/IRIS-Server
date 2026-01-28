package models

type TetraTracker struct {
	BaseTracker
	ISSI string `json:"issi" db:"issi"`
}
