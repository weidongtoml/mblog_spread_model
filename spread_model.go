package mblog_spread_model

import ()

// Ids of all the users considered to be active in the network,
// these users will be randomly selected as seed for spreading tweets.
type userIdList struct {
	list []uint64
	size int
}

func newUserIdList(size int) *userIdList {
	user_id_list := UserQQList{make([]uint64, 0, size), 0}
	return &user_id_list
}

func (user_id_list *userIdList) add(qq uint64) {
	user_id_list.list = append(user_id_list.list, qq)
	user_id_list.size++
}

func (user_id_list *userIdList) randomQQ() uint64 {
	i := rand.Intn(user_id_list.size)
	return user_id_list.list[i]
}

func (user_id_list *userIdList) String() string {
	return fmt.Sprintf("UserQQList[%d]{%v}", user_id_list.size, user_id_list.list)
}


//  Information of a single user
type userInfoMap map[uint64]struct {
	avg_daily_retweets int
	engagement_factor  float32
	neighbors          []uint64
}

func newUserInfoMap (size int) *userInfoMap {
	info_map := new(userInfoMap)
	return info_map
}



// Storing the retweet action of a user, mainly the original poster of the 
// tweets that has been retweeted by the user, and the number of posts of the poster
// retweeted by the current user.
type userRetweetAction map[uint64]struct {
	retweet_count       int
	retweet_probability float32
}

// Retweet information of all the users
type userInteracionMap map[uint64]*userRetweetAction


func newUserInteracionMap (size int) *userInteracionMap {
	interactions := userInteracionMap(make(map[uint64]*interaction, size))
	return &interactions
}

func (interactions *userInteracionMap) String() string {
	str := ""
	for id1, action := range *interactions {
		for id2, v := range *action {
			str += fmt.Sprintf("%d retweets %d: %d[%f]\n", id1, id2, v.retweet_count, v.retweet_probability)
		}
		str += fmt.Sprintf("\n")
	}
	return str
}

func (interactions *userInteracionMap) AddInterctions(origin_id, retweet_id, 
	count uint64) {
	action, found := (*interactions)[retweet_id]
	if !found {
		(*interactions)[retweet_id] = new(userRetweetAction)
		action = (*interactions)[retweet_id]
	}
	_, found = (*action)[origin_id]
	if found {
		//TODO(weidoliang): add error handling and error log
	} else {
		(*retweet_rate)[origin_id].retweet_count = count
	}
}

func (interactions *userInteracionMap) Finalize() {
	for _, action := range *interactions {
		total_retweets := uint64(0)
		for _, r_info := range *action {
			total_retweets += action.retweet_count
		}
		for _, r_info := range *action {
			r_info.retweet_probability = float64(r_info.retweet_count) / float64(total_retweets)
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
	avg_retweet_rate float32
	max_depth        int
}

func LoadSpreadModelData (active_rate_file, interaction_rate_file string) *SpreadModelData {
	num_users := 1000
	user_id_list := NewUserIdList(num_users)
	user_interaction_map := newUserInteracionMap(num_users)
	user_info_map := newUserInfoMap(num_users)

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
