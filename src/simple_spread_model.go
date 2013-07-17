package main

import (
	"fmt"
	"flag"
	"spread_model"
)

func main() {
	var user_active_rate_file = flag.String("active_rate_file",
		"active_rate.txt",
		"File describing the user active rate, each line is of the form QQ<tab>active_rate")

	var user_interaction_rate_file = flag.String("user_interaction_rate_file",
		"user_interaction_rate.txt",
		"File describing the user interaction rate, each line is of the form QQ1<tab>QQ2<tab>RetweetsCount")

	flag.Parse()

	simulator := new(spread_model.Simulator)
	simulator.LoadSpreadModelData(*user_active_rate_file, *user_interaction_rate_file)
	parameters := simulator.GetParameters()
	parameters.Avg_retweet_rate = 0.05
	parameters.Max_depth = 3
	parameters.Is_random_sim = true
	parameters.Random_sim_rounds = 100
	
	result := simulator.RunSimulation()
	fmt.Printf("Average Retweet Count: %f\n", result.GetAverageRetweetCount())
}
