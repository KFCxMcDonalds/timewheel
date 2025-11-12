package timewheel

import "time"


func time2MS(t time.Time) int64 {
	// converte time to unix ms
	return t.UnixMilli()
}

func truncate(t, tickMS int64) int64 {
	// align with interval
	return t - t%tickMS
}

// func ms2DateString(ms int64) string {
// 	return time.UnixMilli(ms).Format("2006-01-02 15:04:05")
// }
