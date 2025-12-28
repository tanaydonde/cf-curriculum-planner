package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"unicode"
	"strings"
)

type CFProblem struct {
	ContestID int `json:"contestId"`
	Index string `json:"index"`
	Name string `json:"name"`
	Rating int `json:"rating"`
	Tags []string `json:"tags"`
}

type CFResponse struct {
	Status string `json:"status"`
	Result struct { Problems []CFProblem `json:"problems"`} `json:"result"`
}

type Problem struct {
    ID string `json:"problem_id"`
	Name string `json:"name"`
    Rating int `json:"rating"`
    Tags []string `json:"tags"`
}

func getProblems() ([]CFProblem, error) {
	resp, err := http.Get("https://codeforces.com/api/problemset.problems")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var apiData CFResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiData); err != nil {
		return nil, err
	}

	return apiData.Result.Problems, nil
}

func cyrillic(s string) bool {
	for _, r := range s {
		if unicode.Is(unicode.Cyrillic, r) {
			return true
		}
	}
	return false
}

func getDisplayName(topic string) string {
	if topic == "tree dp" {
		return "Tree DP"
	} else if topic == "dynamic programming" {
		return "DP"
	}

	words := strings.Fields(topic)
	for i, w := range words {
		if len(w) > 0 {
			words[i] = strings.ToUpper(w[:1]) + w[1:]
		}
	}
	return strings.Join(words, " ")
}

func linkTopics(parent string, child string) {
	query := `
		INSERT INTO topic_dependencies (parent_id, child_id)
		SELECT p.id, c.id FROM topics p, topics c WHERE p.slug = $1 AND c.slug = $2
		ON CONFLICT (parent_id, child_id) DO NOTHING
	`
	_, err := conn.Exec(context.Background(), query, parent, child)
    if err != nil {
        fmt.Printf("error linking %s -> %s: %v\n", parent, child, err)
        return
    }
}

func createRoadMap() {
	// 19 edges in total
	linkTopics("implementation", "ad hoc")
	linkTopics("implementation", "sortings")
	linkTopics("implementation", "data structures")
	linkTopics("implementation", "greedy")
	linkTopics("implementation", "math")
	linkTopics("implementation", "strings")

	linkTopics("sortings", "two pointers")
	linkTopics("sortings", "searching")

	linkTopics("data structures", "searching")
	linkTopics("data structures", "graphs")

	linkTopics("greedy", "dynamic programming")

	linkTopics("math", "advanced math")
	linkTopics("math", "geometry")

	linkTopics("strings", "advanced strings")

	linkTopics("searching", "meet in the middle")

	linkTopics("dynamic programming", "tree dp")

	linkTopics("trees", "tree dp")

	linkTopics("graphs", "advanced graphs")
	linkTopics("graphs", "trees")
}

func createTopics(tagMap map[string]string) {
	uniqueTopics := make(map[string]bool)
	for _, topicSlug := range tagMap {
		uniqueTopics[topicSlug] = true
	}

	for slug := range uniqueTopics {
		query := `
			INSERT INTO topics (slug, display_name)
			VALUES ($1, $2)
			ON CONFLICT (slug) DO UPDATE
			SET display_name = EXCLUDED.display_name
		`
		_, err := conn.Exec(context.Background(), query, slug, getDisplayName(slug))
		if err != nil {
			fmt.Printf("could not save topic %s: %v\n", slug, err)
		}
	}
	// for tree dp
	slug := "tree dp"
	query := `
		INSERT INTO topics (slug, display_name)
		VALUES ($1, $2)
		ON CONFLICT (slug) DO UPDATE
		SET display_name = EXCLUDED.display_name
	`
	_, err := conn.Exec(context.Background(), query, slug, getDisplayName(slug))
	if err != nil {
		fmt.Printf("could not save topic %s: %v\n", slug, err)
	}
}

func saveProblemsToDB(problems []CFProblem, tagMap map[string]string) {
	for _, p := range problems {
		if p.Rating == 0 {
			continue
		}

		if cyrillic(p.Name) {
			continue
		}

		topicSet := make(map[string]bool)
		hasDP := false
		hasTrees := false
		for _, tag := range p.Tags {
			if topic, ok := tagMap[tag]; ok {
				topicSet[topic] = true
				if topic == "dynamic programming" {
					hasDP = true
				}
				if topic == "trees" {
					hasTrees = true
				}
			}
		}
		if hasDP && hasTrees {
			topicSet["tree dp"] = true
		}

		filtered := make([]string, 0, len(topicSet))
		for topic := range topicSet {
			filtered = append(filtered, topic)
		}
		
		if len(filtered) == 0 {
			continue
		}

		problemID := fmt.Sprintf("%d%s", p.ContestID, p.Index)
		
		query := `
			INSERT INTO problems (problem_id, name, rating, tags)
			VALUES ($1, $2, $3, $4)
			ON CONFLICT (problem_id) DO UPDATE
			SET rating = EXCLUDED.rating, tags = EXCLUDED.tags;
		`
		_, err := conn.Exec(context.Background(), query, problemID, p.Name, p.Rating, filtered)
		if err != nil {
			fmt.Printf("could not save problem %s: %v\n", problemID, err)
		}
	}
	//fmt.Println("total problem count:", count)
	fmt.Println("all rated problems saved successfully")
}

func createTables(problems []CFProblem) {
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
	saveProblemsToDB(problems, tagMap)
	createTopics(tagMap)
	createRoadMap()
}

type Node struct {
	ID int `json:"id"`
	Slug string `json:"slug"`
	DisplayName string `json:"display_name"`
}

type Edge struct {
	From int `json:"from"` //parent
	To int `json:"to"` //child
}

type GraphResponse struct {
    Nodes []Node `json:"nodes"`
    Edges []Edge `json:"edges"`
}

func getGraphHandler(w http.ResponseWriter, r *http.Request) {
	var g GraphResponse
	g.Nodes = []Node{}
	g.Edges = []Edge{}

	//getting nodes
	nodeRows, err := conn.Query(context.Background(), "SELECT id, slug, display_name FROM topics")
    if err != nil {
        http.Error(w, "failed to fetch nodes", http.StatusInternalServerError)
        return
    }
    defer nodeRows.Close()

	for nodeRows.Next() {
		var n Node
        if err := nodeRows.Scan(&n.ID, &n.Slug, &n.DisplayName); err == nil {
            g.Nodes = append(g.Nodes, n)
        }
	}

	//getting edges
	edgeRows, err := conn.Query(context.Background(), "SELECT parent_id, child_id FROM topic_dependencies")
    if err != nil {
        http.Error(w, "failed to fetch edges", http.StatusInternalServerError)
        return
    }
    defer edgeRows.Close()

    for edgeRows.Next() {
        var e Edge
        if err := edgeRows.Scan(&e.From, &e.To); err == nil {
            g.Edges = append(g.Edges, e)
        }
    }

	w.Header().Set("Content-Type", "application/json")
    w.Header().Set("Access-Control-Allow-Origin", "*")
    json.NewEncoder(w).Encode(g)
}