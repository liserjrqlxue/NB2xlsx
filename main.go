package main

import (
	"flag"
	"github.com/liserjrqlxue/acmg2015"
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

	// acmg2015 init
	if *acmg {
		acmg2015.AutoPVS1 = *autoPVS1
		var acmgCfg = simpleUtil.HandleError(textUtil.File2Map(*acmgDb, "\t", false)).(map[string]string)
		for k, v := range acmgCfg {
			acmgCfg[k] = filepath.Join(dbPath, v)
		}
		acmg2015.Init(acmgCfg)
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
		loadDbChan   = make(chan bool, 1)
		runAe        = make(chan bool, 1)
		runAvd       = make(chan bool, 1)
		loadDmdChan  = make(chan bool, 1)
		writeDmd     = make(chan bool, 1)
		runQC        = make(chan bool, 1)
		saveMain     = make(chan bool, 1)
		saveBatchCnv = make(chan bool, 1)
	)
	var excel *excelize.File

	// load local db
	{
		if *im {
			loadLocalDb(filepath.Join(etcPath, "已解读数据库.IM.json.aes"), loadDbChan)
		} else {
			loadLocalDb(filepath.Join(etcPath, "已解读数据库.json.aes"), loadDbChan)
		}
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
		excel = initExcel()
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
		fillChan(runQC)
		WriteQC(excel, runQC)
	}

	// CNV
	// QC -> DMD
	fillChan(runQC)
	LoadDmd(excel, loadDmdChan)

	// 补充实验
	WriteAe(excel, runAe)

	// All variant data
	if *im {
		goWriteAvd(excel, "SNV&INDEL", loadDmdChan, runAvd, *all)
	} else {
		goWriteAvd(excel, *avdSheetName, loadDmdChan, runAvd, *all)
	}

	// write CNV after runAvd
	// CNV
	emptyChan(runAvd)
	if *cs {
		*dmdSheetName = "DMD CNV"
	} else {
		goUpdateCNV(excel, writeDmd)
	}
	if *wgs {
		*dmdSheetName = "CNV"
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

	go goWriteBatchCnv(saveBatchCnv)

	emptyChan(runAe, writeDmd)

	go saveMainExcel(excel, *prefix+".xlsx", saveMain)

	// waite excel write done
	emptyChan(saveMain, saveBatchCnv)
	log.Println("All Done")
}

func fillChan(ch ...chan<- bool) {
	for _, bools := range ch {
		bools <- true
	}
}

func emptyChan(ch ...<-chan bool) {
	for _, bools := range ch {
		<-bools
	}
}
