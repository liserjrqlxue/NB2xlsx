package main

import (
	"log"
	"regexp"
	"strings"

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
			碱基改变:  cHgvsAlt(item["cHGVS"]),
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
			患病风险: "携带者",
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

var prefixHBA = map[string]string{
	"3.7":  "-α",
	"4.2":  "-α",
	"SEA":  "--",
	"THAI": "-- ",
	"FIL":  "-- ",
}

func addPrefixHBA(result string) string {
	return prefixHBA[result] + result
}
func updateInfoDBfromHBA(db map[string]Info, item map[string]string) {
	var 最终结果 = "α地贫_最终结果"
	if item[最终结果] != "阴性" {
		var sampleID = item["SampleID"]
		var geneSymbol = "HBA1/HBA2"
		var geneInfo = GeneInfo{
			基因名称: geneSymbol,
			变异:   []MutInfo{},
			疾病:   "α地中海贫血",
			患病风险: 患病风险[item[最终结果]],
			遗传方式: "AR",
		}
		var mutInfo = MutInfo{
			外显子:   ".",
			碱基改变:  addPrefixHBA(item[最终结果]),
			氨基酸改变: ".",
		}
		updateInfoDB(db, geneInfo, mutInfo, sampleID, geneSymbol)
	}

}

func updateInfoDBfromSMA(db map[string]Info, item map[string]string) {
	var 最终结果 = "SMN1 EX7 del最终结果"
	var sampleID = item["SampleID"]
	sampleCount[sampleID]++
	if item[最终结果] == "纯合阳性" || item[最终结果] == "杂合阳性" {
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

func checkTitleName(strArray []string, name string, r, c int) {
	if strArray[c-1] != name {
		log.Fatalf("错误：表头(%d,%d)[%s]!=%s\n", r, c, strArray[c-1], name)
	}
}

func getAfterVerticalbar(str string) string {
	var s = strings.Split(str, "|")
	return strings.TrimSpace(s[len(s)-1])
}

// regexp
var (
	cHGVSalt = regexp.MustCompile(`alt: (\S+) \)`)
)

func cHgvsAlt(cHgvs string) string {
	if m := cHGVSalt.FindStringSubmatch(cHgvs); m != nil {
		return m[1]
	}
	return cHgvs
}

func loadDB() map[string]Info {
	var db = make(map[string]Info)
	// 读解读表
	for n, annoFile := range strings.Split(*anno, ",") {
		log.Printf("信息：开始读取解读表%d[%s]\n", n+1, annoFile)
		var annoExcel = simpleUtil.HandleError(excelize.OpenFile(annoFile)).(*excelize.File)

		getInfoFromAVD(db, annoExcel)
		getInfoFromCNV(db, annoExcel)
		getInfoFromAE(db, annoExcel)
	}
	return db
}

func fillExcel(strSlice [][]string, db map[string]Info, outExcel *excelize.File, sheetName string) {
	for rIdx, strArray := range strSlice {
		if rIdx < titleRowIndex-3 {
			continue
		} else if rIdx == titleRowIndex-3 {
			checkTitleName(strArray, "基因一", rIdx+1, geneNameCIdx)
			if len(strArray) > colLength && strArray[colLength] != "" {
				log.Fatalf("错误：表头(%d,%d)[%s]!=\"\"", rIdx+1, colLength+1, strArray[colLength])
			}
			continue
		} else if rIdx == titleRowIndex-2 {
			checkTitleName(strArray, resultTitle, rIdx+1, resultCIdx)
			checkTitleName(strArray, "基因名称", rIdx+1, geneNameCIdx)
			checkTitleName(strArray, "变异一", rIdx+1, geneNameCIdx+1)
			checkTitleName(strArray, "疾病", rIdx+1, geneNameCIdx+mutLit*mutColCount+1)
			checkTitleName(strArray, "患病风险", rIdx+1, geneNameCIdx+mutLit*mutColCount+2)
			checkTitleName(strArray, "遗传方式", rIdx+1, geneNameCIdx+mutLit*mutColCount+3)
			continue
		} else if rIdx == titleRowIndex-1 {
			checkTitleName(strArray, sampleIDTitle, rIdx+1, sampleIDCIdx)
			checkTitleName(strArray, summaryTitle, rIdx+1, summaryCIdx)
			checkTitleName(strArray, "疾病一", rIdx+1, resultCIdx)
			checkTitleName(strArray, "外显子", rIdx+1, geneNameCIdx+1)
			checkTitleName(strArray, "碱基改变", rIdx+1, geneNameCIdx+2)
			checkTitleName(strArray, "氨基酸改变", rIdx+1, geneNameCIdx+3)
			continue
		}
		rIdx++
		var index = rIdx - titleRowIndex
		if len(strArray) < sampleIDCIdx || strArray[sampleIDCIdx-1] == "" {
			log.Printf("信息：样品%3d为空，跳过\n", index)
			continue
		}
		var sampleID = strArray[sampleIDCIdx-1]
		if len(strArray) > colLength && strArray[colLength] != "" {
			log.Printf("警告：样品%3d[%11s]已有时间戳，跳过\n", index, sampleID)
			continue
		}
		var item, ok = db[sampleID]
		if !ok {
			if sampleCount[sampleID] == 0 {
				log.Printf("警告：样品%3d[%11s]无解读信息，跳过\n", index, sampleID)
				continue
			} else {
				log.Printf("信息：样品%3d[%11s]无疾病\n", index, sampleID)
				item = Info{
					sampleID: sampleID,
					基因检测结果总结: "无疾病",
				}
			}
		}
		if sampleCount[sampleID] > 1 {
			log.Printf("警告：样品%3d[%11s]解读信息重复%d次\n", index, sampleID, sampleCount[sampleID])
		}

		item.sortGene()
		item.resumeGene()

		item.fillExcel(outExcel, sampleID, sheetName, index, rIdx)
	}
}
