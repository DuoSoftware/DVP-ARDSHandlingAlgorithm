// HandlingAlgo project main.go
package main

import (
	"code.google.com/p/gorest"
	"fmt"
	"net/http"
)

type Configuration struct {
	RedisIp          string
	RedisDb          int
	ResourceCSlotUrl string
	RedisPort        string
	Port             string
}

type EnvConfiguration struct {
	RedisIp          string
	RedisDb          string
	ResourceCSlotUrl string
	RedisPort        string
	Port             string
}

type AttributeData struct {
	Attribute  string
	Class      string
	Type       string
	Category   string
	Precentage float64
}

type Resource struct {
	Company               int
	Tenant                int
	Class                 string
	Type                  string
	Category              string
	ResourceId            string
	ResourceAttributeInfo []AttributeData
	OtherInfo             string
}

type CSlotInfo struct {
	Company          int
	Tenant           int
	Class            string
	Type             string
	Category         string
	State            string
	HandlingRequest  string
	ResourceId       string
	SlotId           int
	ObjKey           string
	SessionId        string
	LastReservedTime string
	OtherInfo        string
}

type ReqMetaData struct {
	MaxReservedTime int
	MaxRejectCount  int
}

type ConcurrencyInfo struct {
	RejectCount       int
	LastConnectedTime string
}

func main() {
	fmt.Println("Initializting Main")
	InitiateRedis()
	gorest.RegisterService(new(HandlingAlgo))
	http.Handle("/", gorest.Handle())
	addr := fmt.Sprintf(":%s", port)
	http.ListenAndServe(addr, nil)
}
