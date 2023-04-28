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
		var acmgCfg = simpleUtil.HandleError(textUtil.File2Map(acmgDb, "\t", false)).(map[string]string)
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
		loadDbChan       = make(chan bool, 1)
		writeAeChan      = make(chan bool, 1)
		runAvd           = make(chan bool, 1)
		loadDmdChan      = make(chan bool, 1)
		writeDmdChan     = make(chan bool, 1)
		runQC            = make(chan bool, 1)
		saveMainChan     = make(chan bool, 1)
		saveBatchCnvChan = make(chan bool, 1)
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
		var templateXlsx = mainTemplate
		if *wgs && templateXlsx == filepath.Join(templatePath, "NBS-final.result-批次号_产品编号.xlsx") {
			templateXlsx = filepath.Join(templatePath, "NBS.wgs.xlsx")
		}
		excel = simpleUtil.HandleError(excelize.OpenFile(mainTemplate)).(*excelize.File)
	}
	styleInit(excel)

	// fill sheets
	// local sheet names
	var (
		// Additional Experiments sheet name
		aeSheetName = "补充实验"
		// All samples' snv Excel sheet name
		allSheetName = "Sheet1"
		// All variants data sheet name
		avdSheetName = "All variants data"
		// Bam path sheet name
		bamPathSheetName = "bam文件路径"
		// DMD result sheet name
		dmdSheetName = "CNV"
		// WGS DMD result sheet name
		wgsDmdSheetName = "CNV"
		// Drug sheet name
		drugSheetName = "药物检测结果"
		// QC sheet name
		qcSheetName = "QC"
		// Sample sheet name
		sampleSheetName = "Sample"
		// individual characteristics sheet name
		icSheetName = "个特"
		// Gene ID sheet name
		geneIDSheetName = "基因ID"
		// batchCNV Excel sheet name
		bcSheetName = "Sheet1"
	)
	if *im {
		dmdSheetName = "DMD CNV"
		aeSheetName = "THAL CNV"
		avdSheetName = "SNV&INDEL"
	}
	if *wgs {
		dmdSheetName = "CNV-原始"
		wgsDmdSheetName = "CNV"
	}
	if *cs {
		wgsDmdSheetName = "DMD CNV"
	}

	// Sample
	if *im {
		updateDataFile2Sheet(excel, sampleSheetName, *info, updateSample)
	}
	// bam文件路径
	if *bamPath != "" {
		updateBamPath2Sheet(excel, bamPathSheetName, *bamPath)
	}
	// QC
	if *qc != "" {
		holdChan(runQC)
		WriteQC(excel, qcSheetName, runQC)
	}
	// CNV
	// QC -> DMD
	holdChan(runQC)
	LoadDmd4Sheet(excel, dmdSheetName, loadDmdChan)
	// 补充实验
	WriteAe(excel, aeSheetName, writeAeChan)
	// All variant data
	goWriteAvd(excel, avdSheetName, allSheetName, loadDmdChan, runAvd, *all)

	// CNV
	// write CNV after runAvd
	waitChan(runAvd)
	if !*cs {
		goUpdateCNV(excel, dmdSheetName, writeDmdChan)
	}
	// DMD-lumpy
	updateDataFile2Sheet(excel, wgsDmdSheetName, *lumpy, updateLumpy)
	// DMD-nator
	updateDataFile2Sheet(excel, wgsDmdSheetName, *nator, updateNator)
	// drug, no use
	updateDataFile2Sheet(excel, drugSheetName, *drugResult, updateDrug)
	// 个特
	updateDataList2Sheet(excel, icSheetName, *featureList, updateFeature)
	// 基因ID
	updateDataList2Sheet(excel, geneIDSheetName, *geneIDList, updateGeneID)

	// batchCNV.xlsx
	go goWriteBatchCnv(bcSheetName, saveBatchCnvChan)

	go saveMainExcel(excel, *prefix+".xlsx", saveMainChan, writeAeChan, writeDmdChan)

	// waite excel write done
	waitChan(saveMainChan, saveBatchCnvChan)
	log.Println("All Done")
}

func holdChan(ch ...chan<- bool) {
	for _, bools := range ch {
		bools <- true
	}
}

func waitChan(ch ...<-chan bool) {
	for _, bools := range ch {
		<-bools
	}
}
