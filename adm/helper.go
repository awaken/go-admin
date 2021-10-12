package main

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/GoAdminGroup/go-admin/modules/system"
	"github.com/mgutz/ansi"
)

func cliInfo() {
	fmt.Println("GoAdmin CLI " + system.Version())
	fmt.Println()
}

func checkError(err error) {
	if err != nil {
		panic(err)
	}
}

func getLatestVersion() string {
	http.DefaultClient.Timeout = 3 * time.Second
	res, err := http.Get("https://goproxy.cn/github.com/!go!admin!group/go-admin/@v/list")
	if err != nil || res.Body == nil {
		return ""
	}

	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return ""
	}

	body, err := io.ReadAll(res.Body)
	if err != nil || body == nil {
		return ""
	}

	versionsArr := strings.Split(string(body), "\n")

	return versionsArr[len(versionsArr)-1]
}

func printSuccessInfo(msg string) {
	fmt.Println()
	fmt.Println()
	fmt.Println(ansi.Color(getWord(msg), "green"))
	fmt.Println()
	fmt.Println()
}

func newError(msg string) error {
	return errors.New(getWord(msg))
}
