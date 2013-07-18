package spread_model

import (
	"fmt"
	"math"
	"os"
	"testing"
)

func TestUserIdList(t *testing.T) {
	user_id_list := newUserIdList(100)
	ids := []uint64{1, 2, 3, 4, 5}
	for _, id := range ids {
		user_id_list.add(id)
	}
	for i := 0; i < 20; i++ {
		rand_id := user_id_list.randomId()
		if rand_id < 1 || rand_id > 5 {
			t.Errorf("userIdList.randomId() returns Ids that have not been added [%d],[%v]", rand_id, user_id_list)
		}
	}
	if len(user_id_list.list) != len(ids) {
		t.Errorf("Expected Id list to have %v but got %v", ids, user_id_list.list)
	}
}

func TestUserInfo(t *testing.T) {
	user_info_map := newUserInfoMap(100)

	user_list := []struct {
		id            uint64
		retweet_count uint64
		factor        float32
	}{
		{1, 10, 0.4},
		{2, 20, 0.8},
		{3, 30, 1.2},
		{4, 40, 1.6},
	}

	user_relation := []struct {
		id        uint64
		followers []uint64
	}{
		{1, []uint64{2, 3, 4}},
		{2, []uint64{3, 4}},
		{3, []uint64{4}},
	}

	for _, v := range user_list {
		user_info_map.addUser(v.id, v.retweet_count)
	}

	for _, v := range user_relation {
		for _, f := range v.followers {
			user_info_map.addFollower(v.id, f)
		}
	}

	user_info_map.finalize()

	for _, v := range user_list {
		if !user_info_map.hasUser(v.id) {
			t.Errorf("Expected user_info to have user[%d] but found none.", v.id)
		}
		engagement_factor := user_info_map.engagement_factor(v.id)
		if math.Abs(float64(engagement_factor-v.factor)) > 0.000001 {
			t.Errorf("Expected user[%d] engagement factor to be [%f] but got [%f]",
				v.factor, engagement_factor)
		}
	}

	for _, v := range user_relation {
		followers := user_info_map.followers(v.id)
		followers_are_same := true
		if len(v.followers) != len(*followers) {
			followers_are_same = false
		}
		for i, _ := range *followers {
			if v.followers[i] != (*followers)[i] {
				followers_are_same = false
			}
		}
		if !followers_are_same {
			t.Errorf("Expected followers of [%d] to be %v, but got %v",
				v.followers, *followers)
		}
	}

	min_f, max_f, dist := user_info_map.getEngagementFactorDistribution(0.2)
	expected_dist := []int{1, 0, 1, 0, 1, 0, 1}
	//0.4, 0.8, 1.2, 1.6
	//0.4, 0.6, 0.8, 1.0, 1.2, 1.4, 1.6
	//[1,  	0, 	1,	0,	1,	0,	1]
	if min_f != 0.4 {
		t.Errorf("Expected min_factor to be %f, but got %f", 0.4, min_f)
	}
	if max_f != 1.6 {
		t.Errorf("Expected max_factor to be %f, but got %f", 1.6, max_f)
	}
	dist_is_eqv := (len(expected_dist) == len(*dist))
	if dist_is_eqv {
		for i, _ := range expected_dist {
			if expected_dist[i] != (*dist)[i] {
				dist_is_eqv = false
			}
		}
	}

	if !dist_is_eqv {
		t.Errorf("Expected distribution to be %v, but got %v", expected_dist, *dist)
	}
	
	min_count, max_count, count_dist := user_info_map.getFollowersDistribution(1)
	if min_count != 0 {
		t.Errorf("Expected min count to be %d but got %d", 10, min_count) 
	}
	if max_count != 3 {
		t.Errorf("Expected max count to be %d but got %d", 40, max_count) 
	}
	expected_count_dist := []int{1, 1, 1, 1}
	count_dist_eqv := (len(*count_dist) == len(expected_count_dist))
	if count_dist_eqv {
		for i, _ := range expected_count_dist {
			if expected_count_dist[i] != (*count_dist)[i] {
				count_dist_eqv = false
			}
		}
	}
	if !count_dist_eqv {
		t.Errorf("Expected count distribution to be %v but got %v", expected_count_dist, *count_dist)
	}
}

func TestUserInteractionMap(t *testing.T) {
	user_interaction := []struct {
		reposter_id   uint64
		original_id   uint64
		retweet_count uint64
		retweet_prob  float32
	}{
		{1, 2, 1, 0.2},
		{1, 3, 2, 0.4},
		{1, 4, 2, 0.4},
		{2, 1, 3, 0.3},
		{2, 4, 7, 0.7},
	}

	interaction_map := newUserInteracionMap(100)
	for _, v := range user_interaction {
		interaction_map.addInteractions(v.original_id, v.reposter_id, v.retweet_count)
	}
	interaction_map.finalize()

	for _, v := range user_interaction {
		retweet_prob := interaction_map.getRetweetProb(v.original_id, v.reposter_id)
		if math.Abs(float64(retweet_prob-v.retweet_prob)) > 0.000001 {
			t.Errorf("Expected retweet probability of %d by %d to be %f, but got %f",
				v.original_id, v.reposter_id, v.retweet_prob, retweet_prob)
		}
	}
}

func TestLoadSpreadModelData(t *testing.T) {
	const active_rate_file = "active_rate_test_file.txt"
	const interaction_rate_file = "interaction_rate_test_file.txt"

	user_list := []struct {
		id            uint64
		retweet_count uint64
		factor        float32
	}{
		{1, 10, 0.4},
		{2, 20, 0.8},
		{3, 30, 1.2},
		{4, 40, 1.6},
	}

	user_interaction := []struct {
		reposter_id   uint64
		original_id   uint64
		retweet_count uint64
		retweet_prob  float32
	}{
		{1, 2, 1, 0.2},
		{1, 3, 2, 0.4},
		{1, 4, 2, 0.4},
		{2, 1, 3, 0.3},
		{2, 4, 7, 0.7},
		{3, 4, 1, 0.5},
		{3, 2, 1, 0.5},
		{4, 1, 1, 0.25},
		{4, 2, 2, 0.5},
		{4, 3, 1, 0.25},
	}

	func() {
		active_rate_file_fd, err := os.Create(active_rate_file)
		defer func() {
			active_rate_file_fd.Close()
		}()
		if err != nil {
			t.Fatalf("Failed to create file [%s]", active_rate_file)
		}
		for _, v := range user_list {
			active_rate_file_fd.WriteString(fmt.Sprintf("%d\t%d\n", v.id, v.retweet_count))
		}

		interaction_rate_file_fd, err := os.Create(interaction_rate_file)
		defer func() {
			interaction_rate_file_fd.Close()
		}()
		if err != nil {
			t.Fatalf("Failed to create file [%s]", active_rate_file)
		}
		for _, v := range user_interaction {
			interaction_rate_file_fd.WriteString(fmt.Sprintf("%d\t%d\t%d\n",
				v.reposter_id, v.original_id, v.retweet_count))
		}
	}()

	defer func() {
		os.Remove(active_rate_file)
		os.Remove(interaction_rate_file)
	}()

	var simulator Simulator
	if simulator.LoadSpreadModelData(active_rate_file, interaction_rate_file) {

		user_id_list := simulator.model_data.user_id_list
		for i := 0; i < 20; i++ {
			rand_id := user_id_list.randomId()
			if rand_id < 1 || rand_id > 5 {
				t.Errorf("userIdList.randomId() returns Ids that have not been added [%d],[%v]", rand_id, user_id_list)
			}
		}

		user_info_map := simulator.model_data.user_info_map
		for _, v := range user_list {
			if !user_info_map.hasUser(v.id) {
				t.Errorf("Expected user_info to have user[%d] but found none.", v.id)
			}
			engagement_factor := user_info_map.engagement_factor(v.id)
			if math.Abs(float64(engagement_factor-v.factor)) > 0.000001 {
				t.Errorf("Expected user[%d] engagement factor to be [%f] but got [%f]",
					v.factor, engagement_factor)
			}
		}

		interaction_map := simulator.model_data.user_interact_map
		for _, v := range user_interaction {
			retweet_prob := interaction_map.getRetweetProb(v.original_id, v.reposter_id)
			if math.Abs(float64(retweet_prob-v.retweet_prob)) > 0.000001 {
				t.Errorf("Expected retweet probability of %d by %d to be %f, but got %f",
					v.original_id, v.reposter_id, v.retweet_prob, retweet_prob)
			}
		}

		parameters := simulator.GetParameters()
		parameters.Avg_retweet_rate = 0.5
		parameters.Is_random_sim = false
		parameters.Max_depth = 4

		result := simulator.RunSimulation()

		fmt.Printf("\nAverage Retweet Count for Sim1: %f\n", result.GetAverageRetweetCount())
		fmt.Print("------------------------------------------------------\n\n")

		parameters.Is_random_sim = true
		parameters.Random_sim_rounds = 1000
		result = simulator.RunSimulation()
		fmt.Printf("\nAverage Retweet Count for Sim2: %f\n", result.GetAverageRetweetCount())
		fmt.Print("------------------------------------------------------\n\n")
	} else {
		t.Errorf("Simulator.LoadSpreadModelData(%s,%s) failed", active_rate_file, interaction_rate_file)
	}
}

func TestSimulationResult(t *testing.T) {
	retweet_counts := []int{0, 0, 0, 1, 2, 2, 3, 3, 3, 4, 4, 5, 5, 6, 7, 8, 9, 10, 10, 100, 200, 500, 1000, 10000}
	//retweet_freq := []int{3, 1, 2.,
	var sim_result SimulationResult

	sum := 0
	for _, c := range retweet_counts {
		sim_result.addRetweetCount(c)
		sum += c
	}

	expected_avg_count := float32(sum) / float32(len(retweet_counts))
	if math.Abs(float64(expected_avg_count-sim_result.GetAverageRetweetCount())) > 0.00001 {
		t.Errorf("Expected average count to be %f but got %f\n", expected_avg_count, sim_result.GetAverageRetweetCount())
	}

	intervals := []int{1, 2, 3, 4, 5, 10, 15, 20, 100, 1000}
	expected_distribution := []int{3, 1, 2, 3, 2, 6, 2, 0, 0, 3, 2}

	distribution := sim_result.GetRetweetCountDistribution(&intervals)
	distribution_is_eqv := true
	if len(*distribution) != len(expected_distribution) {
		distribution_is_eqv = false
	}
	for i, _ := range *distribution {
		if (*distribution)[i] != expected_distribution[i] {
			distribution_is_eqv = false
		}
	}
	if !distribution_is_eqv {
		fmt.Printf("Expected count distribution to be %v but got %v", expected_distribution, *distribution)
	}
}
