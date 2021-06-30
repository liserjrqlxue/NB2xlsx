package main

import (
	"fmt"
	"log"

	"github.com/360EntSecGroup-Skylar/excelize/v2"
)

// Info recode sample info and summary
type Info struct {
	sampleID   string
	可疑, 携带     int
	基因检测结果     []string
	基因         []GeneInfo
	geneList   []string
	geneMap    map[string]GeneInfo
	样本快递日期     string
	新筛编号       string
	华大基因检测编号   string
	性别         string
	生化筛查结果     string
	生化筛查结果拟诊   string
	基因检测结果报告日期 string
	基因检测结果总结   string
}

// GeneInfo record gene info
type GeneInfo struct {
	基因名称 string
	变异   []MutInfo
	疾病   string
	患病风险 string
	遗传方式 string
}

// MutInfo recorde mutaiton info
type MutInfo struct {
	外显子   string
	碱基改变  string
	氨基酸改变 string
}

// 按患病风险排序基因
func (item *Info) sortGene() {
	var head, tail []string
	for _, geneSymbol := range item.geneList {
		var geneInfo = item.geneMap[geneSymbol]
		if geneInfo.患病风险 == "可能患病" {
			head = append(head, geneSymbol)
		} else {
			tail = append(tail, geneSymbol)
		}
	}
	item.geneList = append(head, tail...)
}

func (item *Info) resumeGene() {
	for _, geneSymbol := range item.geneList {
		var geneInfo = item.geneMap[geneSymbol]
		item.基因 = append(item.基因, geneInfo)
		switch geneInfo.患病风险 {
		case "可能患病":
			item.可疑++
			item.基因检测结果 = append(item.基因检测结果, geneInfo.疾病+"/可疑")
		case "携带者":
			item.携带++
			item.基因检测结果 = append(item.基因检测结果, geneInfo.疾病+"/携带")
		}
		if item.可疑 > 0 && item.携带 > 0 {
			item.基因检测结果总结 = fmt.Sprintf("%d个疾病可疑+%d个疾病携带", item.可疑, item.携带)
		} else if item.可疑 > 0 {
			item.基因检测结果总结 = fmt.Sprintf("%d个疾病可疑", item.可疑)
		} else if item.携带 > 0 {
			item.基因检测结果总结 = fmt.Sprintf("%d个疾病携带", item.携带)
		} else {
			item.基因检测结果总结 = "无疾病"
		}
	}
	if len(item.geneList) == 0 {
		item.基因检测结果总结 = "无疾病"
	}
}

func (item *Info) fillExcel(outExcel *excelize.File, sampleID, sheetName string, index, rIdx int) {
	WriteCellStr(outExcel, sheetName, summaryCIdx, rIdx, item.基因检测结果总结)
	item.fillGeneResult(outExcel, sampleID, sheetName, index, rIdx)
	item.fillGeneInfo(outExcel, sampleID, sheetName, index, rIdx)
	//WriteCellValue(outExcel, sheetName, colLength+1, rIdx, time.Now().UTC())
	log.Printf("信息：样品%3d[%11s]已写入\n", index, sampleID)
}

// 填充 基因检测结果
func (item *Info) fillGeneResult(outExcel *excelize.File, sampleID, sheetName string, index, rIdx int) {
	for i, 疾病 := range item.基因检测结果 {
		if i > geneLimit {
			log.Printf("警告：样品%3d[%11s]检出疾病超出%d个，疾病[%s]跳过\n", index, sampleID, geneLimit, 疾病)
			break
		}
		WriteCellStr(outExcel, sheetName, resultCIdx+i, rIdx, 疾病)
	}
}

func (item *Info) fillGeneInfo(outExcel *excelize.File, sampleID, sheetName string, index, rIdx int) {
	fmt.Printf("\t%s\n", sampleID)
	for i, geneInfo := range item.基因 {
		if i > geneLimit {
			log.Printf("警告：样品%3d[%11s]检出基因超出%d个，基因[%s]跳过\n", index, sampleID, geneLimit, geneInfo.基因名称)
			break
		}
		WriteCellStr(outExcel, sheetName, geneNameCIdx+i*geneColLength, rIdx, geneInfo.基因名称)
		WriteCellStr(outExcel, sheetName, geneNameCIdx+mutLit*mutColCount+1+i*geneColLength, rIdx, geneInfo.疾病)
		WriteCellStr(outExcel, sheetName, geneNameCIdx+mutLit*mutColCount+2+i*geneColLength, rIdx, geneInfo.患病风险)
		WriteCellStr(outExcel, sheetName, geneNameCIdx+mutLit*mutColCount+3+i*geneColLength, rIdx, geneInfo.遗传方式)
		fmt.Printf("\t\t\t%-8s\t%s\n", geneInfo.基因名称, geneInfo.患病风险)

		geneInfo.fillMutInfo(outExcel, sampleID, sheetName, i, index, rIdx)
	}
}

func (geneInfo *GeneInfo) fillMutInfo(outExcel *excelize.File, sampleID, sheetName string, i, index, rIdx int) {
	for j, mutInfo := range geneInfo.变异 {
		if j > 1 {
			log.Printf("警告：样品%3d[%11s]在基因[%s]检出变异超出2个，变异[%s]跳过", index, sampleID, geneInfo.基因名称, mutInfo.碱基改变)
			break
		}
		WriteCellStr(outExcel, sheetName, geneNameCIdx+1+i*geneColLength+j*mutColCount, rIdx, mutInfo.外显子)
		WriteCellStr(outExcel, sheetName, geneNameCIdx+2+i*geneColLength+j*mutColCount, rIdx, mutInfo.碱基改变)
		WriteCellStr(outExcel, sheetName, geneNameCIdx+3+i*geneColLength+j*mutColCount, rIdx, mutInfo.氨基酸改变)
		fmt.Printf("\t\t\t\t\t%s\n", mutInfo.碱基改变)
	}
}
