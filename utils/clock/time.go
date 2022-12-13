package clock

import (
	"math/rand"
	"strconv"
	"time"
)

// time format alias
const (
	TimeFmtStrYear   = "2006"
	TimeFmtStrMonth  = "01"
	TimeFmtStrDay    = "02"
	TimeFmtStrHour   = "15"
	TimeFmtStrMinute = "04"
	TimeFmtStrSecond = "05"
)

// TimeGeneralLayout general time layout: YYYY-MM-DD HH:ii:ss
func TimeGeneralLayout() string {
	layout := TimeFmtStrYear + "-" + TimeFmtStrMonth + "-" + TimeFmtStrDay + " "
	layout += TimeFmtStrHour + ":" + TimeFmtStrMinute + ":" + TimeFmtStrSecond

	return layout
}

// RandByTime random base time
func RandByTime(t *time.Time) int64 {
	var timeInt int64

	if t == nil {
		timeInt = time.Now().UnixNano()
	} else {
		timeInt = t.UnixNano()
	}

	return rand.New(rand.NewSource(timeInt)).Int63()
}

// ParseTime parse general time layout to int64.
func ParseTime(timeStr string) (int64, error) {
	t, err := time.Parse(TimeGeneralLayout(), timeStr)
	if err != nil {
		return -1, err
	}
	return t.Unix(), nil
}

// TimeStampToStr convert timestamp to string.
func TimeStampToStr(timestamp int64) string {
	return time.Unix(timestamp, 0).Format(TimeGeneralLayout())
}

// CurrTimeStr get general layout time str: YYYY-MM-DD HH:ii:ss
func CurrTimeStr() string {
	return time.Now().Format(TimeGeneralLayout())
}

// CurrTimeStamp return current timestamp
func CurrTimeStamp() int64 {
	return time.Now().Unix()
}

// CurrTimeStampString return current timestamp string
func CurrTimeStampString() string {
	return strconv.FormatInt(CurrTimeStamp(), 10)
}

// CurrTimestampInMilli get current timestamp in millisecond
func CurrTimestampInMilli() int64 {
	return time.Now().UnixNano() / (1000 * 1000) // ms
}

// CurrTimestampInMilliString return 1441007112776 in millisecond
func CurrTimestampInMilliString() string {
	return strconv.FormatInt(CurrTimestampInMilli(), 10)
}

// TimeStr2MillsecondUnix general time string to millisecond
func TimeStr2MillsecondUnix(timeStr string) int64 {
	times, _ := time.ParseInLocation(TimeGeneralLayout(), timeStr, time.Local)
	return times.UnixNano() / 1000000
}

// TimeStr2SecondUnix time str to timestamp seconds
func TimeStr2SecondUnix(timeStr string) int64 {
	times, _ := time.ParseInLocation(TimeGeneralLayout(), timeStr, time.Local)
	return times.Unix()
}
