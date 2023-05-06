package main

import (
	"log"

	"github.com/liserjrqlxue/goUtil/textUtil"
)

func loadSamplesInfo(lims, detail, info string) (limsDb, detailDb, infoDb map[string]map[string]string) {
	limsDb = make(map[string]map[string]string)
	detailDb = make(map[string]map[string]string)
	infoDb = make(map[string]map[string]string)

	if lims == "" {
		log.Println("skip lims.info for absence")
	} else {
		var limsTitle = textUtil.File2Array(limsHeader)
		for _, line := range textUtil.File2Slice(lims, "\t") {
			var item = make(map[string]string)
			for i := range line {
				item[limsTitle[i]] = line[i]
			}
			limsDb[item["MAIN_SAMPLE_NUM"]] = item
		}

	}

	if detail == "" {
		log.Println("skip detail for absence")
	} else {
		for _, line := range textUtil.File2Slice(detail, "\t") {
			var db = make(map[string]string)
			var sampleID = line[0]
			db["productCode"] = line[1]
			db["hospital"] = line[2]
			detailDb[sampleID] = db
		}
	}

	if info == "" {
		log.Println("skip info.txt for absence")
	} else {
		infoDb, _ = textUtil.File2MapMap(info, "sampleID", "\t", nil)
	}

	return
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
