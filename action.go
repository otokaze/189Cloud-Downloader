package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"runtime"
	"strings"

	"189Cloud-Downloader/dao"
	"189Cloud-Downloader/model"
	"189Cloud-Downloader/utils"

	"github.com/otokaze/go-kit/log"
	"github.com/otokaze/go-kit/printcolor"
	"github.com/peterh/liner"
	"github.com/urfave/cli/v2"
)

var (
	d    = dao.New()
	dirs = map[string]*model.Dir{}

	user    *model.UserInfo
	share   *model.ShareInfo
	current *model.Dir
)

func loginAction(ctx *cli.Context) (err error) {
	if cookie := ctx.String("cookie"); cookie != "" {
		d.LoginWithCookie(ctx.Context, cookie)
		var info interface{}
		if info, err = d.GetLoginedInfo(ctx.Context); err != nil {
			return
		}
		user = info.(*model.UserInfo)
		log.Info("user(%+v)", user)
	} else {
		if ctx.Args().Len() < 2 {
			log.Error("请输入格式正确的账号密码！格式：%s", ctx.Command.ArgsUsage)
			return
		}
		if user, err = d.Login(ctx.Context, ctx.Args().Get(0), ctx.Args().Get(1)); err != nil {
			log.Error("登陆失败！请确保账号密码正确。")
			return
		}
	}
	if current == nil {
		current = &model.Dir{ID: "-11", Name: "全部文件", IsFolder: true, IsHome: true}
		dirs[current.GetID()] = current
	}
	printcolor.Blue("登陆成功！你好，%s\n", user.GetName())
	return
}

func logoutAction(ctx *cli.Context) (err error) {
	return d.Logout(ctx.Context)
}

func versionAction(ctx *cli.Context) (err error) {
	if version == "" {
		version = "null"
	}
	println(ctx.App.Name, "version", version, runtime.GOOS)
	return
}

func userInfoAction(ctx *cli.Context) (err error) {
	var info interface{}
	if info, err = d.GetLoginedInfo(ctx.Context, ctx.Bool("all")); err != nil {
		printcolor.Red("%s\n", err.Error())
		return
	}
	if ctx.Bool("all") {
		println(info.(string))
		return
	}
	var userinfo = info.(*model.UserInfo)
	println("UserId:", userinfo.UserId)
	println("UserAccount:", userinfo.GetName())
	println("已用容量:", utils.FormatFileSize(userinfo.UsedSize))
	println("可用容量:", utils.FormatFileSize(userinfo.Quota-userinfo.UsedSize))
	println("总容量:", utils.FormatFileSize(userinfo.Quota))
	user = userinfo
	return
}

func shareAction(ctx *cli.Context) (err error) {
	if ctx.Args().Len() < 1 {
		log.Error("请输入格式正确的分享链接！ 格式：%s", ctx.Command.ArgsUsage)
		return
	}
	var (
		shareCode  string
		accessCode = ctx.Args().Get(1)
	)
	if shareCode = utils.ParseShareCode(ctx.Args().Get(0)); shareCode == "" {
		err = fmt.Errorf("没有找到ShareCode，请联系作者更新脚本！")
		printcolor.Red("%v\n", err)
		return
	}
	if share, err = d.GetShareInfo(ctx.Context, shareCode, accessCode); err != nil {
		printcolor.Red("%v\n", err)
		return
	}
	current = &model.Dir{ID: share.FileID, Name: share.FileName}
	dirs[current.GetID()] = current
	return
}

func lsAction(ctx *cli.Context) (err error) {
	var isLong bool
	if key := ctx.Context.Value("isLong"); key != nil {
		isLong = key.(bool)
	}
	var (
		pn, ps = ctx.Int("pn"), ctx.Int("ps")
		order  = ctx.String("order")
		path   = ctx.Args().Get(0)
	)
	if current == nil {
		println("当前没有目录可载入！请尝试登陆或者加载分享链接。")
		return
	}
	if path == "" && current != nil {
		path = current.GetID()
	}
	var _dirs []*model.Dir
	if path == "~" || current.IsHome {
		if _dirs, err = d.GetHomeDirList(ctx.Context, pn, ps, order, path); err != nil {
			log.Error("d.GetHomeDirList() pn(%d) ps(%d) folderId(%s) error(%v)", pn, ps, path, err)
			return
		}
	} else if share != nil {
		if _dirs, err = d.GetShareDirList(ctx.Context, share, pn, ps, order, path); err != nil {
			log.Error("d.GetShareDirList() pn(%d) ps(%d) folderId(%s) error(%v)", pn, ps, path, err)
			return
		}
	} else {
		if path != "" {
			printcolor.Red("no such file or directory: %s\n", path)
		}
		return
	}
	for idx, dir := range _dirs {
		var format string
		if isLong {
			if dir.IsFolder {
				format = fmt.Sprintf("[D]%s\t%s\t%s\t%s", dir.GetID(), utils.FormatFileSize(dir.FileListSize), dir.CreateDate, dir.Name)
			} else {
				format = fmt.Sprintf("[F]%s\t%s\t%s\t%s", dir.GetID(), utils.FormatFileSize(dir.Size), dir.CreateDate, dir.Name)
			}
			if idx != len(_dirs)-1 {
				format += "\n"
			}
		} else {
			format = dir.Name
			if idx != len(_dirs)-1 {
				format += "\t"
			}
		}
		if dir.IsFolder {
			printcolor.Blue(format)
		} else {
			print(format)
		}
		dirs[dir.GetID()] = dir
	}
	print("\n")
	return
}

func llAction(ctx *cli.Context) (err error) {
	var key interface{} = "isLong"
	ctx.Context = context.WithValue(ctx.Context, key, true)
	return lsAction(ctx)
}

func cdAction(ctx *cli.Context) (err error) {
	var args []string
	if argsInf := ctx.Context.Value("args"); argsInf != nil {
		args = argsInf.([]string)
	}
	if len(args) < 1 {
		log.Error("请输入正确的目标路径的ID！ 格式：%s", ctx.Command.ArgsUsage)
		return
	}
	var path = args[0]
	if path == "/" {
		if current.IsHome {
			path = "-11"
		} else {
			path = share.FileID
		}
	} else if path == "../" || path == ".." {
		if current.ParentID == "" {
			log.Error("当前已在顶级目录！")
			return
		}
		path = current.GetParentID()
	} else if path == "~" {
		path = "-11"
	} else if path == "share" {
		path = share.FileID
	}
	if dirs[path] == nil {
		printcolor.Red("找不到目录: %s，请先尝试`ls`遍历。\n", path)
		return
	}
	if !dirs[path].IsFolder {
		printcolor.Red("%s 不是个目录，无法执行`cd`命令。\n", path)
		return
	}
	current = dirs[path]
	return
}

func pwdAction(ctx *cli.Context) (err error) {
	if current == nil {
		println("/")
		return
	}
	if current.ParentID == nil {
		fmt.Printf("/%s\n", current.Name)
		return
	}
	var (
		path   = current.Name
		parent = current.GetParentID()
	)
	for dirs[parent] != nil {
		path = dirs[parent].Name + "/" + path
		parent = dirs[parent].GetParentID()
	}
	fmt.Printf("/%s\n", path)
	return
}

func getAction(ctx *cli.Context) (err error) {
	var fileId string
	if fileId = ctx.Args().Get(0); fileId == "" {
		log.Error("请输入要下载的文件！格式：%s", ctx.Command.ArgsUsage)
		return
	}
	var toPath string
	if toPath = ctx.Args().Get(1); toPath == "" {
		toPath, _ = os.Getwd()
	} else {
		toPath = strings.TrimRight(toPath, "/")
	}
	var c int
	if c = ctx.Int("c"); c <= 0 {
		if c = ctx.Int("concurrency"); c <= 0 {
			c = 1
		}
	}
	if user == nil {
		log.Error("请先登陆189账号！")
		return
	}
	var dir *model.Dir
	if fileId == "." || fileId == "./" &&
		current != nil {
		dir = current
	} else if fileId == "../" &&
		dirs[current.GetParentID()] == nil {
		log.Error("%s 已经是根目录！", current.GetID())
		return
	} else {
		dir = dirs[fileId]
	}
	if dir == nil {
		log.Error("找不到任何文件！")
		return
	}
	var fn func(*model.Dir, string)
	fn = func(dir *model.Dir, path string) {
		if dir.IsFolder {
			var dirs []*model.Dir
			if dir.IsHome {
				dirs, _ = d.GetHomeDirAll(ctx.Context, dir.GetID())
			} else {
				dirs, _ = d.GetShareDirAll(ctx.Context, share, dir.GetID())
			}
			for _, d := range dirs {
				fn(d, path+"/"+dir.Name)
			}
			return
		}
		if url, _ := d.GetDownloadURL(ctx.Context, dir.GetID(), share); url != "" {
			if err = d.Download(ctx.Context, url, path, c, ctx.String("tmp")); err == nil {
				return
			}
			file, _ := os.OpenFile("./189Cloud-Downloader.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
			log.Infow(file, "File(%+v), 下载失败！", dir)
		}
	}
	fn(dir, toPath)
	return
}

func afterAction(ctx *cli.Context) (err error) {
	line := liner.NewLiner()
	line.SetCtrlCAborts(true)
	defer line.Close()
	line.SetCompleter(func(line string) (cs []string) {
		for _, c := range ctx.App.Commands {
			if strings.HasPrefix(c.Name, strings.ToLower(line)) {
				cs = append(cs, c.Name)
			}
		}
		return
	})
	for {
		var processName string
		if current != nil {
			processName = current.GetShortName()
		}
		var input string
		if input, err = line.Prompt(processName + "> "); err != nil {
			if err == liner.ErrPromptAborted {
				err = nil
				return
			}
			log.Error("line.Prompt() error(%v)", err)
			continue
		}
		line.AppendHistory(input)
		input = strings.TrimSpace(input)
		args := strings.Split(input, " ")
		if args[0] == "" {
			continue
		}
		var command *cli.Command
		if command = ctx.App.Command(args[0]); command == nil {
			printcolor.Red("找不到指令: %s\n", args[0])
			continue
		}
		var fset = flag.NewFlagSet(args[0], flag.ContinueOnError)
		for _, f := range command.Flags {
			f.Apply(fset)
		}
		if !command.SkipFlagParsing {
			if err = fset.Parse(args[1:]); err != nil {
				if err == flag.ErrHelp {
					err = nil
					continue
				}
				log.Error("fs.Parse(%v) error(%v)", args[1:], err)
				continue
			}
		}
		var key interface{} = "args"
		nCtx := cli.NewContext(ctx.App, fset, ctx)
		nCtx.Context = context.WithValue(nCtx.Context, key, args[1:])
		nCtx.Command = command
		command.Action(nCtx)
	}
}
