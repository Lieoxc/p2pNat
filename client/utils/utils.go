package utils

import (
	"crypto/rand"
	"math"
	"math/big"
)

// RandPort 生成区间范围内的随机端口
func RandPort(min, max int64) int {
	if min > max {
		panic("the min is greater than max!")
	}
	if min < 0 {
		f64Min := math.Abs(float64(min))
		i64Min := int64(f64Min)
		result, _ := rand.Int(rand.Reader, big.NewInt(max+1+i64Min))
		return int(result.Int64() - i64Min)
	}
	result, _ := rand.Int(rand.Reader, big.NewInt(max-min+1))
	return int(min + result.Int64())
}
