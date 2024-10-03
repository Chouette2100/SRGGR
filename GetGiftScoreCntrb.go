// Copyright © 2024 chouette.21.00@gmail.com
// Released under the MIT license
// https://opensource.org/licenses/mit-license.php
package main

import (
	//	"flag"
	"fmt"
	//	"strconv"

	//	"io"
	//	"log"
	//	"os"

	//	"strings"
	//	"strconv"
	"time"

	"net/http"

	"github.com/go-gorp/gorp"

	//	"github.com/dustin/go-humanize"

	//	"github.com/Chouette2100/exsrapi"
	"github.com/Chouette2100/srapi"
	"github.com/Chouette2100/srdblib"
)

// ユーザー（リスナー）のルームのギフト獲得数に対する貢献ランキングを取得しデータベースに格納する
func GetGiftScoreCntrb(
	client *http.Client,
	dbmap *gorp.DbMap,
	tnow time.Time,
	t1 int,
	t2 int,
) (
	err error,
) {

	//	貢献ランキング取得の対象となるGiftidを求める
	sqlst := "select grid, startedat from giftranking "
	sqlst += " where cntrblst = 1 "
	sqlst += " and startedat < ? and endedat > ? "
	//	rows, errt := dbmap.Select(srdblib.GiftRanking{}, sqlst, tnow, tnow.Add(-24 * time.Hour))
	rows, errt := dbmap.Select(srdblib.GiftRanking{}, sqlst, tnow, tnow.Add(-120*time.Hour))
	if errt != nil {
		err = fmt.Errorf("Select(GiftRanking{}, ...): %w", errt)
		return
	}
	//	if len(rows) == 0 {
	//		err = fmt.Errorf("GetGiftScoreCntrb: no giftid found")
	//		return
	//	}

	for _, v := range rows {
		giftid := v.(*srdblib.GiftRanking).Grid
		startedat := v.(*srdblib.GiftRanking).Startedat

		sqlst = "select userno, userid from user "
		sqlst += " where userno in (select distinct userno from giftscore where giftid = ? and ts between ? and ? )"
		rows, errt = dbmap.Select(srdblib.User{}, sqlst, giftid, startedat, tnow)
		if errt != nil {
			err = fmt.Errorf("Select(User{}, ...): %w", errt)
			return
		}

		for _, w := range rows {
			userno := w.(*srdblib.User).Userno
			userid := w.(*srdblib.User).Userid

			gsc, errt := srapi.ApiCdnGiftRankingContribution(client, giftid, userid)
			if errt != nil {
				err = fmt.Errorf("ApiCdnGiftRankingContribution: %w", errt)
				return
			}

			for _, x := range gsc.RankingList {

				err = srdblib.InserIntoGiftScoreCntrb(client, dbmap, giftid, userno, &x, tnow)
				if err != nil {
					err = fmt.Errorf("InsertIntoGiftScoreCntrb(): %w", err)
					return
				}
			}
		}
	}

	return
}
