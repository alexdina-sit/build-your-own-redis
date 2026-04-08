package main

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

func readArray(respArray string) []string {
	splitedString := strings.Split(respArray, "\r\n")
	var returnArray []string

	for _, elem := range splitedString[2:] {
		if elem == "" || (len(elem) > 1 && elem[0] == 36) {
			continue
		}
		returnArray = append(returnArray, elem)
	}
	return returnArray
}

func handleSet(arr []string) string {
	var expireType string
	var expireTime int

	if len(arr) > 4 {
		expireType = strings.ToUpper(arr[3])

		intValue, err := strconv.Atoi(arr[4])
		if err != nil {
			fmt.Println("Error while converting your expire value", err)
		}
		expireTime = intValue
	}

	mu.Lock()
	mmap[arr[1]] = &Item{
		Value:      arr[2],
		ExpireType: expireType,
		ExpireTime: expireTime,
		CreateDate: time.Now(),
	}
	mu.Unlock()

	return "+OK\r\n"
}

func handleGet(arr []string) string {
	mu.RLock()
	item, prs := mmap[arr[1]]
	mu.RUnlock()

	if !prs {
		return "$-1\r\n"
	}
	expireType := strings.ToUpper(item.ExpireType)
	timeSpent := time.Since(item.CreateDate)

	if (expireType == "EX" && timeSpent.Seconds() > float64(item.ExpireTime)) ||
		(expireType == "PX" && timeSpent.Milliseconds() > int64(item.ExpireTime)) {

		mu.Lock()
		delete(mmap, arr[1])
		mu.Unlock()

		return "$-1\r\n"
	}

	return fmt.Sprintf("$%d\r\n%s\r\n", len(item.Value), item.Value)
}
