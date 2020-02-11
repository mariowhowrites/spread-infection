package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
	"strings"
	"sync"
)

type probabilities [3]float64
type coordinatePair [2]float64
type treeList []coordinatePair
type ringList []treeList
type resultMap map[coordinatePair]probabilities

type resultPair struct {
	Tree          coordinatePair
	Probabilities probabilities
}

func main() {
	SpreadInfection()
}

// SpreadInfection is the main function
func SpreadInfection() {
	worldWidth, err := strconv.ParseFloat(os.Args[1], 64)
	if err != nil {
		fmt.Println(err)
		return
	}

	exePath, err := os.Executable()
	if err != nil {
		fmt.Println(err)
		return
	}
	lastIndex := strings.LastIndex(exePath, string(os.PathSeparator)) + 1
	exePath = exePath[:lastIndex]

	// Ring List
	ringListJSON, err := ioutil.ReadFile(exePath + "ring_list.json")
	if err != nil {
		fmt.Println(err)
		return
	}
	var ringsList ringList
	err = json.Unmarshal(ringListJSON, &ringsList)
	if err != nil {
		fmt.Println(err)
		return
	}

	// Prob SC List
	probSCListJSON, err := ioutil.ReadFile(exePath + "prob_SC_list.json")
	if err != nil {
		fmt.Println(err)
		return
	}
	var probSCList []float64
	err = json.Unmarshal(probSCListJSON, &probSCList)
	if err != nil {
		fmt.Println(err)
		return
	}

	// Prob PF List
	probPFListJSON, err := ioutil.ReadFile(exePath + "prob_PF_list.json")
	if err != nil {
		fmt.Println(err)
		return
	}
	var probPFList []float64
	err = json.Unmarshal(probPFListJSON, &probPFList)
	if err != nil {
		fmt.Println(err)
		return
	}

	// Prob HN List
	probHNListJSON, err := ioutil.ReadFile(exePath + "prob_HN_list.json")
	if err != nil {
		fmt.Println(err)
		return
	}
	var probHNList []float64
	err = json.Unmarshal(probHNListJSON, &probHNList)
	if err != nil {
		fmt.Println(err)
		return
	}

	// Potential Trees
	potentialTreesJSON, err := ioutil.ReadFile(exePath + "potential_trees.json")
	if err != nil {
		fmt.Println(err)
		return
	}
	var potentialTrees treeList
	err = json.Unmarshal(potentialTreesJSON, &potentialTrees)
	if err != nil {
		fmt.Println(err)
		return
	}

	// Newly Infected Trees
	treesNewlyInfectedJSON, err := ioutil.ReadFile(exePath + "trees_newly_infected.json")
	if err != nil {
		fmt.Println(err)
		return
	}
	var treesNewlyInfected treeList
	err = json.Unmarshal(treesNewlyInfectedJSON, &treesNewlyInfected)
	if err != nil {
		fmt.Println(err)
		return
	}

	// Losing Infection Trees
	treesLosingInfectionJSON, err := ioutil.ReadFile(exePath + "trees_losing_infection.json")
	if err != nil {
		fmt.Println(err)
		return
	}
	var treesLosingInfection treeList
	err = json.Unmarshal(treesLosingInfectionJSON, &treesLosingInfection)
	if err != nil {
		fmt.Println(err)
		return
	}

	potentialMap := make(map[coordinatePair]bool)

	for _, tree := range potentialTrees {
		potentialMap[tree] = true
	}

	potentialTrees = nil

	doneChannel := make(chan bool)
	var resultsMap sync.Map

	for ringIndex := range ringsList {
		ring := ringsList[ringIndex]

		var probSC float64
		if ringIndex < len(probSCList) {
			probSC = probSCList[ringIndex]
		} else {
			probSC = 0
		}

		probPF := probPFList[ringIndex]
		probHN := probHNList[ringIndex]

		probs := probabilities{probSC, probPF, probHN}

		go processRing(doneChannel, &resultsMap, &potentialMap, &ring, probs, &treesNewlyInfected, &treesLosingInfection, worldWidth)
	}

	ringCount := len(ringsList)
	processedRings := 0

ResultProcessLoop:
	for {
		select {
		case <-doneChannel:
			processedRings++

			if processedRings == ringCount {
				break ResultProcessLoop
			}
		}
	}

	type exportPair struct {
		Tree          treeList
		Probabilities probabilities
	}

	exportsSlice := make([]exportPair, 0, 0)

	resultsMap.Range(func(resultTreeInterface interface{}, probsInterface interface{}) bool {
		resultTree := resultTreeInterface.(coordinatePair)
		probs := probsInterface.(probabilities)

		exportsSlice = append(exportsSlice, exportPair{
			Tree:          []coordinatePair{resultTree},
			Probabilities: probs,
		})

		return true
	})

	resultsJSON, err := json.Marshal(exportsSlice)
	os.Stdout.Write(resultsJSON)
}

func processRing(
	doneChannel chan bool,
	resultsMap *sync.Map,
	potentialTrees *map[coordinatePair]bool,
	ring *treeList,
	probs probabilities,
	treesNewlyInfected *treeList,
	treesLosingInfection *treeList,
	worldWidth float64,
) {
	for _, neighbor := range *ring {
		for _, infectedTree := range *treesNewlyInfected {
			absoluteTree := coordinatePair{neighbor[0] + infectedTree[0], neighbor[1] + infectedTree[1]}

			if absoluteTree[0] < 0 {
				absoluteTree[0] = absoluteTree[0] + worldWidth
			}

			if absoluteTree[0] > worldWidth-1 {
				absoluteTree[0] = absoluteTree[0] - worldWidth
			}

			if absoluteTree[1] < -0.5 {
				absoluteTree[1] = absoluteTree[1] + worldWidth
			}

			if absoluteTree[1] > worldWidth-1 {
				absoluteTree[1] = absoluteTree[1] - worldWidth
			}

			_, found := (*potentialTrees)[absoluteTree]

			if found {
				processNewlyInfectedTree(resultPair{Tree: absoluteTree, Probabilities: probs}, resultsMap)
			}
		}

		for _, losingInfectionTree := range *treesLosingInfection {
			absoluteTree := coordinatePair{neighbor[0] + losingInfectionTree[0], neighbor[1] + losingInfectionTree[1]}

			if absoluteTree[0] < 0 {
				absoluteTree[0] = absoluteTree[0] + worldWidth
			}

			if absoluteTree[0] > worldWidth-1 {
				absoluteTree[0] = absoluteTree[0] - worldWidth
			}

			if absoluteTree[1] < -0.5 {
				absoluteTree[1] = absoluteTree[1] + worldWidth
			}

			if absoluteTree[1] > worldWidth-1 {
				absoluteTree[1] = absoluteTree[1] - worldWidth
			}

			_, found := (*potentialTrees)[absoluteTree]

			if found {
				processLosingInfectionTree(resultPair{Tree: absoluteTree, Probabilities: probs}, resultsMap)
			}

		}
	}

	doneChannel <- true
}

func processLosingInfectionTree(losingInfectionTree resultPair, resultsMap *sync.Map) {
	var losingInfectionProbs probabilities

	for i, probability := range losingInfectionTree.Probabilities {
		losingInfectionProbs[i] = -probability
	}

	probsInterface, found := (*resultsMap).LoadOrStore(losingInfectionTree.Tree, losingInfectionProbs)

	if found == true {
		probs := probsInterface.(probabilities)

		reduceProbabilities(probs, losingInfectionTree, resultsMap)
	}
}

func processNewlyInfectedTree(infectedTree resultPair, resultsMap *sync.Map) {
	probsInterface, found := (*resultsMap).LoadOrStore(infectedTree.Tree, infectedTree.Probabilities)

	if found == true {
		probs := probsInterface.(probabilities)

		increaseProbabilities(probs, infectedTree, resultsMap)
	}
}

func reduceProbabilities(probs probabilities, losingInfectionTree resultPair, resultsMap *sync.Map) {
	(*resultsMap).Store(losingInfectionTree.Tree, probabilities{
		probs[0] - losingInfectionTree.Probabilities[0],
		probs[1] - losingInfectionTree.Probabilities[1],
		probs[2] - losingInfectionTree.Probabilities[2],
	})
}

func increaseProbabilities(probs probabilities, infectedTree resultPair, resultsMap *sync.Map) {
	(*resultsMap).Store(infectedTree.Tree, probabilities{
		infectedTree.Probabilities[0] + probs[0],
		infectedTree.Probabilities[1] + probs[1],
		infectedTree.Probabilities[2] + probs[2],
	})
}
