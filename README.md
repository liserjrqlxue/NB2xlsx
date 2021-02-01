# NB2xlsx
[![GoDoc](https://godoc.org/github.com/liserjrqlxue/NB2xlsx?status.svg)](https://pkg.go.dev/github.com/liserjrqlxue/NB2xlsx) 
[![Go Report Card](https://goreportcard.com/badge/github.com/liserjrqlxue/NB2xlsx)](https://goreportcard.com/report/github.com/liserjrqlxue/NB2xlsx)

安馨可生信注释格式

- [ ] [All variants data](#all-variants-data)
  * [x] [过滤](#过滤)
  * [X] [标签](#标签)
  * [x] [疾病数据库](#疾病数据库)
  * [x] [已解读数据库](#已解读数据库)
  * [ ] [Other Columns](#other-columns)
    - [x] [LOF](#lof)
    - [ ] [遗传模式判读](#遗传模式判读)
- [x] [lims.info](#limsinfo)
- [ ] [QC](#qc)
  - [x] [common](#common)
  - [ ] [others](#others)
- [ ] [excel 格式](#excel-格式)
  * [x] [DataValidation](#datavalidation)
  * [ ] [Background Color](#background-color)

## All variants data

### 过滤
```text
三 输出到解读表位点调整
满足以下任一一个条件就输出到sheet1：
1. 161基因上Clinvar的标签是Pathogenic或者Likely_pathogenic或者Pathogenic/Likely_pathogenic或者HGMD的标签是DM或者DM？或者DM/DM? 并且ESP6500/1000G/ExAC和GnomeAD总人频以及ExAC的East Asian、GnomAD的East Asian频率≤0.05，去除自动化致病等级是B，LB的变异（保留clinvar是P，LP，P/LP的位点）。
2. 161基因上Clinvar/HGMD数据库外ESP6500/1000G/ExAC和GnomeAD总人频以及ExAC的East Asian、GnomAD的East Asian的频率≤0.01，并且变异包括missense/nonsense/frameshift/cds-ins/cds-del/coding-synon/init-loss/ncRNA/splice-3/splice-5/20位以内的intron/SpliceAI预测影响剪切的ESP6500/1000G/ExAC和GnomeAD总人频以及ExAC的East Asian、GnomAD的East Asian的频率≤0.01的intron
3. 已解读数据库内位点
```
```text
第三条的具体条件如下：

（1）保留已解读数据库内位点
（2）过滤 etc/gene.list.txt 之外的基因
（3）过滤 "ESP6500 AF/1000G AF/ExAC AF/GnomeAD AF/ExAC EAS AF、GnomAD EAS AF" > 0.05
（4）过滤自动化致病等级是B，LB的变异（不包括clinvar是P，LP，P/LP的位点）。
（5）"ClinVar Significance" 是 Pathogenic 或者 Likely_pathogenic 或者 Pathogenic/Likely_pathogenic 的保留
（6）"HGMD Pred" 是 DM 或者 DM? 或者 DM/DM? 的保留
（7）"Function"列为 missense/nonsense/frameshift/cds-ins/cds-del/cds-indel/stop-loss/span/altstart/coding-synon/init-loss/ncRNA/splice-3/splice-5/splice+10/splice-10/splice+20/splice-20，且 "ESP6500 AF/1000G AF/ExAC AF/GnomeAD AF/ExAC EAS AF、GnomAD EAS AF" <= 0.01 的保留
（8）"Function"列为intron且SpliceAI预测影响剪切且"ESP6500 AF/1000G AF/ExAC AF/GnomeAD AF/ExAC EAS AF、GnomAD EAS AF" <= 0.01 的保留
```
1.  **保留** 已解读数据库内位点
2.  **过滤** "Gene Symbol" no in `etc/gene.list.txt`
3.  **定义** 频率列表 `avdAfList` ["ESP6500 AF","1000G AF","ExAC AF","GnomAD AF","ExAC EAS AF","GnomAD EAS AF",]
4.  **过滤** `avdAfList` > 0.05
5.  **保留** "ClinVar Significance" in ["Pathogenic","Likely_pathogenic","Pathogenic/Likely_pathogenic"]
6.  **过滤** "ACMG" in ["B","LB"]
7.  **保留** "HGMD Pred" in ["DM","DM?","DM/DM?"]
8.  **过滤** `avdAfList` > 0.01
9.  **保留** "Function" = "intron" 且 "SpliceAI Pred" = "D"
10. **过滤** "Function" in `etc/function.exclude.txt`
11. **保留** 剩余

### 标签

|遗传模式|P/LP*|compositeP|Zygosity|Function|自动化判断|Definition|ClinVar| HGMD |lowAF|VarCount| CNV |标签|CNV标签|
|-------|-----|----------|--------|--------|---------|----------|-------|------|-----|---------|----|----|-------|
|     AD|true |          |        |        |         |          |       |      |true |         |    |   1|       |
|     AR|true |          |Hom     |        |         |          |       |      |     |         |    |   1|       |
|     AR|true |          |Het     |        |         |          |       |      |     |LPL>1    |    |   1|       |
|     AR|true |          |Het     |        |         |          |       |      |     |PLPVUS>1 |    |   1|       |
|     AR|true |          |Het     |        |!PLPVUS  |          |       |      |     |PLP==1   |    |   1|       |
|     AR|false|          |        |        |VUS      |          |       |      |     |hetPLP==1|    |   1|       |
|     AD|false|true      |        |        |         |PLPVUS    |       |      |true |         |    |   2|       |
|     AR|     |true      |Hom     |        |VUS      |          |       |      |     |         |    |   2|       |
|     AR|     |true      |        |        |VUS      |          |       |      |     |HpVUS>1  |    |   2|       |
|     AR|true |          |        |        |         |          |       |      |     |VUS==0   |cnv |   3|3      |
|     AR|     |true      |        |        |VUS      |          |       |      |     |         |cnv |   3|3      |
|     AD|     |          |        |        |         |          |       |      |     |         |cnv |    |4      |
|     AR|     |          |        |        |         |          |       |      |     |         |cnv0|    |4      |
|       |     |          |        |        |         |P/LP      |       |      |     |         |    |   5|       |
|       |     |          |        |LoF     |         |          |       |      |     |         |    |   5|       |
|       |     |          |        |        |         |          |P/LP   |      |     |         |    |   5|       |
|       |     |          |        |        |         |          |!B/LB  |DM/DM?|     |         |    |   5|       |

### 疾病数据库
```text
临床新生儿注释表shee1的Q列“Gene Symbol”与疾病库的C列“基因”相匹配，
匹配上的在sheet1表的BQ列“疾病中文名”输出疾病库的D列“疾病”，sheet1表的BR列“遗传模式”输出疾病库的E列“遗传模式”
```
```text
模板excel的CD列“疾病简介“疾病库中的的“疾病简介”
```
key1|key2|note
-|-|-
Gene Symbol|基因|main key
疾病中文名|疾病|
遗传模式|遗传模式|
疾病简介|疾病简介|

### 已解读数据库
```text
第三附件1中的BS列Database，在已解读数据库内并且已解读数据库的DU列是否是包装位点记录为“是”：标记NBS-in
                         在已解读数据库内并且已解读数据库的DU列是否是包装位点无记录：标记NBS-out
                         不在已解读数据库：标记.
第五附件1中的BW列报告类别，同孕前，数据库内包装的变异（BS列Database标记NBS-in）标记正式报告；数据库外的烈性（LOF列为YES）且低频(GnomAD≤1%，且千人≤1%)：标记补充报告
```
```text
sheet1里面的CC列“参考文献”，提取的是已解读数据库中的DM列“Reference”的内容
```
```shell script
2. 模板excel的CE列“位点关联疾病”匹配位点数据库CU列“Disease“
3. 模板excel的CF列“位点关联遗传模式“匹配位点数据库CV列” 遗传模式“
4. 模板excel的CG列“Evidence New + Check“匹配位点数据库DU列” 证据项“
5. 模板excel的CH列“Definition“匹配位点数据库CX列” Definition“
6. 位点数据库中的DV列“是否是包装位点“为“是”的在注释表中是正式报告
```

key1|key2|note
-|-|-
Transcript|Transcript|main key 1
cHGVS|cHGVS|main key 2
参考文献|Reference|
位点关联疾病|Disease|
位点关联遗传模式|遗传模式|
Evidence New + Check|证据项|
Definition|Definition|
Database||NBS-in\|NBS-out\|.
报告类别||正式报告\|补充报告
||是否是包装位点|

### Other Columns
key1|key2|note
-|-|-
ClinVar星级|ClinVar Number of gold stars|
1000Gp3 AF|1000G AF|
1000Gp3 EAS AF|1000G EAS AF|
引物设计|anno.PrimerDesign(item)|

#### LOF
```text
第二附件1中的BL列LOF同孕前：nonsense、frameshift、splice-3、splice-5类型且低频(GnomAD≤1%，且千人≤1%)，标记YES，否则标记NO。
```
[`updateLOF`](../367051a760349aac7a4b236ca081340d086c10bd/main.go#L361)
key|value
-|-
LOF|['YES','NO']
#### 遗传模式判读
```text
 遗传模式判读列输出两种：携带者和可能患病
1、输出“可能患病”有以下情况：
1) 基因与疾病遗传方式AR/AR;AR，检出单个基因1个致病变异纯合突变。
“遗传方式”列为AR/AR;AR，“基因”列检出1个致病变异，“杂合性”列为hom。
(备注：AR;AR是指一个基因对应两种疾病，两种疾病遗传方式均为AR)
2) 基因与疾病遗传方式AR，检出单个基因≥2个致病变异杂合突变。
“遗传方式”列为AR，“基因”列检出≥2个致病变异，“杂合性”列为het。
3) 基因与疾病遗传方式AD或者AD,AR，检出单个基因1个致病变异纯合突变或者杂合突变。
“遗传方式”列为AD或者AD,AR，“基因”列检出1个致病变异，“杂合性”列为hom或het。
4) 基因与疾病遗传方式XL，男性检出单个基因1个致病变异半合突变；女性检出单个基因1个致病变异纯合突变或者杂合突变。
男性：“遗传方式”列为XL，“基因”列检出1个致病变异，“杂合性”列为hemi。
女性：
女性：“遗传方式”列为XL，“基因”列检出1个致病变异，“杂合性”列为hom或het。-
5) 遗传方式Maternal Inheritance，单个基因上检出1个致病变异同质性突变或者异质性突变。
“遗传方式”列为Maternal Inheritance，“基因”列检出1个致病变异，“杂合性”列为hom或者het。
2、输出“携带者”有以下情况：
1）基因与疾病遗传方式AR，检出单个基因1个致病变异het。
“遗传方式”列为AR，“基因”列检出1个致病变异，“杂合性”列为het。
2）基因与疾病遗传方式AR;AR，检出单个基因1个致病变异het。
“遗传方式”列为AR;AR，“基因”列检出1个致病变异，“杂合性”列为het。
```

遗传模式|杂合性|个数|样品性别|遗传模式判读
-|-|-|-|-
['AR','AR/AR']|Hom|>=1||可能患病
['AR','AR/AR']|Het|=1||携带者
['AR','AR/AR']|Het|>1||可能患病
['AD','AD,AR']|['Hom','Het]|>=1||可能患病
['XL']|Hemi|>=1|Male|可能患病
['XL']|['Hom','Het']|>=1|Female|可能患病

## lims.info

key1|key2|note
-|-|-
SampleID|MAIN_SAMPLE_NUM|main key
期数|HYBRID_LIBRARY_NUM|
flow ID|FLOW_ID|
产品编码|PRODUCT_CODE|
产品名称|PRODUCT_NAME|
产品编码_产品名称||产品编码+'_'+产品名称

## QC
```text
1. Order为序号：1，2，3，。。。
2. Sample列对应下机QC的 “Sample”列
3. Q20列对应下机QC的“Q20”
4. Q30列对应下机QC的”Q30”
5. AverageDepth对应下机QC的“Target Mean Depth[RM DUP]:”列
6. Depth>=10(%)对应下机QC的“Target coverage >=10X percentage[RM DUP]:”列
7. Coverage(%)对应下机QC的“Target coverage[RM DUP]:“列
8. GC(%)对应下机QC的“Total mapped GC Rate:“
9. Target coverage >=20X percentage对应下机QC的“Target coverage >=20X percentage[RM DUP]:“列
10. mitochondria Target Mean Depth[RM DUP]对应下机QC的“Target Mean Depth[RM DUP]: (mitochondria)“列
11. Gender为性别
12. RESULT为C-K列的一个判断，比如提示性别不一致，GC含量高，如果都合格就输出“YES“
13. 解读人和审核人为空
14. 产品编号对应临床新生儿的产品编号DX1968，多中心是DX1964
```
```text
11. Gender为性别，取下机QC的“gender_analysed”列
12. RESULT为C-K列的一个判断，比如提示性别不一致，GC含量高，如果都合格就输出“YES“
详细规则见下面玉婧的邮件
14. 产品编号对应临床新生儿的产品编号DX1968，多中心是DX1964
如微信沟通的方式实现
```
### common
[`etc/QC.txt`](etc/QC.txt)

### others
title|key|note
-|-|-
Order|i+1|index+1
产品编号|lims["PRODUCT_CODE"]|from `lims.info`
RESULT|RESULT|


## excel 格式
### DataValidation
```text
（1）CNV sheet的Y列“报告类别”有下拉选项：正式报告和补充报告
（2）CNV sheet的AA列“杂合性”有下拉选项：Hom和Het和Hemi
（3） CNV sheet的AB列“disGroup”有下拉选项：A和B
（4）CNV sheet的AC列”突变类型”有下拉选项：Pathogenic和Likely pathogenic和VUS
（5）补充实验sheet的I列“β地贫_最终结果”有下拉选项：阴性和SEA-HPFH和Chinese和SEA-HPFH；SEA-HPFH和Chinese；Chinese和SEA-HPFH；Chinese
（6）补充实验sheet的J列“α地贫_最终结果”有下拉选项：阴性和3.7和SEA和4.2和THAI和FIL和3.7;3.7和4.2;4.2和SEA;SEA和3.7;4.2和3.7;SEA和3.7;THAI和3.7;FIL和4.2;SEA和 4.2;THAI和4.2;FIL和SEA;THAI和SEA;FIL和THAI;THAI和THAI;FIL和FIL;FIL
（7）补充实验sheet的M列“SMN1 EX7 del最终结果”有下拉选项：阴性和杂合阳性和纯合阳性和杂合灰区和纯合灰区
```
etc/drop.list.txt 包含对应列的下拉表

### Background Color
```text
4. 需要验证位点的一行字体标记红色
5. 正式报告的位点一行标记蓝色底纹，补充报告位点一行标记绿色底纹
```

### HyperLink
```text
2.1Sheet1 All variants data的内容如下
2.	“reads_picture”需要链接reads图，需要链接的位点要同时满足以下2个条件：
（1）正式报告或者补充报告或者clinvar收录是P,LP,P/LP或者HGMD收录DM,DM?,DM/DM?或者库内解读过的
（2）SNV:depth≤40或者A.Ratio≤0.4；Indel:depth≤60或者A.Ratio≤0.45
```
```text
2.3 Sheet3 补充实验的内容如下
1.	“β地贫_最终结果”“α地贫_最终结果”CNE图链接的添加
```
```go
item["HyperLink"] = filepath.Join(*batch+".result_batCNV-dipin", "chr11_chr16_chrX_cnemap", item["SampleID"]+"_W30S25_cne.jpg")
item["β地贫_最终结果_HyperLink"] = item["HyperLink"]
item["α地贫_最终结果_HyperLink"] = item["HyperLink"]
```
