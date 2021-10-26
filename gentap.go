// AGPL License
// Copyright (c) 2021 ysicing <i@ysicing.me>

package main

import (
	"flag"
	"fmt"
	"github.com/ergoapi/exgin"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/webhooks/v6/github"
	"github.com/ysicing/ergo/pkg/util/ssh"
	"log"
)

var port int

func main()  {
	flag.IntVar(&port, "p", 8080, "port")
	flag.Parse()
	hook, _ := github.New(github.Options.Secret(""))
	g := exgin.Init(true)
	g.Any("/webhooks", func(c *gin.Context) {
		payload, err := hook.Parse(c.Request, github.ReleaseEvent, github.PullRequestEvent)
		if err != nil {
			exgin.GinsData(c, nil, nil)
			return
		}
		switch payload.(type) {
		case github.ReleasePayload:
			log.Println("收到release hook")
			err := ssh.RunCmd("/bin/bash", "/root/homebrew-tap/genergo.sh")
			if err != nil {
				exgin.GinsData(c, nil, err)
				return
			}
		}
		exgin.GinsData(c, nil, nil)
	})
	addr := fmt.Sprintf(":%v", port)
	g.Run(addr)
}
