package main

import (
	"flag"
	"log"
	"os"

	"github.com/360EntSecGroup-Skylar/excelize/v2"
	"github.com/liserjrqlxue/goUtil/osUtil"

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
		runDmd       = make(chan bool, 1)
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

	if *bamPath != "" {
		for i, path := range textUtil.File2Array(*bamPath) {
			var axis = simpleUtil.HandleError(excelize.CoordinatesToCellName(1, i+1)).(string)
			simpleUtil.CheckErr(
				excel.SetCellStr(
					*bamPathSheetName,
					axis,
					path,
				),
			)
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
		runDmd <- true
		WriteDmd(excel, runDmd)
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
		WriteAvd(excel, runDmd, runAvd, *all)
	}

	// drug
	if *drugResult != "" {
		log.Println("Start load Drug")
		var drugDb = make(map[string]map[string]map[string]string)
		var drug, _ = textUtil.File2MapArray(*drugResult, "\t", nil)
		for _, item := range drug {
			var sampleID = item["样本编号"]
			item["SampleID"] = sampleID
			updateABC(item)
			var drugName = item["药物名称"]
			var sampleDrug, ok1 = drugDb[sampleID]
			if !ok1 {
				sampleDrug = make(map[string]map[string]string)
				drugDb[sampleID] = sampleDrug
			}
			if _, ok := sampleDrug[drugName]; !ok {
				item["SampleID"] = sampleID
				sampleDrug[drugName] = item
			}
		}
	}

	{
		runAvd <- true
		saveBatchCnv <- true
		go writeBatchCnv(saveBatchCnv)
	}

	{
		runAe <- true
		saveMain <- true
		go func() {
			log.Printf("excel.SaveAs(\"%s\")\n", *prefix+".xlsx")
			simpleUtil.CheckErr(excel.SaveAs(*prefix + ".xlsx"))
			log.Println("Save main Done")
			<-saveMain
		}()
	}

	// waite excel write done
	{
		saveMain <- true
		saveBatchCnv <- true
	}
	log.Println("All Done")
}
