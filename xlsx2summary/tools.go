package main

import (
	"fmt"

	"github.com/360EntSecGroup-Skylar/excelize/v2"
	"github.com/liserjrqlxue/goUtil/simpleUtil"
)

func WriteCellStr(excel *excelize.File, sheetName string, col, row int, value string) {
	simpleUtil.CheckErr(
		excel.SetCellStr(
			sheetName,
			simpleUtil.HandleError(excelize.CoordinatesToCellName(col, row)).(string),
			value,
		),
	)
}

func WriteCellValue(excel *excelize.File, sheetName string, col, row int, value interface{}) {
	simpleUtil.CheckErr(
		excel.SetCellValue(
			sheetName,
			simpleUtil.HandleError(excelize.CoordinatesToCellName(col, row)).(string),
			value,
		),
	)
}

type updateInfo func(db map[string]Info, item map[string]string)

func updateInfoDB(db map[string]Info, geneInfo0 GeneInfo, mutInfo MutInfo, sampleID, geneSymbol string) {
	var info, ok1 = db[sampleID]
	if !ok1 {
		info = Info{
			sampleID: sampleID,
			geneMap:  make(map[string]GeneInfo),
		}
	}

	var geneInfo, ok2 = info.geneMap[geneSymbol]
	if !ok2 {
		geneInfo = geneInfo0
		info.geneList = append(info.geneList, geneSymbol)
	}

	geneInfo.变异 = append(geneInfo.变异, mutInfo)
	fmt.Printf("[%9s]\t[%s]\t%+v\n", sampleID, geneInfo.基因名称, geneInfo.变异)

	info.geneMap[geneSymbol] = geneInfo

	db[sampleID] = info
}

func updateInfoDBfromAVD(db map[string]Info, item map[string]string) {
	if item["报告类别"] == "正式报告" {
		var sampleID = item["SampleID"]
		var geneSymbol = item["Gene Symbol"]
		var geneInfo = GeneInfo{
			基因名称: geneSymbol,
			变异:   []MutInfo{},
			疾病:   item["疾病中文名"],
			患病风险: item["遗传模式判读"],
			遗传方式: item["遗传模式"],
		}
		var mutInfo = MutInfo{
			外显子:   item["ExIn_ID"],
			碱基改变:  item["cHGVS"],
			氨基酸改变: getAfterVerticalbar(item["pHGVS"]),
		}
		updateInfoDB(db, geneInfo, mutInfo, sampleID, geneSymbol)
	}
}

func updateInfoDBfromCNV(db map[string]Info, item map[string]string) {
	if item["报告类别"] == "正式报告" {
		var sampleID = item["#sample"]
		var geneSymbol = item["gene"]
		var geneInfo = GeneInfo{
			基因名称: geneSymbol,
			变异:   []MutInfo{},
			疾病:   "杜氏肌营养不良",
			患病风险: "携带",
			遗传方式: "XLR",
		}
		if (item["gender"] == "M" && item["杂合性"] == "Hemi") || (item["gender"] == "F" && item["杂合性"] == "Hom") {
			geneInfo.患病风险 = "可能患病"
		}
		var mutInfo = MutInfo{
			外显子:   item["exon"],
			碱基改变:  item["核苷酸变化"],
			氨基酸改变: ".",
		}
		updateInfoDB(db, geneInfo, mutInfo, sampleID, geneSymbol)
	}
}

func updateInfoDBfromHBB(db map[string]Info, item map[string]string) {
	var 最终结果 = "β地贫_最终结果"
	if item[最终结果] != "阴性" {
		var sampleID = item["SampleID"]
		var geneSymbol = "HBB"
		var geneInfo = GeneInfo{
			基因名称: geneSymbol,
			变异:   []MutInfo{},
			疾病:   "β地中海贫血",
			患病风险: 患病风险[item[最终结果]],
			遗传方式: "AR",
		}
		var mutInfo = MutInfo{
			外显子:   ".",
			碱基改变:  item[最终结果],
			氨基酸改变: ".",
		}
		updateInfoDB(db, geneInfo, mutInfo, sampleID, geneSymbol)
	}
}

func updateInfoDBfromHBA(db map[string]Info, item map[string]string) {
	var 最终结果 = "α地贫_最终结果"
	if item[最终结果] != "阴性" {
		var sampleID = item["SampleID"]
		var geneSymbol = "HBA1/HBA2"
		var geneInfo = GeneInfo{
			基因名称: geneSymbol,
			变异:   []MutInfo{},
			疾病:   "β地中海贫血",
			患病风险: 患病风险[item[最终结果]],
			遗传方式: "AR",
		}
		var mutInfo = MutInfo{
			外显子:   ".",
			碱基改变:  item[最终结果],
			氨基酸改变: ".",
		}
		updateInfoDB(db, geneInfo, mutInfo, sampleID, geneSymbol)
	}

}

func updateInfoDBfromSMA(db map[string]Info, item map[string]string) {
	var 最终结果 = "SMN1 EX7 del最终结果"
	if item[最终结果] == "纯合阳性" || item[最终结果] == "杂合阳性" {
		var sampleID = item["SampleID"]
		var geneSymbol = "SMN1"
		var geneInfo = GeneInfo{
			基因名称: geneSymbol,
			变异:   []MutInfo{},
			疾病:   "脊髓性肌萎缩症",
			患病风险: 患病风险[item[最终结果]],
			遗传方式: "AR",
		}
		var mutInfo = MutInfo{
			外显子:   ".",
			碱基改变:  "EX7 DEL",
			氨基酸改变: ".",
		}
		updateInfoDB(db, geneInfo, mutInfo, sampleID, geneSymbol)
	}
}

func updateInfoDBfromAE(db map[string]Info, item map[string]string) {
	updateInfoDBfromHBB(db, item)
	updateInfoDBfromHBA(db, item)
	updateInfoDBfromSMA(db, item)
}

func GetInfo(db map[string]Info, annoExcel *excelize.File, sheetName string, fn updateInfo) {
	var strSlice = simpleUtil.HandleError(annoExcel.GetRows(sheetName)).([][]string)
	var title []string
	for i, strArray := range strSlice {
		if i == 0 {
			title = strArray
			continue
		}
		var item = make(map[string]string)
		for j := range title {
			if j < len(strArray) {
				item[title[j]] = strArray[j]
			}
		}
		fn(db, item)
	}
}

func getInfoFromAE(db map[string]Info, annoExcel *excelize.File) {
	GetInfo(db, annoExcel, "补充实验", updateInfoDBfromAE)
}

func getInfoFromCNV(db map[string]Info, annoExcel *excelize.File) {
	GetInfo(db, annoExcel, "CNV", updateInfoDBfromCNV)
}

func getInfoFromAVD(db map[string]Info, annoExcel *excelize.File) {
	GetInfo(db, annoExcel, "All variants data", updateInfoDBfromAVD)
}
