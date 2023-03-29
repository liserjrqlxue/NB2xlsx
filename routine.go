package main

import (
	"fmt"
	"log"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/liserjrqlxue/acmg2015"
	"github.com/liserjrqlxue/anno2xlsx/v2/anno"
	"github.com/liserjrqlxue/goUtil/simpleUtil"
	"github.com/liserjrqlxue/goUtil/textUtil"
	"github.com/xuri/excelize/v2"
)

func getAvd(fileName string, dbChan chan<- []map[string]string, throttle, writeAll chan bool, all bool) {
	log.Printf("load avd[%s]\n", fileName)
	var (
		avd, _   = textUtil.File2MapArray(fileName, "\t", nil)
		sampleID = filepath.Base(fileName)

		allExcelPath = strings.Join([]string{*prefix, "all", sampleID, "xlsx"}, ".")
		allTitle     = textUtil.File2Array(*allColumns)

		subFlag = false

		geneHash = make(map[string]string)

		inheritDb = make(map[string]map[string]int)

		filterData []map[string]string
	)

	if len(avd) > 0 && avd[0]["SampleID"] != "" {
		sampleID = avd[0]["SampleID"]
		allExcelPath = strings.Join([]string{*prefix, "all", sampleID, "xlsx"}, ".")
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
			filterData = append(filterData, item)
		}
	}

	dbChan <- filterData

	if all {
		writeAll <- true
		goWriteSampleAvd(allExcelPath, *allSheetName, allTitle, avd, writeAll)
	}

	<-throttle
}

// goWriteSampleAvd write data to sheetName of excelName
func goWriteSampleAvd(excelName, sheetName string, title []string, data []map[string]string, done chan bool) {
	var (
		excel = excelize.NewFile()
		rIdx  = 1
	)
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
	<-done
}

// goWriteAvd write AVD sheet to excel
func goWriteAvd(excel *excelize.File, runDmd, runAvd chan bool, all bool) {
	log.Println("Write AVD Start")
	var avdArray []string
	if *avdFiles != "" {
		avdArray = strings.Split(*avdFiles, ",")
	}
	if *avdList != "" {
		avdArray = append(avdArray, textUtil.File2Array(*avdList)...)
	}
	if len(avdArray) > 0 {
		log.Println("Start load AVD")

		// acmg
		if *acmg {
			acmg2015.AutoPVS1 = *autoPVS1
			var acmgCfg = simpleUtil.HandleError(textUtil.File2Map(*acmgDb, "\t", false)).(map[string]string)
			for k, v := range acmgCfg {
				acmgCfg[k] = filepath.Join(dbPath, v)
			}
			acmg2015.Init(acmgCfg)
		}

		// wait runDmd done
		runDmd <- true

		var (
			runWrite   = make(chan bool, 1)
			throttle   = make(chan bool, *threshold)
			writeExcel = make(chan bool, *threshold)
			size       = len(avdArray)
			dbChan     = make(chan []map[string]string, size)
		)

		// goroutine writeAvd in case block getAvd since more avd than threshold
		runWrite <- true
		go writeAvd(excel, dbChan, size, runWrite)
		for _, fileName := range avdArray {
			throttle <- true
			go getAvd(fileName, dbChan, throttle, writeExcel, all)
		}
		// wait writeAvd done
		runWrite <- true
		for i := 0; i < *threshold; i++ {
			throttle <- true
			writeExcel <- true
		}
	} else {
		log.Println("Write AVD Skip")
	}
	log.Println("Write AVD Done")
	<-runAvd
}

func addChr(chr string) string {
	return "Chr" + strings.Replace(
		strings.Replace(chr, "chr", "", 1),
		"Chr", "", 1,
	)
}

func writeAvd(excel *excelize.File, dbChan chan []map[string]string, size int, throttle chan bool) {
	var sheetName = *avdSheetName
	if *im {
		sheetName = "SNV&INDEL"
	}
	var rows = simpleUtil.HandleError(excel.GetRows(sheetName)).([][]string)
	var title = rows[0]
	var rIdx = len(rows)
	var count = 0
	for avd := range dbChan {
		for _, item := range avd {
			rIdx++
			updateINDEX(item, "D", rIdx)
			var sampleID = item["SampleID"]
			item["sampleID"] = sampleID
			if *im {
				updateInfo(item, sampleID)
				updateColumns(item, sheetTitleMap[sheetName])
				writeRow(excel, sheetName, item, sheetTitle[sheetName], rIdx)
			} else {
				if *cs {
					item["LOF"] = ""
					item["disGroup"] = item["PP_disGroup"]
					if top1kGene[item["Gene Symbol"]] {
						item["是否国内（际）包装变异"] = "国内包装基因"
					}
					item["Database"] = ""
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
					if item["报告类别"] == "" || item["报告类别"] == "." {
						item["报告类别"] = "正式报告"
					}
					// style
					item["报告类别-原始"] = item["报告类别"]

					updateInfo(item, sampleID)
				}
				writeRow(excel, sheetName, item, title, rIdx)
			}
		}
		count++
		if count == size {
			close(dbChan)
		}
	}
	<-throttle
}

// LoadDmd load DMD sheet to excel
func LoadDmd(excel *excelize.File, throttle chan bool) {
	log.Println("Load DMD Start")
	var dmdArray []string
	if *dmdFiles != "" {
		dmdArray = strings.Split(*dmdFiles, ",")
	}
	if *dmdList != "" {
		dmdArray = append(dmdArray, textUtil.File2Array(*dmdList)...)
	}
	if len(dmdArray) > 0 {
		loadDmd(excel, dmdArray)
	} else {
		log.Println("Load DMD Skip")
	}
	log.Println("Load DMD Done")
	<-throttle
}

func loadDmd(excel *excelize.File, dmdArray []string) {
	var sheetName = *dmdSheetName
	if *im {
		sheetName = "DMD CNV"
	}
	var rows = simpleUtil.HandleError(excel.GetRows(sheetName)).([][]string)
	//var title = rows[0]
	var rIdx = len(rows)
	for _, fileName := range dmdArray {
		var dmd, _ = textUtil.File2MapArray(fileName, "\t", nil)
		for _, item := range dmd {
			rIdx++
			updateDmd(item)
			var (
				sampleID = item["#Sample"]
				gene     = item["gene"]
				CN       = strings.Split(item["CopyNum"], ";")[0]
				cn       float64
				err      error
			)
			if CN == ">4" {
				cn = 5
				log.Printf("treat CopyNum[%s] as 5\n", item["CopyNum"])
			} else {
				cn, err = strconv.ParseFloat(strings.Split(item["CopyNum"], ";")[0], 64)
				if err != nil {
					cn = 3
					log.Printf("treat CopyNum[%s] as 3:%+v\n", item["CopyNum"], err)
				}
			}
			updateSampleGeneInfo(cn, sampleID, gene)
			addDiseases2Cnv(item, multiDiseaseSep, gene)
			updateINDEX(item, "D", rIdx)
			if *im {
				addDatabase2Cnv(item)
				updateColumns(item, sheetTitleMap[sheetName])
			}
			DmdCnv = append(DmdCnv, item)
			//writeRow(excel, sheetName, item, title, rIdx)
		}
	}
}

func goUpdateCNV(excel *excelize.File, throttle chan bool) {
	if *im {
		*dmdSheetName = "DMD CNV"
	}
	updateData2Sheet(excel, *dmdSheetName, DmdCnv, updateDMDCNV)
	<-throttle
}

// WriteAe write AE sheet to excel
func WriteAe(excel *excelize.File, throttle chan bool) {
	log.Println("Write AE Start")
	var db = make(map[string]map[string]string)
	if *dipinResult != "" {
		var dipin, _ = textUtil.File2MapArray(*dipinResult, "\t", nil)
		for _, item := range dipin {
			updateDipin(item, db)
		}
	}
	if *smaResult != "" {
		var smaXlsx = excelize.NewFile()
		var smaTitle = textUtil.File2Array(filepath.Join(etcPath, "title.sma.txt"))
		var sma, _ = textUtil.File2MapArray(*smaResult, "\t", nil)
		writeTitle(smaXlsx, "Sheet1", smaTitle)
		for i, item := range sma {
			updateSma(item, db)
			writeRow(smaXlsx, "Sheet1", item, smaTitle, i+2)
		}
		simpleUtil.CheckErr(smaXlsx.SaveAs(*prefix + ".SMA_result.xlsx"))
	}
	if *sma2Result != "" {
		var sma, _ = textUtil.File2MapArray(*sma2Result, "\t", nil)
		for _, item := range sma {
			updateSma2(item, db)
		}
	}
	if len(db) > 0 {
		writeAe(excel, db)
	} else {
		log.Println("Write AE Skip")
	}
	log.Println("Write AE Done")
	<-throttle
}

func writeAe(excel *excelize.File, db map[string]map[string]string) {
	if *im {
		*aeSheetName = "THAL CNV"
	}
	var rows = simpleUtil.HandleError(excel.GetRows(*aeSheetName)).([][]string)
	var title = rows[0]
	var rIdx = len(rows)
	for _, item := range db {
		rIdx++
		updateAe(item)
		updateINDEX(item, "D", rIdx)
		var sampleID = item["SampleID"]
		item["sampleID"] = sampleID
		if *im {
			updateInfo(item, sampleID)
			for _, s := range []string{"THAL CNV", "SMN1 CNV"} {
				updateColumns(item, sheetTitleMap[s])
				writeRow(excel, s, item, sheetTitle[s], rIdx)
			}
		} else {
			if *cs {
				item["sex"] = item["Sex"]
				updateInfo(item, sampleID)
			}
			writeRow(excel, *aeSheetName, item, title, rIdx)
		}
	}
}

// WriteQC write QC sheet to excel
func WriteQC(excel *excelize.File, throttle chan bool) {
	log.Println("Write QC Start")
	if *cs {
		var qcMaps, _ = textUtil.File2MapArray(*qc, "\t", nil)
		writeQC(excel, qcMaps)
	} else {
		writeQC(excel, loadQC(*qc))
	}
	log.Println("Write QC Done")
	<-throttle
}

func updateINDEX(item map[string]string, col string, index int) {
	item["解读人"] = fmt.Sprintf("=INDEX(任务单!O:O,MATCH(%s%d,任务单!I:I,0),1)", col, index)
	item["审核人"] = fmt.Sprintf("=INDEX(任务单!P:P,MATCH(%s%d,任务单!I:I,0),1)", col, index)
}

var (
	multiDiseaseSep = "\n"
	//batchCnvDiseaseTitle = []string{
	//	"疾病中文名",
	//	"遗传模式",
	//}
)

func goWriteBatchCnv(throttle chan bool) {
	var sheetName = "Sheet1"
	var bcExcel = simpleUtil.HandleError(excelize.OpenFile(*bcTemplate)).(*excelize.File)

	updateData2Sheet(bcExcel, sheetName, BatchCnv, updateBatchCNV)

	//var lastCellName = simpleUtil.HandleError(excelize.CoordinatesToCellName(len(BatchCnvTitle), len(BatchCnv)+1)).(string)
	//simpleUtil.CheckErr(bcExcel.AddTable(sheetName, "A1", lastCellName, `{"table_style":"TableStyleMedium9"}`), "bcExcel.AddTable Error!")

	simpleUtil.CheckErr(bcExcel.SaveAs(*prefix+".batchCNV.xlsx"), "bcExcel.SaveAs Error!")

	<-throttle
}

func saveMainExcel(excel *excelize.File, path string, throttle chan bool) {
	log.Printf("excel.SaveAs(\"%s\")\n", path)
	simpleUtil.CheckErr(excel.SaveAs(path))
	log.Println("Save main Done")
	<-throttle
}
