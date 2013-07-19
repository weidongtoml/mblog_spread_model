#!/bin/sh

pushd ..
export GOPATH=`pwd`
popd

pushd ./spread_model
go install
popd

go build
mv src mblog_spread_model
nohup ./mblog_spread_model --active_rate_file="./user_retweet_counts_gt_10_p_day.txt" --user_interaction_rate_file="./interactions_gt_2.txt" > sim_output.txt 2>&1 &
#./mblog_spread_model --active_rate_file="./user_retweet_counts_gt_10_p_day.txt" --user_interaction_rate_file="./interactions_gt_2.txt" 

