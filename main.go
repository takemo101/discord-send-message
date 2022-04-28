package main

import (
	"context"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"time"

	"github.com/chromedp/cdproto/emulation"
	"github.com/chromedp/chromedp"
	"github.com/chromedp/chromedp/kb"
	"github.com/jinzhu/configor"
)

// メッセージ送信対象チャンネル
type TargetChannel struct {
	Server  string
	Channel string
}

// Discordアカウント
type DiscordAccount struct {
	Email    string
	Password string
}

// メッセージ設定
type MessageSetting struct {
	Target  TargetChannel
	Account DiscordAccount
	Message string
}

// メッセージ設定生成
func NewMessageSetting(
	target TargetChannel,
	account DiscordAccount,
	message string,
) MessageSetting {
	return MessageSetting{
		target,
		account,
		message,
	}
}

func main() {
	// 設定ファイルのを無名構造体で読み込み
	config := struct {
		Account struct {
			Email    string
			Password string
		}
		Default struct {
			Server  string
			Channel string
			Message string
		}
	}{}
	if err := configor.Load(&config, "./config.yaml"); err != nil {
		log.Fatal(err)
	}

	// コマンド引数
	message := flag.String("message", config.Default.Message, "message")
	server := flag.String("server", config.Default.Server, "server")
	channel := flag.String("channel", config.Default.Channel, "channel")
	flag.Parse()

	// chromedpコンテキストを生成
	ctx, cancel := chromedp.NewContext(
		context.Background(),
		chromedp.WithLogf(log.Printf),
	)
	defer cancel()

	// タイムアウトを設定
	ctx, cancel = context.WithTimeout(ctx, 50*time.Second)
	defer cancel()

	// スクリーンショットのバッファ
	var picbuf []byte

	// 処理開始時間
	startTime := time.Now()

	// メッセージ送信実行！
	if err := executeSendMessageAction(ctx, &picbuf, NewMessageSetting(
		TargetChannel{
			*server,
			*channel,
		},
		DiscordAccount{
			config.Account.Email,
			config.Account.Password,
		},
		*message,
	)); err != nil {
		log.Fatal(err)
	}

	// スクリーンショットをファイル出力
	if err := ioutil.WriteFile(
		"./screenshot/"+time.Now().Format("2006-01-02_15:04:05")+".png",
		picbuf,
		0644,
	); err != nil {
		log.Fatal(err)
	}

	fmt.Printf("処理時間 : %f 秒\n", time.Since(startTime).Seconds())
}

// メッセージ送信の実行
func executeSendMessageAction(
	ctx context.Context,
	picbuf *[]byte,
	setting MessageSetting,
) error {
	return chromedp.Run(ctx,
		emulation.SetUserAgentOverride("DiscordScraper 1.0"),
		loginTasks(setting.Account),
		sendMessageTasks(setting.Message, setting.Target),
		chromedp.FullScreenshot(picbuf, 50),
	)
}

// ログインするためのタスク
func loginTasks(account DiscordAccount) chromedp.Tasks {
	return chromedp.Tasks{
		chromedp.Navigate("https://discord.com/login"),
		chromedp.SendKeys("input[name=email]", account.Email),
		chromedp.SendKeys("input[name=password]", account.Password),
		chromedp.Click("button[type=submit]"),
	}
}

// メッセージ送信するためのタスク
func sendMessageTasks(
	message string,
	target TargetChannel,
) chromedp.Tasks {
	return chromedp.Tasks{
		chromedp.Click(
			"div[aria-label="+target.Server+"]",
			chromedp.NodeVisible,
		),
		chromedp.Click(
			"li[data-dnd-name="+target.Channel+"]",
			chromedp.NodeVisible,
		),
		chromedp.Click("div[data-slate-node=element]"),
		chromedp.KeyEvent(message + kb.Enter),
		chromedp.ActionFunc(func(ctx context.Context) error {
			// なぜかキーイベントで入力を行うと入力前に処理が終了してしまうので数秒待つ
			time.Sleep(3 * time.Second)
			return nil
		}),
	}
}
