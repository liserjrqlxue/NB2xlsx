package main

import (
	"github.com/liserjrqlxue/goUtil/simpleUtil"
	"github.com/liserjrqlxue/goUtil/textUtil"
	"github.com/xuri/excelize/v2"
	"log"
	"strings"
)

func loadAvdList() (avdArray []string) {
	if *avdFiles != "" {
		avdArray = strings.Split(*avdFiles, ",")
	}
	if *avdList != "" {
		avdArray = append(avdArray, textUtil.File2Array(*avdList)...)
	}
	return avdArray
}

// goWriteAvd write AVD sheet to excel
func goWriteAvd(excel *excelize.File, runDmd, runAvd chan bool, all bool) {
	log.Println("Write AVD Start")
	var (
		avdArray   = loadAvdList()
		runWrite   = make(chan bool, 1)
		throttle   = make(chan bool, *threshold)
		writeExcel = make(chan bool, *threshold)
		size       = len(avdArray)
		dbChan     = make(chan []map[string]string, size)
	)

	if size == 0 {
		log.Println("Write AVD Skip")
		<-runAvd
		return
	}

	log.Println("Start load AVD")

	// wait runDmd done
	// goroutine writeAvd in case block getAvd since more avd than threshold
	wait(runDmd, runWrite)
	go writeAvd(excel, dbChan, size, runWrite)

	for _, fileName := range avdArray {
		throttle <- true
		go getAvd(fileName, dbChan, throttle, writeExcel, all)
	}
	wait(runWrite)

	for i := 0; i < *threshold; i++ {
		wait(throttle, writeExcel)
	}
	log.Println("Write AVD Done")
	<-runAvd
}

func writeAvd(excel *excelize.File, dbChan chan []map[string]string, size int, throttle chan bool) {
	var sheetName = *avdSheetName
	if *im {
		sheetName = "SNV&INDEL"
	}
	var (
		rows  = simpleUtil.HandleError(excel.GetRows(sheetName)).([][]string)
		title = rows[0]
		rIdx  = len(rows)
		count = 0
	)
	if *im {
		title = sheetTitle[sheetName]
	}

	for avd := range dbChan {
		for _, item := range avd {
			rIdx++
			updateINDEX(item, "D", rIdx)
			var sampleID = item["SampleID"]
			item["sampleID"] = sampleID
			if *wgs {
				updateInfo(item, sampleID)
			} else if *im {
				updateInfo(item, sampleID)
				updateColumns(item, sheetTitleMap[sheetName])
			} else if *cs {
				updateInfo(item, sampleID)

				item["LOF"] = ""
				item["disGroup"] = item["PP_disGroup"]
				if top1kGene[item["Gene Symbol"]] {
					item["是否国内（际）包装变异"] = "国内包装基因"
				}
				item["Database"] = "."
				switch item["Auto ACMG + Check"] {
				case "P":
					item["Auto ACMG + Check"] = "Pathogenic"
					item["Database"] = "DX1605"
					item["是否是库内位点"] = "是"
				case "LP":
					item["Auto ACMG + Check"] = "Likely pathogenic"
					item["Database"] = "DX1605"
					item["是否是库内位点"] = "是"
				case "", ".":
					item["Auto ACMG + Check"] = "待解读"
				}
				if item["Auto ACMG + Check"] == "" || item["Auto ACMG + Check"] == "." {
					item["Auto ACMG + Check"] = "待解读"
				}
				item["突变类型"] = item["Auto ACMG + Check"]
				item["报告类别"] = "正式报告"
				// style
				item["报告类别-原始"] = item["报告类别"]
				item["遗传模式"] = strings.Replace(item["遗传模式"], "[n]", ",", -1)
			} else {
				updateABC(item, sampleID)
			}
			writeRow(excel, sheetName, item, title, rIdx)
		}
		count++
		if count == size {
			close(dbChan)
		}
	}
	<-throttle
}
