package model

const (
	NoDataValue float64 = -1
)

type MetricsSummary struct {
	GRPCRPCTotal          float64
	GRPCFailurePercentage float64
	GRPCAll               ValueByLabel
	GRPCByMethod          map[string]ValueByLabel

	HTTPRequestTotal      float64
	HTTPFailurePercentage float64
	HTTPAll               ValueByLabel
	HTTPByPath            map[string]ValueByLabel

	VirtualUserStartedTotal float64
	VirtualUserFailedTotal  float64
}

type ValueByLabel map[string]float64

const (
	GRPCRPCsKey              = "RPCs"
	GRPCFailurePercentageKey = "FailurePercentage"
	GRPCLatencyAvgKey        = "LatencyAvg"
	GRPCSentBytesAvgKey      = "SentBytesAvg"
	GRPCReceivedBytesAvgKey  = "ReceivedBytesAvg"

	HTTPRequestsKey          = "Requests"
	HTTPFailurePercentageKey = "FailurePercentage"
	HTTPLatencyAvgKey        = "LatencyAvg"
	HTTPSentBytesAvgKey      = "SentBytesAvg"
	HTTPReceivedBytesAvgKey  = "ReceivedBytesAvg"
)
