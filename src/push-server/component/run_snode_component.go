package component

import (
	"connect-server/utils/config"
	"fmt"
)

type RunWithSnode struct {
	Flag bool
}

var runWithSnode *RunWithSnode

func (this *RunWithSnode) Init() error {
	config := config.NewConfig()
	config.Read("conf/push-server.conf")
	flag, err := config.GetInt("RunWithSnode")
	if err != nil {
		fmt.Println("Error in RunWithSnode.Init() ", err)
		return err
	}
	if flag == 0 {
		this.Flag = false
	} else {
		this.Flag = true
	}
	return nil
}

func GetRunWithSnode() *RunWithSnode {
	if runWithSnode == nil {
		runWithSnode = &RunWithSnode{}
		err := runWithSnode.Init()
		if err != nil {
			fmt.Println(err)
		}
	}
	return runWithSnode
}
