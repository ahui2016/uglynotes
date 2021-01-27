package model

import (
	"crypto/rand"
	"math/big"
	"strconv"
	"time"
)

// IncreaseID 用来记录自动生成 ID 的状态，便于生成特有的自增 ID.
// 该 ID 由年份与自增数两部分组成，分别取两个部分的 36 进制, 转字符串后拼接而成。
type IncreaseID struct {
	Year  int
	Count int
}

// FirstID 生成初始 id, 当且仅当程序每一次使用时（数据库为空时）使用该函数，
// 之后应使用 Increase 函数来获得新 id.
func FirstID() IncreaseID {
	nowYear := time.Now().Year()
	return IncreaseID{nowYear, 0}
}

// ParseID 把字符串形式的 id 转换为 IncreaseID.
// (有“万年虫”问题，但是当然，这个问题可以忽略。)
func ParseID(strID string) (id IncreaseID, err error) {
	strYear := strID[:3] // 可以认为年份总是占前三个字符
	strCount := strID[3:]
	year, err := strconv.ParseInt(strYear, 36, 0)
	if err != nil {
		return id, err
	}
	count, err := strconv.ParseInt(strCount, 36, 0)
	if err != nil {
		return id, err
	}
	id.Year = int(year)
	id.Count = int(count)
	return
}

// Increase 使 id 自增一次，输出自增后的新 id.
// 如果当前年份大于 id 中的年份，则年份进位，Count 重新计数。
// 否则，年份不变，Count 加一。
func (id IncreaseID) Increase() IncreaseID {
	nowYear := time.Now().Year()
	if nowYear > id.Year {
		return IncreaseID{nowYear, 1}
	}
	return IncreaseID{id.Year, id.Count + 1}
}

// String 返回 id 的字符串形式。
func (id IncreaseID) String() string {
	year := strconv.FormatInt(int64(id.Year), 36)
	count := strconv.FormatInt(int64(id.Count), 36)
	return year + count
}

// RandomID 返回一个随机字符串，与 IncreaseID 无关。
func RandomID() string {
	var max int64 = 100_000_000
	n, err := rand.Int(rand.Reader, big.NewInt(max))
	if err != nil {
		panic(err)
	}
	timestamp := time.Now().Unix()
	idInt64 := timestamp*max + n.Int64()
	return strconv.FormatInt(idInt64, 36)
}

// TimeID 返回一个基于时间的 ID, 与 IncreaseID 无关。
func TimeID() string {
	unixMicro := time.Now().UnixNano() / 1000
	return strconv.FormatInt(unixMicro, 36)
}

func NextTimeID() string {
	time.Sleep(100 * time.Microsecond)
	return TimeID()
}
