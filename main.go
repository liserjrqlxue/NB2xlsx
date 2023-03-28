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

func getI18n(key string) string {
	var value, ok = I18n[key][i18n]
	if ok {
		return value
	}
	return key
}

func main() {
	if osUtil.FileExists(*gender) {
		log.Printf("load gender map from %s", *gender)
		genderMap = simpleUtil.HandleError(textUtil.File2Map(*gender, "\t", false)).(map[string]string)
	}

	I18n, _ = textUtil.File2MapMap(filepath.Join(etcPath, "i18n.txt"), "CN", "\t", nil)

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

	if *cs {
		for _, s := range textUtil.File2Array(filepath.Join(etcPath, "TOP1K.BB.gene.name.txt")) {
			top1kGene[s] = true
		}
	}

	if *im {
		updateColumName()
	}

	updateSheetTitleMap()

	if *im {
		initExcel(excel)

		// Sample
		if *info != "" {
			updateDataFile2Sheet(excel, "Sample", *info, updateSample)
		}
	} else {
		var templateXlsx = *template
		if *wgs && templateXlsx == filepath.Join(templatePath, "NBS-final.result-批次号_产品编号.xlsx") {
			templateXlsx = filepath.Join(templatePath, "NBS.wgs.xlsx")
		}
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
	if !*cs {
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
	}

	if *cs {
		// DMD-lumpy
		if *lumpy != "" {
			updateDataFile2Sheet(excel, "DMD", *lumpy, updateLumpy)
		}
		// DMD-nator
		if *nator != "" {
			updateDataFile2Sheet(excel, "DMD", *nator, updateDMD)
		}
	} else { // NBS WGS
		// DMD-lumpy
		if *lumpy != "" {
			updateDataFile2Sheet(excel, "CNV", *lumpy, updateDMD)
		}
		// DMD-nator
		if *nator != "" {
			updateDataFile2Sheet(excel, "CNV", *nator, updateDMD)
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
