// AGPL License
// Copyright (c) 2021 ysicing <i@ysicing.me>

package main

import (
	"flag"
	"fmt"
	"strings"

	"github.com/ergoapi/util/exgin"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/webhooks/v6/github"
	"github.com/sirupsen/logrus"
	"github.com/ysicing/ergo/pkg/util/ssh"
)

var port int

func main() {
	flag.IntVar(&port, "p", 8080, "port")
	flag.Parse()
	hook, _ := github.New(github.Options.Secret(""))
	g := exgin.Init(&exgin.Config{
		Cors:    true,
		Metrics: true,
		Debug:   true,
	})
	g.Use(exgin.ExLog(), exgin.ExRecovery(), exgin.ExTraceID())
	g.Any("/webhooks", func(c *gin.Context) {
		payload, err := hook.Parse(c.Request, github.ReleaseEvent, github.PullRequestEvent, github.PushEvent)
		if err != nil {
			logrus.Errorf("webhooks: %v", err)
			exgin.GinsData(c, nil, err)
			return
		}
		switch payload.(type) {
		case github.PushPayload:
			push := payload.(github.PushPayload)
			fullname := push.Repository.FullName
			logrus.Infof("[push]收到 %s", fullname)
			if strings.Contains(fullname, "devops-handbook") {
				exgin.GinsData(c, nil, nil)
				err := ssh.RunCmd("/bin/bash", "/root/blog/update.sh")
				if err != nil {
					logrus.Errorf("[push]更新失败 %v", err)
				}
				return
			}
		case github.ReleasePayload:
			release := payload.(github.ReleasePayload)
			fullname := release.Repository.FullName
			v := release.Release.TagName
			logrus.Infof("[release]收到 %s, release: %s", fullname, v)
			if strings.Contains(fullname, "ergo") {
				err := ssh.RunCmd("/bin/bash", "/root/homebrew-tap/genergo.sh")
				if err != nil {
					logrus.Errorf("[release]更新失败 %v", err)
					exgin.GinsData(c, nil, err)
					return
				}
			} else if strings.Contains(fullname, "kube-resource") {
				err := ssh.RunCmd("/bin/bash", "/root/homebrew-tap/kr.sh")
				if err != nil {
					logrus.Errorf("[kube-resource] %v", err)
					exgin.GinsData(c, nil, err)
					return
				}
			}
		default:
			logrus.Errorf("[webhooks]收到未知类型 %v", payload)
			exgin.GinsData(c, nil, fmt.Errorf("未知的事件类型"))
			return
		}
		exgin.GinsData(c, nil, nil)
	})
	g.GET("/auto", func(c *gin.Context) {
		exgin.GinsData(c, nil, nil)
		err := ssh.RunCmd("/bin/bash", "/root/blog/update.sh")
		if err != nil {
			logrus.Errorf("[push]更新失败 %v", err)
		}
	})
	addr := fmt.Sprintf(":%v", port)
	g.Run(addr)
}
