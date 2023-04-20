package main

import (
	"github.com/liserjrqlxue/acmg2015"
	"github.com/liserjrqlxue/goUtil/stringsUtil"
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
func goWriteAvd(excel *excelize.File, sheetName string, runDmd, runAvd chan bool, all bool) {
	log.Println("Write AVD Start")
	var (
		avdArray = loadAvdList()
		runWrite = make(chan bool, 1)
		throttle = make(chan bool, *threshold)
		size     = len(avdArray)
		dbChan   = make(chan []map[string]string, size)
	)

	if size == 0 {
		log.Println("Write AVD Skip")
		fillChan(runAvd)
		return
	}

	log.Println("Start load AVD")

	// wait runDmd done
	emptyChan(runDmd)
	log.Println("Wait DMD Done")
	// go loadAvd -> dbChan -> go writeAvd
	go writeAvd(excel, sheetName, dbChan, size, runWrite)
	for _, fileName := range avdArray {
		go loadAvd(fileName, dbChan, throttle, all)
	}
	emptyChan(runWrite)
	for i := 0; i < *threshold; i++ {
		fillChan(throttle)
	}

	fillChan(runAvd)
	log.Println("Write AVD Done")
}

func writeAvd(excel *excelize.File, sheetName string, dbChan chan []map[string]string, size int, throttle chan<- bool) {
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
		count++
		for _, item := range avd {
			rIdx++
			writeAvdRow(excel, sheetName, rIdx, item, title)
		}
		// stop channel range
		if count == size {
			close(dbChan)
		}
	}
	fillChan(throttle)
}

func writeAvdRow(excel *excelize.File, sheetName string, rIdx int, item map[string]string, title []string) {
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

func loadAvd(fileName string, dbChan chan<- []map[string]string, throttle chan bool, all bool) {
	// block threads
	fillChan(throttle)

	log.Printf("load avd[%s]\n", fileName)

	var (
		data, _  = textUtil.File2MapArray(fileName, "\t", nil)
		sampleID = filepath.Base(fileName)

		allTitle = textUtil.File2Array(*allColumns)

		subFlag = false

		geneHash = make(map[string]string)

		inheritDb = make(map[string]map[string]int)

		filterData []map[string]string
	)

	if len(data) == 0 {
		log.Printf("skip avd of [%s]:[%s] for 0 line", sampleID, fileName)
	}

	if data[0]["SampleID"] != "" {
		sampleID = data[0]["SampleID"]
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
		SampleGeneInfo[sampleID] = geneInfo
	}

	// cycle 1
	for _, item := range data {
		updateAvd(item, subFlag)
		updateFromAvd(item, geneHash, geneInfo, sampleID, subFlag)
		if *cs {
			// 烈性突变
			anno.UpdateSnvTier1(item)

			// 遗传相符
			item["Zygosity"] = anno.ZygosityFormat(item["Zygosity"])
			anno.InheritCheck(item, inheritDb)
		}
	}

	// cycle 2
	for _, item := range data {
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
		writeSampleAvd(allExcelPath, *allSheetName, allTitle, data)
	}

	dbChan <- filterData

	// release threads
	emptyChan(throttle)
}

func updateAvd(item map[string]string, subFlag bool) {
	updateABC(item, item["SampleID"])
	item["HGMDorClinvar"] = "否"
	if isHGMD[item["HGMD Pred"]] || isClinVar[item["ClinVar Significance"]] {
		item["HGMDorClinvar"] = "是"
	}
	item["ClinVar星级"] = item["ClinVar Number of gold stars"]
	item["1000Gp3 AF"] = item["1000G AF"]
	item["1000Gp3 EAS AF"] = item["1000G EAS AF"]
	var gene = item["Gene Symbol"]
	item["gene+cHGVS"] = gene + ":" + item["cHGVS"]
	item["gene+pHGVS3"] = gene + ":" + item["pHGVS3"]
	item["gene+pHGVS1"] = gene + ":" + item["pHGVS1"]
	anno.Score2Pred(item)
	updateLOF(item)
	updateDisease(item)

	if *im {
		item["报告类别"] = "否"
		item["In BGI database"] = "否"
	}

	// reads_picture
	// reads_picture_HyperLink
	readsPicture(item)
	// Function
	anno.UpdateFunction(item)
	// AF -1 -> .
	updateAf(item)

	if *acmg {
		acmg2015.AddEvidences(item)
		item["自动化判断"] = acmg2015.PredACMG2015(item, *autoPVS1)
		anno.UpdateAutoRule(item)
	}
	if *cs {
		floatFormat(item, afFloatFormatArray, 6)
		// remove trailing zeros
		floatFormat(item, afFloatFormatArray, -1)
		var (
			repeatHit     bool
			homologousHit bool
			start         = stringsUtil.Atoi(item["Start"])
			end           = stringsUtil.Atoi(item["Stop"])
			hits          []string
		)
		for _, region := range repeatRegion {
			if region.gene == item["Gene Symbol"] && start > region.start && end < region.end {
				repeatHit = true
			}
		}
		for _, region := range homologousRegion {
			if region.chr == item["#Chr"] && start > region.start && end < region.end {
				homologousHit = true
			}
		}
		if repeatHit {
			hits = append(hits, "重复区域变异")
		}
		if homologousHit {
			hits = append(hits, "同源区域变异")
		}
		item["需验证的变异"] = strings.Join(hits, ";")
		item["#Chr"] = addChr(item["#Chr"])
	} else {
		annoLocaDb(item, localDb, subFlag)
	}
	item["exonCount"] = exonCount[item["Transcript"]]
	item["引物设计"] = anno.PrimerDesign(item)
	item["验证"] = ifCheck(item)
}

// writeSampleAvd read data and write to sheetName of excelName
func writeSampleAvd(excelName, sheetName string, title []string, data []map[string]string) {
	var (
		excel = excelize.NewFile()
		rIdx  = 1
	)
	if *im {
		title = sheetTitle["SNV&INDEL"]
	}
	excel.NewSheet(sheetName)
	writeTitle(excel, sheetName, title)
	for _, item := range data {
		if *im {
			if geneIMListMap[item["Gene Symbol"]] {
				rIdx++
				writeRow(excel, sheetName, item, title, rIdx)
			}
		} else {
			rIdx++
			writeRow(excel, sheetName, item, title, rIdx)
		}
	}
	log.Printf("excel.SaveAs(\"%s\") with %d variants\n", excelName, rIdx-1)
	simpleUtil.CheckErr(excel.SaveAs(excelName))
}
