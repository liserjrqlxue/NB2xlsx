package main

import "fmt"

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

// 按患病风险排序基因
func (item Info) sortGene() {
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

func (item Info) resumeGene() {
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
