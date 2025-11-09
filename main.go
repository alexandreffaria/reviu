package main

import (
	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/launcher"
)

func main() {
	u := launcher.New().Headless(false).MustLaunch() // auto-find/download Chromium
	rod.New().ControlURL(u).MustConnect().
		MustPage("https://google.com").
		MustWaitLoad()
}
