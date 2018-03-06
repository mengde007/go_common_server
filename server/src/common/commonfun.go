package common

import (
	"bytes"
	gp "code.google.com/p/goprotobuf/proto"
	"code.google.com/p/snappy-go/snappy"
	"crypto/rc4"
	// "csvcfg"
	"encoding/binary"
	"encoding/gob"
	"encoding/json"
	// "errors"
	"fmt"
	"hash/crc32"
	"io"
	"logger"
	"math/rand"
	// "path"
	// "proto"
	// "rpc"
	// "sort"
	"strconv"
	"strings"
	"sync/atomic"
	"time"
)

func EncodeMessage(value gp.Message) (result []byte, err error) {
	//ts("KVWrite", table, uid)
	//defer te("KVWrite", table, uid)

	buf, err := gp.Marshal(value)

	if err != nil {
		return
	}

	result, err = snappy.Encode(nil, buf)

	return
}

func DecodeMessage(value []byte, result gp.Message) (err error) {
	var dst []byte

	dst, err = snappy.Decode(nil, value)

	if err != nil {
		return
	}

	err = gp.Unmarshal(dst, result)

	return
}

//唯一id生成
var uuid uint32 = 0

// UUID() provides unique identifier strings.
func GenUUID(sid uint8) string {
	b := make([]byte, 16)

	t := time.Now().Unix()
	tmpid := uint16(atomic.AddUint32(&uuid, 1))

	b[0] = byte(sid)
	b[1] = byte(0)
	b[2] = byte(tmpid)
	b[3] = byte(tmpid >> 8)

	b[4] = byte(t)
	b[5] = byte(t >> 8)
	b[6] = byte(t >> 16)
	b[7] = byte(t >> 24)

	c, _ := rc4.NewCipher([]byte{0x0c, b[2], b[3], b[0]})
	c.XORKeyStream(b[8:], b[:8])

	return fmt.Sprintf("%x-%x-%x-%x-%x", b[:4], b[4:6], b[6:8], b[8:12], b[12:])
}

func GenUUIDWith32(sid uint8) string {
	b := make([]byte, 16)

	t := time.Now().Unix()
	tmpid := uint16(atomic.AddUint32(&uuid, 1))

	b[0] = byte(sid)
	b[1] = byte(0)
	b[2] = byte(tmpid)
	b[3] = byte(tmpid >> 8)

	b[4] = byte(t)
	b[5] = byte(t >> 8)
	b[6] = byte(t >> 16)
	b[7] = byte(t >> 24)

	c, _ := rc4.NewCipher([]byte{0x0c, b[2], b[3], b[0]})
	c.XORKeyStream(b[8:], b[:8])

	return fmt.Sprintf("%x%x%x%x%x", b[:4], b[4:6], b[6:8], b[8:12], b[12:])
}

func CheckUUID(uid string) bool {
	if len(uid) != 36 {
		return false
	}

	b := make([]uint32, 5)

	_, err := fmt.Sscanf(uid, "%x-%x-%x-%x-%x", &b[0], &b[1], &b[2], &b[3], &b[4])
	if err != nil {
		return false
	}

	info1 := make([]byte, 4)
	binary.BigEndian.PutUint32(info1, b[0])

	info2 := make([]byte, 4)
	binary.BigEndian.PutUint16(info2[:2], uint16(b[1]))
	binary.BigEndian.PutUint16(info2[2:], uint16(b[2]))

	c, _ := rc4.NewCipher([]byte{0x0c, info1[2], info1[3], info1[0]})

	tmp := make([]byte, 4)

	c.XORKeyStream(tmp, info1)

	if binary.BigEndian.Uint32(tmp) != b[3] {
		return false
	}

	c.XORKeyStream(tmp, info2)

	if binary.BigEndian.Uint32(tmp) != b[4] {
		return false
	}

	return true
}

func GenMailId() string {
	return "mail-" + GenUUID(0)
}

//是否是同一天（后面的sec表示几点为分隔线，如凌晨4点则传入4*60*60）
func IsTheSameDay(utime1, utime2 uint32, sec int) bool {
	time1 := time.Unix(int64(utime1)-int64(sec), 0)
	time2 := time.Unix(int64(utime2)-int64(sec), 0)

	return time1.YearDay() == time2.YearDay() && time1.Year() == time2.Year()
}

//是否是同一周（后面的秒数表示从星期天0点开始算的偏移秒数）
func IsTheSameWeek(utime1, utime2 uint32, sec int) bool {
	time1 := time.Unix(int64(utime1)-int64(sec)+24*3600, 0)
	time2 := time.Unix(int64(utime2)-int64(sec)+24*3600, 0)
	//iosweek是按周一0点分隔的
	year1, week1 := time1.ISOWeek()
	year2, week2 := time2.ISOWeek()
	return year1 == year2 && week1 == week2
}

//取几天后几时几分几秒的时间点
func GetNextDayHourTime(curTime int64, day, hour, min, sec int, btoday bool) int64 {
	if day <= 0 {
		panic("get next time day must > 0")
		return 0
	}

	timeCur := time.Unix(curTime, 0)
	if btoday &&
		(hour > timeCur.Hour() ||
			(hour == timeCur.Hour() && min > timeCur.Minute()) ||
			(hour == timeCur.Hour() && min == timeCur.Minute() && sec > timeCur.Second())) {
		day = timeCur.Day() + day - 1
	} else {
		day = timeCur.Day() + day
	}

	timeRet := time.Date(timeCur.Year(), timeCur.Month(), day, hour, min, sec, 0, time.Local)

	return timeRet.Unix()
}

//判断当前时间是否在时间区间内(暂时只给宝石购买奖励次数使用，其他情况不要使用该函数)
//beginTime,endTime格式为：2016.3.12.12   年.月.日.时   分秒默认为0
func IsBetweenTwoDays(curTime uint32, beginTime string, endTime string) bool {
	beginTimeArr := strings.Split(beginTime, ".")
	endTimeArr := strings.Split(endTime, ".")
	if len(endTimeArr) != 4 {
		return false
	}
	if len(beginTimeArr) != 4 {
		return false
	}

	beginYear, _ := strconv.Atoi(beginTimeArr[0])
	beginMonth, _ := strconv.Atoi(beginTimeArr[1])
	beginDay, _ := strconv.Atoi(beginTimeArr[2])
	beginHour, _ := strconv.Atoi(beginTimeArr[3])

	month1 := time.Month(beginMonth)

	timeBegin := time.Date(beginYear, month1, beginDay, beginHour, 0, 0, 0, time.Local)

	endYear, _ := strconv.Atoi(endTimeArr[0])
	endMonth, _ := strconv.Atoi(endTimeArr[1])
	endDay, _ := strconv.Atoi(endTimeArr[2])
	endHour, _ := strconv.Atoi(endTimeArr[3])

	month2 := time.Month(endMonth)
	timeEnd := time.Date(endYear, month2, endDay, endHour, 0, 0, 0, time.Local)

	if int64(curTime) >= timeBegin.Unix() && int64(curTime) <= timeEnd.Unix() {
		return true
	}
	return false
}

// 计算两次时间之间过了几天 以传入刷新时间为两天分界
// sec为刷新时间,入凌晨4点就传入4*60*60
func GetPassDayNum(lastTime, curTime uint32, sec int) uint32 {
	// 这个函数为考虑时区问题 已经废弃 调用下面的函数:GetPassDayNumSuper
	if curTime <= lastTime {
		// 不同服务器切换登录 curTime 有可能小于lastTime
		return uint32(0)
	}
	day1 := (int64(lastTime) - int64(sec)) / (24 * 3600)
	day2 := (int64(curTime) - int64(sec)) / (24 * 3600)
	// logger.Error("############### GetPassDayNum: ", day2, day1, curTime, lastTime, sec)
	return uint32(day2 - day1)
}

func GetPassDayNumSuper(lastTime, curTime uint32, sec int) uint32 {
	// if curTime <= lastTime {
	// 	// 不同服务器切换登录 curTime 有可能小于lastTime
	// 	return uint32(0)
	// }
	// time1 := time.Unix(int64(curTime)-int64(sec), 0)
	// time2 := time.Unix(int64(lastTime)-int64(sec), 0)

	// logger.Error("############### GetPassDayNumSuper: lastday curday", time1.YearDay(), time1.Hour(), time2.YearDay(), time2.Hour(), sec)
	// theDay := uint32(0)
	// if time1.Year() == time2.Year() {
	// 	theDay = uint32(time1.YearDay() - time2.YearDay())
	// 	return theDay
	// } else {
	// 	//持续超过一年的情况会有问题
	// 	theDay = uint32((time1.Year()-time2.Year())*365 + (time1.YearDay() - time2.YearDay()))
	// 	return theDay
	// }
	// return uint32(0)
	if curTime <= lastTime {
		// 不同服务器切换登录 curTime 有可能小于lastTime
		return uint32(0)
	}
	_, offsetTime := time.Now().Zone()
	day1 := (int64(lastTime) - int64(sec) + int64(offsetTime)) / (24 * 3600)
	day2 := (int64(curTime) - int64(sec) + int64(offsetTime)) / (24 * 3600)
	// logger.Error("############### GetPassDayNum: ", day2, day1, curTime, lastTime, sec)
	return uint32(day2 - day1)
}

//锁定相关
var nid uint32 = 0

func GenLockMessage(sid uint8, tid uint8, value uint8) uint64 {

	tmpid := uint8(atomic.AddUint32(&nid, 1))

	return uint64(time.Now().Unix()) | uint64(tmpid)<<32 | uint64(value)<<40 | uint64(tid)<<48 | uint64(sid)<<56
}

func ParseLockMessage(lid uint64) (sid uint8, tid uint8, value uint8, t uint32, tmpid uint8) {
	return uint8(lid >> 56), uint8(lid >> 48), uint8(lid >> 40), uint32(lid), uint8(lid >> 32)
}

//排序
func BubbleSort(values []uint32) {
	flag := true
	for i := 0; i < len(values)-1; i++ {
		flag = true

		for j := 0; j < len(values)-i-1; j++ {
			if values[j] > values[j+1] {
				values[j], values[j+1] = values[j+1], values[j]
				flag = false
			}
		}
		if flag == true {
			break
		}
	}
}

//排序
func BubbleSortExtra(values []int32) {
	flag := true
	for i := 0; i < len(values)-1; i++ {
		flag = true

		for j := 0; j < len(values)-i-1; j++ {
			if values[j] > values[j+1] {
				values[j], values[j+1] = values[j+1], values[j]
				flag = false
			}
		}
		if flag == true {
			break
		}
	}
}

//随机函数
func RandomNumber(start uint32, stop uint32) uint32 {
	if start > stop {
		start, stop = stop, start
	}

	//前闭后开
	total := stop - start + 1

	//同一时刻调用多次返回值一样
	//var randSource rand.Source = rand.NewSource(time.Now().Unix())
	//ran := rand.New(randSource)

	return uint32(rand.Intn(int(total))) + start
}

func RandomWeightTable(table map[interface{}]uint32) interface{} {
	nSum := uint32(0)
	for _, v := range table {
		nSum += v
	}

	//传入错误
	if nSum == 0 {
		logger.Error("RandomWeightTable sum wrong", nSum, table)
		return nil
	}

	nLuckNum := RandomNumber(1, nSum)
	nSum = uint32(0)
	for k, v := range table {
		nSum += v
		if nLuckNum <= nSum {
			return k
		}
	}

	logger.Error("RandomWeightTable failed", nLuckNum, nSum)
	return nil
}

//加解密
func GobEncode(arg interface{}) ([]byte, error) {
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	if err := enc.Encode(arg); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func GobDecode(arg []byte, out interface{}) error {
	buf := bytes.NewBuffer(arg)
	dec := gob.NewDecoder(buf)

	return dec.Decode(out)
}

//json加解密
func JsonEncode(arg interface{}) ([]byte, error) {
	return json.Marshal(arg)
}

func JsonDecode(arg []byte, out interface{}) error {
	return json.Unmarshal(arg, out)
}

//hash
func MakeHash(key string) uint32 {
	ieee := crc32.NewIEEE()
	io.WriteString(ieee, key)
	return ieee.Sum32()
}

//加载配置表
func LoadCharacterConfigFiles() {

}

//取本周零点时间点
func GetThisSundayTime() int64 {

	//7天
	timeNow := time.Now().Unix()
	weekday := time.Now().Weekday()
	dayAfter := (time.Monday - weekday + 7) % 7
	if 0 == dayAfter {
		dayAfter = 7
	}

	retTime := GetNextDayHourTime(int64(timeNow), int(dayAfter), 0, 0, 0, true)
	return retTime
}

//实现一个深度拷贝函数
func DeepCopy(dst, src interface{}) error {
	var buf bytes.Buffer
	if err := gob.NewEncoder(&buf).Encode(src); err != nil {
		return err
	}
	return gob.NewDecoder(bytes.NewBuffer(buf.Bytes())).Decode(dst)
}

func DaysOfMonth(year int, month int) (days int32) {
	if month != 2 {
		if month == 4 || month == 6 || month == 9 || month == 11 {
			days = 30

		} else {
			days = 31
		}
	} else {
		if ((year%4) == 0 && (year%100) != 0) || (year%400) == 0 {
			days = 29
		} else {
			days = 28
		}
	}
	return
}
