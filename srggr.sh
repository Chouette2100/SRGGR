#!/bin/sh
cd /home/chouette/MyProject/Showroom/SRGGR
# 人気ライバーランキング【予選】' 486
# 新人スタートダッシュ 490
# 人気ジャンルライバーランキング-アイドル王 494,495,496,497,498,499,500,501,502,503,504
# 最強ファンランキング206
# 期間限定ランキング【9/22】 512,515
# 期間限定ランキング【9/25】 513,516
# 期間限定ランキング【9/28】 514,517
# 人気ライバーランキング【決勝 Sリーグ】 491,492,493
export SOPS_AGE_KEY_FILE=/home/chouette/.config/age/key.txt
./SRGGR -campaignid "omatsuri2026(final)" -limit 50 -giftid 1228,1229 > /dev/null 2>> error.log
date >> error.log
