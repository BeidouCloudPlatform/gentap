// AGPL License
// Copyright (c) 2021 ysicing <i@ysicing.me>

package main

import (
	"flag"
	"fmt"
	"strings"

	"github.com/ergoapi/exgin"
	"github.com/ergoapi/zlog"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/webhooks/v6/github"
	"github.com/ysicing/ergo/pkg/util/ssh"
)

var port int

func init() {
	cfg := zlog.Config{
		Simple:      true,
		WriteLog:    false,
		ServiceName: "gentap",
	}
	zlog.InitZlog(&cfg)
}

func main() {
	flag.IntVar(&port, "p", 8080, "port")
	flag.Parse()
	hook, _ := github.New(github.Options.Secret(""))
	g := exgin.Init(true)
	g.Use(exgin.ExLog(), exgin.ExRecovery(), exgin.ExCors())
	g.Any("/webhooks", func(c *gin.Context) {
		payload, err := hook.Parse(c.Request, github.ReleaseEvent, github.PullRequestEvent, github.PushEvent)
		if err != nil {
			zlog.Error("webhooks: %v", err)
			exgin.GinsData(c, nil, err)
			return
		}
		switch payload.(type) {
		case github.PushPayload:
			push := payload.(github.PushPayload)
			fullname := push.Repository.FullName
			zlog.Info("[push]收到 %s", fullname)
			if strings.Contains(fullname, "devops-handbook") {
				exgin.GinsData(c, nil, nil)
				err := ssh.RunCmd("/bin/bash", "/root/blog/update.sh")
				if err != nil {
					zlog.Error("[push]更新失败 %v", err)
				}
				return
			}
		case github.ReleasePayload:
			release := payload.(github.ReleasePayload)
			fullname := release.Repository.FullName
			v := release.Release.TagName
			zlog.Info("[release]收到 %s, release: %s", fullname, v)
			if strings.Contains(fullname, "ergo") {
				err := ssh.RunCmd("/bin/bash", "/root/homebrew-tap/genergo.sh")
				if err != nil {
					zlog.Error("[release]更新失败 %v", err)
					exgin.GinsData(c, nil, err)
					return
				}
			} else if strings.Contains(fullname, "kube-resource") {
				err := ssh.RunCmd("/bin/bash", "/root/homebrew-tap/kr.sh")
				if err != nil {
					zlog.Error("[kube-resource] %v", err)
					exgin.GinsData(c, nil, err)
					return
				}
			}
		default:
			zlog.Error("[webhooks]收到未知类型 %v", payload)
			exgin.GinsData(c, nil, fmt.Errorf("未知的事件类型"))
			return
		}
		exgin.GinsData(c, nil, nil)
	})
	addr := fmt.Sprintf(":%v", port)
	g.Run(addr)
}
