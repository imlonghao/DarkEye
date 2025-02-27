package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/zsdevX/DarkEye/common"
	"github.com/zsdevX/DarkEye/zoomeye"
	"strconv"
	"strings"
)

type zoomEyeRuntime struct {
	Module
	parent *RequestContext

	api     string
	search  string
	page    string
	flagSet *flag.FlagSet
	cmd     []string
}

var (
	zoomEye               = "zoomEye"
	zoomEyeRuntimeOptions = &zoomEyeRuntime{
		flagSet: flag.NewFlagSet(zoomEye, flag.ExitOnError),
	}
)

func (zoom *zoomEyeRuntime) Start(ctx context.Context) {
	z := zoomeye.New()
	z.Query = strings.TrimSpace(zoom.search)
	z.ApiKey = zoom.api
	z.Pages = zoom.page
	zoom.parent.taskId ++
	common.Log("zoom.start", fmt.Sprintf("获取页面范围%s", z.Pages), common.INFO)
	z.ErrChannel = make(chan string, 10)
	go func() {
		for {
			select {
			case m, ok := <-z.ErrChannel:
				if !ok {
					return
				}
				fmt.Println(m)
			default:
			}
		}
	}()
	if matches := z.Run(zoom.parent.ctx); matches != nil {
		for _, m := range matches {
			e := &analysisEntity{
				Task:            strconv.Itoa(zoom.parent.taskId),
				Ip:              m.Ip,
				Port:            strconv.Itoa(m.Port),
				Country:         m.Country,
				Service:         m.Service,
				Url:             m.Url,
				Title:           m.Title,
				WebServer:       m.App,
				WebResponseCode: int32(m.HttpCode),
				Hostname:        m.Hostname,
				Os:              m.Os,
				Device:          m.Device,
				Banner:          m.Banner,
				Version:         m.Version,
				ExtraInfo:       m.ExtraInfo,
				RDns:            m.RDns,
				Isp:             m.Isp,
			}
			analysisRuntimeOptions.upInsertEnt(e)
		}
		analysisRuntimeOptions.PrintCurrentTaskResult()
	}

}

func (z *zoomEyeRuntime) Init(requestContext *RequestContext) {
	z.parent = requestContext
	z.flagSet.StringVar(&z.api,
		"api", "you-key", "API-KEY")
	z.flagSet.StringVar(&z.search,
		"search", "ip:8.8.8.8", "https://www.zoomeye.org/")
	z.flagSet.StringVar(&z.page,
		"page", "1-5", "返回查询页面范围(每页20条):开始页-结束页")
}

func (z *zoomEyeRuntime) ValueCheck(value string) (bool, error) {
	if v, ok := zoomEyeValue[value]; ok {
		if isDuplicateArg(value, z.parent.CmdArgs) {
			if value == "-api" || value == "-page" {
				return false, fmt.Errorf("参数重复")
			}
		}
		return v, nil
	}
	return false, fmt.Errorf("无此参数")
}

func (z *zoomEyeRuntime) CompileArgs(cmd []string, os []string) error {
	if cmd != nil {
		search := []string{"-search"}
		ret, s := z.buildQuery(cmd)
		if s == "" {
			return fmt.Errorf("搜索参数为空")
		}
		search = append(search, s)
		ret = append(ret, search...)
		if err := z.flagSet.Parse(ret); err != nil {
			return err
		}
		z.flagSet.Parsed()
	} else {
		if err := z.flagSet.Parse(os); err != nil {
			return err
		}
	}
	return nil
}

func (a *zoomEyeRuntime) saveCmd(cmd []string) {
	a.cmd = cmdSave(cmd)
}

func (a *zoomEyeRuntime) restoreCmd() []string {
	return cmdRestore(a.cmd)
}

func (z *zoomEyeRuntime) buildQuery(cmd []string) ([]string, string) {
	ret := make([]string, 0)
	s := ""
	for _, c := range cmd {
		if strings.HasPrefix(c, "-api") || strings.HasPrefix(c, "-page") {
			ret = append(ret, strings.SplitN(c, " ", 2)...)
		} else {
			blocks := strings.Split(c, ":")
			rule := strings.TrimSpace(blocks[0])
			if len(blocks) == 2 {
				rule += ":" + strings.TrimSpace(blocks[1])
			}
			if strings.HasSuffix(s, "+") || strings.HasSuffix(s, "-") {
				s += rule
			} else {
				s += " " + rule
			}
		}
	}
	s = strings.TrimSpace(s)
	return ret, s
}

func (z *zoomEyeRuntime) Usage() {
	fmt.Println(fmt.Sprintf("Usage of %s:", zoomEye))
	fmt.Println("Options:")
	z.flagSet.VisitAll(func(f *flag.Flag) {
		var opt = "  -" + f.Name
		fmt.Println(opt)
		fmt.Println(fmt.Sprintf("		%v (default '%v')", f.Usage, f.DefValue))
	})
}
