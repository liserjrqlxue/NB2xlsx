package main

import (
	"flag"
	"github.com/liserjrqlxue/goUtil/osUtil"
	"github.com/liserjrqlxue/goUtil/simpleUtil"
	"github.com/liserjrqlxue/goUtil/textUtil"
	"github.com/liserjrqlxue/version"
	"github.com/xuri/excelize/v2"
	"log"
	"os"
	"path/filepath"
)

func init() {
	version.LogVersion()
	// flag
	flag.Parse()
	if *prefix == "" {
		if *batch != "" {
			*prefix = *batch
		} else {
			flag.Usage()
			log.Println("-prefix are required!")
			os.Exit(1)
		}
	}
}

func main() {
	if osUtil.FileExists(*gender) {
		log.Printf("load gender map from %s", *gender)
		genderMap = simpleUtil.HandleError(textUtil.File2Map(*gender, "\t", false)).(map[string]string)
	}

	var (
		localDb      = make(chan bool, 1)
		runAe        = make(chan bool, 1)
		runAvd       = make(chan bool, 1)
		loadDmd      = make(chan bool, 1)
		writeDmd     = make(chan bool, 1)
		runQC        = make(chan bool, 1)
		saveMain     = make(chan bool, 1)
		saveBatchCnv = make(chan bool, 1)
		excel        *excelize.File
	)

	// load local db
	{
		localDb <- true
		loadLocalDb(localDb)
	}

	loadDb()

	if *lims != "" {
		limsInfo = loadLimsInfo()
	}

	if *batchCNV != "" {
		loadBatchCNV(*batchCNV)
	}

	if *info != "" {
		imInfo, _ = textUtil.File2MapMap(*info, "sampleID", "\t", nil)
	}

	if *im {
		var productMap, _ = textUtil.File2MapMap(filepath.Join(etcPath, "product.txt"), "productCode", "\t", nil)
		var typeMode = make(map[string]bool)
		for _, m := range imInfo {
			typeMode[productMap[m["ProductID"]]["productType"]] = true
		}
		if typeMode["CN"] && typeMode["EN"] {
			log.Fatalln("Conflict for CN or EN!")
		} else if typeMode["CN"] {
			columnName = "字段-一体机中文"
		} else if typeMode["EN"] {
			columnName = "字段-一体机英文"
		} else {
			log.Fatalln("No CN or EN!")
		}

		var imSheetList = []string{
			"Sample",
			"QC",
			"SNV&INDEL",
			"DMD CNV",
			"THAL CNV",
			"SMN1 CNV",
		}

		excel = excelize.NewFile()
		for _, s := range imSheetList {
			excel.NewSheet(s)
			var titleMaps, _ = textUtil.File2MapArray(filepath.Join(templatePath, s+".txt"), "\t", nil)
			var titleMap = make(map[string]string)
			var title []string
			for _, m := range titleMaps {
				title = append(title, m[columnName])
				titleMap[m["Raw"]] = m[columnName]
			}
			sheetTitle[s] = title
			sheetTitleMap[s] = titleMap
			writeTitle(excel, s, title)
		}
		excel.DeleteSheet("Sheet1")
		styleInit(excel)

		// Sample
		if *info != "" {
			updateDataFile2Sheet(excel, "Sample", *info, updateSample)
		}
	} else {
		excel = simpleUtil.HandleError(excelize.OpenFile(*template)).(*excelize.File)
		styleInit(excel)

		// bam文件路径
		if *bamPath != "" {
			updateBamPath(excel, *bamPath)
		}
	}

	// QC
	if *qc != "" {
		runQC <- true
		WriteQC(excel, runQC)
	}

	// CNV
	// QC -> DMD
	{
		runQC <- true
		loadDmd <- true
		LoadDmd(excel, loadDmd)
	}

	// 补充实验
	{
		runAe <- true
		WriteAe(excel, runAe)
	}

	// All variant data
	{
		localDb <- true
		runAvd <- true
		goWriteAvd(excel, loadDmd, runAvd, *all)
	}

	// write CNV after runAvd
	// CNV
	{
		runAvd <- true
		writeDmd <- true
		goUpdateCNV(excel, writeDmd)
	}

	if !*im {
		// drug, no use
		if *drugResult != "" {
			updateDataFile2Sheet(excel, *drugSheetName, *drugResult, updateABC)
		}

		// 个特
		if *featureList != "" {
			updateDataList2Sheet(excel, "个特", *featureList, updateFeature)
		}

		// 基因ID
		if *geneIDList != "" {
			updateDataList2Sheet(excel, "基因ID", *geneIDList, updateABC)
		}

		// DMD-lumpy
		if *lumpy != "" {
			updateDataFile2Sheet(excel, "DMD-lumpy", *lumpy, updateDMD)
		}
		// DMD-nator
		if *nator != "" {
			updateDataFile2Sheet(excel, "DMD-nator", *nator, updateDMD)
		}
	}

	{
		saveBatchCnv <- true
		go goWriteBatchCnv(saveBatchCnv)
	}

	{
		runAe <- true
		writeDmd <- true
		saveMain <- true
		go saveMainExcel(excel, *prefix+".xlsx", saveMain)
	}

	// waite excel write done
	{
		saveMain <- true
		saveBatchCnv <- true
	}
	log.Println("All Done")
}
