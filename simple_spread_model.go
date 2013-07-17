package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"math/rand"
	"os"
	"strconv"
	"strings"
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


	fmt.Printf("User Network:\n%v\n", user_network)
	fmt.Println("-------------------------------------")
	fmt.Printf("User List:\n%v\n", user_list)
	fmt.Println("-------------------------------------")
	fmt.Printf("%v", user_interactions)
}
