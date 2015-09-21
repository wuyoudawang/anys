package lottery

import (
	"fmt"

	"anys/config"
	"anys/instantiate/lottery/model"
	"anys/jobs"
	"anys/log"
	"anys/pkg/utils"
)

const (
	ModuleName = "lottery"
)

type lotteryConf struct {
	lotteries map[string]*Lottery
	currName  string
}

func (l *lotteryConf) GetAllLottery() map[string]*Lottery {
	return l.lotteries
}

func (l *lotteryConf) GetLottery(name string) (*Lottery, error) {
	if lty, exists := l.lotteries[name]; exists {
		return lty, nil
	}
	return nil, fmt.Errorf("this lottery '%s' is not exist", name)
}

var (
	commands = []*config.Command{
		{config.BLOCK, "lottery", lotteryBlock},
		{config.MORE1, "hedging", setHedging},
		{config.MORE1, "property", setProperty},
		{config.MORE1, "count", setCount},
		{config.BLOCK, "entity", entityBlock},
		{config.MORE1, "method", createMethod},
	}

	LotteryModule = &config.Module{
		Commands: commands,

		Create_conf: createConf,
		Init_conf:   initConf,

		Init_module: initModule,
	}
)

func init() {
	config.RegisterModule(ModuleName, LotteryModule)
}

func createConf(c *config.Config) {
	conf := &lotteryConf{}
	c.SetConf(config.GetModule(ModuleName), conf)
}

func initConf(c *config.Config) error {
	conf := GetConf(c)
	conf.lotteries = make(map[string]*Lottery)

	return nil
}

func initModule(c *config.Config) error {
	conf := GetConf(c)
	for name, lty := range conf.lotteries {
		err := lty.load(name)

		if err != nil {
			log.Warning("invaild lottery '%s'", lty.name)
			delete(conf.lotteries, lty.name)
			continue
		}
	}

	return nil
}

func installJobs(c *config.Config) {
	eng := jobs.GetConf(c).GetEngine()

	lotteries := GetConf(c).GetAllLottery()
	for _, lty := range lotteries {
		eng.Register(eng.NewJob(&lotteryJob{lty}, fmt.Sprintf("%s-DRAW-JOB", lty.name)))
		eng.Register(eng.NewJob(&issueJob{lty.GetLotteryModel()}, fmt.Sprintf("%s-ISSUE-JOB", lty.name)))
		eng.Register(eng.NewJob(&issueErrorJob{lty.GetLotteryModel()}, fmt.Sprintf("%s-ISSUEERROR-JOB", lty.name)))
	}
}

func GetConf(c *config.Config) *lotteryConf {
	conf := c.GetConf(config.GetModule(ModuleName))
	return conf.(*lotteryConf)
}

func lotteryBlock(c *config.Config) error {

	return c.Block()
}

func entityBlock(c *config.Config) error {
	args, err := c.CheckArgs(5)
	if err != nil {
		return err
	}

	name := args[1]
	conf := GetConf(c)
	a, err := c.Int(args[2])
	b, err := c.Int(args[3])
	d, err := c.Int(args[4])
	if err != nil {
		return err
	}

	lottery := NewLottery(a, b, d)
	lottery.name = name
	conf.lotteries[name] = lottery
	conf.currName = name
	return c.Block()
}

func setHedging(c *config.Config) error {
	args, err := c.CheckArgs(2)
	if err != nil {
		return err
	}

	conf := GetConf(c)
	conf.lotteries[conf.currName].hedging, err = c.Float64(args[1])
	return err
}

func setProperty(c *config.Config) error {
	args, err := c.CheckArgs(2)
	if err != nil {
		return err
	}

	conf := GetConf(c)
	conf.lotteries[conf.currName].cvt.property(args[1])
	return err
}

func setCount(c *config.Config) error {
	args, err := c.CheckArgs(2)
	if err != nil {
		return err
	}

	conf := GetConf(c)
	conf.lotteries[conf.currName].count, err = c.Int(args[1])
	return err
}

func createMethod(c *config.Config) error {
	args, err := c.CheckArgs(3)
	if err != nil {
		return err
	}

	conf := GetConf(c)
	methodName := utils.UcWords(args[1])
	jsonStr := args[2]
	l := conf.lotteries[conf.currName]
	method, err := NewMethod(jsonStr)
	if err != nil {
		return err
	}

	fn, exists := utils.FindFunc(method.getFuncName(methodName), method)
	if !exists {
		return fmt.Errorf("can't find this method '%s'", methodName)
	}

	// callFn := func(bet string, p *model.Projects) (error, int) {
	// 	assertFn := fn.(func(*Lottery, string, *model.Projects) (error, int))
	// 	return assertFn(l, bet, p)
	// }
	l.Register(methodName, fn.(func(*Lottery, string, *model.Projects) (error, int)))

	return nil
}
