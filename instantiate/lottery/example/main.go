package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"regexp"
	"strings"

	"anys/config"
	"anys/instantiate/lottery"
	"anys/instantiate/lottery/model"
	"anys/log"
	"anys/pkg/db"
	"anys/pkg/utils"
	"time"
)

var test_data = []struct {
	funcname   string
	code       string
	modes      string
	omodel     string
	totalprice string
	multiple   string
}{
	// 时时彩
	// {"PrevTwoNum", "1|1|*|*|*;0,1,2,3,4,5,6,7,8,9|0,1,2,3,4,5,6,7,8,9|*|*|*", "1", "1950", "60", "1"},
	// {"OneNum", "0,1,2,3,4,5,6,7,8,9|*|*|*|*", "1", "1950", "60", "1"},
	// {"PrevThreeNum", "1|1|2|*|*;0,1,2,3,4,5,6,7,8,9|0,1,2,3,4,5,6,7,8,9|3|*|*", "1", "1950", "60", "1"},
	// {"PrevThreeTwoNum", "2,3,8", "1", "1950", "60", "1"},
	// {"PrevThreeOneNum", "3,1", "1", "1950", "60", "1"},
	// {"LastThreeTwoNum", "0,2,6,7,8", "1", "1950", "60", "1"},
	// {"AllComNum", "0,3|3|3|3|2;4|3|3|3|2", "1", "1950", "60", "1"},

	// {"All120Num", "0,1,2,3,4,5;1,1,2,3,4", "1", "1950", "60", "1"},
	// {"All60Num", "1,2|1,3,4,5;1,2|1,3,4", "1", "1950", "60", "1"},
	// {"All30Num", "1,2,3|1;1,2,3|4", "1", "1950", "60", "1"},
	// {"All20Num", "3|1,0,4;3|1,2,4;", "1", "1950", "60", "1"},
	// {"All10Num", "3|1;3|1,2,4;", "1", "1950", "60", "1"},
	// {"All5Num", "3|1,2,4", "1", "1950", "60", "1"},

	// {"FourNum", "2|3|3|5", "1", "1950", "60", "1"},
	// {"FourComNum", "2|3|3|5", "1", "1950", "60", "1"},
	{"Four24Num", "0,1,2,3,4,5,6,7,8,9", "1", "1950", "60", "1"},
	// {"Four12Numerical", "0,1,2,3,4,5,6,7,8,9|0,1,2,3,4,5,6,7,8,9", "1", "1950", "60", "1"},
	// {"Four6Numerical", "3,4", "1", "1950", "60", "1"},
	// {"Four4Numerical", "3|1,2,4", "1", "1950", "60", "1"},

	// {"PrevTwoSizeOdd", "1|1", "1", "1950", "60", "1"},
	// {"EverythingGood", "1,3", "1", "1950", "60", "1"},
	// {"PairGood", "1,3,5,7,8", "1", "1950", "60", "1"},
	// {"ThreeGood", "0,1,2,3,4,5,6,7,8,9", "1", "1950", "60", "1"},
	// {"FourGood", "3", "1", "1950", "60", "1"},

	// {"PrevMixThreeNum", "213", "1", "1950", "60", "1"},
	// {"AllNumerical", "0,1,2,3,4,5,6|0,1,2,3,4,5,6|0,1,2,3,4,5,6|0,1,2,3,4,5,6|0,1,2,3,4,6", "1", "1950", "60", "1"},
	// {"All20Numerical", "0,1,2,3,4,5,6,7,8,9|0,1,2,3,4,5,6,7,8,9", "1", "1950", "60", "1"},
	// {"PrevThreeSumCom", "21", "1", "1950", "60", "1"},
	// {"LastChooseThreeNumerical", "0,1,2;0,1,2,3,4,5,6,7,8,9", "1", "1950", "60", "1"},
	// {"PrevThreeSixNumerical", "0,1,2,3,4,5,6,7,8,9", "1", "1950", "60", "1"},
	// {"PrevTwoTwoNumerical", "0,1,2,3,4,5,6,7,8,9", "1", "1950", "60", "1"},

	// 福彩
	// {"AllNum", "1|2|3", "1", "1950", "60", "1"},
	// {"AllSumNum", "6", "1", "1950", "60", "1"},
	// {"ThreeComNum", "1,3", "1", "1950", "60", "1"},
	// {"SixComNum", "1,2,3", "1", "1950", "60", "1"},
	// {"AllComSum", "6", "1", "1950", "60", "1"},
	// {"PrevTwoSumCom", "1,2,3,4,5,6,7,8,9,10,11,12,13,14,15,16,17", "1", "1950", "60", "1"},

	// {"MixNum", `001,002,003`, "1", "1950", "60", "1"},
	// {"AnyoneNum", "1", "1", "1950", "60", "1"},
	// {"AnyTwoNum", "1,3", "1", "1950", "60", "1"},
	// {"OneNum", "2|*|*", "1", "1950", "60", "1"},
	// {"PrevTwoSizeOdd", "1|1", "1", "1950", "60", "1"},

	// {"TwoNum", "2|1|*", "1", "1950", "60", "1"},
	// {"LastTwoComNum", "2,3", "1", "1950", "60", "1"},

	// 十一选五
	// {"MidSize", "04;05;06;07", "1", "1950", "60", "1"},
	// {"OneNum", "*|*|03,06,07,09,10|*|*", "1", "1950", "60", "1"},
	// {"TreeNum", "1|2|3", "1", "1950", "60", "1"},
	// {"PrevChooseThreeNumerical", "01,02,03;02,03,07,08,11", "1", "1950", "60", "1"},
	// {"Odd", "3X2;2X1", "1", "1950", "60", "1"},

	// {"PrevTwoNum", "01,04,07,11|02,03,05,08,09|*|*|*", "1", "1950", "60", "1"},
	// {"PrevTreeOneNum", "01,04,07,11", "1", "1950", "60", "1"},
	// {"AnyoneNum", "02;03", "1", "1950", "60", "1"},
	// {"AnyTwoNum", "02,03;04,02", "1", "1950", "60", "1"},

	// 排列三
	// {"ThreeNum", "1|1|1;1|2|3", "1", "1950", "60", "1"},
	// {"ThreeSum", "6", "1", "1950", "60", "1"},
	// {"ThreeComNum", "1,3", "1", "1950", "60", "1"},
	// {"SixComNum", "1,2,3", "1", "1950", "60", "1"},
	// {"AllComSum", "6", "1", "1950", "60", "1"},
	// {"AnyoneNum", "0,1,2,3,4,5,6,7,8,9", "1", "1950", "60", "1"},
	// {"AnyTwoNum", "0,1,2,3,4,5,6,7,8,9", "1", "1950", "60", "1"},

	// {"PrevTwoComNum", "0,1,2,3,4,5,6,7,8,9", "1", "1950", "60", "1"},
	// {"PrevTwoSizeOdd", "1|1", "1", "1950", "60", "1"},
}

func main() {
	var (
		count      = flag.Int("c", 1, "每种玩法运行多少条记录?")
		methodName = flag.String("m", "", "特定执行某种方法")
	)

	flag.Parse()

	c := &config.Config{}
	initMaster(c)

	lcf := lottery.GetConf(c)
	lty, err := lcf.GetLottery("ssj")
	if err != nil {
		panic(err.Error())
	}

	before := time.Now().UnixNano()
	id := 1
	for j := 0; j < *count; j++ {

		data := getData()
		for _, item := range data {

			p := model.NewProjects()
			p.SetData("projectid", id)
			p.SetData("customname", item["funcname"])
			p.SetData("code", item["code"])
			p.SetData("modes", item["modes"])
			p.SetData("multiple", item["multiple"])
			p.SetData("omodel", item["omodel"])
			p.SetData("totalprice", item["totalprice"])

			if *methodName != "" && utils.UcWords(p.GetString("customname")) != utils.UcWords(*methodName) {
				continue
			}

			err = lty.Dispatch(p)
			if err != nil {
				fmt.Println(err)
				return
			}

			id++
		}
	}
	after := time.Now().UnixNano()
	t := float64(after-before) / float64(time.Second)
	fmt.Printf("finish:%f \n", t)

	fmt.Println("请正确输入想要获取所输入号码的中奖情况：")
	fmt.Println("输入q或quit退出命令模式！！")
	in := bufio.NewReader(os.Stdin)
	for {
		bs, err := in.ReadSlice('\r')
		if err == io.EOF {
			break
		}

		line := string(bs[:len(bs)-1])
		line = strings.TrimSpace(line)
		if line == "quit" || line == "q" {
			break
		}

		txt := fmt.Sprintf("^[0-9]{%d}$", lty.GetLen())
		reg := regexp.MustCompile(txt)
		if reg.MatchString(line) {
			key, _ := lty.GenerateKey(line)
			fmt.Println("总派奖金额:", lty.GetTotalReward(key))

			records := lty.GetRecords(key)
			for _, r := range records {
				fmt.Println(r)
			}
		} else {
			fmt.Println(len(line))
			fmt.Printf("号码必须是长度%d的数字:%s\n", lty.GetLen(), line)
		}

	}

	lty.Reduce()
	winNum := lty.Draw()
	fmt.Println("开奖号码：", winNum)
	key, _ := lty.GenerateKey(winNum)
	fmt.Println("中奖金额：", lty.GetTotalReward(key))
	fmt.Println("投注金额：", lty.GetGross())

	exitMaster(c)

}

func getData() []map[string]interface{} {
	rd, _ := os.Open("./test.data")
	buf := bufio.NewReader(rd)

	var data []map[string]interface{}
	for {
		line, _, err := buf.ReadLine()
		if err == io.EOF {
			break
		}

		line = bytes.TrimSpace(line)
		if len(line) > 0 && line[0] != '#' {
			item := make(map[string]interface{})
			if err := json.Unmarshal(line, &item); err == nil {
				data = append(data, item)
			} else {
				fmt.Println(err)
			}
		}
	}

	return data

}

func initMaster(c *config.Config) {
	loadCoreMudule(c)

	err := c.SortModules()
	if err != nil {
		panic(err.Error())
	}

	c.CreateConfModules()
	c.InitConfModules()

	err = c.Parse("../../../conf/example.conf")
	if err != nil {
		panic(err.Error())
	}

	c.InitModules()
}

func exitMaster(c *config.Config) {
	c.ExitModules()
}

func loadCoreMudule(c *config.Config) {
	c.LoadModule(db.ModuleName)
	c.LoadModule(log.ModuleName)
	c.LoadModule(lottery.ModuleName)
}
