package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/360EntSecGroup-Skylar/excelize/v2"
	"github.com/liserjrqlxue/goUtil/simpleUtil"
	"github.com/liserjrqlxue/goUtil/textUtil"
	"github.com/liserjrqlxue/version"
)

type Info struct {
	sampleID   string
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

// os
var (
	ex, _   = os.Executable()
	exPath  = filepath.Dir(ex)
	etcPath = filepath.Join(exPath, "..", "etc")
)

// flag
var (
	anno = flag.String(
		"anno",
		"",
		"anno excel, comma as sep",
	)
	input = flag.String(
		"input",
		"",
		"info dir",
	)
	prefix = flag.String(
		"prefix",
		"",
		"prefix of output, output -prefix.-tag.xlsx",
	)
	tag = flag.String(
		"tag",
		time.Now().Format("2006-01-02"),
		"tag of output, default is date[2006-01-02]",
	)
)

var 患病风险 map[string]string

var sampleCount = make(map[string]int)

func main() {
	version.LogVersion()
	flag.Parse()
	if *anno == "" || *input == "" {
		flag.Usage()
		log.Printf("-anno/-input required!")
		os.Exit(1)
	}
	if *prefix == "" {
		*prefix = *input
	}

	患病风险 = simpleUtil.HandleError(textUtil.File2Map(filepath.Join(etcPath, "患病风险.txt"), "\t", false)).(map[string]string)

	var outExcel = simpleUtil.HandleError(excelize.OpenFile(*input)).(*excelize.File)
	var db = make(map[string]Info)
	// 读解读表
	for n, annoFile := range strings.Split(*anno, ",") {
		log.Printf("信息：开始读取解读表%d[%s]\n", n+1, annoFile)
		var annoExcel = simpleUtil.HandleError(excelize.OpenFile(annoFile)).(*excelize.File)

		getInfoFromAVD(db, annoExcel)
		getInfoFromCNV(db, annoExcel)
		getInfoFromAE(db, annoExcel)
	}

	// load sample info
	var sheetName = "159基因结果汇总"

	var strSlice = simpleUtil.HandleError(
		simpleUtil.HandleError(
			excelize.OpenFile(*input),
		).(*excelize.File).GetRows(sheetName),
	).([][]string)

	var colLength = 105
	var titleRowIndex = 4
	var sampleIDCIdx = 4
	var resultCIdx = 15
	var resultTitle = "基因检测结果拟诊疾病（疾病/风险）"
	var geneLimit = 6
	var geneNameCIdx = resultCIdx + geneLimit
	var geneColCount = 4
	var mutLit = 2
	var mutColCount = 5
	var geneColLength = geneColCount + mutLit*mutColCount

	var appendColName = simpleUtil.HandleError(excelize.ColumnNumberToName(colLength + 1)).(string)
	simpleUtil.CheckErr(
		outExcel.SetColWidth(
			sheetName,
			appendColName,
			appendColName,
			10,
		),
	)
	simpleUtil.CheckErr(
		outExcel.SetColStyle(
			sheetName,
			appendColName,
			simpleUtil.HandleError(outExcel.NewStyle(`{"fill":{"type":"pattern","color":["#E0EBF5"],"pattern":1}, "number_format": 14}`)).(int),
		),
	)

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
			checkTitleName(strArray, "华大基因检测编号", rIdx+1, sampleIDCIdx)
			checkTitleName(strArray, "疾病一", rIdx+1, resultCIdx)
			checkTitleName(strArray, "外显子", rIdx+1, geneNameCIdx+1)
			checkTitleName(strArray, "碱基改变", rIdx+1, geneNameCIdx+2)
			checkTitleName(strArray, "氨基酸改变", rIdx+1, geneNameCIdx+3)
			continue
		}
		rIdx++
		var index = rIdx - titleRowIndex
		var sampleID = strArray[sampleIDCIdx-1]
		if len(strArray) > colLength && strArray[colLength] != "" {
			log.Printf("警告：样品%3d[%11s]已有时间戳，跳过\n", index, sampleID)
			continue
		}
		var item, ok = db[sampleID]
		if !ok {
			log.Printf("警告：样品%3d[%11s]无解读信息，跳过\n", index, sampleID)
			continue
		}
		if sampleCount[sampleID] > 1 {
			log.Printf("警告：样品%3d[%11s]解读信息重复%d次\n", index, sampleID, sampleCount[sampleID])
		}
		// 按患病风险排序基因
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

		for _, geneSymbol := range item.geneList {
			var geneInfo = item.geneMap[geneSymbol]
			item.基因 = append(item.基因, geneInfo)
			item.基因检测结果 = append(item.基因检测结果, geneInfo.疾病)
		}

		// 填充 基因检测结果
		for i, 疾病 := range item.基因检测结果 {
			if i > geneLimit {
				log.Printf("警告：样品%3d[%11s]检出疾病超出%d个，疾病[%s]跳过\n", index, sampleID, geneLimit, 疾病)
				break
			}
			WriteCellStr(outExcel, sheetName, resultCIdx+i, rIdx, 疾病)
		}
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
			fmt.Printf("\t\t\t%8s\t%s\n", geneInfo.基因名称, geneInfo.患病风险)
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
		WriteCellValue(outExcel, sheetName, colLength+1, rIdx, time.Now().UTC())
		log.Printf("信息：样品%3d[%11s]已写入\n", index, sampleID)
	}
	var outputPath = fmt.Sprintf("%s.%s.xlsx", *prefix, *tag)
	simpleUtil.CheckErr(outExcel.SaveAs(outputPath))
	log.Printf("信息：保存到[%s]\n", outputPath)
}

func getAfterVerticalbar(str string) string {
	var s = strings.Split(str, "|")
	return strings.TrimSpace(s[len(s)-1])
}
