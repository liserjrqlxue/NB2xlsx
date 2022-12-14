package main

import (
	"flag"
	"github.com/liserjrqlxue/goUtil/osUtil"
	"github.com/xuri/excelize/v2"
	"log"
	"os"

	"github.com/liserjrqlxue/goUtil/simpleUtil"
	"github.com/liserjrqlxue/goUtil/textUtil"
	"github.com/liserjrqlxue/version"
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
	)

	// load local db
	{
		localDb <- true
		loadLocalDb(localDb)
	}

	loadDb()

	limsInfo = loadLimsInfo()

	if *batchCNV != "" {
		loadBatchCNV(*batchCNV)
	}

	var excel = simpleUtil.HandleError(excelize.OpenFile(*template)).(*excelize.File)
	styleInit(excel)

	// bam文件路径
	if *bamPath != "" {
		updateBamPath(excel, *bamPath)
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

	// drug, no use
	if *drugResult != "" {
		updateDrug(excel, *drugResult)
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
