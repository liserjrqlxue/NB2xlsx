package main

import (
	"fmt"
	"log"
	"path/filepath"
	"strings"

	"github.com/360EntSecGroup-Skylar/excelize/v2"
	"github.com/liserjrqlxue/goUtil/simpleUtil"
	"github.com/liserjrqlxue/goUtil/textUtil"
)

func getAvd(fileName string, dbChan chan<- []map[string]string, throttle, writeExcel chan bool) {
	log.Printf("load avd[%s]\n", fileName)
	var avd, _ = textUtil.File2MapArray(fileName, "\t", nil)
	var sampleID = filepath.Base(fileName)
	if len(avd) == 0 {
		writeExcel <- true
		go func() {
			// all snv
			var allExcel = excelize.NewFile()
			allExcel.NewSheet(*allSheetName)
			var allTitle = textUtil.File2Array(*allColumns)
			writeTitle(allExcel, *allSheetName, allTitle)
			log.Printf("excel.SaveAs(\"%s\")\n", strings.Join([]string{*prefix, "all", sampleID, "xlsx"}, "."))
			simpleUtil.CheckErr(allExcel.SaveAs(strings.Join([]string{*prefix, "all", sampleID, "xlsx"}, ".")))
			<-writeExcel
		}()
		dbChan <- avd
		<-throttle
		return
	}
	if avd[0]["SampleID"] != "" {
		sampleID = avd[0]["SampleID"]
	}

	var geneHash = make(map[string]string)
	var geneInfo, ok = SampleGeneInfo[sampleID]
	if !ok {
		geneInfo = make(map[string]*GeneInfo)
	}
	for _, item := range avd {
		updateAvd(item)
		if item["filterAvd"] == "Y" {
			var info, ok = geneInfo[item["Gene Symbol"]]
			if !ok {
				info = new(GeneInfo).new(item)
				geneInfo[item["Gene Symbol"]] = info
			} else {
				info.count(item)
			}
			if *gender == "M" || genderMap[sampleID] == "M" {
				updateGeneHash(geneHash, item, "M")
			} else if *gender == "F" || genderMap[sampleID] == "F" {
				updateGeneHash(geneHash, item, "F")
			}
		}
	}

	var filterAvd []map[string]string
	for _, item := range avd {
		if item["filterAvd"] == "Y" {
			var info, ok = geneInfo[item["Gene Symbol"]]
			if !ok {
				log.Fatalf("geneInfo build error:\t%+v\n", geneInfo)
			} else {
				item["Database"] = info.getTag(item)
			}
			item["遗传模式判读"] = geneHash[item["Gene Symbol"]]
			filterAvd = append(filterAvd, item)
		}
	}
	writeExcel <- true
	go func() {
		// all snv
		var allExcel = excelize.NewFile()
		allExcel.NewSheet(*allSheetName)
		var allTitle = textUtil.File2Array(*allColumns)
		writeTitle(allExcel, *allSheetName, allTitle)
		var rIdx0 = 1
		for _, item := range avd {
			rIdx0++
			writeRow(allExcel, *allSheetName, item, allTitle, rIdx0)
		}
		log.Printf("excel.SaveAs(\"%s\")\n", strings.Join([]string{*prefix, "all", sampleID, "xlsx"}, "."))
		simpleUtil.CheckErr(allExcel.SaveAs(strings.Join([]string{*prefix, "all", sampleID, "xlsx"}, ".")))
		<-writeExcel
	}()
	dbChan <- filterAvd
	<-throttle
}

func writeAvd(excel *excelize.File, dbChan chan []map[string]string, num int, throttle chan bool) {
	log.Println("Write AVD Start")
	var sheetName = *avdSheetName
	var rows = simpleUtil.HandleError(excel.GetRows(sheetName)).([][]string)
	var title = rows[0]
	var rIdx = len(rows)
	var count = 0
	for avd := range dbChan {
		for _, item := range avd {
			rIdx++
			updateINDEX(item, rIdx)
			writeRow(excel, sheetName, item, title, rIdx)
		}
		count++
		if count == num {
			close(dbChan)
		}
	}
	log.Println("Write AVD Done")
	<-throttle
}

func writeDmd(excel *excelize.File, dmdArray []string, throttle chan bool) {
	log.Println("Write DMD Start")
	var sheetName = *dmdSheetName
	var rows = simpleUtil.HandleError(excel.GetRows(sheetName)).([][]string)
	var title = rows[0]
	var rIdx = len(rows)
	for _, fileName := range dmdArray {
		var dmd, _ = textUtil.File2MapArray(fileName, "\t", nil)
		for _, item := range dmd {
			rIdx++
			updateDmd(item)
			updateINDEX(item, rIdx)
			writeRow(excel, sheetName, item, title, rIdx)
		}
	}
	log.Println("Write DMD Done")
	<-throttle
}

func writeAe(excel *excelize.File, db map[string]map[string]string, throttle chan bool) {
	log.Println("Write AE Start")
	var rows = simpleUtil.HandleError(excel.GetRows(*aeSheetName)).([][]string)
	var title = rows[0]
	var rIdx = len(rows)
	for _, item := range db {
		rIdx++
		updateAe(item)
		updateINDEX(item, rIdx)
		writeRow(excel, *aeSheetName, item, title, rIdx)
	}
	log.Println("Write AE Done")
	<-throttle
}

func updateINDEX(item map[string]string, rIdx int) {
	item["解读人"] = fmt.Sprintf("=INDEX('任务单（空sheet）'!O:O,MATCH(D%d&MID($C%d,1,6),'任务单（空sheet）'!$R:$R,0),1)", rIdx, rIdx)
	item["审核人"] = fmt.Sprintf("=INDEX('任务单（空sheet）'!P:P,MATCH(D%d&MID($C%d,1,6),'任务单（空sheet）'!$R:$R,0),1)", rIdx, rIdx)
}
