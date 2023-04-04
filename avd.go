package main

import (
	"log"
	"path/filepath"
	"strings"

	"github.com/liserjrqlxue/anno2xlsx/v2/anno"
	"github.com/liserjrqlxue/goUtil/simpleUtil"
	"github.com/liserjrqlxue/goUtil/textUtil"
	"github.com/xuri/excelize/v2"
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
	wait(runDmd)
	go writeAvd(excel, dbChan, size, runWrite)

	for _, fileName := range avdArray {
		go getAvd(fileName, dbChan, throttle, writeExcel, all)
	}
	for i := 0; i < *threshold; i++ {
		wait(throttle, writeExcel)
	}
	waitWrite(runWrite)
	log.Println("Write AVD Done")
	<-runAvd
}

func writeAvd(excel *excelize.File, dbChan chan []map[string]string, size int, throttle chan<- bool) {
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
	throttle <- true
}

func getAvd(fileName string, dbChan chan<- []map[string]string, throttle, writeAll chan bool, all bool) {
	// block threads
	throttle <- true

	log.Printf("load avd[%s]\n", fileName)

	var (
		avd, _   = textUtil.File2MapArray(fileName, "\t", nil)
		sampleID = filepath.Base(fileName)

		allTitle = textUtil.File2Array(*allColumns)

		subFlag = false

		geneHash = make(map[string]string)

		inheritDb = make(map[string]map[string]int)

		filterData []map[string]string
	)

	if len(avd) > 0 && avd[0]["SampleID"] != "" {
		sampleID = avd[0]["SampleID"]
	}
	var allExcelPath = strings.Join([]string{*prefix, "all", sampleID, "xlsx"}, ".")
	if *cs {
		allExcelPath = filepath.Join(*annoDir, sampleID+"_vcfanno.xlsx")
		allTitle = textUtil.File2Array(filepath.Join(templatePath, "vcfanno.txt"))
	}

	var details, ok1 = sampleDetail[sampleID]
	if ok1 && details["productCode"] == "DX1968" && details["hospital"] == "南京市妇幼保健院" {
		subFlag = true
	}

	var geneInfo, ok = SampleGeneInfo[sampleID]
	if !ok {
		geneInfo = make(map[string]*GeneInfo)
	}

	// cycle 1
	for _, item := range avd {
		updateAvd(item, subFlag)
		updateFromAvd(item, geneHash, geneInfo, sampleID)
		if *cs {
			// 烈性突变
			anno.UpdateSnvTier1(item)

			// 遗传相符
			item["Zygosity"] = anno.ZygosityFormat(item["Zygosity"])
			anno.InheritCheck(item, inheritDb)
		}
	}

	// cycle 2
	for _, item := range avd {
		if *cs {
			item["遗传相符"] = anno.InheritCoincide(item, inheritDb, false)
			filterData = append(filterData, item)
		} else if item["filterAvd"] == "Y" {
			var info, ok = geneInfo[item["Gene Symbol"]]
			if !ok {
				log.Fatalf("geneInfo build error:\t%+v\n", geneInfo)
			} else {
				if !geneExcludeListMap[item["Gene Symbol"]] {
					item["Database"] = info.getTag(item)
				}
			}
			item["遗传模式判读"] = geneHash[item["Gene Symbol"]]
			if subFlag && !deafnessGeneList[item["Gene Symbol"]] && item["遗传模式判读"] == "携带者" && item["报告类别-原始"] == "正式报告" {
				item["报告类别-原始"] = "补充报告"
			}
			filterData = append(filterData, item)
		}
	}

	if all {
		wait(writeAll)
		goWriteSampleAvd(allExcelPath, *allSheetName, allTitle, avd, writeAll)
	}

	dbChan <- filterData

	// release threads
	<-throttle
}
