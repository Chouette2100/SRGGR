// Copyright © 2024 chouette.21.00@gmail.com
// Released under the MIT license
// https://opensource.org/licenses/mit-license.php
package main

import (
	"flag"
	"fmt"
	"strconv"

	"io"
	"log"
	"os"

	"strings"
	//	"strconv"
	"time"

	"net/http"

	"github.com/go-gorp/gorp"

	//	"github.com/dustin/go-humanize"

	"github.com/Chouette2100/exsrapi"
	"github.com/Chouette2100/srapi"
	"github.com/Chouette2100/srdblib"
)

/*

00AA00	新規作成

*/

const Version = "00AA00"

// ユーザーギフトランキングを取得しデータベースに格納する
//
//	ここでいうユーザーとは視聴者のことを意味する
func GetViewerGiftScore(client *http.Client, dbmap *gorp.DbMap, tnow time.Time, limit int) (err error) {

	cugr, err := srapi.ApiCdnUserGiftRanking(client, 206, limit)
	if err != nil {
		err = fmt.Errorf("srapi.ApiCdnUserGiftRanking() returned error. %w", err)
		return err
	}
	l := len(cugr.RankingList)
	log.Printf("GetViewerGiftScore() %d Users\n", l)
	for i := 0; i < l; i++ {
		err = srdblib.InserIntoViewerGiftScore(
			client,
			dbmap,
			206,
			&cugr.RankingList[i],
			tnow,
		)
		if err != nil {
			err = fmt.Errorf("srdblib.InserIntoViewerGiftScore() returned error. %w", err)
			return err
		}
	}
	return nil

}

// 指定されたギフトコードのギフトランキングを取得しデータベースに格納する
func GetGiftScore(client *http.Client, dbmap *gorp.DbMap, tnow time.Time, giftid int, limit int) (err error) {

	cgr, err := srapi.ApiCdnGiftRanking(client, giftid, limit)
	if err != nil {
		err = fmt.Errorf("srapi.ApiCdnGiftRanking() returned error. %w", err)
		return err
	}

	l := len(cgr.RankingList)
	log.Printf("GetGiftScore() Giftid: %d  %d Rooms\n", giftid, l)
	for i := 0; i < l; i++ {
		err = srdblib.InserIntoGiftScore(
			client,
			dbmap,
			giftid,
			&cgr.RankingList[i],
			tnow,
		)
		if err != nil {
			err = fmt.Errorf("srdblib.InserIntoGiftScore() returned error. %w", err)
			return err
		}
	}
	return nil
}

// ギフトランキングを読み込みデータベースに書き込む
//
//	ギフトランキング　=> struct GiftScorer => table giftscore, user
//	ユーザーギフトランキング　=> struct ViewerGiftScorer => table viewergiftscore, viewer
//
// cronで実行することを前提としている
func main() {

	var (
		//      |コード|名称|補足|
		//      |---|---|---|
		//      |486|人気ライバーランキング|GetGiftScore()|
		//      |490|新人スタートダッシュ|〃|
		//      |494|アイドル|〃|
		//      |495|俳優|〃|
		//      |496|アナウンサー|〃|
		//      |497|グローバル|〃|
		//      |498|声優|〃|
		//      |499|芸人|〃|
		//      |500|タレント|〃|
		//      |501|ライバー|〃|
		//      |502|モデル|〃|
		//      |503|バーチャル|〃|
		//      |504|アーティスト|〃|
		//      |206|ユーザーギフトランキング|GetViewerGiftScore()|
		giftid = flag.String("giftid", "", "string flag")
		limit = flag.Int("limit", 500, "int flag")
	)

	//	ログ出力を設定する
	logfile, err := exsrapi.CreateLogfile(Version, srdblib.Version)
	if err != nil {
		panic("cannnot open logfile: " + err.Error())
	}
	defer logfile.Close()
	// log.SetOutput(logfile)
	log.SetOutput(io.MultiWriter(logfile, os.Stdout))

	flag.Parse()

	log.Printf("param -giftid: %s -limit: %d\n", *giftid, *limit)

	//	データベースとの接続をオープンする。
	var dbconfig *srdblib.DBConfig
	dbconfig, err = srdblib.OpenDb("DBConfig.yml")
	if err != nil {
		err = fmt.Errorf("srdblib.OpenDb() returned error. %w", err)
		log.Printf("%s\n", err.Error())
		return
	}
	if dbconfig.UseSSH {
		defer srdblib.Dialer.Close()
	}
	defer srdblib.Db.Close()

	log.Printf("********** Dbhost=<%s> Dbname = <%s> Dbuser = <%s> Dbpw = <%s>\n",
		(*dbconfig).DBhost, (*dbconfig).DBname, (*dbconfig).DBuser, (*dbconfig).DBpswd)

	//	gorpの初期設定を行う
	dial := gorp.MySQLDialect{Engine: "InnoDB", Encoding: "utf8mb4"}
	srdblib.Dbmap = &gorp.DbMap{Db: srdblib.Db, Dialect: dial, ExpandSliceArgs: true}

	srdblib.Dbmap.AddTableWithName(srdblib.User{}, "user").SetKeys(false, "Userno")
	srdblib.Dbmap.AddTableWithName(srdblib.GiftScore{}, "giftscore").SetKeys(false, "Giftid", "Ts", "Userno")
	srdblib.Dbmap.AddTableWithName(srdblib.Viewer{}, "viewer").SetKeys(false, "Viewerid")
	srdblib.Dbmap.AddTableWithName(srdblib.ViewerGiftScore{}, "viewergiftscore").SetKeys(false, "Giftid", "Ts", "Viewerid")

	//      cookiejarがセットされたHTTPクライアントを作る
	client, jar, err := exsrapi.CreateNewClient("anonymous")
	if err != nil {
		err = fmt.Errorf("CreateNewClient() returned error. %w", err)
		log.Printf("%s\n", err.Error())
		return
	}
	//      すべての処理が終了したらcookiejarを保存する。
	defer jar.Save() //	忘れずに！

	if *giftid == "" {
		log.Printf("giftid is empty\n")
		return
	}
	cgida := strings.Split(*giftid, ",")

	for _, cgid := range cgida {
		gid, err := strconv.Atoi(cgid)
		if err != nil {
			log.Printf("strconv() returned error. %s\n", err.Error())
			continue
		}
		tnow := time.Now().Truncate(time.Second)
		if gid == 206 {
			err = GetViewerGiftScore(client, srdblib.Dbmap, tnow, *limit)
			if err != nil {
				log.Printf("%s\n", err.Error())
				continue
			}
		} else {
			err = GetGiftScore(client, srdblib.Dbmap, tnow, gid, *limit)
			if err != nil {
				log.Printf("%s\n", err.Error())
				continue
			}
		}
	}

}
