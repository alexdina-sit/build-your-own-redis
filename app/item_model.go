package main

import "time"

type Item struct {
	Value      string
	ExpireType string
	ExpireTime int
	CreateDate time.Time
}
