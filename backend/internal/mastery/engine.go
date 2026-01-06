package mastery

import (
	"container/list"
	"context"
	"encoding/json"
	"fmt"
	"math"
	"math/rand"
	"net/http"
	"regexp"
	"sort"
	"strconv"
	"time"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
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

func syncUser(conn *pgxpool.Pool, handle string, tagMap map[string]string, ancestry models.AncestryMap) error {
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

	fmt.Println("fetched all the problems needed to update. now inserting...")

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

	problemUpserts := make([]ProblemUpsert, 0, len(problemHistory))
	binAgg := make(map[BinKey]*BinAgg)

    for id, subs := range problemHistory {
        var firstOK *CFSubmission
        attempts := 0
        for i := len(subs) - 1; i >= 0; i-- {
			if subs[i].Verdict == "COMPILATION_ERROR" || subs[i].Verdict == "SKIPPED" || subs[i].Verdict == "TESTING" {
				continue
			}
            attempts++
            if subs[i].Verdict == "OK" {
                firstOK = &subs[i]
                break
            }
        }

		if firstOK != nil {
			solvedAt := time.Unix(firstOK.CreationTimeSeconds, 0).UTC()

			problemUpserts = append(problemUpserts, ProblemUpsert{
				ProblemID: id, Status: "solved", T: solvedAt,
			})

            sub := Submission{
                ID: id,
                Rating: firstOK.Problem.Rating,
                Attempts: attempts,
                TopicSlugs: getTopicSlugs(firstOK.Problem.Tags, tagMap),
                SolvedAt: solvedAt,
            }
            
            accumulateSubmission(binAgg, sub, tagMap, ancestry)
        } else {
			last := subs[0]
			lastAt := time.Unix(last.CreationTimeSeconds, 0).UTC()
			problemUpserts = append(problemUpserts, ProblemUpsert{
				ProblemID: id, Status: "unsolved", T: lastAt,
			})
		}
	}

	//updating user_problems
	err = bulkUpsertUserProblems(tx, handle, problemUpserts)
	if err != nil {
		return err
	}
	
	//updating user_interval_stats
	err = bulkUpsertUserIntervalStats(tx, handle, binAgg)
	if err != nil {
		return err
	}

	topics, err := loadAllTopicBins(tx, handle, tagMap)
	if err != nil {
		return err
	}
	if err := fillAllTopicMasteryBatch(tx, handle, nowBinIdx, topics); err != nil {
		return err
	}
    return tx.Commit(context.Background())
}

func accumulateSubmission(binAgg map[BinKey]*BinAgg, sub Submission, tagMap map[string]string, ancestry models.AncestryMap) {
	base := getBaseRating(sub.Rating, sub.Attempts)
	binIdx := getAbsoluteBinIdx(sub.SolvedAt)
	for topic := range getTopics(tagMap) {
		m := getMultiplier(topic, sub, ancestry)
		if m <= 0 {
			continue
		}
		credit := base * m

		key := BinKey{Topic: topic, BinIdx: binIdx}
		a := binAgg[key]
		if a == nil {
			a = &BinAgg{}
			binAgg[key] = a
		}
		a.Credits = append(a.Credits, credit)
		a.Multipliers = append(a.Multipliers, m)
	}
}

func bulkUpsertUserProblems(tx pgx.Tx, handle string, problemUpserts []ProblemUpsert) error {
	if len(problemUpserts) == 0 {
		return nil
	}
	var b pgx.Batch
	for _, pu := range problemUpserts {
		b.Queue(`
			INSERT INTO user_problems (handle, problem_id, status, last_attempted_at)
			VALUES ($1, $2, $3, $4)
			ON CONFLICT (handle, problem_id) DO UPDATE SET
				status = CASE
					WHEN user_problems.status = 'solved' THEN 'solved'
					ELSE EXCLUDED.status
				END,
				last_attempted_at = GREATEST(user_problems.last_attempted_at, EXCLUDED.last_attempted_at)
		`, handle, pu.ProblemID, pu.Status, pu.T)
	}

	br := tx.SendBatch(context.Background(), &b)
	defer br.Close()
	for range problemUpserts {
		if _, err := br.Exec(); err != nil {
			return err
		}
	}

	return nil
}

func bulkUpsertUserIntervalStats(tx pgx.Tx, handle string, binAgg map[BinKey]*BinAgg) error {
	if len(binAgg) == 0 {
		return nil
	}
	topics := make([]string, 0, len(binAgg))
	bins := make([]int32, 0, len(binAgg))
	for k := range binAgg {
		topics = append(topics, k.Topic)
		bins = append(bins, int32(k.BinIdx))
	}

	rows, err := tx.Query(context.Background(), `
		SELECT s.topic_slug, s.bin_idx, s.credits, s.multipliers
		FROM user_interval_stats s
		JOIN UNNEST($2::text[], $3::int[]) AS u(topic_slug, bin_idx)
		ON s.topic_slug = u.topic_slug AND s.bin_idx = u.bin_idx
		WHERE s.handle = $1
	`, handle, topics, bins)

	if err != nil {
		return err
	}

	defer rows.Close()

	for rows.Next() {
		if err := rows.Err(); err != nil {
			return err
		}
		var topic string
		var binIdx int
		var credits, multipliers []float64
		if err := rows.Scan(&topic, &binIdx, &credits, &multipliers); err != nil {
			return err
		}

		key := BinKey{Topic: topic, BinIdx: binIdx}
		a := binAgg[key]
		if a == nil {
			continue
		}

		a.Credits = append(credits, a.Credits...)
		a.Multipliers = append(multipliers, a.Multipliers...)
	}

	var b2 pgx.Batch
	for key, a := range binAgg {
		attributes := make([]SolveAttributes, 0, len(a.Credits))
		for i := range a.Credits {
			attributes = append(attributes, SolveAttributes{
				BaseRating:  a.Credits[i] / a.Multipliers[i],
				Multiplier:  a.Multipliers[i],
			})
		}
		newBinScore := calculateIntervalBin(attributes)

		b2.Queue(`
			INSERT INTO user_interval_stats (handle, topic_slug, bin_idx, bin_score, credits, multipliers, last_updated)
			VALUES ($1, $2, $3, $4, $5, $6, NOW())
			ON CONFLICT (handle, topic_slug, bin_idx) DO UPDATE SET
				bin_score = EXCLUDED.bin_score,
				credits = EXCLUDED.credits,
				multipliers = EXCLUDED.multipliers,
				last_updated = NOW()
		`, handle, key.Topic, key.BinIdx, newBinScore, a.Credits, a.Multipliers)
	}

	br2 := tx.SendBatch(context.Background(), &b2)
	defer br2.Close()
	for range binAgg {
		if _, err := br2.Exec(); err != nil {
			return err
		}
	}

	return nil
}

func updateSubmissionFull(conn *pgxpool.Pool, handle string, problem ProblemSolveInput, tagMap map[string]string, ancestry models.AncestryMap) error {
	var problemStatus string
    err := conn.QueryRow(context.Background(), 
        "SELECT status FROM user_problems WHERE handle=$1 AND problem_id=$2", 
        handle, problem.ProblemID).Scan(&problemStatus)

    if err == nil && problemStatus == "solved" {
        return fmt.Errorf("problem %s already solved", problem.ProblemID)
    }

	sub, err := hydrateSubmission(handle, problem, tagMap)
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
	topics, err := loadAllTopicBins(tx, handle, tagMap)
	if err != nil {
		return err
	}
	if err := refreshAllTopicMasteryBatch(tx, handle, nowBinIdx, topics); err != nil {
		return err
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
        handle, submission.ID, submission.SolvedAt.UTC())
    if err != nil {
        return err
    }

	topics := getTopics(tagMap)
	zeroCount := 0
	totalTopicCount := 0
	for topic := range topics {
		multiplier := getMultiplier(topic, submission, ancestry)

		if multiplier > 0 {
			credit := base * multiplier

			err := updateBinStats(tx, handle, topic, solveBinIdx, credit, multiplier)
			if err != nil {
				return err
			}
		} else {
			zeroCount++
		}
		totalTopicCount++
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

func getAllStats(conn *pgxpool.Pool, handle string, tagMap map[string]string) (map[string]MasteryResult, error) {
	topics := getTopics(tagMap)

	tx, err := conn.Begin(context.Background())
    if err != nil {
        return nil, err
    }
    defer tx.Rollback(context.Background())

	mastery := make(map[string]MasteryResult)

	for topic := range topics {
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

func loadAllTopicBins(tx pgx.Tx, handle string, tagMap map[string]string) (map[string]map[int]float64, error) {
	topicsSet := getTopics(tagMap)
	out := make(map[string]map[int]float64, len(topicsSet))
	for topic := range topicsSet {
		out[topic] = make(map[int]float64)
	}

	rows, err := tx.Query(context.Background(), `
		SELECT topic_slug, bin_idx, bin_score
		FROM user_interval_stats
		WHERE handle = $1
	`, handle)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var topic string
		var idx int
		var score float64
		if err := rows.Scan(&topic, &idx, &score); err != nil {
			return nil, err
		}

		if _, ok := out[topic]; !ok {
			continue
		}

		out[topic][idx] = score
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return out, nil
}

func fillAllTopicMasteryBatch(tx pgx.Tx, handle string, nowBinIdx int, topics map[string]map[int]float64) error {
	var b pgx.Batch
	for topic, binMap := range topics {
		scores := getTopicScoresArr(nowBinIdx, binMap)
		res := calculateMasteryScore(scores)

		b.Queue(`
			INSERT INTO user_topic_stats (handle, topic_slug, mastery_score, peak_score, last_updated)
			VALUES ($1, $2, $3, $4, NOW())
			ON CONFLICT (handle, topic_slug) DO UPDATE SET
				mastery_score = EXCLUDED.mastery_score,
				peak_score = EXCLUDED.peak_score,
				last_updated = NOW()
		`, handle, topic, res.Current, res.Peak)
	}

	br := tx.SendBatch(context.Background(), &b)
	defer br.Close()

	for range topics {
		if _, err := br.Exec(); err != nil {
			return err
		}
	}
	return nil
}

func refreshAllTopicMasteryBatch(tx pgx.Tx, handle string, nowBinIdx int, topics map[string]map[int]float64) error {
	var b pgx.Batch
	for topic, binMap := range topics {
		scores := getTopicScoresArr(nowBinIdx, binMap)
		cur := calculateMasteryCurrentScore(scores)

		b.Queue(`
			INSERT INTO user_topic_stats (handle, topic_slug, mastery_score, peak_score, last_updated)
			VALUES ($1, $2, $3, $3, NOW())
			ON CONFLICT (handle, topic_slug) DO UPDATE SET
				mastery_score = EXCLUDED.mastery_score,
				peak_score = GREATEST(user_topic_stats.peak_score, EXCLUDED.mastery_score),
				last_updated = NOW()
		`, handle, topic, cur)
	}

	br := tx.SendBatch(context.Background(), &b)
	defer br.Close()

	for range topics {
		if _, err := br.Exec(); err != nil {
			return err
		}
	}
	return nil
}

func getTopicScoresArr(currentBinIdx int, binMap map[int]float64) []float64 {
	var scores []float64
	if len(binMap) == 0 {
		return scores
	}
	minIdx := currentBinIdx
	for idx := range binMap {
		if idx < minIdx {
			minIdx = idx
		}
	}

	for i := currentBinIdx; i >= minIdx; i-- {
		scores = append(scores, binMap[i])
	}
	return scores
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

func hydrateSubmission(handle string, problem ProblemSolveInput, tagMap map[string]string) (Submission, error) {
	url := fmt.Sprintf("https://codeforces.com/api/user.status?handle=%s", handle)
	resp, err := http.Get(url)
	if err != nil {
		return Submission{}, fmt.Errorf("misc fail: %w", err)
	}
	defer resp.Body.Close()

	var data CFUserResponse
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return Submission{}, fmt.Errorf("misc fail: %w", err)
	}

	if data.Status != "OK" {
		return Submission{}, fmt.Errorf("misc fail")
	}

	re := regexp.MustCompile(`^(\d+)([A-Za-z0-9]+)$`)
	matches := re.FindStringSubmatch(problem.ProblemID)
	if len(matches) != 3 {
		return Submission{}, fmt.Errorf("invalid problem id: %s", problem.ProblemID)
	}
	targetContestID, _ := strconv.Atoi(matches[1])
	targetIndex := matches[2]

	var problemSubs []CFSubmission
	for _, s := range data.Result {
		if s.Problem.ContestID == targetContestID && s.Problem.Index == targetIndex {
			problemSubs = append(problemSubs, s)
		}
	}

	attempts := 0
	var firstOK *CFSubmission

	for i := len(problemSubs) - 1; i >= 0; i-- {
		s := problemSubs[i]
		
		if s.Verdict == "COMPILATION_ERROR" || s.Verdict == "SKIPPED" || s.Verdict == "TESTING" {
			continue
		}
		
		attempts++
		
		if s.Verdict == "OK" {
			firstOK = &s
			break
		}
	}
	if firstOK == nil {
		return Submission{}, fmt.Errorf("problem %s has not been solved", problem.ProblemID)
	}

	rating := firstOK.Problem.Rating

	tags := getTopicSlugs(firstOK.Problem.Tags, tagMap)

    return Submission{
        ID: problem.ProblemID,
        Rating: rating,
        Attempts: attempts,
        TopicSlugs: tags,
        TimeSpentMinutes: problem.TimeSpentMinutes,
        SolvedAt: time.Unix(firstOK.CreationTimeSeconds, 0),
    }, nil
}

func recommendProblem(conn *pgxpool.Pool, handle string, topic string, targetInc int, k int) ([]CFProblemOutput, error) {
	userRatings := make(map[string]int)
	
	rows, err := conn.Query(context.Background(), `
		SELECT topic_slug, CAST(mastery_score AS INTEGER) 
		FROM user_topic_stats 
		WHERE handle = $1`, handle)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch user stats: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var slug string
		var score int
		if err := rows.Scan(&slug, &score); err != nil {
			return nil, err
		}
		userRatings[slug] = score
	}

	currentMainRating := max(userRatings[topic], 800)

	targetRating := currentMainRating + targetInc
	targetRating = max(targetRating, 800)
	
	query := `
		SELECT problem_id, name, rating, tags
		FROM problems p
		WHERE $1 = ANY(tags)
		AND rating BETWEEN $2 AND $3
		AND NOT EXISTS (
			SELECT 1 FROM user_problems up
			WHERE up.handle = $4 
			AND up.problem_id = p.problem_id
			AND up.status = 'solved'
		)
		ORDER BY ABS(rating - $5) ASC
		LIMIT 200
	`

	minRating := max(targetRating - 200, 800)
	maxRating := targetRating + 200

	pRows, err := conn.Query(context.Background(), query, topic, minRating, maxRating, handle, targetRating)
	if err != nil {
		return nil, err
	}
	defer pRows.Close()

	var candidates []CFProblemOutput
	for pRows.Next() {
		var p CFProblemOutput
		if err := pRows.Scan(&p.ID, &p.Name, &p.Rating, &p.Tags); err != nil {
			return nil, err
		}
		candidates = append(candidates, p)
	}

	finalRecommendations := make([]CFProblemOutput, 0, k)
	added := make(map[string]bool)

	margins := []int{50, 100, 150, 200, 300, 500, 1000}

	for _, margin := range margins {
		if len(finalRecommendations) >= k {
			break
		}
		for _, problem := range candidates {
			if len(finalRecommendations) >= k {
				break
			}
			if added[problem.ID] {
				continue
			}
			canAdd := true
			
			for _, tag := range problem.Tags {
				if tag == topic {
					continue
				}

				curTopicRating := max(userRatings[tag], 800)

				if problem.Rating > (curTopicRating + margin) {
					canAdd = false
					break
				}
			}

			if canAdd {
				finalRecommendations = append(finalRecommendations, problem)
				added[problem.ID] = true
			}
		}
	}
	return finalRecommendations, nil
}

func recommendDailyProblem(conn *pgxpool.Pool, handle string) (CFProblemOutput, error) {

	type TopicStat struct {
		Slug string
		Current int
		Decay int
	}

	rows, err := conn.Query(context.Background(), `
		SELECT topic_slug, 
		       CAST(mastery_score AS INTEGER) as current, 
		       CAST(peak_score - mastery_score AS INTEGER) as decay
		FROM user_topic_stats 
		WHERE handle = $1 AND mastery_score > 0`, handle)
	if err != nil {
		return CFProblemOutput{}, fmt.Errorf("failed to fetch stats: %w", err)
	}
	defer rows.Close()

	var activeTopics []TopicStat
	for rows.Next() {
		var t TopicStat
		if err := rows.Scan(&t.Slug, &t.Current, &t.Decay); err != nil {
			return CFProblemOutput{}, err
		}
		activeTopics = append(activeTopics, t)
	}

	fallback := func() (CFProblemOutput, error) {
		res, err := recommendProblem(conn, handle, "implementation", 100, 1)
		if err != nil {
			return CFProblemOutput{}, err
		}
		if len(res) > 0 {
			return res[0], nil
		}
		return CFProblemOutput{}, fmt.Errorf("no problems found")
	}

	if len(activeTopics) == 0 {
		return fallback()
	}

	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	roll := r.Intn(100)

	if roll < 50 {
		sort.Slice(activeTopics, func(i int, j int) bool {
			if activeTopics[i].Decay != activeTopics[j].Decay {
				return activeTopics[i].Decay > activeTopics[j].Decay
			}
			return activeTopics[i].Current < activeTopics[j].Current
		})
	} else if roll < 80 {
		sort.Slice(activeTopics, func(i int, j int) bool {
			if activeTopics[i].Decay != activeTopics[j].Decay {
				return activeTopics[i].Decay < activeTopics[j].Decay
			}
			return activeTopics[i].Current > activeTopics[j].Current
		})
	} else {
		r.Shuffle(len(activeTopics), func(i int, j int) {
			activeTopics[i], activeTopics[j] = activeTopics[j], activeTopics[i]
		})
	}

	count := min(3, len(activeTopics))
	candidates := activeTopics[:count]

	r.Shuffle(len(candidates), func(i, j int) {
		candidates[i], candidates[j] = candidates[j], candidates[i]
	})

	for _, topic := range candidates {
		recommendations, err := recommendProblem(conn, handle, topic.Slug, 100, 1)
		
		if err == nil && len(recommendations) > 0 {
			return recommendations[0], nil
		}
	}
	return fallback()
}

func getLastKSolves(conn *pgxpool.Pool, handle string, k int, status string) ([]CFSolveOutput, error ) {
	query := `
        SELECT p.problem_id, p.name, p.rating, p.tags, up.last_attempted_at 
        FROM user_problems up
        JOIN problems p ON up.problem_id = p.problem_id
        WHERE up.handle = $1 AND up.status = $3
        ORDER BY up.last_attempted_at DESC 
        LIMIT $2
    `

	rows, err := conn.Query(context.Background(), query, handle, k, status)
	if err != nil {
        return nil, err
    }
	defer rows.Close()

	var recentSolves []CFSolveOutput
    for rows.Next() {
        var p CFSolveOutput
        if err := rows.Scan(&p.ID, &p.Name, &p.Rating, &p.Tags, &p.SolvedAt); err != nil {
            return nil, err
        }
        recentSolves = append(recentSolves, p)
    }

	if err := rows.Err(); err != nil {
        return nil, err
    }

	return recentSolves, nil
}