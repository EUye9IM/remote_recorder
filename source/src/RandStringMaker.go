package main

import (
	"math/rand"
	"time"
)

func init() {
	rand.Seed(time.Now().UnixNano() + int64(rand.Intn(100)))
}

type RandStringMaker struct {
	set_len  int
	str_len  int
	byte_set []byte
}

func (rsm *RandStringMaker) Set(str string, strlen int) {
	rsm.byte_set = []byte(str)
	rsm.set_len = len(rsm.byte_set)
	rsm.str_len = strlen
}
func (rsm RandStringMaker) Get() string {
	res := []byte{}
	for i := 0; i < rsm.str_len; i++ {
		res = append(res, rsm.byte_set[rand.Intn(rsm.set_len)])
	}
	return string(res)
}
