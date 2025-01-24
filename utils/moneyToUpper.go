package utils

import (
	"math"
	"regexp"
	"strconv"
)

func AmountConvert(pMoney float64, pRound bool) string {
	var NumberUpper = []string{"壹", "贰", "叁", "肆", "伍", "陆", "柒", "捌", "玖", "零"}
	var Unit = []string{"分", "角", "圆", "拾", "佰", "仟", "万", "拾", "佰", "仟", "亿", "拾", "佰", "仟"}
	var regex = [][]string{
		{"零拾", "零"}, {"零佰", "零"}, {"零仟", "零"}, {"零零零", "零"}, {"零零", "零"},
		{"零角零分", "整"}, {"零分", "整"}, {"零角", "零"}, {"零亿零万零元", "亿元"},
		{"亿零万零元", "亿元"}, {"零亿零万", "亿"}, {"零万零元", "万元"}, {"万零元", "万元"},
		{"零亿", "亿"}, {"零万", "万"}, {"拾零圆", "拾元"}, {"零圆", "元"}, {"零零", "零"}}
	str, DigitUpper, UnitLen, round := "", "", 0, 0
	if pMoney == 0 {
		return "零"
	}
	if pMoney < 0 {
		str = "负"
		pMoney = math.Abs(pMoney)
	}
	if pRound {
		round = 2
	} else {
		round = 1
	}
	digitByte := []byte(strconv.FormatFloat(pMoney, 'f', round+1, 64)) //注意币种四舍五入
	UnitLen = len(digitByte) - round

	for _, v := range digitByte {
		if UnitLen >= 1 && v != 46 {
			s, _ := strconv.ParseInt(string(v), 10, 0)
			if s != 0 {
				DigitUpper = NumberUpper[s-1]

			} else {
				DigitUpper = "零"
			}
			str = str + DigitUpper + Unit[UnitLen-1]
			UnitLen = UnitLen - 1
		}
	}
	for i := range regex {
		reg := regexp.MustCompile(regex[i][0])
		str = reg.ReplaceAllString(str, regex[i][1])
	}
	if string(str[0:3]) == "元" {
		str = str[3:]
	}
	if string(str[0:3]) == "零" {
		str = str[3:]
	}
	return str
}
