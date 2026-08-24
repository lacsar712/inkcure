package model

import "time"

const (
	DefaultLeaseTTL        = 30 * time.Second
	DwellWindow            = 5 * time.Minute
	IgnitionDelayWindow    = 15 * time.Second
	InkvatSwellSettleWindow  = 45 * time.Second
	UvbankWarmupWindow = 2 * time.Minute
	FeedwaterRampWindow    = 30 * time.Second
	MaxInkvatLevelPercent    = 95.0
	MinInkvatLevelPercent    = 15.0
	TripInkvatLowPercent     = 10.0
	TripInkvatHighPercent    = 98.0
	NormalSteamPressurePSI = 1800.0
	MaxSteamPressurePSI    = 2000.0
	MinLampframeO2Percent    = 2.5
	MaxLampframeO2Percent    = 6.0
	DefaultJournalCapacity = 512
)
