package main

import (
	"path/filepath"

	"github.com/liserjrqlxue/goUtil/textUtil"
	"golang.org/x/text/encoding/simplifiedchinese"
)

var limsHeader = filepath.Join(etcPath, "lims.info.header")
var limsTitle = textUtil.File2Array(limsHeader)

func loadLimsInfo() map[string]map[string]string {
	if *lims != "" {
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
	return nil
}

func updateABC(item map[string]string) {
	var info = limsInfo[item["SampleID"]]
	item["期数"] = info["HYBRID_LIBRARY_NUM"]
	item["flow ID"] = info["FLOW_ID"]
	var productName, err = simplifiedchinese.GB18030.NewDecoder().String(info["PRODUCT_NAME"])
	if err == nil {
		item["产品编码_产品名称"] = info["PRODUCT_CODE"] + "_" + productName
	} else {
		item["产品编码_产品名称"] = info["PRODUCT_CODE"] + "_" + info["PRODUCT_NAME"]
	}
}
