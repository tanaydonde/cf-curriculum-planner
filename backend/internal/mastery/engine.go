package mastery

import (
	"container/list"
	"context"
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/tanaydonde/cf-curriculum-planner/backend/internal/models"
)

const N = 14

func GetTagMap() map[string]string {
	tagMap := map[string]string{
		// foundation
		"implementation": "implementation",
		"brute force": "implementation",

		// ad-hoc
		"constructive algorithms": "ad hoc",

		// sorting
		"sortings": "sortings",

		// two pointers
		"two pointers": "two pointers",

		// searching
		"binary search": "searching",
		"ternary search": "searching",
		"divide and conquer": "searching",

		// meet-in-the-middle
		"meet-in-the-middle": "meet in the middle",

		// greedy
		"greedy": "greedy",

		// math + advanced math
		"math": "math",
		"number theory": "math",
		"combinatorics": "math",
		"matrices": "math",
		"probabilities": "math",
		"fft": "advanced math",
		"chinese remainder theorem": "advanced math",

		// geometry
		"geometry": "geometry",

		// graphs + advanced graphs
		"graphs": "graphs",
		"dfs and similar": "graphs",
		"shortest paths": "graphs",
		"dsu": "graphs",
		"flows": "advanced graphs",
		"graph matchings": "advanced graphs",
		"2-sat": "advanced graphs",

		// trees
		"trees": "trees",

		// strings + advanced strings
		"strings": "strings",
		"hashing": "strings",
		"string suffix structures": "advanced strings",

		// data structures
		"data structures": "data structures",
		"bitmasks": "data structures",

		// dp
		"dp": "dynamic programming",
	}
	return tagMap
}

func BuildAncestryMap(nodes []models.Node, edges []models.Edge) models.AncestryMap {
	ancestry := make(models.AncestryMap)

	idToSlug := make(map[int]string)
	adjlist := make(map[int][]int)
	for _, node := range nodes {
		idToSlug[node.ID] = node.Slug
		ancestry[node.Slug] = make(map[string]int)
		ancestry[node.Slug][node.Slug] = 0
	}
	
	for _, edge := range edges {
		adjlist[edge.To] = append(adjlist[edge.To], edge.From)
	}

	//bfs through each node
	for _, node := range nodes {

		type pair struct {
			id int
			dist int
		}

		q := list.New()
		q.PushBack(pair{node.ID, 0})

		//bfs
		for q.Len() > 0 {
			elem := q.Front()
			cur := elem.Value.(pair)
			q.Remove(elem)

			curSlug := idToSlug[cur.id]
			ancestry[node.Slug][curSlug] = cur.dist
			
			for _, neighbor := range adjlist[cur.id] {
				neighborSlug := idToSlug[neighbor]
				if _, ok := ancestry[node.Slug][neighborSlug]; !ok {
					ancestry[node.Slug][neighborSlug] = cur.dist + 1
					q.PushBack(pair{neighbor, cur.dist+1})
				}
			}
		}
	}

	return ancestry
}

//calculates B
func getBaseRating(rating int, attempts int) float64 {
	if attempts <= 1 {
		return float64(rating)
	}

	const k = 0.1

	modifier := 0.5 + 0.5*math.Exp(-k*float64(attempts-1))
	
	return float64(rating) * modifier
}

func getBaseRatingTime(rating int, attempts int, timeSpentMinutes int) float64 {
	base := getBaseRating(rating, attempts)
	const (
		tAvg = 45
		floor = 0.85
		k = 10
	)

	speedFactor := (tAvg + k)/(float64(timeSpentMinutes) + k)
	
	speedMultiplier := floor + (1-floor)*speedFactor
	return base * speedMultiplier
}

//calculates M given a B(j) and multipliers(j) for all j in the interval
func calculateIntervalBin(solves []SolveAttributes) float64 {
	if len(solves) == 0 {
		return 0
	}

	var p float64 //max of c
	credits := make([]float64, len(solves)) //c array
	multiplier := make([]float64, len(solves)) //multipliers array

	for i, solve := range solves {
		credits[i] = solve.BaseRating * solve.Multiplier
		multiplier[i] = solve.Multiplier
		if credits[i] > p {
			p = credits[i]
		}
	}

	if p == 0 {
		return 0
	}

	var numerator, denominator float64

	const K = 1.5 //confidence constant

	for i, solve := range solves {
		weight := math.Pow((credits[i]/p), 3)

		numerator += credits[i] * weight
		denominator += solve.Multiplier * weight
	}
	denominator = math.Max(denominator, K)
	score := numerator/denominator

	return score
}

func getMultiplier(targetTopic string, submission Submission, ancestry models.AncestryMap) float64 {
	multiplier := float64(0)
	minDist := -1
	for _, topic := range submission.TopicSlugs {
		if dist, ok := ancestry[topic][targetTopic]; ok {
			if minDist == -1 || dist < minDist {
				minDist = dist
			}
		}
	}

	if minDist != -1 {
		multiplier = math.Pow(0.75, float64(minDist))
	}
	return multiplier
}

//calculates mastery score (cur and peak) given slice of interval scores
func calculateMasteryScore(binScores []float64) MasteryResult {
	currentScore := calculateMasteryCurrentScore(binScores)
	peakScore := currentScore
	for i := 1; i < len(binScores); i++ {
		score := calculateMasteryCurrentScore(binScores[i:])
		if score > peakScore {
			peakScore = score
		}
	}
	return MasteryResult{currentScore, peakScore}
}

func calculateMasteryCurrentScore(binScores []float64) float64 {
	if len(binScores) == 0 {
		return 0
	}

	var p float64
	for _, score := range binScores {
		if score > p {
			p = score
		}
	}

	if p == 0 {
		return 0
	}

	const lambda = 0.05
	const K = 1.2

	var numerator float64
	var denominator float64

	for i, score := range binScores {
		timeWeight := math.Exp(-lambda * float64(i))
		qualityWeight := math.Pow(score/p, 3)

		totalWeight := timeWeight * qualityWeight

		numerator += score * totalWeight
		denominator += totalWeight
	}
	if denominator == 0 {
		return 0
	}
	return numerator/math.Max(denominator, K)
}

//returns index of bin given a time
func getAbsoluteBinIdx(t time.Time) int {
    return int(t.Unix() / int64(N*86400))
}

//returns all topics
func getTopics(tagMap map[string]string) map[string]bool {
	topics := make(map[string]bool)
	topics["tree dp"] = true
	for _, topic := range tagMap {
		topics[topic] = true
	}
	return topics
}

//returns all topic slugs for a problem given a slice of its tags
func getTopicSlugs(problemTags []string, tagMap map[string]string) []string {
	tagSlugMap := make(map[string]bool)
	tree := false
	dp := false
	for _, topic := range problemTags {
		if tag, ok := tagMap[topic]; ok {
			tagSlugMap[tag] = true
			if tag == "trees" {
				tree = true
			}
			if tag == "dynamic programming" {
				dp = true
			}
		}
	}
	if tree && dp {
		tagSlugMap["tree dp"] = true
	}
	var tagSlug []string
	for tag := range tagSlugMap {
		tagSlug = append(tagSlug, tag)
	}
	return tagSlug
}

func syncUser(conn *pgx.Conn, handle string, tagMap map[string]string, ancestry models.AncestryMap) error {
	url := fmt.Sprintf("https://codeforces.com/api/user.status?handle=%s", handle)
    resp, err := http.Get(url)
    if err != nil {
		return err
	}
    defer resp.Body.Close()

	var data CFUserResponse
    json.NewDecoder(resp.Body).Decode(&data)

	if data.Status == "FAILED" {
        return fmt.Errorf("handle '%s' not found or invalid", handle)
    }

	//gets problems already solved
    existingSolved := make(map[string]bool)
    rows, _ := conn.Query(context.Background(), "SELECT problem_id FROM user_problems WHERE handle = $1 AND status = 'solved'", handle)
    for rows.Next() {
        var id string
        rows.Scan(&id)
        existingSolved[id] = true
    }
    rows.Close()

	//fills problemHistory which contains information about all problems the user attempted which isn't in our db
    problemHistory := make(map[string][]CFSubmission)
    for _, s := range data.Result {
        id := fmt.Sprintf("%d%s", s.Problem.ContestID, s.Problem.Index)
        if !existingSolved[id] {
            problemHistory[id] = append(problemHistory[id], s)
        }
    }

	nowBinIdx := getAbsoluteBinIdx(time.Now())

	tx, err := conn.Begin(context.Background())
    if err != nil {
		return err
	}
    defer tx.Rollback(context.Background())

    for id, subs := range problemHistory {
        var firstOK *CFSubmission
        attempts := 0
        for i := len(subs) - 1; i >= 0; i-- {
            attempts++
            if subs[i].Verdict == "OK" {
                firstOK = &subs[i]
                break
            }
        }
		//if the user solved it call updateSubmission to update the problems solved
		if firstOK != nil {
            sub := Submission{
                ID: id,
                Rating: firstOK.Problem.Rating,
                Attempts: attempts,
                TopicSlugs: getTopicSlugs(firstOK.Problem.Tags, tagMap),
                SolvedAt: time.Unix(firstOK.CreationTimeSeconds, 0),
            }
            
            err := updateSubmission(tx, handle, sub, tagMap, ancestry)
			if err != nil {
				return err
			}
        } else { //if the user didn't solve it, add into unsolved problems
			firstOK = &subs[0]
			_, err = tx.Exec(context.Background(), `
			INSERT INTO user_problems (handle, problem_id, status, last_attempted_at)
			VALUES ($1, $2, 'unsolved', $3)
			ON CONFLICT (handle, problem_id) DO UPDATE SET
        		status = CASE 
            		WHEN user_problems.status = 'solved' THEN 'solved'
            		ELSE EXCLUDED.status
        		END,
        	last_attempted_at = GREATEST(user_problems.last_attempted_at, EXCLUDED.last_attempted_at)`,
			handle, id, time.Unix(firstOK.CreationTimeSeconds, 0))
			if err != nil {
				return err
			}
		}
	}

    for topic := range getTopics(tagMap) {
        fillTopicMastery(tx, handle, topic, nowBinIdx)
    }
    return tx.Commit(context.Background())
}

func updateSubmissionFull(conn *pgx.Conn, handle string, problem ProblemSolveInput, tagMap map[string]string, ancestry models.AncestryMap) error {
	sub, err := hydrateSubmission(conn, problem)
    if err != nil {
        return err
    }
	tx, err := conn.Begin(context.Background())
    if err != nil {
		return err
	}
    defer tx.Rollback(context.Background())

	err = updateSubmission(tx, handle, sub, tagMap, ancestry)
	if err != nil {
		return err
	}

	nowBinIdx := getAbsoluteBinIdx(time.Now())
	
	topics := getTopics(tagMap)
    for topic := range topics {
        if err := refreshTopicMastery(tx, handle, topic, nowBinIdx); err != nil {
            return err
        }
    }

	return tx.Commit(context.Background())
}

//given a submission and handle, it updates all topics in the db
func updateSubmission(tx pgx.Tx, handle string, submission Submission, tagMap map[string]string, ancestry models.AncestryMap) error {
	var base float64
	if submission.TimeSpentMinutes > 0 {
		base = getBaseRatingTime(submission.Rating, submission.Attempts, submission.TimeSpentMinutes)
	} else {
		base = getBaseRating(submission.Rating, submission.Attempts)
	}
	solveBinIdx := getAbsoluteBinIdx(submission.SolvedAt)

	_, err := tx.Exec(context.Background(), `
        INSERT INTO user_problems (handle, problem_id, status, last_attempted_at)
        VALUES ($1, $2, 'solved', $3)
        ON CONFLICT (handle, problem_id) DO UPDATE SET
            status = 'solved',
            last_attempted_at = EXCLUDED.last_attempted_at`,
        handle, submission.ID, submission.SolvedAt)
    if err != nil {
        return err
    }

	topics := getTopics(tagMap)
	for topic := range topics {
		multiplier := getMultiplier(topic, submission, ancestry)

		if multiplier > 0 {
			credit := base * multiplier

			err := updateBinStats(tx, handle, topic, solveBinIdx, credit, multiplier)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

//adds credit, multiplier into (handle, topic, binIdx) and updates the bin score
func updateBinStats(tx pgx.Tx, handle string, topic string, binIdx int, credit float64, multiplier float64) error {
	var credits, multipliers []float64
	err := tx.QueryRow(context.Background(), `
		SELECT credits, multipliers FROM user_interval_stats 
		WHERE handle = $1 AND topic_slug = $2 AND bin_idx = $3`,
		handle, topic, binIdx).Scan(&credits, &multipliers)

	if err != nil && err != pgx.ErrNoRows {
		return err
	}

	credits = append(credits, credit)
	multipliers = append(multipliers, multiplier)

	var attributes []SolveAttributes
	for i := range credits {
		attributes = append(attributes, SolveAttributes{BaseRating: credits[i] / multipliers[i], Multiplier: multipliers[i]})
	}
	newBinScore := calculateIntervalBin(attributes)

	_, err = tx.Exec(context.Background(), `
		INSERT INTO user_interval_stats (handle, topic_slug, bin_idx, bin_score, credits, multipliers, last_updated)
		VALUES ($1, $2, $3, $4, $5, $6, NOW())
		ON CONFLICT (handle, topic_slug, bin_idx) DO UPDATE SET
			bin_score = EXCLUDED.bin_score,
			credits = EXCLUDED.credits,
			multipliers = EXCLUDED.multipliers,
			last_updated = NOW()`,
		handle, topic, binIdx, newBinScore, credits, multipliers)

	return err
}

func refreshAndGetAllStats(conn *pgx.Conn, handle string, tagMap map[string]string) (map[string]MasteryResult, error) {
	currentBinIdx := getAbsoluteBinIdx(time.Now())
	topics := getTopics(tagMap)

	tx, err := conn.Begin(context.Background())
    if err != nil {
        return nil, err
    }
    defer tx.Rollback(context.Background())

	mastery := make(map[string]MasteryResult)

	for topic := range topics {
		//refreshing
		err = refreshTopicMastery(tx, handle, topic, currentBinIdx)
		if err != nil {
			return nil, err
		}

		//getting
		cur, peak, err := getUserTopicStats(tx, handle, topic)
		if err != nil {
			return nil, err
		}
		mastery[topic] = MasteryResult{cur, peak}
	}
	if err := tx.Commit(context.Background()); err != nil {
		return nil, err
	}
	return mastery, nil
}

func getTopicScoresArr(tx pgx.Tx, handle string, topic string, currentBinIdx int) ([]float64, error) {
	var scores []float64
	rows, err := tx.Query(context.Background(), `
		SELECT bin_idx, bin_score FROM user_interval_stats 
		WHERE handle = $1 AND topic_slug = $2 ORDER BY bin_idx DESC`, 
		handle, topic)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	binMap := make(map[int]float64)
	minIdx := currentBinIdx
	for rows.Next() {
		var idx int
		var score float64
		rows.Scan(&idx, &score)
		binMap[idx] = score
		if idx < minIdx {
			minIdx = idx
		}
	}

	if len(binMap) == 0 {
		return scores, nil
	}

	for i := currentBinIdx; i >= minIdx; i-- {
		scores = append(scores, binMap[i])
	}
	return scores, nil
}

func fillTopicMastery(tx pgx.Tx, handle string, topic string, currentBinIdx int) error {
	scores, err := getTopicScoresArr(tx, handle, topic, currentBinIdx)
	if err != nil {
		return err
	}

	result := calculateMasteryScore(scores)

	_, err = tx.Exec(context.Background(), `
		INSERT INTO user_topic_stats (handle, topic_slug, mastery_score, peak_score, last_updated)
		VALUES ($1, $2, $3, $4, NOW())
		ON CONFLICT (handle, topic_slug) DO UPDATE SET
			mastery_score = EXCLUDED.mastery_score,
			peak_score = EXCLUDED.peak_score,
			last_updated = NOW()`,
		handle, topic, result.Current, result.Peak)

	return err
}

func refreshTopicMastery(tx pgx.Tx, handle string, topic string, currentBinIdx int) error {
	scores, err := getTopicScoresArr(tx, handle, topic, currentBinIdx)
	if err != nil {
		return err
	}

	result := calculateMasteryCurrentScore(scores)

	_, err = tx.Exec(context.Background(), `
		INSERT INTO user_topic_stats (handle, topic_slug, mastery_score, peak_score, last_updated)
		VALUES ($1, $2, $3, $3, NOW())
		ON CONFLICT (handle, topic_slug) DO UPDATE SET
			mastery_score = EXCLUDED.mastery_score,
			peak_score = GREATEST(user_topic_stats.peak_score, EXCLUDED.mastery_score),
			last_updated = NOW()`,
		handle, topic, result)

	return err
}

func getUserTopicStats(tx pgx.Tx, handle string, topic string) (float64, float64, error) {
	var cur, peak float64
	query := `SELECT mastery_score, peak_score FROM user_topic_stats WHERE handle = $1 AND topic_slug = $2`
	
	err := tx.QueryRow(context.Background(), query, handle, topic).Scan(&cur, &peak)
	if err != nil {
		if err == pgx.ErrNoRows {
			return 0, 0, nil
		}
		return 0, 0, err
	}
	return cur, peak, nil
}

func hydrateSubmission(conn *pgx.Conn, problem ProblemSolveInput) (Submission, error) {
	var rating int
    var tags []string

    // One query to get everything we need
    query := `SELECT rating, tags FROM problems WHERE problem_id = $1`
    
    err := conn.QueryRow(context.Background(), query, problem.ProblemID).Scan(&rating, &tags)
    if err != nil {
        if err == pgx.ErrNoRows {
            return Submission{}, fmt.Errorf("problem %s not found in database", problem.ProblemID)
        }
        return Submission{}, err
    }

    return Submission{
        ID: problem.ProblemID,
        Rating: rating,
        Attempts: problem.Attempts,
        TopicSlugs: tags,
        TimeSpentMinutes: problem.TimeSpentMinutes,
        SolvedAt: time.Now(),
    }, nil
}