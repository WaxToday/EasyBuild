package main

import (
	"fmt"
	"flag"
	"os"
	"strings"
	"path/filepath"
	"bytes"
	"os/exec"
)

// 编译信息结构体
type BuildInfo struct {
	os string
	arch []int	// 对应架构数组下标
	class string
}

// 参数结构体
type Args struct {
	GoFile string
	Style string
	OS string
	Version string
	Help bool
}

var args Args = Args{}

// 初始化参数
func init() {
	flag.StringVar(&args.GoFile, "f", "", "编译的目标文件，留空则自动检测当前目录下唯一go文件，多个文件时则需手动指定")
	flag.StringVar(&args.Style, "s", "pc", `编译风格，默认为pc
值：
    all: 全部操作系统
    pc: windows、linux、darwin
    unix: aix、freebsd、illumos、netbsd、openbsd、plan9、solaris
    mobile: android、ios 
    web: js、wasip1
    p2p: dragonfly
`)
	flag.StringVar(&args.OS, "o", "", "指定的操作系统名称，")
	flag.StringVar(&args.Version, "v", "1.0", "版本号，默认为1.0")
	flag.BoolVar(&args.Help, "h", false, "显示帮助")
	flag.Usage = usage

	// 格式化参数
	flag.Parse()

	// 显示帮助
	if args.Help {
		flag.Usage()
		os.Exit(0)
	}

	// 获取编译的文件
	args.GoFile = getTargetFile(args.GoFile)

	// 获取编译的类型
	args.Style = getStyle(args.Style)

	// 获取操作系统名称
	args.OS = getOS(args.OS)
}

func usage() {
	fmt.Printf(`EasyBuild Version: 1.0
Usage: easybuild [-h help]

Options:
`)
	flag.PrintDefaults()
}

func main() {
	// 架构数组
	var ARCH []string = []string{
		"386",
		"amd64",
		"arm",
		"arm64",
		"loong64",
		"mips",
		"mips64",
		"mips64le",
		"mipsle",
		"ppc64",
		"ppc64le",
		"riscv64",
		"s390x",
		"wasm",
	}

	var buildTree []BuildInfo

	buildTree = append(buildTree, BuildInfo{"aix", []int{9}, "unix"})
	buildTree = append(buildTree, BuildInfo{"android", []int{0, 1, 2, 3}, "mobile"})
	buildTree = append(buildTree, BuildInfo{"darwin", []int{1, 3}, "pc"})
	buildTree = append(buildTree, BuildInfo{"dragonfly", []int{1}, "p2p"})
	buildTree = append(buildTree, BuildInfo{"freebsd", []int{0, 1, 2}, "unix"})
	buildTree = append(buildTree, BuildInfo{"illumos", []int{1}, "unix"})
	buildTree = append(buildTree, BuildInfo{"ios", []int{3}, "mobile"})
	buildTree = append(buildTree, BuildInfo{"js", []int{13}, "web"})
	buildTree = append(buildTree, BuildInfo{"linux", []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12}, "pc"})
	buildTree = append(buildTree, BuildInfo{"netbsd", []int{0, 1, 2}, "unix"})
	buildTree = append(buildTree, BuildInfo{"openbsd", []int{0, 1, 2, 3}, "unix"})
	buildTree = append(buildTree, BuildInfo{"plan9", []int{0, 1, 2}, "unix"})
	buildTree = append(buildTree, BuildInfo{"solaris", []int{1}, "unix"})
	buildTree = append(buildTree, BuildInfo{"wasip1", []int{13}, "web"})
	buildTree = append(buildTree, BuildInfo{"windows", []int{0, 1, 2, 3}, "pc"})
	
	for _, item := range buildTree {
		if (args.OS == "" && item.class == args.Style) || args.OS == item.os {
			for i := range item.arch {
				buildFile(args.GoFile, item.os, ARCH[i], args.Version)
			}
		}
	}
}

func buildFile(gofile string, osname string, arch string, version string) {
	extend := ""
	if osname == "windows" {
		extend = ".exe"
	}

	osenv := fmt.Sprintf("GOOS=%s", osname)
	archenv := fmt.Sprintf("GOARCH=%s", arch)
	outputName := fmt.Sprintf("%s/%s_%s_%s%s", version, strings.TrimSuffix(gofile, ".go"), osname, arch, extend)

	cmd := exec.Command("go", "build", "-o", outputName, gofile)
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Env = append(os.Environ(),
		osenv,
		archenv,
	)
	err := cmd.Run()
	if err != nil {
		fmt.Printf("[x] Build %s - %s .... Error! %s\n", osname, arch, err)
		fmt.Println(err)
	}else{
		fmt.Printf("[+] Build %s - %s .... success!\n", osname, arch)
	}
}

/* 获取编译目标文件 */
func getTargetFile(filename string) string {
	if (filename == ""){
		// 文件名为空，获取当前目录下唯一一个go文件
		dir, err := os.Getwd()
		if err != nil {
			fmt.Println(err)
			os.Exit(0)
		}

		entries, err := os.ReadDir(dir)
		if err != nil {
			fmt.Println(err)
			os.Exit(0)
		}

		ext := ".go" // 假设我们要找的是.txt文件
		var filteredFiles []string
		for _, entry := range entries {
			if !entry.IsDir() && filepath.Ext(entry.Name()) == ext {
				filteredFiles = append(filteredFiles, entry.Name())
			}
		}

		if len(filteredFiles) == 0 {
			fmt.Println("当前目录下不存在go文件")
			os.Exit(0)
		}

		if len(filteredFiles) > 1 {
			fmt.Println("当前目录下存在多个go文件:")
			for _, filename := range filteredFiles {
				fmt.Println(filename)
			}
			fmt.Println("请使用[-f filename.go]指定需要编译的文件")
			os.Exit(0)
		}

		filename = filteredFiles[0]
	}else{
		// 判断后缀是否为.go
		if !strings.HasSuffix(filename, ".go") {
			fmt.Println("非*.go文件！")
			os.Exit(0)
		}

		// 判断文件是否存在
		info, err := os.Stat(filename)
		if os.IsNotExist(err) || info.IsDir() {
			fmt.Println("文件不存在！")
			os.Exit(0)
		}
	}

	return filename
}

/* 获取编译风格 */
func getStyle(style string) string {
	styleList := []string{"all", "pc", "unix", "mobile", "web", "p2p"}
	style = strings.ToLower(style)

	if isStringInList(style, styleList) {
		return style
	}

	return "pc"
}

/* 获取操作系统名称 */
func getOS(os string) string {
	osList := []string{"aix", "android", "darwin", "dragonfly", "freebsd", "illumos", "ios", "js", "linux", "netbsd", "openbsd", "plan9", "solaris", "wasip1", "windows"}
	os = strings.ToLower(os)

	if isStringInList(os, osList) {
		return os
	}

	return ""
}

/* 判断字符串是否在数组中 */
func isStringInList(target string, stringList []string) bool {
	for _, val := range stringList {
		if target == val {
			return true
		}
	}

	return false
}