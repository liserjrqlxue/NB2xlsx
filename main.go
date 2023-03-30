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

func getI18n(v, k string) string {
	var value, ok = I18n[k+"."+v][i18n]
	if !ok {
		value, ok = I18n[v][i18n]
	}
	if ok {
		return value
	}
	return v
}

func main() {
	if osUtil.FileExists(*gender) {
		log.Printf("load gender map from %s", *gender)
		genderMap = simpleUtil.HandleError(textUtil.File2Map(*gender, "\t", false)).(map[string]string)
	}

	I18n, _ = textUtil.File2MapMap(filepath.Join(etcPath, "i18n.txt"), "CN", "\t", nil)

	// un-block channel bool
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
	var excel *excelize.File

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
		parseProductCode()
	}
	loadDiseaseDb(i18n)

	updateSheetTitleMap()

	if *im {
		initExcel(excel)
	} else {
		var templateXlsx = *template
		if *wgs && templateXlsx == filepath.Join(templatePath, "NBS-final.result-批次号_产品编号.xlsx") {
			templateXlsx = filepath.Join(templatePath, "NBS.wgs.xlsx")
		}
		excel = simpleUtil.HandleError(excelize.OpenFile(*template)).(*excelize.File)
	}
	styleInit(excel)

	if *im {
		// Sample
		if *info != "" {
			updateDataFile2Sheet(excel, "Sample", *info, updateSample)
		}
	} else {
		// bam文件路径
		if *bamPath != "" {
			updateBamPath(excel, *bamPath)
		}
	}

	// QC
	if *qc != "" {
		wait(runQC)
		WriteQC(excel, runQC)
	}

	// CNV
	// QC -> DMD
	wait(runQC, loadDmd)
	LoadDmd(excel, loadDmd)

	// 补充实验
	wait(runAe)
	WriteAe(excel, runAe)

	// All variant data
	wait(localDb, runAvd)
	goWriteAvd(excel, loadDmd, runAvd, *all)

	// write CNV after runAvd
	// CNV
	wait(runAvd)
	if *cs {
		*dmdSheetName = "DMD CNV"
	} else {
		wait(writeDmd)
		goUpdateCNV(excel, writeDmd)
	}
	// DMD-lumpy
	if *lumpy != "" {
		updateDataFile2Sheet(excel, *dmdSheetName, *lumpy, updateLumpy)
	}
	// DMD-nator
	if *nator != "" {
		updateDataFile2Sheet(excel, *dmdSheetName, *nator, updateNator)
	}

	// drug, no use
	if *drugResult != "" {
		updateDataFile2Sheet(excel, *drugSheetName, *drugResult, updateDrug)
	}

	// 个特
	if *featureList != "" {
		updateDataList2Sheet(excel, "个特", *featureList, updateFeature)
	}

	// 基因ID
	if *geneIDList != "" {
		updateDataList2Sheet(excel, "基因ID", *geneIDList, updateGeneID)
	}

	wait(saveBatchCnv)
	go goWriteBatchCnv(saveBatchCnv)

	wait(runAe, writeDmd, saveMain)
	go saveMainExcel(excel, *prefix+".xlsx", saveMain)

	// waite excel write done
	wait(saveMain, saveBatchCnv)
	log.Println("All Done")
}

func wait(ch ...chan<- bool) {
	for _, bools := range ch {
		bools <- true
	}
}
