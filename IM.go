package main

import (
	"github.com/liserjrqlxue/goUtil/textUtil"
	"github.com/xuri/excelize/v2"
	"log"
	"path/filepath"
)

func parseProductCode() {
	var productMap, _ = textUtil.File2MapMap(filepath.Join(etcPath, "product.txt"), "productCode", "\t", nil)
	var typeMode = make(map[string]bool)
	for _, m := range imInfo {
		typeMode[productMap[m["ProductID"]]["productType"]] = true
	}
	if typeMode["CN"] && typeMode["EN"] {
		log.Fatalln("Conflict for CN or EN!")
	} else if typeMode["CN"] {
		i18n = "CN"
		columnName = "字段-一体机中文"
	} else if typeMode["EN"] {
		i18n = "EN"
		columnName = "字段-一体机英文"
	} else {
		log.Fatalln("No CN or EN!")
	}
}

func updateSheetTitleMap() {
	for _, s := range imSheetList {
		var titleMaps, _ = textUtil.File2MapArray(filepath.Join(templatePath, s+".txt"), "\t", nil)
		var titleMap = make(map[string]string)
		var title []string
		for _, m := range titleMaps {
			title = append(title, m[columnName])
			titleMap[m["Raw"]] = m[columnName]
		}
		sheetTitle[s] = title
		sheetTitleMap[s] = titleMap
	}
}

func initExcel() *excelize.File {
	var excel = excelize.NewFile()
	for _, s := range imSheetList {
		excel.NewSheet(s)
		var titleMaps, _ = textUtil.File2MapArray(filepath.Join(templatePath, s+".txt"), "\t", nil)
		var title []string
		for _, m := range titleMaps {
			title = append(title, m[columnName])
		}
		writeTitle(excel, s, title)
	}
	excel.DeleteSheet("Sheet1")
	return excel
}
