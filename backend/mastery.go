package main

import (
	"container/list"
	"math"
	"time"
)

type AncestryMap map[string]map[string]int;

type SolveAttributes struct {
	BaseRating float64
	Multiplier float64
}

type Submission struct {
	ID int
	Rating int
	Attempts int
	TopicSlugs []string
	SolvedAt time.Time
}

type MasteryResult struct {
	Current float64
	Peak    float64
}

//builds the distance map
func BuildAncestryMap(nodes []Node, edges []Edge) AncestryMap {
	ancestry := make(AncestryMap)

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
func GetBaseRating(rating int, attempts int) float64 {
	if attempts <= 1 {
		return float64(rating)
	}

	const k = 0.1

	modifier := 0.5 + 0.5*math.Exp(-k*float64(attempts-1))
	
	return float64(rating) * modifier
}

//calculates M
func CalculateIntervalBin(solves []SolveAttributes) float64 {
	if len(solves) == 0 {
		return 0
	}

	var p float64 //max of c
	credits := make([]float64, len(solves)) //c array

	for i, solve := range solves {
		credits[i] = solve.BaseRating * solve.Multiplier
		if credits[i] > p {
			p = credits[i]
		}
	}

	if p == 0 {
		return 0
	}

	var numerator float64
	var denominator float64

	const K = 1.5 //confidence constant

	for i, solve := range solves {
		weight := math.Pow((credits[i]/p), 3)

		numerator += credits[i] * weight
		denominator += solve.Multiplier * weight
	}
	denominator = math.Max(denominator, K)

	return numerator/denominator
}

//computed M(i, T) given T and the array of submissions at interval i. uses CalculateIntervalBin and GetBaseRating
func GetTopicIntervalScore(targetTopic string, intervalSubmissions []Submission, ancestry AncestryMap) float64 {
	var attributes []SolveAttributes
	
	for _, submission := range intervalSubmissions {
		minDist := -1
		for _, topic := range submission.TopicSlugs {
			if dist, ok := ancestry[topic][targetTopic]; ok {
				if minDist == -1 || dist < minDist {
					minDist = dist
				}
			}
		}

		if minDist != -1 {
			base := GetBaseRating(submission.Rating, submission.Attempts)
			multipler := math.Pow(0.75, float64(minDist))
			attributes = append(attributes, SolveAttributes{base, multipler})
		}
	}
	return CalculateIntervalBin(attributes)
}

//calculates mastery score (cur and peak) given slice of interval scores
func CalculateMasteryScore(binScores []float64) MasteryResult {
	if len(binScores) == 0 {
		return MasteryResult{0, 0}
	}

	const lambda = 0.2

	var numerator float64
	var denominator float64

	var peak float64

	for i, score := range binScores {
		W := math.Exp(-lambda * float64(i))
		numerator += score * W
		denominator += W
		if score > peak {
			peak = score
		}
	}
	if denominator == 0 {
		return MasteryResult{0, 0}
	}
	return MasteryResult{numerator/denominator, peak}
}

//takes all submissions and an int n and groups them into n-day intervals
func GetBinnedSubmissions(submissions []Submission, n int) [][]Submission {
	if len(submissions) == 0 {
		return [][]Submission{}
	}

	now := time.Now()

	binToSub := make(map[int][]Submission)
	maxBinIdx := 0
	for _, sub := range submissions {
		days := int(now.Sub(sub.SolvedAt).Hours() / 24)
		
		binIdx := days / n
		
		if binIdx < 0 {
			binIdx = 0
		} 
		
		binToSub[binIdx] = append(binToSub[binIdx], sub)
		
		if binIdx > maxBinIdx {
			maxBinIdx = binIdx
		}
	}

	bins := make([][]Submission, maxBinIdx + 1)
	for i := 0; i <= maxBinIdx; i++ {
		bins[i] = binToSub[i]
	}
	return bins
}

//returns a map, mapping each topic to its current mastery score and peak mastery score
func CalculateAllTopicMasteries(topics []string, submissions []Submission, ancestry AncestryMap, n int) map[string] MasteryResult{
	results := make(map[string]MasteryResult)

	binnedSubs := GetBinnedSubmissions(submissions, n)

	for _, topicSlug := range topics {
		var binScores []float64

		for _, intervalSubs := range binnedSubs {
			score := GetTopicIntervalScore(topicSlug, intervalSubs, ancestry)
			binScores = append(binScores, score)
		}

		lastIdx := -1
        for i := len(binScores) - 1; i >= 0; i-- {
            if binScores[i] > 0 {
                lastIdx = i
                break
            }
        }

		if lastIdx == -1 {
            results[topicSlug] = MasteryResult{0, 0}
            continue
        }
		filteredBins := binScores[:lastIdx+1]

		results[topicSlug] = CalculateMasteryScore(filteredBins)
	}

	return results
}