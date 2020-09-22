package main

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/360EntSecGroup-Skylar/excelize/v2"
	"github.com/liserjrqlxue/acmg2015"
	"github.com/liserjrqlxue/anno2xlsx/v2/anno"
	"github.com/liserjrqlxue/goUtil/simpleUtil"
	"github.com/liserjrqlxue/goUtil/textUtil"
)

func loadDb() {
	// load gene list
	for _, key := range textUtil.File2Array(*geneList) {
		geneListMap[key] = true
	}
	// load function exclude list
	for _, key := range textUtil.File2Array(*functionExclude) {
		functionExcludeMap[key] = true
	}

	// load disease database
	diseaseDb, _ = simpleUtil.Slice2MapMapArrayMerge(
		simpleUtil.HandleError(
			simpleUtil.HandleError(
				excelize.OpenFile(*diseaseExcel),
			).(*excelize.File).
				GetRows(*diseaseSheetName),
		).([][]string),
		"基因",
		"/",
	)
	// load 已解读数据库
	localDb, _ = simpleUtil.Slice2MapMapArray(
		simpleUtil.HandleError(
			simpleUtil.HandleError(
				excelize.OpenFile(*localDbExcel),
			).(*excelize.File).
				GetRows(*localDbSheetName),
		).([][]string),
		"Transcript", "cHGVS",
	)

	// load drop list
	for k, v := range simpleUtil.HandleError(textUtil.File2Map(*dropList, "\t", false)).(map[string]string) {
		dropListMap[k] = strings.Split(v, ",")
	}

}

var formulaTitle = map[string]bool{
	"解读人": true,
	"审核人": true,
}

var (
	isClinVar = map[string]bool{
		"Pathogenic":                   true,
		"Likely_pathogenic":            true,
		"Pathogenic/Likely_pathogenic": true,
	}
	isHGMD = map[string]bool{
		"DM":     true,
		"DM?":    true,
		"DM/DM?": true,
	}
)

func gt(s string, tv float64) bool {
	var af, err = strconv.ParseFloat(s, 64)
	if err == nil && af > tv {
		return true
	}
	return false
}

func filterAvd(item map[string]string) bool {
	var mainKey = item["Transcript"] + "\t" + item["cHGVS"]
	if _, ok := localDb[mainKey]; ok {
		return true
	}
	if !geneListMap[item["Gene Symbol"]] {
		return false
	}
	if isClinVar[item["ClinVar Significance"]] || isHGMD[item["HGMD Pred"]] {
		return true
	}
	if functionExcludeMap[item["Function"]] {
		return false
	}
	if gt(item["GnomAD AF"], 0.01) {
		return false
	}
	return true
}

var LOF = map[string]bool{
	"nonsense":   true,
	"frameshift": true,
	"splice-3":   true,
	"splice-5":   true,
}

func updateLOF(item map[string]string) {
	if !LOF[item["Function"]] || gt(item["GnomAD AF"], 0.01) || gt(item["1000G AF"], 0.01) {
		item["LOF"] = "NO"
	} else {
		item["LOF"] = "YES"
	}
}

func updateDisease(item map[string]string) {
	var gene = item["Gene Symbol"]
	var disease, ok = diseaseDb[gene]
	if ok {
		item["疾病中文名"] = disease["疾病"]
		item["遗传模式"] = disease["遗传模式"]
	}
}

func Interpret(item map[string]string, geneCount map[string]int) {

}

func updateAvd(item map[string]string, rIdx int) {
	item["1000Gp3 AF"] = item["1000G AF"]
	item["1000Gp3 EAS AF"] = item["1000G EAS AF"]
	var gene = item["Gene Symbol"]
	item["gene+cHGVS"] = gene + ":" + item["cHGVS"]
	item["gene+pHGVS3"] = gene + ":" + item["pHGVS3"]
	item["gene+pHGVS1"] = gene + ":" + item["pHGVS1"]
	anno.Score2Pred(item)
	updateLOF(item)
	updateDisease(item)
	var mainKey = item["Transcript"] + "\t" + item["cHGVS"]
	var db, ok = localDb[mainKey]
	if ok {
		if db["是否是包装位点"] == "是" {
			item["Database"] = "NBS-in"
			item["报告类别"] = "正式报告"
		} else {
			item["Database"] = "NBS-out"
		}
		item["Definition"] = db["Definition"]
		item["参考文献"] = db["Reference"]
	} else {
		item["Database"] = "."
		if item["LOF"] == "YES" {
			item["报告类别"] = "补充报告"
		}
	}
	acmg2015.AddEvidences(item)
	item["ACMG"] = acmg2015.PredACMG2015(item, *autoPVS1)
	anno.UpdateAutoRule(item)
	item["解读人"] = fmt.Sprintf("=INDEX('任务单（空sheet）'!O:O,MATCH(D%d&MID($C%d,1,6),'任务单（空sheet）'!$R:$R,0),1)", rIdx, rIdx)
	item["审核人"] = fmt.Sprintf("=INDEX('任务单（空sheet）'!P:P,MATCH(D%d&MID($C%d,1,6),'任务单（空sheet）'!$R:$R,0),1)", rIdx, rIdx)
}

func updateDmd(item map[string]string, rIdx int) {
	item["#sample"] = item["#Sample"]
	item["OMIM"] = item["Disease"]
	if item["Significant"] != "YES" {
		item["Significant"] = "NO"
	}
	/*
		item["depth_rate"]=item["batch_control"]
		item["others_rate"]=item["all_control"]
		item["G/H"]=item["Mean_Ratio"]
		item["Zscore"]=item["Median_Ratio"]
		var omimWebsite="http://omim.org/search/?index=entry&start=1&limit=10&sort=score+desc%2C+prefix_sort+desc&search="+item["gene"]
		item["omimWebsite"]=omimWebsite
	*/
	// primerDesign
	var exId = item["exon"]
	var cdsId = item["exon"]
	var ratioVal, err = strconv.ParseFloat(item["Mean_Ratio"], 64)
	if err != nil {
		ratioVal = 0
	}
	if ratioVal >= 1.3 && ratioVal < 1.8 {
		item["primerDesign"] = item["gene"] + "; " + item["NM"] + "; " + exId + " DUP; - ;" + exId + "; " + cdsId + "; Het"
	} else if ratioVal >= 1.8 {
		if item["chr"] == "chrX" {
			item["primerDesign"] = item["gene"] + "; " + item["NM"] + "; " + exId + " DUP; - ;" + exId + "; " + cdsId + "; Hemi"
		} else {
			item["primerDesign"] = item["gene"] + "; " + item["NM"] + "; " + exId + " DUP; - ;" + exId + "; " + cdsId + "; Hom"
		}
	} else if ratioVal >= 0.2 && ratioVal <= 0.75 {
		item["primerDesign"] = item["gene"] + "; " + item["NM"] + "; " + exId + " DEL; - ;" + exId + "; " + cdsId + "; Het"
	} else if ratioVal < 0.2 {
		if item["chr"] == "chrX" {
			item["primerDesign"] = item["gene"] + "; " + item["NM"] + "; " + exId + " DEL; - ;" + exId + "; " + cdsId + "; Hemi"
		} else {
			item["primerDesign"] = item["gene"] + "; " + item["NM"] + "; " + exId + " DEL; - ;" + exId + "; " + cdsId + "; Hom"
		}
	} else {
		item["primerDesign"] = "-"
	}
	item["解读人"] = fmt.Sprintf("=INDEX('任务单（空sheet）'!O:O,MATCH(D%d&MID($C%d,1,6),'任务单（空sheet）'!$R:$R,0),1)", rIdx, rIdx)
	item["审核人"] = fmt.Sprintf("=INDEX('任务单（空sheet）'!P:P,MATCH(D%d&MID($C%d,1,6),'任务单（空sheet）'!$R:$R,0),1)", rIdx, rIdx)
}

func updateDipin(item map[string]string, db map[string]map[string]string) {
	var sampleID = item["sample"]
	var info, ok = db[sampleID]
	if !ok {
		info = item
	}
	var qc, aResult, bResult string
	if item["QC"] != "pass" {
		qc = "_等验证"
	}
	if item["chr11"] == "N" {
		bResult = "阴性"
	} else {
		bResult = item["chr11"]
	}
	if item["chr16"] == "N" {
		aResult = "阴性"
	} else {
		aResult = item["chr16"]
	}
	info["SampleID"] = item["sample"]
	info["地贫_QC"] = item["QC"]
	info["β地贫_chr11"] = item["chr11"]
	info["α地贫_chr16"] = item["chr16"]
	info["β地贫_最终结果"] = bResult + qc
	info["α地贫_最终结果"] = aResult + qc
	db[sampleID] = info
}

func updateSma(item map[string]string, db map[string]map[string]string) {
	var sampleID = item["SampleID"]
	var info, ok = db[sampleID]
	if !ok {
		info = item
	}
	var result, qc, qcResult string
	var Categorization = item["SMN1_ex7_cn"]
	var QC = item["qc"]
	if Categorization == "1.5" || Categorization == "1" || QC != "1" {
		qcResult = "_等验证"
	}
	switch Categorization {
	case "0":
		result = "纯阳性"
	case "0.5":
		result = "纯合灰区"
	case "1":
		result = "杂合阳性"
	case "1.5":
		result = "杂合灰区"
	default:
		result = "阴性"
	}
	if QC == "1" {
		qc = "Pass"
	} else {
		qc = "Fail"
	}
	info["SMN1_检测结果"] = result
	info["SMN1_质控结果"] = qc
	info["SMN1 EX7 del最终结果"] = result + qcResult
	db[sampleID] = info
}

func updateAe(item map[string]string, rIdx int) {
	item["F8int1h-1.5k&2k最终结果"] = "检测范围外"
	item["F8int22h-10.8k&12k最终结果"] = "检测范围外"
	item["解读人"] = fmt.Sprintf("=INDEX('任务单（空sheet）'!O:O,MATCH(D%d&MID($C%d,1,6),'任务单（空sheet）'!$R:$R,0),1)", rIdx, rIdx)
	item["审核人"] = fmt.Sprintf("=INDEX('任务单（空sheet）'!P:P,MATCH(D%d&MID($C%d,1,6),'任务单（空sheet）'!$R:$R,0),1)", rIdx, rIdx)
}

func writeRow(excel *excelize.File, sheetName string, item map[string]string, title []string, rIdx int) {
	for j, k := range title {
		var axis = simpleUtil.HandleError(excelize.CoordinatesToCellName(j+1, rIdx)).(string)
		if formulaTitle[k] {
			simpleUtil.CheckErr(excel.SetCellFormula(sheetName, axis, item[k]))
		} else {
			simpleUtil.CheckErr(excel.SetCellValue(sheetName, axis, item[k]))
		}
		var list, ok = dropListMap[k]
		if ok {
			var dvRange = excelize.NewDataValidation(true)
			dvRange.Sqref = axis
			simpleUtil.CheckErr(dvRange.SetDropList(list))
			simpleUtil.CheckErr(excel.AddDataValidation(sheetName, dvRange))
		}
	}
}

func writeTitle(excel *excelize.File, sheetName string, title []string) {
	for j, k := range title {
		var axis = simpleUtil.HandleError(excelize.CoordinatesToCellName(j+1, 1)).(string)
		simpleUtil.CheckErr(excel.SetCellValue(sheetName, axis, k))
	}
}
