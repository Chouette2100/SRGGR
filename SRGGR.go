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
00AB00	新規作成（機能再検討、バグ修正）
00AC00	GetGiftScoreCntrb()をあらたに作成する（ギフトランキング貢献ランキングデータの保存）
00AD00	「修羅の道ランキング」(Giftid=13）に対応する
00AE00	貢献ランキング取得の指定を"giftid=-1"での一括指定から、"giftid=-491,-492"の形式にする

*/

const Version = "00AE00"

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

	cgr := new(srapi.CdnGiftRanking)
	if giftid == 13 {
		cgr, err = srapi.ApiCdnSeasonAwardRanking(client, giftid, limit)
	} else {
		cgr, err = srapi.ApiCdnGiftRanking(client, giftid, limit)
	}
	if err != nil {
		err = fmt.Errorf("srapi.ApiCdnGiftRanking() returned error. %w", err)
		return
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
			return
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
		//      | 13|修羅の道ランキング|GetGiftScore()|
		//      |206|ユーザーギフトランキング|GetViewerGiftScore()|
		//      |486|人気ライバーランキング【予選】|GetGiftScore()|
		//      |490|新人スタートダッシュ|〃|
		//      |491|人気ライバーランキング【決勝Sリーグ】|〃|
		//      |492|人気ライバーランキング【決勝Rリーグ】|〃|
		//      |493|人気ライバーランキング【決勝S・Rリーグ総合】|〃|
		//      |494|人気ジャンルライバーランキング・アイドル|〃|
		//      |495|人気ジャンルライバーランキング・俳優|〃|
		//      |496|人気ジャンルライバーランキング・アナウンサー|〃|
		//      |497|人気ジャンルライバーランキング・グローバル|〃|
		//      |498|人気ジャンルライバーランキング・声優|〃|
		//      |499|人気ジャンルライバーランキング・芸人|〃|
		//      |500|人気ジャンルライバーランキング・タレント|〃|
		//      |501|人気ジャンルライバーランキング・ライバー|〃|
		//      |502|人気ジャンルライバーランキング・モデル|〃|
		//      |503|人気ジャンルライバーランキング・バーチャル|〃|
		//      |504|人気ジャンルライバーランキング・アーティスト|〃|
		//      |505|ふぁぼきゅんランキング|〃|
		//      |512|Sリーグ期間限定ランキング【9/22】|〃|
		//      |513|Sリーグ期間限定ランキング【9/25】|〃|
		//      |514|Sリーグ期間限定ランキング【9/28】|〃|
		//      |515|Rリーグ期間限定ランキング【9/22】|〃|
		//      |516|Rリーグ期間限定ランキング【9/25】|〃|
		//      |517|Rリーグ期間限定ランキング【9/28】|〃|
		campaignid = flag.String("campaignid", "", "string flag")
		giftid = flag.String("giftid", "", "string flag")
		limit  = flag.Int("limit", 500, "int flag")
	)

	//	ログ出力を設定する
	logfile, err := exsrapi.CreateLogfile(Version, srdblib.Version)
	if err != nil {
		panic("cannnot open logfile: " + err.Error())
	}
	defer logfile.Close()
	// log.SetOutput(logfile)
	log.SetOutput(io.MultiWriter(logfile, os.Stdout))

	fileenv := "Env.yml"
	err = exsrapi.LoadConfig(fileenv, &srdblib.Env)
	if err != nil {
		err = fmt.Errorf("exsrapi.Loadconfig(): %w", err)
		log.Printf("%s\n", err.Error())
		return
	}

	flag.Parse()

	log.Printf("param -campaignid: %s -giftid: %s -limit: %d\n", *campaignid, *giftid, *limit)

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
	srdblib.Dbmap.AddTableWithName(srdblib.Userhistory{}, "userhistory").SetKeys(false, "Userno", "Ts")
	srdblib.Dbmap.AddTableWithName(srdblib.GiftScore{}, "giftscore").SetKeys(false, "Giftid", "Ts", "Userno")
	srdblib.Dbmap.AddTableWithName(srdblib.Viewer{}, "viewer").SetKeys(false, "Viewerid")
	srdblib.Dbmap.AddTableWithName(srdblib.ViewerHistory{}, "viewerhistory").SetKeys(false, "Viewerid", "Ts")
	srdblib.Dbmap.AddTableWithName(srdblib.GiftRanking{}, "giftranking").SetKeys(false, "Campaignid","Grid")
	srdblib.Dbmap.AddTableWithName(srdblib.ViewerGiftScore{}, "viewergiftscore").SetKeys(false, "Giftid", "Ts", "Viewerid")
	srdblib.Dbmap.AddTableWithName(srdblib.GiftScoreCntrb{}, "giftscorecntrb").SetKeys(false, "Giftid", "Ts", "Userno", "Viewerid")

	//      cookiejarがセットされたHTTPクライアントを作る
	client, jar, err := exsrapi.CreateNewClient("anonymous")
	if err != nil {
		err = fmt.Errorf("CreateNewClient() returned error. %w", err)
		log.Printf("%s\n", err.Error())
		return
	}
	//      すべての処理が終了したらcookiejarを保存する。
	defer jar.Save() //	忘れずに！

	if *campaignid == "" {
		log.Printf("campaignid is empty\n")
		return
	}

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
		if gid < 0 {
			err = GetGiftScoreCntrb(client, srdblib.Dbmap, tnow, *campaignid, -gid)
			if err != nil {
				log.Printf("%s\n", err.Error())
			}
			continue
		} else if gid == 206 {
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
