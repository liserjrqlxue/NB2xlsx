package main

func geneCnv(item map[string]string) (cnv, cnv0 map[string]bool) {
	cnv = make(map[string]bool)
	cnv0 = make(map[string]bool)
	return
}

var LOFofPLP = map[string]bool{
	"nonsense":   true,
	"frameshift": true,
	"stop-gain":  true,
	"span":       true,
	"altstart":   true,
	"init-loss":  true,
	"splice-3":   true,
	"splice-5":   true,
}

func isPLP(item map[string]string) bool {
	if item["Definition"] == "P" || item["Definition"] == "LP" {
		return true
	}
	if LOFofPLP[item["Function"]] {
		return true
	}
	if isClinVar[item["ClinVar Significance"]] {
		return true
	}
	if notClinVar[item["ClinVar Significance"]] {
		return false
	}
	if isHGMD[item["HGMD Pred"]] {
		return true
	}
	return false
}

type GeneInfo struct {
	基因               string
	遗传模式             string
	性别               string
	PLP, hetPLP, VUS int
	cnv, cnv0        bool
	tag3, tag4       bool
}

func (info *GeneInfo) new(item map[string]string) *GeneInfo {
	info.基因 = item["Gene Symbol"]
	info.遗传模式 = geneInheritance[info.基因]
	info.count(item)
	return info
}

func (info *GeneInfo) count(item map[string]string) {
	if item["自动化判断"] == "VUS" {
		info.VUS++
	}
	if isPLP(item) {
		item["P/LP*"] = "1"
		info.PLP++
		if item["Zygosity"] == "Het" {
			info.hetPLP++
		}
	}
}

func (info *GeneInfo) getTag(item map[string]string) (tag string) {
	tag += 标签1(item, info)
	tag += 标签2(item, info)
	tag += 标签3(item, info)
	tag += 标签4(item, info)
	tag += 标签5(item, info)
	return
}

func 标签1(item map[string]string, geneInfo *GeneInfo) string {
	var (
		遗传模式     = geneInfo.遗传模式
		性别       = geneInfo.性别
		PLP      = geneInfo.PLP
		hetPLP   = geneInfo.hetPLP
		cnv      = geneInfo.cnv
		VUS      = geneInfo.VUS
		自动化判断    = item["自动化判断"]
		Zygosity = item["Zygosity"]
	)
	if 遗传模式 == "AD" || 遗传模式 == "AD,AR" || 遗传模式 == "AD,SMu" || 遗传模式 == "Mi" || ((遗传模式 == "XL" || 遗传模式 == "YL") && 性别 == "M") {
		if item["P/LP*"] == "1" {
			return "1"
		}
	}
	if 遗传模式 == "AR" || 遗传模式 == "AR/AR" || (遗传模式 == "XL" && 性别 == "F") {
		if item["P/LP*"] != "1" {
			if 自动化判断 == "VUS" && hetPLP >= 1 {
				return "1"
			}
			return ""
		} else {
			if Zygosity == "Hom" {
				return "1"
			}
			if PLP > 1 {
				return "1"
			}
			switch VUS {
			case 0:
				if cnv {
					return "1"
				}
			case 1:
				if 自动化判断 != "VUS" {
					return "1"
				}
			default:
				return "1"
			}
		}
	}
	return ""
}

var (
	spliceList = map[string]bool{
		"splice+10": true,
		"splice-10": true,
		"splice+20": true,
		"splice-20": true,
	}
	cdsList = map[string]bool{
		"cds-del":   true,
		"cds-ins":   true,
		"cds-indel": true,
		"stop-loss": true,
	}
	af0List = map[string]bool{
		"ESP6500 AF":    true,
		"1000G AF":      true,
		"ExAC AF":       true,
		"ExAC EAS AF":   true,
		"GnomAD AF":     true,
		"GnomAD EAS AF": true,
	}
)

func 标签2(item map[string]string, geneInfo *GeneInfo) string {
	if item["自动化判断"] != "VUS" {
		return ""
	}
	var (
		遗传模式     = geneInfo.遗传模式
		性别       = geneInfo.性别
		VUS      = geneInfo.VUS
		function = item["Function"]
	)
	if 遗传模式 == "AD" || 遗传模式 == "AD,AR" || 遗传模式 == "AD,SMu" || 遗传模式 == "Mi" || ((遗传模式 == "XL" || 遗传模式 == "YL") && 性别 == "M") {
		if item["P/LP*"] != "2" {
			return ""
		}
		if 遗传模式 == "AD" || 遗传模式 == "AD,AR" || 遗传模式 == "AD,SMu" || 遗传模式 == "XL" || 遗传模式 == "YL" {
			for af := range af0List {
				if gt(item[item[af]], 2e-5) {
					return ""
				}
			}
		}
		if cdsList[function] && item["RepeatTag"] == "" {
			return "2"
		}
		if spliceList[function] {
			if item["SpliceAI Pred"] == "D" {
				return "2"
			}
		} else {
			if item["PP3"] == "1" {
				return "2"
			}
		}
	}
	if 遗传模式 == "AR" || 遗传模式 == "AR/AR" || (遗传模式 == "XL" && 性别 == "F") {
		if item["Zygosity"] == "Hom" || VUS > 1 {
			if cdsList[function] && item["RepeatTag"] == "" {
				return "2"
			}
			if spliceList[function] {
				if item["SpliceAI Pred"] == "D" {
					return "2"
				}
			} else {
				if item["PP3"] == "1" {
					return "2"
				}
			}
		}
	}
	return ""
}

func 标签3(item map[string]string, geneInfo *GeneInfo) string {
	if item["VUS"] == "0" {
		return ""
	}
	var (
		遗传模式     = geneInfo.遗传模式
		性别       = geneInfo.性别
		cnv      = geneInfo.cnv
		function = item["Function"]
	)
	if !cnv {
		return ""
	}
	if 遗传模式 == "AR" || 遗传模式 == "AR/AR" || (遗传模式 == "XL" && 性别 == "F") {
		if cdsList[function] && item["RepeatTag"] == "" {
			geneInfo.tag3 = true
			return "3"
		}
		if spliceList[function] {
			if item["SpliceAI Pred"] == "D" {
				geneInfo.tag3 = true
				return "3"
			}
		} else {
			if item["PP3"] == "1" {
				geneInfo.tag3 = true
				return "3"
			}
		}
	}
	return ""
}

func 标签4(item map[string]string, geneInfo *GeneInfo) string {
	var (
		遗传模式 = geneInfo.遗传模式
		性别   = geneInfo.性别
		cnv  = geneInfo.cnv
		cnv0 = geneInfo.cnv0
	)
	if 遗传模式 == "AD" || 遗传模式 == "AD,AR" || 遗传模式 == "AD,SMu" || 遗传模式 == "Mi" || ((遗传模式 == "XL" || 遗传模式 == "YL") && 性别 == "M") {
		if cnv {
			geneInfo.tag4 = true
			return "4"
		}
	}
	if 遗传模式 == "AR" || 遗传模式 == "AR/AR" || (遗传模式 == "XL" && 性别 == "F") {
		if cnv0 {
			geneInfo.tag4 = true
			return "4"
		}
	}
	return ""
}

func 标签5(item map[string]string, geneInfo *GeneInfo) string {
	if item["P/LP*"] == "0" || item["Zygosity"] != "Het" {
		return ""
	}
	var (
		遗传模式 = geneInfo.遗传模式
		性别   = geneInfo.性别
		VUS  = geneInfo.VUS
		PLP  = geneInfo.PLP
		cnv  = geneInfo.cnv
	)
	if 遗传模式 == "AR" || 遗传模式 == "AR/AR" || (遗传模式 == "XL" && 性别 == "F") {
		if PLP == 1 && VUS == 0 && !cnv {
			return "5"
		}
	}
	return ""
}
