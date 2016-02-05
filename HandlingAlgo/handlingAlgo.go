package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/DuoSoftware/gorest"
	"github.com/fzzy/radix/redis"
	"io/ioutil"
	"net/http"
	"strings"
	"time"
)

const layout = "2006-01-02T15:04:05Z07:00"

type HandlingAlgo struct {
	gorest.RestService `root:"/HandlingAlgo/" consumes:"application/json" produces:"application/json"`
	singleResource     gorest.EndPoint `method:"GET" path:"/Single/{ReqClass:string}/{ReqType:string}/{ReqCategory:string}/{SessionId:string}/{ResourceIds:string}" output:"string"`
}

func (handlingAlgo HandlingAlgo) SingleResource(ReqClass, ReqType, ReqCategory, SessionId, ResourceIds string) string {
	ch := make(chan string)
	fmt.Println(ResourceIds)
	byt := []byte(ResourceIds)
	var resourceIds []string
	json.Unmarshal(byt, &resourceIds)
	go SingleHandling(ReqClass, ReqType, ReqCategory, SessionId, resourceIds, ch)
	var result = <-ch
	close(ch)
	return result

}

func ReserveSlot(slotInfo CSlotInfo) bool {
	url := fmt.Sprintf("http://%s/DVP/API/1.0.0.0/ARDS/resource/%s/concurrencyslot", CreateHost(ardsServiceHost, ardsServicePort), slotInfo.ResourceId)
	fmt.Println("URL:>", url)

	slotInfoJson, _ := json.Marshal(slotInfo)
	var jsonStr = []byte(slotInfoJson)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonStr))
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		//panic(err)
		return false
	}
	defer resp.Body.Close()

	fmt.Println("response Status:", resp.Status)
	fmt.Println("response Headers:", resp.Header)
	body, _ := ioutil.ReadAll(resp.Body)
	result := string(body)
	fmt.Println("response Body:", result)
	if result == "OK" {
		fmt.Println("Return true")
		return true
	}

	fmt.Println("Return false")
	return false
}

func ClearSlotOnMaxRecerved(reqClass, reqType, reqCategory string, resObj Resource, metaData ReqMetaData) {
	client, err := redis.Dial("tcp", redisIp)
	errHndlr(err)
	defer client.Close()

	// select database
	r := client.Cmd("select", redisDb)
	errHndlr(r.Err)
	var tagArray = make([]string, 8)

	tagArray[0] = fmt.Sprintf("company_%d", resObj.Company)
	tagArray[1] = fmt.Sprintf("tenant_%d", resObj.Tenant)
	tagArray[2] = fmt.Sprintf("class_%s", reqClass)
	tagArray[3] = fmt.Sprintf("type_%s", reqType)
	tagArray[4] = fmt.Sprintf("category_%s", reqCategory)
	tagArray[5] = fmt.Sprintf("state_%s", "Reserved")
	tagArray[6] = fmt.Sprintf("resourceid_%s", resObj.ResourceId)
	tagArray[7] = fmt.Sprintf("objtype_%s", "CSlotInfo")

	tags := fmt.Sprintf("tag:*%s*", strings.Join(tagArray, "*"))
	fmt.Println(tags)
	reservedSlots, _ := client.Cmd("keys", tags).List()

	for _, tagKey := range reservedSlots {
		strslotKey, _ := client.Cmd("get", tagKey).Str()
		fmt.Println(strslotKey)

		strslotObj, _ := client.Cmd("get", strslotKey).Str()
		fmt.Println(strslotObj)

		var slotObj CSlotInfo
		json.Unmarshal([]byte(strslotObj), &slotObj)

		fmt.Println("Datetime Info" + slotObj.LastReservedTime)
		t, _ := time.Parse(layout, slotObj.LastReservedTime)
		t1 := int(time.Now().Sub(t).Seconds())
		t2 := metaData.MaxReservedTime
		fmt.Println(fmt.Sprintf("Time Info T1: %d", t1))
		fmt.Println(fmt.Sprintf("Time Info T2: %d", t2))
		if t1 > t2 {
			slotObj.State = "Available"
			slotObj.OtherInfo = "ClearReserved"

			ReserveSlot(slotObj)
		}
	}
}

func GetReqMetaData(_company, _tenent int, _class, _type, _category string) ReqMetaData {
	client, err := redis.Dial("tcp", redisIp)
	errHndlr(err)
	defer client.Close()

	// select database
	r := client.Cmd("select", redisDb)
	errHndlr(r.Err)
	key := fmt.Sprintf("ReqMETA:%d:%d:%s:%s:%s", _company, _tenent, _class, _type, _category)
	fmt.Println(key)
	strMetaObj, _ := client.Cmd("get", key).Str()
	fmt.Println(strMetaObj)

	var metaObj ReqMetaData
	json.Unmarshal([]byte(strMetaObj), &metaObj)

	return metaObj
}

func GetConcurrencyInfo(_company, _tenant int, _resId, _class, _type, _category string) ConcurrencyInfo {
	client, err := redis.Dial("tcp", redisIp)
	errHndlr(err)
	defer client.Close()

	// select database
	r := client.Cmd("select", redisDb)
	errHndlr(r.Err)
	key := fmt.Sprintf("ConcurrencyInfo:%d:%d:%s:%s:%s:%s", _company, _tenant, _resId, _class, _type, _category)
	fmt.Println(key)
	strCiObj, _ := client.Cmd("get", key).Str()
	fmt.Println(strCiObj)

	var ciObj ConcurrencyInfo
	json.Unmarshal([]byte(strCiObj), &ciObj)

	return ciObj
}

func GetResourceState(_company, _tenant int, _resId string) string {
	client, err := redis.Dial("tcp", redisIp)
	errHndlr(err)
	defer client.Close()

	// select database
	r := client.Cmd("select", redisDb)
	errHndlr(r.Err)
	key := fmt.Sprintf("ResourceState:%d:%d:%s", _company, _tenant, _resId)
	fmt.Println(key)
	strResStateObj, _ := client.Cmd("get", key).Str()
	fmt.Println(strResStateObj)

	return strResStateObj
}

func AppendIfMissing(dataList []string, i string) []string {
	for _, ele := range dataList {
		if ele == i {
			return dataList
		}
	}
	return append(dataList, i)
}
