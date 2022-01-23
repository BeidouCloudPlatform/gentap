// AGPL License
// Copyright (c) 2021 ysicing <i@ysicing.me>

package main

import (
	"flag"
	"fmt"
	"log"
	"strings"

	"github.com/ergoapi/exgin"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/webhooks/v6/github"
	"github.com/ysicing/ergo/pkg/util/ssh"
)

var port int

func main() {
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
			release := payload.(github.ReleasePayload)
			fullname := release.Repository.FullName
			v := release.Release.TagName
			log.Println("收到 ", fullname, " release: ", v)
			if strings.Contains(fullname, "ergo") {
				err := ssh.RunCmd("/bin/bash", "/root/homebrew-tap/genergo.sh")
				if err != nil {
					exgin.GinsData(c, nil, err)
					return
				}
			} else if strings.Contains(fullname, "kube-resource") {
				err := ssh.RunCmd("/bin/bash", "/root/homebrew-tap/kr.sh")
				if err != nil {
					exgin.GinsData(c, nil, err)
					return
				}
			}
		}
		exgin.GinsData(c, nil, nil)
	})
	addr := fmt.Sprintf(":%v", port)
	g.Run(addr)
}
