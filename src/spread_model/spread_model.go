package spread_model

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"math/rand"
	"os"
	"strconv"
	"strings"
)

// Ids of all the users considered to be active in the network,
// these users will be randomly selected as seed for spreading tweets.
type userIdList struct {
	list []uint64
	size int
}

func newUserIdList(size int) *userIdList {
	user_id_list := userIdList{make([]uint64, 0, size), 0}
	return &user_id_list
}

func (user_id_list *userIdList) add(qq uint64) {
	user_id_list.list = append(user_id_list.list, qq)
	user_id_list.size++
}

func (user_id_list *userIdList) randomId() uint64 {
	i := rand.Intn(user_id_list.size)
	return user_id_list.list[i]
}

func (user_id_list *userIdList) String() string {
	return fmt.Sprintf("UserQQList[%d]{%v}", user_id_list.size, user_id_list.list)
}

//  Information of a single user
type userInfo struct {
	avg_daily_retweets uint64
	engagement_factor  float32
	followers          []uint64
}

type userInfoMap map[uint64]*userInfo

func newUserInfoMap(size int) *userInfoMap {
	user_info_map := userInfoMap(make(map[uint64]*userInfo, size))
	return &user_info_map
}

func (user_info_map *userInfoMap) hasUser(id uint64) bool {
	_, found := (*user_info_map)[id]
	return found
}

func (user_info_map *userInfoMap) addUser(id uint64, avg_retweets uint64) {
	user_info, found := (*user_info_map)[id]
	if found {
		log.Printf("Error: duplicated user [%d], overriding pervioud info", id)
	} else {
		(*user_info_map)[id] = new(userInfo)
		user_info = (*user_info_map)[id]
	}
	user_info.avg_daily_retweets = avg_retweets
}

func (user_info_map *userInfoMap) addFollower(id, follower_id uint64) {
	user_info, found := (*user_info_map)[id]
	if found {
		user_info.followers = append(user_info.followers, follower_id)
	}
}

func (user_info_map *userInfoMap) followers(id uint64) *[]uint64 {
	return &(*user_info_map)[id].followers
}

func (user_info_map *userInfoMap) engagement_factor(id uint64) float32 {
	return (*user_info_map)[id].engagement_factor
}

func (user_info_map *userInfoMap) finalize() {
	total_retweets := uint64(0)
	user_count := 0
	for _, user_info := range *user_info_map {
		total_retweets += user_info.avg_daily_retweets
		user_count++
	}
	avg_daily_retweets := float32(total_retweets) / float32(user_count)
	for _, user_info := range *user_info_map {
		user_info.engagement_factor = float32(user_info.avg_daily_retweets) / avg_daily_retweets
	}
}

// Storing the retweet action of a user, mainly the original poster of the 
// tweets that has been retweeted by the user, and the number of posts of the poster
// retweeted by the current user.
type userAction struct {
	retweet_count       uint64
	retweet_probability float32
}

type userRetweetAction map[uint64]*userAction

// Retweet information of all the users
type userInteractionMap map[uint64]*userRetweetAction

func newUserInteracionMap(size int) *userInteractionMap {
	interactions := userInteractionMap(make(map[uint64]*userRetweetAction, size))
	return &interactions
}

func (interactions *userInteractionMap) String() string {
	str := ""
	for id1, action := range *interactions {
		for id2, v := range *action {
			str += fmt.Sprintf("%d retweets %d: %d[%f]\n", id1, id2, v.retweet_count, v.retweet_probability)
		}
		str += fmt.Sprintf("\n")
	}
	return str
}

func (interactions *userInteractionMap) addInteractions(origin_id, retweet_id,
	count uint64) {
	_, found := (*interactions)[retweet_id]
	if !found {
		new_retweet_action := userRetweetAction(make(map[uint64]*userAction, 10))
		(*interactions)[retweet_id] = &new_retweet_action
	}
	user_retweet_action, _ := (*interactions)[retweet_id]

	user_action, found := (*user_retweet_action)[origin_id]
	if found {
		log.Printf("Error, userInteractionMap.addInteractions encountered duplicated action pair[%d][%d], old information will be over-written",
			origin_id, retweet_id)
		(*user_action).retweet_count = count
	} else {
		(*user_retweet_action)[origin_id] = &userAction{count, float32(0)}
	}
}

func (interactions *userInteractionMap) getRetweetProb(origin_id, retweet_id uint64) float32 {
	user_retweet_action, found := (*interactions)[retweet_id]
	if !found {
		log.Printf("Cannot find [%d]", retweet_id)
		return float32(0)
	}
	retweet_info, found := (*user_retweet_action)[origin_id]
	if !found {
		log.Printf("Cannot find [%d][%d]", origin_id, retweet_id)
		return float32(0)
	}
	return retweet_info.retweet_probability
}

func (interactions *userInteractionMap) finalize() {
	for _, user_retweet_action := range *interactions {
		total_retweets := uint64(0)
		for _, r_info := range *user_retweet_action {
			total_retweets += r_info.retweet_count
		}
		for _, r_info := range *user_retweet_action {
			r_info.retweet_probability = float32(r_info.retweet_count) / float32(total_retweets)
		}
	}
}

// Simulation Data needed for simulation of the spread model
type SpreadModelData struct {
	user_id_list      *userIdList
	user_info_map     *userInfoMap
	user_interact_map *userInteractionMap
}

// Parameters for the simulation
type SimulationParameters struct {
	Avg_retweet_rate  float32
	Max_depth         int
	Is_random_sim     bool
	Random_sim_rounds int
}

// Structure for holding result of the current simulation
type SimulationResult struct {
	num_retweets []int
}

func (simulation_result *SimulationResult) addRetweetCount(count int) {
	simulation_result.num_retweets = append(simulation_result.num_retweets, count)
}

func (simulation_result *SimulationResult) GetAverageRetweetCount() float32 {
	sum := 0
	for _, v := range simulation_result.num_retweets {
		sum += v
	}
	return float32(sum)/float32(len(simulation_result.num_retweets))
}

type Simulator struct {
	model_data *SpreadModelData
	parameter  *SimulationParameters
}

func (simulator *Simulator) GetParameters () *SimulationParameters {
	if simulator.parameter == nil {
		simulator.parameter = new(SimulationParameters)
	}
	return simulator.parameter
}

// Load simulation data from the given files.
func (simulator *Simulator) LoadSpreadModelData(active_rate_file, interaction_rate_file string) bool {
	num_users := 1000
	user_id_list := newUserIdList(num_users)
	user_info_map := newUserInfoMap(num_users)
	user_interaction_map := newUserInteracionMap(num_users)

	u_active_rate_f, err := os.Open(active_rate_file)
	u_interact_rate_f, err := os.Open(interaction_rate_file)
	defer func() {
		u_active_rate_f.Close()
		u_interact_rate_f.Close()
	}()

	if err != nil {
		log.Fatalf("Failed to open file [%s]: %s", active_rate_file, err)
		return false
	}

	if err != nil {
		log.Fatalf("Failed to open file [%s]: %s", interaction_rate_file, err)
		return false
	}

	// Load User Activity Rate
	active_rate_reader := bufio.NewReader(u_active_rate_f)
	for {
		line, err := active_rate_reader.ReadString('\n')
		if err != nil {
			break
		}
		tokens := strings.Fields(line)
		user_id, err := strconv.ParseUint(tokens[0], 10, 64)
		if err != nil {
			log.Printf("Invalid QQ number: [%s]", tokens[0])
		}
		avg_retweets, err := strconv.ParseUint(tokens[1], 10, 64)
		if err != nil {
			log.Printf("Invalid active rate: [%s]", tokens[1])
		}

		if user_info_map.hasUser(user_id) {
			log.Printf("Error, duplicate Id[%d] in activity file", user_id)
		} else {
			user_id_list.add(user_id)
			user_info_map.addUser(user_id, avg_retweets)
		}
	}

	// Load User Interaction Rate
	interact_rate_reader := bufio.NewReader(u_interact_rate_f)
	for {
		line, err := interact_rate_reader.ReadString('\n')
		if err != nil {
			if err != io.EOF {
				log.Printf("Error while reading file [%s] : %s\n",
					interaction_rate_file,
					err)
			}
			break
		}
		tokens := strings.Fields(line)
		id_repost, err := strconv.ParseUint(tokens[0], 10, 64)
		if err != nil {
			log.Println("Invalid Id number: [%s]", tokens[0])
		}
		id_original, err := strconv.ParseUint(tokens[1], 10, 64)
		if err != nil {
			log.Println("Invalid Id number: [%s]", tokens[1])
		}
		retweet_count, err := strconv.ParseUint(tokens[2], 10, 64)
		if err != nil {
			log.Println("Invalid active rate: [%s]", tokens[2])
		}
		user_interaction_map.addInteractions(id_original, id_repost, retweet_count)
	}
	user_interaction_map.finalize()
	user_info_map.finalize()

	simulator.model_data = &SpreadModelData{user_id_list, user_info_map, user_interaction_map}
	return true
}

// Runs the Spread Model Simulation and returns the simulation result
func (simulator *Simulator) RunSimulation() *SimulationResult {
	param := simulator.parameter
	id_list := simulator.model_data.user_id_list

	simulation_result := new(SimulationResult)
	if param.Is_random_sim {
		round := 0
		for round < param.Random_sim_rounds {
			round++
			id := id_list.randomId()
			num_retweets := simulator.runSingleSpread(id)
			simulation_result.addRetweetCount(num_retweets)
		}
	} else {
		for _, id := range id_list.list {
			num_retweets := simulator.runSingleSpread(id)
			simulation_result.addRetweetCount(num_retweets)
		}
	}
	return simulation_result
}

func (simulator *Simulator) runSingleSpread(id uint64) int {
	user_info_map := simulator.model_data.user_info_map
	retweet_prob := simulator.parameter.Avg_retweet_rate * user_info_map.engagement_factor(id)
	rnd := rand.Float32()
	users_retweeted := make(map[uint64]bool)
	if rnd < retweet_prob {
		users_retweeted[id] = true
		for _, follower_id := range *user_info_map.followers(id) {
			simulator.runReweet(id, follower_id, 0, &users_retweeted)
		}
	}
	return len(users_retweeted)
}

func (simulator *Simulator) runReweet(post_id, follower_id uint64, depth int, users_retweeted *map[uint64]bool) {
	if depth > simulator.parameter.Max_depth || (*users_retweeted)[follower_id] {
		return
	}

	user_info_map := simulator.model_data.user_info_map
	interactions := simulator.model_data.user_interact_map
	retweet_rate := simulator.parameter.Avg_retweet_rate * user_info_map.engagement_factor(follower_id) * interactions.getRetweetProb(post_id, follower_id)

	if rand.Float32() < retweet_rate {
		(*users_retweeted)[follower_id] = true
		for _, f_follow_id := range *user_info_map.followers(follower_id) {
			simulator.runReweet(follower_id, f_follow_id, depth+1, users_retweeted)
		}
	}
}

// TODO(weidoliang): Intialize Random Seed
func Init() {
}
