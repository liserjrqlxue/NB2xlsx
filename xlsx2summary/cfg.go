package main

var (
	colLength     = 105
	titleRowIndex = 4
	sampleIDCIdx  = 4
	sampleIDTitle = "华大基因检测编号"
	summaryCIdx   = 14
	summaryTitle  = "基因检测结果总结"
	resultCIdx    = 15
	resultTitle   = "基因检测结果拟诊疾病（疾病/风险）"
	geneLimit     = 6
	geneNameCIdx  = resultCIdx + geneLimit
	geneColCount  = 4
	mutLit        = 2
	mutColCount   = 5
	geneColLength = geneColCount + mutLit*mutColCount
)
