package main

import (
	"path/filepath"

	"github.com/liserjrqlxue/goUtil/textUtil"
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
