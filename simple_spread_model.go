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
)

type Neighbour struct {
	qq             uint64
	retweet_counts uint64
	retweet_rate   float64
}

type User struct {
	active_rate float64
	neighbors   []Neighbour
}

func NewUser(active_rate float64) *User {
	var user User
	user.active_rate = active_rate
	return &user
}

func (user *User) AppendNeighbour(qq, retweet_count uint64) {
	user.neighbors = append(user.neighbors, Neighbour{qq, retweet_count, 0.0})
}

func (user *User) FinalizeRetweetRate() {
	total_retweets := uint64(0)
	for _, n := range user.neighbors {
		total_retweets += n.retweet_counts
	}
	for i, _ := range user.neighbors {
		user.neighbors[i].retweet_rate = float64(user.neighbors[i].retweet_counts) / float64(total_retweets)
	}
}

// A UserNetwork represents all the users in the network and the activity rate
// retweet rate, and such information of each user.
type UserNetwork map[uint64]*User

func NewUserNetwork(size int) UserNetwork {
	return make(map[uint64]*User, size)
}

// TODO(weidoliang): This is incorrect, fix it
func (user_network *UserNetwork) FinalizeNetwork() {
	for _, user := range *user_network {
		user.FinalizeRetweetRate()
	}
}

func (user_network *UserNetwork) String() string {
	str := ""
	for qq, user := range *user_network {
		str += fmt.Sprintf("%d {%v}\n", qq, user)
	}
	return str
}

// A UserQQList contains all the QQ number currently in the network
type UserQQList struct {
	list []uint64
	size int
}

func NewUserQQList(size int) *UserQQList {
	qq_list := UserQQList{make([]uint64, 0, size), 0}
	return &qq_list
}

func (user_qq_list *UserQQList) Add(qq uint64) {
	user_qq_list.list = append(user_qq_list.list, qq)
	user_qq_list.size++
}

func (user_qq_list *UserQQList) RandomQQ() uint64 {
	i := rand.Intn(user_qq_list.size)
	return user_qq_list.list[i]
}

func (user_qq_list *UserQQList) String() string {
	return fmt.Sprintf("UserQQList[%d]{%v}", user_qq_list.size, user_qq_list.list)
}

// Structure for storing retweet information
type RetweetInfo struct {
	retweet_count	uint64
	retweet_rate	float64
}

type UserRetweetRate map[uint64]*RetweetInfo

type UserInteractions map[uint64]*UserRetweetRate

func NewUserRetweetRate () *UserRetweetRate {
	user_retweet_rate := UserRetweetRate(make(map[uint64]*RetweetInfo))
	return &user_retweet_rate
}

func NewUserInteractions (size int) *UserInteractions {
	interactions := UserInteractions(make(map[uint64]*UserRetweetRate, size))
	return &interactions
}

func (interactions *UserInteractions) String() string {
	str := ""
	for qq, value := range *interactions {
		str += fmt.Sprintf("%d: %v\n", qq, value)
	}
	return str
}

func (retweet_rate *UserRetweetRate) String() string {
	str := ""
	for qq, value := range *retweet_rate {
		str += fmt.Sprintf("%d: %v\t", qq, *value)
	}
	return str
}


func (interactions *UserInteractions) AddInterctions(qq_origin, qq_retweet, 
	count uint64) {
	retweet_rate, found := (*interactions)[qq_retweet]
	if !found {
		(*interactions)[qq_retweet] = NewUserRetweetRate()
		retweet_rate = (*interactions)[qq_retweet]
	}
	_, found = (*retweet_rate)[qq_origin]
	if found {
		//TODO(weidoliang): add error handling and error log
	} else {
		(*retweet_rate)[qq_origin] = & RetweetInfo{retweet_count: count}
	}
}

func (interactions *UserInteractions) Finalize() {
	for _, user_retweet_rate := range *interactions {
		total_retweets := uint64(0)
		for _, retweet_info := range *user_retweet_rate {
			total_retweets += retweet_info.retweet_count
		}
		for _, r_info := range *user_retweet_rate {
			r_info.retweet_rate = float64(r_info.retweet_count) / float64(total_retweets)
		}
	}
}


// Load an mblog network from the given files.
func LoadUserNetworkFromFile(active_rate_file, interaction_rate_file string) (*UserNetwork, *UserQQList, *UserInteractions) {
	user_network := NewUserNetwork(1000)
	user_qq_list := NewUserQQList(1000)
	user_interactions := NewUserInteractions(1000)

	u_active_rate_f, err := os.Open(active_rate_file)
	if err != nil {
		fmt.Println("Failed to open file [%s]: %s", active_rate_file, err)
		os.Exit(-1)
	}

	u_interact_rate_f, err := os.Open(interaction_rate_file)
	if err != nil {
		fmt.Println("Failed to open file [%s]: %s", interaction_rate_file, err)
		os.Exit(-1)
	}

	defer func() {
		u_active_rate_f.Close()
		u_interact_rate_f.Close()
	}()

	// Load User Activity Rate
	active_rate_reader := bufio.NewReader(u_active_rate_f)
	for {
		line, err := active_rate_reader.ReadString('\n')
		if err != nil {
			break
		}
		tokens := strings.Fields(line)
		qq, err := strconv.ParseUint(tokens[0], 10, 64)
		if err != nil {
			fmt.Println("Invalid QQ number: [%s]", tokens[0])
		}
		active_rate, err := strconv.ParseFloat(tokens[1], 64)
		if err != nil {
			fmt.Println("Invalid active rate: [%s]", tokens[1])
		}

		_, found := user_network[qq]
		if found {
			fmt.Printf("Error, duplicate QQ[%d] in activity file", qq)
		} else {
			user_network[qq] = NewUser(active_rate)
			user_qq_list.Add(qq)
		}
	}

	// Load User Interaction Rate
	interact_rate_reader := bufio.NewReader(u_interact_rate_f)
	for {
		line, err := interact_rate_reader.ReadString('\n')
		if err != nil {
			if err != io.EOF {
				fmt.Printf("Error while reading file [%s] : %s\n",
					interaction_rate_file,
					err)
			}
			break
		}
		tokens := strings.Fields(line)
		qq_repost, err := strconv.ParseUint(tokens[0], 10, 64)
		if err != nil {
			fmt.Println("Invalid QQ number: [%s]", tokens[0])
		}
		qq_original, err := strconv.ParseUint(tokens[1], 10, 64)
		if err != nil {
			fmt.Println("Invalid QQ number: [%s]", tokens[1])
		}
		retweet_count, err := strconv.ParseUint(tokens[2], 10, 64)
		if err != nil {
			fmt.Println("Invalid active rate: [%s]", tokens[2])
		}
		user, found := user_network[qq_original]
		if found {
			user.AppendNeighbour(qq_repost, retweet_count)
			user_interactions.AddInterctions(qq_original, qq_repost, retweet_count)
		} else {
			// Not an active user, ignore this interaction
		}
	}
	user_network.FinalizeNetwork()
	user_interactions.Finalize()

	return &user_network, user_qq_list, user_interactions
}

type SimulationParam struct {
	max_depth		int
	retweet_factor	float32
}
/*
// Runs the spread simulation 
func RunSpreadSimulation(param *SimulationParam, user_network *UserNetwork, 
	user_list *UserQQList, user_interactions *UserInteractions) float64 {
	rounds = 100
	total_retweets := 0
	for i := 0; i < rounds; i++ {
		init_qq := user_list.RandomQQ()
		runSpread(param, user_network, user_list, user_interactions, init_qq, 0);
	}
	return float64(total_retweets)/float64(rounds)
}

func runSpread(param *SimulationParam, user_network *UserNetwork, 
	user_list *UserQQList, init_qq uint64, depth int) uint64 {
	total_retweets := 0
	if depth < param.max_depth {
		user, found := (*user_network)[init_qq]
		if found {
			do_retweet := true
			if depth > 0 {
				
			}
		} else {
			//User is inactive, we just skip it
		}
	}
	return total_retweets
}
*/
func main() {
	var user_active_rate_file = flag.String("active_rate_file",
		"active_rate.txt",
		"File describing the user active rate, each line is of the form QQ<tab>active_rate")

	var user_interaction_rate_file = flag.String("user_interaction_rate_file",
		"user_interaction_rate.txt",
		"File describing the user interaction rate, each line is of the form QQ1<tab>QQ2<tab>RetweetsCount")

	flag.Parse()

	user_network, user_list, user_interactions := LoadUserNetworkFromFile(*user_active_rate_file, *user_interaction_rate_file)

	fmt.Printf("User Network:\n%v\n", user_network)
	fmt.Println("-------------------------------------")
	fmt.Printf("User List:\n%v\n", user_list)
	fmt.Println("-------------------------------------")
	fmt.Printf("%v", user_interactions)
}
