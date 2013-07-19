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
	
	
	fmt.Printf("Loading data from files [%s],[%s]..\n", *user_active_rate_file, *user_interaction_rate_file)
	
	simulator.LoadSpreadModelData(*user_active_rate_file, *user_interaction_rate_file)
	
	fmt.Printf("Done\n")
	
	simulator.PrintDataStatistics()
	
	parameters := simulator.GetParameters()
	
	//TODO(weidoliang): iterate througn different avg_retweet_rate and depth to produce 
	//results under different parameters
	
	avg_rates := []float32{0.01, 0.02, 0.03, 0.04, 0.05, 0.06, 0.07, 0.08, 0.09, 0.1, 0.2, 0.3, 0.4, 0.5, 0.7, 1.0}
	max_depth := []int{2, 3, 4, 5, 6, 7}
	score_distribution := []int{1, 2, 3, 4, 5, 10, 50, 100, 1000}

	//parameters.Is_random_sim = true
	//parameters.Random_sim_rounds = 100
	run_simulation := true
	
	if run_simulation {
		parameters.Is_random_sim = false
		for _, r := range avg_rates {
			for _, d := range max_depth {
				parameters.Avg_retweet_rate = r
				parameters.Max_depth = d
		
				fmt.Printf("Runing simulation with Parameters: %v\n...", *parameters)
				
				result := simulator.RunSimulation()
				avg_retweet := result.GetAverageRetweetCount()
				retweet_dist := result.GetRetweetCountDistribution(&score_distribution)
				
				fmt.Printf("Average Retweet Count: %f\n", avg_retweet)
				fmt.Printf("Score distribution: %v\n", *retweet_dist)
				fmt.Printf("---------------------------------------------------------\n")
			}
		}
	}
}
