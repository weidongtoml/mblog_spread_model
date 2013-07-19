package spread_model

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"math"
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

func (user_info_map *userInfoMap) size() int {
	return len(*user_info_map)
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
	v, found := (*user_info_map)[id]
	if found {  
		return v.engagement_factor
	} else {
		return float32(0)
	}
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

func (user_info_map *userInfoMap) getEngagementFactorDistribution(resolution float32) (float32, float32, *[]int) {
	min_factor := float32(1.0)
	max_factor := float32(0)
	u_map := map[uint64]*userInfo(*user_info_map)
	for _, v := range u_map {
		f := (*v).engagement_factor
		if f > max_factor {
			max_factor = f
		}
		if f < min_factor {
			min_factor = f
		}
	}
	dist_size := int((max_factor-min_factor)/resolution + 1)
	dist := make([]int, dist_size)

	for _, v := range u_map {
		f := (*v).engagement_factor
		index := int((f - min_factor) / resolution)
		dist[index]++
	}

	return min_factor, max_factor, &dist
}

func (user_info_map *userInfoMap) getFollowersDistribution(resolution int) (int, int, *[]int) {
	min_count := 10000
	max_count := 0
	u_map := map[uint64]*userInfo(*user_info_map)
	for _, v := range u_map {
		s := len((*v).followers)
		if s < min_count {
			min_count = s
		}
		if s > max_count {
			max_count = s
		}
	}
	dist_size := int((max_count-min_count)/resolution + 1)
	dist := make([]int, dist_size)

	for _, v := range u_map {
		s := len((*v).followers)
		index := int((s - min_count) / resolution)
		dist[index]++
	}

	return min_count, max_count, &dist
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
// [retweeter_id][original_poster_id] {retweet_count, retweet_probability}
type userInteractionMap map[uint64]*userRetweetAction

func newUserInteracionMap(size int) *userInteractionMap {
	interactions := userInteractionMap(make(map[uint64]*userRetweetAction, size))
	return &interactions
}

func (interactions *userInteractionMap) getCoActionRatioDistribution(resolution float32) (float32, float32, *[]int) {
	co_action_ratios := make([]float32, 0)
	min_co_ratio := float32(math.MaxFloat32)
	max_co_ratio := float32(0)
	max_co_found_so_far := float32(0)
	for reposter_id, action := range *interactions {
		for original_id, v := range *action {
			ratio := float32(math.MaxFloat32)
			action_2, found := (*interactions)[original_id]
			if found { 
				v2, found := (*action_2)[reposter_id]
				if found && v2.retweet_count > 0 {
					original_count := v2.retweet_count
					reposter_count := v.retweet_count
					ratio = float32(reposter_count) / float32(original_count)
					
					if ratio > max_co_found_so_far {
						max_co_found_so_far = ratio
					}		
				}
			}
			co_action_ratios = append(co_action_ratios, ratio)
			
			if ratio < min_co_ratio {
				min_co_ratio = ratio
			}
			if ratio > max_co_ratio {
				max_co_ratio = ratio 
			}
		}
	}
	max_co_found_so_far += resolution
	dist_size := int(float32(max_co_found_so_far - min_co_ratio)/resolution)+1
	dist := make([]int, dist_size)
	for _, v := range co_action_ratios {
		index := int((v - min_co_ratio)/resolution)
		if v == max_co_ratio {
			index = int((max_co_found_so_far - min_co_ratio)/resolution)
		}
		dist[index]++
	}
	return min_co_ratio, max_co_ratio, &dist
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

func (interactions *userInteractionMap) size() int {
	return len(*interactions)
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

func (spread_model_data *SpreadModelData) PrintDataStatistics() {
	num_unique_users := spread_model_data.user_id_list.size

	user_info := spread_model_data.user_info_map
	engage_factor_resolution := float32(0.1)

	min_factor, max_factor, factor_dist := user_info.getEngagementFactorDistribution(engage_factor_resolution)

	follow_count_resolution := 1
	min_followers, max_followers, follower_dist := user_info.getFollowersDistribution(follow_count_resolution)

	co_ratio_resolution := float32(1.0)
	min_co_ratio, max_co_ratio, co_ratio_dist := spread_model_data.user_interact_map.getCoActionRatioDistribution(co_ratio_resolution)

	fmt.Printf("------------------- Data Statistics --------------------------\n")
	fmt.Printf("Number of unique users: %d\n", num_unique_users)
	fmt.Printf("User Engagement Factor Statistics:\n")
	fmt.Printf("\tmin: %f, max: %f\n", min_factor, max_factor)
	fmt.Printf("\tDistribution (resolution: %f): \n", engage_factor_resolution)
	scale := min_factor
	for _, v := range *factor_dist {
		if v > 0 {
			fmt.Printf("\t\t[%f, %f) = %d\n", scale-engage_factor_resolution, scale, v)
		}
		scale += engage_factor_resolution
	}

	fmt.Printf("User Followers Statistics:\n")
	fmt.Printf("\tdirected interaction pairs: %d\n", user_info.size())
	fmt.Printf("\tmin: %d, max: %d\n", min_followers, max_followers)
	fmt.Printf("\tDistribution (resolution: %d):\n", follow_count_resolution)
	follow_scale := min_followers
	for _, v := range *follower_dist {
		if v > 0 {
			fmt.Printf("\t\t[%d, %d) = %d\n", follow_scale-follow_count_resolution, follow_scale, v)
		}
		follow_scale += follow_count_resolution
	}
	
	fmt.Printf("User CoAction Ratio Statistics:\n")
	fmt.Printf("min: %f, max: %f\n", min_co_ratio, max_co_ratio)
	fmt.Printf("\tDistribution (resolution: %f):\n", co_ratio_resolution)
	ratio_scale := min_co_ratio
	for _, v:= range *co_ratio_dist {
		if v > 0 {
			fmt.Printf("\t\t[%f, %f) = %d\n", ratio_scale-co_ratio_resolution, ratio_scale, v)
		}
		ratio_scale += co_ratio_resolution
	}

	fmt.Printf("---------------------------------------------------------------\n")
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
	return float32(sum) / float32(len(simulation_result.num_retweets))
}

//Takes an interval and returns the corresponding frequency, e.g.
// []int{1, 2, 3, 4, 5, 10, 15 } means
// [-inf, 1}, [1, 2}, [2, 3}, [3, 4}, [4, 5}, [5, 10}, [10, 15}, [15, +inf}
func (simulation_result *SimulationResult) GetRetweetCountDistribution(intervals *[]int) *[]int {
	freq := make([]int, len(*intervals)+1)
	for _, v := range simulation_result.num_retweets {
		ind := 0
		if v >= (*intervals)[0] {
			for _, ind_v := range *intervals {
				if v >= ind_v {
					ind++
				} else {
					break
				}
			}

		}
		freq[ind]++
	}
	return &freq
}

type Simulator struct {
	model_data *SpreadModelData
	parameter  *SimulationParameters
}

func (simulator *Simulator) GetParameters() *SimulationParameters {
	if simulator.parameter == nil {
		simulator.parameter = new(SimulationParameters)
	}
	return simulator.parameter
}

func (simulator *Simulator) PrintDataStatistics() {
	simulator.model_data.PrintDataStatistics()
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
		user_info_map.addFollower(id_original, id_repost)
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
	
	if user_info_map == nil {
		panic("user_info_map is nil")
	}
	if interactions == nil {
		panic("interactions is nil")
	}
	if simulator.parameter == nil {
		panic("parameter is nil")
	}
	
	g_avg_retweet_rate := simulator.parameter.Avg_retweet_rate
	u_engagement_factor := user_info_map.engagement_factor(follower_id)
	u_retweet_prob := interactions.getRetweetProb(post_id, follower_id)
	retweet_rate :=  g_avg_retweet_rate * u_engagement_factor * u_retweet_prob

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
