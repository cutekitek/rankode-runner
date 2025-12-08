package utils

import "strconv"

func MustParseInt(s string) int {
	val, err := strconv.Atoi(s)
	if err != nil {
		panic(err)
	}
	return val
}

func MustParseFloat64(s string) float64 {
	val, er := strconv.ParseFloat(s, 64)
	if er != nil {
		panic(er)
	}
	return val
}
