package main

import (
	"path/filepath"

	"github.com/liserjrqlxue/goUtil/textUtil"
)

var limsHeader = filepath.Join(etcPath, "lims.info.header")
var limsTitle = textUtil.File2Array(limsHeader)

func loadLimsInfo() map[string]map[string]string {
	var limsSlice = textUtil.File2Slice(*lims, "\t")
	var db = make(map[string]map[string]string)
	for _, info := range limsSlice {
		var item = make(map[string]string)
		for i := range info {
			item[limsTitle[i]] = info[i]
		}
		db[item["MAIN_SAMPLE_NUM"]] = item
	}
	return db
}

func updateABC(item map[string]string) {
	var sampleID = item["SampleID"]
	if *gender == "M" || genderMap[sampleID] == "M" {
		item["Sex"] = "M"
	} else if *gender == "F" || genderMap[sampleID] == "F" {
		item["Sex"] = "F"
	}
	var info = limsInfo[sampleID]
	item["期数"] = info["HYBRID_LIBRARY_NUM"]
	item["flow ID"] = info["FLOW_ID"]
	item["产品编码_产品名称"] = info["PRODUCT_CODE"] + "_" + info["PRODUCT_NAME"]
}
