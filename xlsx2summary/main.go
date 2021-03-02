package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/360EntSecGroup-Skylar/excelize/v2"
	"github.com/liserjrqlxue/goUtil/simpleUtil"
	"github.com/liserjrqlxue/version"
)

// os
var (
	ex, _        = os.Executable()
	exPath       = filepath.Dir(ex)
	dbPath       = filepath.Join(exPath, "..", "db")
	etcPath      = filepath.Join(exPath, "..", "etc")
	templatePath = filepath.Join(exPath, "..", "template")
)

type Info struct {
	sampleID string
	基因检测结果   []string
	基因       []GeneInfo
	geneList []string
	geneMap  map[string]GeneInfo
}

type GeneInfo struct {
	基因名称 string
	变异   []MutInfo
	疾病   string
	患病风险 string
	遗传方式 string
}
type MutInfo struct {
	外显子   string
	碱基改变  string
	氨基酸改变 string
}

// flag
var (
	input = flag.String(
		"input",
		"",
		"input excel",
	)
	infoDir = flag.String(
		"infoDir",
		"",
		"info dir",
	)
	output = flag.String(
		"output",
		"",
		"output excel, default is -input.结果汇总.xlsx")
	template = flag.String(
		"template",
		filepath.Join(templatePath, "结果汇总.xlsx"),
		"template of 结果汇总",
	)
)

func main() {
	version.LogVersion()
	flag.Parse()
	if *input == "" || *infoDir == "" {
		flag.Usage()
		fmt.Printf("-input/-infoDir required!")
		os.Exit(1)
	}
	if *output == "" {
		*output = *input + ".结果汇总.xlsx"
	}

	var inExcel = simpleUtil.HandleError(excelize.OpenFile(*input)).(*excelize.File)
	var outExcel = simpleUtil.HandleError(excelize.OpenFile(*template)).(*excelize.File)
	var strSlice = simpleUtil.HandleError(inExcel.GetRows("All variants data")).([][]string)
	var sampleList []string
	var db = make(map[string]Info)
	var title []string
	for i, strArray := range strSlice {
		if i == 0 {
			title = strArray
			continue
		}
		var item = make(map[string]string)
		for j := range title {
			item[title[j]] = strArray[j]
		}
		var sampleID = item["SampleID"]
		var info, ok1 = db[sampleID]
		if !ok1 {
			info = Info{
				sampleID: sampleID,
				geneMap:  make(map[string]GeneInfo),
			}
			sampleList = append(sampleList, sampleID)
		}
		var geneSymbol = item["Gene Symbol"]
		var geneInfo, ok2 = info.geneMap[geneSymbol]
		if !ok2 {
			geneInfo = GeneInfo{
				基因名称: geneSymbol,
				变异:   []MutInfo{},
				疾病:   item["疾病中文名"],
				患病风险: item["遗传模式判读"],
				遗传方式: item["遗传模式"],
			}
			info.geneList = append(info.geneList, geneSymbol)
		}
		var mutInfo = MutInfo{
			外显子:   item["ExIn_ID"],
			碱基改变:  item["cHGVS"],
			氨基酸改变: item["pHGVS"],
		}
		geneInfo.变异 = append(geneInfo.变异, mutInfo)
		info.geneMap[geneSymbol] = geneInfo
		db[sampleID] = info
	}
	var rIdx = 5
	var sheetName = "159基因结果汇总"
	for _, sampleID := range sampleList {
		var item = db[sampleID]
		for _, geneSymbol := range item.geneList {
			var geneInfo = item.geneMap[geneSymbol]
			item.基因 = append(item.基因, geneInfo)
			item.基因检测结果 = append(item.基因检测结果, geneInfo.疾病)
		}
		simpleUtil.CheckErr(
			outExcel.SetCellValue(
				sheetName,
				simpleUtil.HandleError(excelize.CoordinatesToCellName(4, rIdx)).(string),
				sampleID,
			),
		)
		for i, v := range item.基因检测结果 {
			simpleUtil.CheckErr(
				outExcel.SetCellValue(
					sheetName,
					simpleUtil.HandleError(excelize.CoordinatesToCellName(10+i, rIdx)).(string),
					v,
				),
			)
		}
		// 15
		for i, geneInfo := range item.基因 {
			if i > 4 {
				break
			}
			fmt.Printf("%s\t%d\t%s\n", sampleID, i, geneInfo.基因名称)
			simpleUtil.CheckErr(
				outExcel.SetCellValue(
					sheetName,
					simpleUtil.HandleError(excelize.CoordinatesToCellName(15+i*10, rIdx)).(string),
					geneInfo.基因名称,
				),
			)
			simpleUtil.CheckErr(
				outExcel.SetCellValue(
					sheetName,
					simpleUtil.HandleError(excelize.CoordinatesToCellName(22+i*10, rIdx)).(string),
					geneInfo.疾病,
				),
			)
			simpleUtil.CheckErr(
				outExcel.SetCellValue(
					sheetName,
					simpleUtil.HandleError(excelize.CoordinatesToCellName(23+i*10, rIdx)).(string),
					geneInfo.患病风险,
				),
			)
			simpleUtil.CheckErr(
				outExcel.SetCellValue(
					sheetName,
					simpleUtil.HandleError(excelize.CoordinatesToCellName(24+i*10, rIdx)).(string),
					geneInfo.遗传方式,
				),
			)
			for j, mutInfo := range geneInfo.变异 {
				if j > 1 {
					break
				}
				fmt.Printf("\t%d:%d\t%s\n", i, j, mutInfo.碱基改变)
				simpleUtil.CheckErr(
					outExcel.SetCellValue(
						sheetName,
						simpleUtil.HandleError(excelize.CoordinatesToCellName(16+i*10+j*3, rIdx)).(string),
						mutInfo.外显子,
					),
				)
				simpleUtil.CheckErr(
					outExcel.SetCellValue(
						sheetName,
						simpleUtil.HandleError(excelize.CoordinatesToCellName(17+i*10+j*3, rIdx)).(string),
						mutInfo.碱基改变,
					),
				)
				simpleUtil.CheckErr(
					outExcel.SetCellValue(
						sheetName,
						simpleUtil.HandleError(excelize.CoordinatesToCellName(18+i*10+j*3, rIdx)).(string),
						mutInfo.氨基酸改变,
					),
				)
			}
		}
		rIdx++
	}
	simpleUtil.CheckErr(outExcel.SaveAs(*output))
}
