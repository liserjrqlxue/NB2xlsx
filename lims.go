package main

import (
	"log"
	"path/filepath"

	"github.com/liserjrqlxue/goUtil/textUtil"
)

var limsHeader = filepath.Join(etcPath, "lims.info.header")
var limsTitle = textUtil.File2Array(limsHeader)

func loadLimsInfo(path string) map[string]map[string]string {
	var db = make(map[string]map[string]string)
	if path == "" {
		log.Println("skip lims.info for absence")
		return db
	}

	for _, line := range textUtil.File2Slice(path, "\t") {
		var item = make(map[string]string)
		for i := range line {
			item[limsTitle[i]] = line[i]
		}
		db[item["MAIN_SAMPLE_NUM"]] = item
	}
	return db
}

func updateABC(item map[string]string, sampleID string) {
	//var sampleID = item["SampleID"]
	if *gender == "M" || genderMap[sampleID] == "M" {
		item["Sex"] = "M"
	} else if *gender == "F" || genderMap[sampleID] == "F" {
		item["Sex"] = "F"
	}
	var db = limsInfo[sampleID]
	item["期数"] = db["HYBRID_LIBRARY_NUM"]
	item["flow ID"] = db["FLOW_ID"]
	item["产品编码_产品名称"] = db["PRODUCT_CODE"] + "_" + db["PRODUCT_NAME"]
}
