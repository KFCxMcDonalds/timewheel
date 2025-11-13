package timewheel

import "time"


func time2ms(t time.Time) int64 {
	// converte time to unix ms
	return t.UnixMilli()
}

func ms2time(ms int64) time.Time {
	// converte unix ms to time
	return time.UnixMilli(ms)
}

func truncate(t, tickMS int64) int64 {
	// align with interval
	return t - t%tickMS
}

// func ms2DateString(ms int64) string {
// 	return time.UnixMilli(ms).Format("2006-01-02 15:04:05")
// }
