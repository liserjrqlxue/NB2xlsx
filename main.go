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
		var acmgCfg = simpleUtil.HandleError(textUtil.File2Map(acmgDbList, "\t", false)).(map[string]string)
		for k, v := range acmgCfg {
			acmgCfg[k] = filepath.Join(dbPath, v)
		}
		acmg2015.Init(acmgCfg)
	}

	I18n, _ = textUtil.File2MapMap(i18nTxt, "CN", "\t", nil)

	// load local db
	{
		if *im {
			loadLocalDb(jsonAesIM)
		} else {
			loadLocalDb(jsonAes)
		}
	}

	loadDb()

	log.Println("init done")
}

func main() {
	if osUtil.FileExists(*gender) {
		log.Printf("load gender map from %s", *gender)
		genderMap = simpleUtil.HandleError(textUtil.File2Map(*gender, "\t", false)).(map[string]string)
	}

	var (
		// un-block channel bool
		saveMainChan     = make(chan bool, 1)
		saveBatchCnvChan = make(chan bool, 1)

		excel *excelize.File
	)

	// load sample detail
	if *detail != "" {
		var details = textUtil.File2Slice(*detail, "\t")
		for _, line := range details {
			var db = make(map[string]string)
			var sampleID = line[0]
			db["productCode"] = line[1]
			db["hospital"] = line[2]
			sampleDetail[sampleID] = db
		}
	}

	// limsInfo for updateABC and updateQC
	limsInfo = loadLimsInfo(*lims)

	var batchCnvDb = loadBatchCNV(*batchCNV)

	if *info != "" {
		imInfo, _ = textUtil.File2MapMap(*info, "sampleID", "\t", nil)
	}

	go createExcel(excel, *prefix+".xlsx", *mode, *all, saveMainChan)

	// batchCNV.xlsx
	var bcSheetName = "Sheet1"
	go goWriteBatchCnv(bcSheetName, batchCnvDb, saveBatchCnvChan)

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
