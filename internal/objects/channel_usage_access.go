package objects

// ChannelUsageAccess uses stable account IDs, not API keys. Nil preserves legacy public access.
type ChannelUsageAccess struct {
	Mode    ChannelUsageAccessMode `json:"mode"`
	UserIDs []int                  `json:"userIDs"`
}

type ChannelUsageAccessMode string

const (
	ChannelUsageAccessAll      ChannelUsageAccessMode = "all"
	ChannelUsageAccessSelected ChannelUsageAccessMode = "selected"
)
