#!/bin/bash
cd ~chouette/MyProject/Showroom/SRGGR
# 人気ライバーランキング【予選】' 486
# 新人スタートダッシュ 490
# 人気ジャンルライバーランキング-アイドル王 494,495,496,497,498,499,500,501,502,503,504
# 最強ファンランキング206
# 期間限定ランキング【9/22】 512,515
# 期間限定ランキング【9/25】 513,516
# 期間限定ランキング【9/28】 514,517
# 人気ライバーランキング【決勝 Sリーグ】 491,492,493
env DBNAME=showroom DBUSER=xxxxxx DBPW=xxxxxxxxxx ./SRGGR -limit 10 -giftid 486,490,494,495,496,497,498,499,500,501,502,503,504,206 > /dev/null 2>> error.log
date >> error.log
