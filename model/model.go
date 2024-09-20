package model

import "time"

type Dish struct {
	Id        int
	Name      string
	Url       string
	Created   time.Time
	UsedCount int
	LastUsage time.Time
}

type UsageStats struct {
	Count     int
	DaysSince int
}

type TemplateDish struct {
	Id           int
	Name         string
	Url          string
	Created      time.Time
	UsageStats   UsageStats
	UsageOptions UsageOptions
}

type UsageOptions struct {
	Today      UsageOption
	Yesterday  UsageOption
	WithinWeek UsageOption
}

type UsageOption struct {
	Id   int64
	Name string
}
