package main

type DrugResult struct {
	样本编号 string
	药物分类 string
	药物名称 string
	英文名  string
	用药建议 string
	结果说明 string
}

type DrugGene struct {
	检测基因   string
	代谢型    string
	阳性结果说明 string
}

type DrugVar struct {
	检测位点, 影响类型, 证据等级, 参考基因型, Ref, Var string
	检测结果                              string
	分条结果说明                            string
	分条用药建议                            string
	参考来源                              string
}
