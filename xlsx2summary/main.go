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
	simpleUtil.CheckErr(
		outExcel.SetColWidth(
			sheetName,
			"BM",
			"BM",
			10,
		),
	)
	simpleUtil.CheckErr(
		outExcel.SetColStyle(
			sheetName,
			"BM",
			simpleUtil.HandleError(outExcel.NewStyle(`{"fill":{"type":"pattern","color":["#E0EBF5"],"pattern":1}, "number_format": 14}`)).(int),
		),
	)
	var strSlice = simpleUtil.HandleError(
		simpleUtil.HandleError(
			excelize.OpenFile(*input),
		).(*excelize.File).GetRows(sheetName),
	).([][]string)

	var titleRowIndex = 4
	var colLength = 64
	var sampleIDcIdx = 4
	for rIdx, strArray := range strSlice {
		if rIdx < titleRowIndex-1 {
			continue
		} else if rIdx == titleRowIndex-1 {
			if strArray[sampleIDcIdx-1] != "华大基因检测编号" {
				log.Fatalf("错误：表头(%d,%d)[%s]！=%s\n", titleRowIndex, sampleIDcIdx, strArray[sampleIDcIdx-1], "华大基因检测编号")
			}
			continue
		}
		rIdx++
		var index = rIdx - titleRowIndex
		var sampleID = strArray[sampleIDcIdx-1]
		if len(strArray) > colLength && strArray[colLength] != "" {
			log.Printf("警告：样品%3d[%11s]已有时间戳，跳过\n", index, sampleID)
			continue
		}
		var item, ok = db[sampleID]
		if !ok {
			log.Printf("警告：样品%3d[%11s]无解读信息，跳过\n", index, sampleID)
			continue
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
		for i, v := range item.基因检测结果 {
			WriteCellStr(outExcel, sheetName, 10+i, rIdx, v)
		}
		fmt.Printf("\t%s\n", sampleID)
		for i, geneInfo := range item.基因 {
			if i > 4 {
				log.Printf("警告：样品%3d[%11s]检出基因超出5个，基因[%s]跳过\n", index, sampleID, geneInfo.基因名称)
				break
			}
			WriteCellStr(outExcel, sheetName, 15+i*10, rIdx, geneInfo.基因名称)
			WriteCellStr(outExcel, sheetName, 22+i*10, rIdx, geneInfo.疾病)
			WriteCellStr(outExcel, sheetName, 23+i*10, rIdx, geneInfo.患病风险)
			WriteCellStr(outExcel, sheetName, 24+i*10, rIdx, geneInfo.遗传方式)
			fmt.Printf("\t\t\t%8s\t%s\n", geneInfo.基因名称, geneInfo.患病风险)
			for j, mutInfo := range geneInfo.变异 {
				if j > 1 {
					log.Printf("警告：样品%3d[%11s]在基因[%s]检出变异超出2个，变异[%s]跳过", index, sampleID, geneInfo.基因名称, mutInfo.碱基改变)
					break
				}
				WriteCellStr(outExcel, sheetName, 16+i*10+j*3, rIdx, mutInfo.外显子)
				WriteCellStr(outExcel, sheetName, 17+i*10+j*3, rIdx, mutInfo.碱基改变)
				WriteCellStr(outExcel, sheetName, 18+i*10+j*3, rIdx, mutInfo.氨基酸改变)
				fmt.Printf("\t\t\t\t\t%s\n", mutInfo.碱基改变)
			}
		}
		WriteCellValue(outExcel, sheetName, 65, rIdx, time.Now().UTC())
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
