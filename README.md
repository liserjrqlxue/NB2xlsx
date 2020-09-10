# NB2xlsx
安馨可生信注释格式

- [ ] [All variants data](#all-variants-data)
  * [x] [过滤](#过滤)
  * [x] [疾病数据库](#疾病数据库)
  * [x] [已解读数据库](#已解读数据库)
  * [Other Columns](#other-columns)
    - [x] [LOF](#lof)
    - [ ] [遗传模式判读](#遗传模式判读)
- [ ] [excel 格式](#excel-格式)
  * [x] [DataValidation](#datavalidation)

## All variants data

### 过滤
```
一：现在输出到Sheet1的逻辑（也就是下面邮件的第一个要求）更改如下：
满足以下任一一个条件就输出到sheet1
1.     164基因上Clinvar的标签是Pathogenic或者Likely_pathogenic或者Pathogenic/Likely_pathogenic
2.     164基因上HGMD的标签是DM或者DM？或者DM/DM?
3.     164基因上Clinvar/HGMD数据库外GnomAD频率≤0.01，并且变异类型不包括intron、promoter、no-change、UTR区变异
4.     已解读数据库内位点

```
1. 保留已解读数据库内位点
2. 过滤 etc/gene.list.txt 之外的基因
3. "ClinVar Significance" 是 Pathogenic 或者 Likely_pathogenic 或者 Pathogenic/Likely_pathogenic 的保留
4. "HGMD Pred" 是 DM 或者 DM? 或者 DM/DM? 的保留
5. "Function" 不在 etc/function.exclude.txt 中，且 "GnomAD AF" <= 0.01 的保留

### 疾病数据库
```
临床新生儿注释表shee1的Q列“Gene Symbol”与疾病库的C列“基因”相匹配，
匹配上的在sheet1表的BQ列“疾病中文名”输出疾病库的D列“疾病”，sheet1表的BR列“遗传模式”输出疾病库的E列“遗传模式”
```
key1|key2|note
-|-|-
Gene Symbol|基因|main key
疾病中文名|疾病|
遗传模式|遗传模式|

### 已解读数据库
```
第三附件1中的BS列Database，在已解读数据库内并且已解读数据库的DU列是否是包装位点记录为“是”：标记NBS-in
                         在已解读数据库内并且已解读数据库的DU列是否是包装位点无记录：标记NBS-out
                         不在已解读数据库：标记.
第四附件1中的BU列Definition，填写的是已解读数据库中CW列Definition的致病等级
第五附件1中的BW列报告类别，同孕前，数据库内包装的变异（BS列Database标记NBS-in）标记正式报告；数据库外的烈性（LOF列为YES）且低频(GnomAD≤1%，且千人≤1%)：标记补充报告
```
```
sheet1里面的CC列“参考文献”，提取的是已解读数据库中的DM列“Reference”的内容
```

key1|key2|note
-|-|-
Transcript|Transcript|main key 1
cHGVS|cHGVS|main key 2
Definition|Definition|
参考文献|Reference|
Database||NBS-in\|NBS-out\|.
报告类别||正式报告\|补充报告
||是否是包装位点|

### Other Columns
key1|key2|note
1000Gp3 AF|1000G AF|
1000Gp3 EAS AF|1000G EAS AF|
#### LOF
```
第二附件1中的BL列LOF同孕前：nonsense、frameshift、splice-3、splice-5类型且低频(GnomAD≤1%，且千人≤1%)，标记YES，否则标记NO。
```
[`updateLOF`](../367051a760349aac7a4b236ca081340d086c10bd/main.go#L361)
key|value
-|-
LOF|YES\|NO
#### 遗传模式判读
```
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

## excel 格式
### DataValidation
```
（1）CNV sheet的Y列“报告类别”有下拉选项：正式报告和补充报告
（2）CNV sheet的AA列“杂合性”有下拉选项：Hom和Het和Hemi
（3） CNV sheet的AB列“disGroup”有下拉选项：A和B
（4）CNV sheet的AC列”突变类型”有下拉选项：Pathogenic和Likely pathogenic和VUS
（5）补充实验sheet的I列“β地贫_最终结果”有下拉选项：阴性和SEA-HPFH和Chinese和SEA-HPFH；SEA-HPFH和Chinese；Chinese和SEA-HPFH；Chinese
（6）补充实验sheet的J列“α地贫_最终结果”有下拉选项：阴性和3.7和SEA和4.2和THAI和FIL和3.7;3.7和4.2;4.2和SEA;SEA和3.7;4.2和3.7;SEA和3.7;THAI和3.7;FIL和4.2;SEA和 4.2;THAI和4.2;FIL和SEA;THAI和SEA;FIL和THAI;THAI和THAI;FIL和FIL;FIL
（7）补充实验sheet的M列“SMN1 EX7 del最终结果”有下拉选项：阴性和杂合阳性和纯合阳性和杂合灰区和纯合灰区
```
etc/drop.list.txt 包含对应列的下拉表

