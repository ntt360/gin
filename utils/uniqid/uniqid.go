package uniqid

import (
	"math"
	"math/rand"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/bwmarrin/snowflake"
	"github.com/rs/xid"
)

var (
	snow     *snowflake.Node
	snowNode = os.Getpid() % 1024
	snowOnce sync.Once
)

// 10进制和62进制转换字典
const dict = "6Qw7fWXgTjkcordIKLvBstRSD3n904U5e8ZMOPluhJmiNxVyzY12AFGHabpCEq"

// From10To62 基于自增 10 进制转为62进制
func From10To62(num int64) string {
	//dict := "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	var str62 []byte
	for {
		var result byte
		var tmp []byte

		number := num % 62
		result = dict[number]

		// 临时变量，为了追加到头部
		tmp = append(tmp, result)

		str62 = append(tmp, str62...)
		num = num / 62

		if num == 0 {
			break
		}
	}
	return string(str62)
}

// From62To10 62 bit to 10 bit
func From62To10(str62 string) int64 {
	var pos int
	var number int64
	l := len(str62)

	for i := 0; i < l; i++ {
		pos = strings.IndexAny(dict, str62[i:i+1])
		number = int64(math.Pow(62, float64(l-i-1))*float64(pos)) + number
	}
	return number
}

// Snowflake snowflake global init
func Snowflake() (snowflake.ID, error) {
	if snow != nil {
		return snow.Generate(), nil
	}

	var err error
	snowOnce.Do(func() {
		snow, err = snowflake.NewNode(int64(snowNode))
	})
	if err != nil {
		return 0, err
	}

	return snow.Generate(), nil
}

// Snowflake62 算法，转换成62进制字符串，减少存储长度
func Snowflake62() (string, error) {
	if snow != nil {
		return From10To62(snow.Generate().Int64()), nil
	}

	var err error
	snowOnce.Do(func() {
		snow, err = snowflake.NewNode(int64(snowNode))
	})
	if err != nil {
		return "", err
	}

	// Generate a snowflake ID.
	return From10To62(snow.Generate().Int64()), nil
}

// Shuffle the string
func Shuffle(origin string) string {
	arr := []byte(origin)
	rand.Seed(time.Now().UnixNano())
	rand.Shuffle(len(arr), func(i, j int) {
		arr[i], arr[j] = arr[j], arr[i]
	})

	return string(arr)
}

// LogID 根据xid库生成logId
func LogID(prefix string) string {
	guid := xid.New()
	if len(prefix) != 0 {
		return prefix + guid.String()
	}
	return guid.String()
}
