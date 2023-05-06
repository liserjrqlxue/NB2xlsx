package main

import (
	"flag"
	"github.com/liserjrqlxue/acmg2015"
	"github.com/liserjrqlxue/goUtil/osUtil"
	"github.com/liserjrqlxue/goUtil/simpleUtil"
	"github.com/liserjrqlxue/goUtil/textUtil"
	"github.com/liserjrqlxue/version"
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

	modeType = ModeMap[*mode]

	// acmg2015 init
	if *acmg {
		log.Println("ACMG2015 init")
		acmg2015.AutoPVS1 = *autoPVS1
		var acmgCfg = simpleUtil.HandleError(textUtil.File2Map(acmgDbList, "\t", false)).(map[string]string)
		for k, v := range acmgCfg {
			acmgCfg[k] = filepath.Join(dbPath, v)
		}
		acmg2015.Init(acmgCfg)
	}

	I18n, _ = textUtil.File2MapMap(i18nTxt, "CN", "\t", nil)

	// load local db
	switch modeType {
	case NBSP:
		loadLocalDb(jsonAes)
	case NBSIM:
		loadLocalDb(jsonAesIM)
	case WGSNB:
		loadLocalDb(jsonAesWGS)
	}

	loadDb(modeType)

	log.Println("init done")
}

func main() {
	if osUtil.FileExists(*gender) {
		log.Printf("load gender map from %s", *gender)
		genderMap = simpleUtil.HandleError(textUtil.File2Map(*gender, "\t", false)).(map[string]string)
	}

	// limsInfo for updateABC and updateQC
	// sampleDetail for subFlag
	// imInfo for parseProductCode, updateInfo, updateQC
	limsInfo, sampleDetail, imInfo = loadSamplesInfo(*lims, *detail, *info)
	if modeType == NBSIM {
		parseProductCode()
	}
	loadDiseaseDb(i18n, modeType)

	createMainExcel(*prefix+".xlsx", modeType, *all)
	// batchCNV -> SampleGeneInfo,batchCNV.xlsx
	useBatchCNV(*batchCNV, *prefix+".batchCNV.xlsx", "Sheet1", modeType)

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
