package entity

import "time"

type Link struct {
	ID         int64
	ShortID    string
	Href       string
	CreatedAt  time.Time
	UsageCount int64
	UsageAt    time.Time
}
