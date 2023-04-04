package main

import (
	"github.com/liserjrqlxue/acmg2015"
	"github.com/liserjrqlxue/goUtil/simpleUtil"
	"github.com/liserjrqlxue/goUtil/textUtil"
	"github.com/xuri/excelize/v2"
	"log"
	"path/filepath"
	"strings"
)

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
