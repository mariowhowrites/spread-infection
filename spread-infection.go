package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
	"strings"
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
	// startTime := time.Now()

	// sampleTree := coordinatePair{0, -0.5}
	worldWidth, err := strconv.ParseFloat(os.Args[1], 64)
	if err != nil {
		fmt.Println(err)
		return
	}

	// var worldWidth float64
	// worldWidth = 1000

	exePath, err := os.Executable()
	if err != nil {
		fmt.Println(err)
		return
	}
	lastIndex := strings.LastIndex(exePath, string(os.PathSeparator)) + 1
	exePath = exePath[:lastIndex]

	// exePath := "/Users/mariovega/go/infected-trees/"
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

	newlyInfectedChannel := make(chan []resultPair)
	losingInfectionChannel := make(chan []resultPair)
	doneChannel := make(chan bool)

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

		go processRing(doneChannel, newlyInfectedChannel, losingInfectionChannel, &ring, probs, &treesNewlyInfected, &treesLosingInfection, worldWidth)
	}

	ringCount := len(ringsList)
	processedRings := 0
	resultsMap := make(resultMap, 0)

ResultProcessLoop:
	for {
		select {
		case infectedTrees := <-newlyInfectedChannel:
			processNewlyInfectedTrees(&infectedTrees, &resultsMap)
		case losingInfectionTrees := <-losingInfectionChannel:
			processLosingInfectionTrees(&losingInfectionTrees, &resultsMap)
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

	exportsSlice := make([]exportPair, 0, len(resultsMap))

	for resultTree, probs := range resultsMap {
		exportsSlice = append(exportsSlice, exportPair{
			Tree:          []coordinatePair{resultTree},
			Probabilities: probs,
		})
	}

	// elapsed := time.Since(startTime)
	// fmt.Println("time elapsed: ", elapsed)

	resultsJSON, err := json.Marshal(exportsSlice)
	os.Stdout.Write(resultsJSON)
}

func processRing(
	doneChannel chan bool,
	newlyInfectedChannel chan []resultPair,
	losingInfectionChannel chan []resultPair,
	ring *treeList,
	probs probabilities,
	treesNewlyInfected *treeList,
	treesLosingInfection *treeList,
	worldWidth float64,
) {
	for _, neighbor := range *ring {
		var potentialInfectedTrees []resultPair
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

			potentialInfectedTrees = append(potentialInfectedTrees, resultPair{
				Tree:          absoluteTree,
				Probabilities: probs,
			})
		}
		newlyInfectedChannel <- potentialInfectedTrees

		var potentialLosingInfectionTrees []resultPair

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

			potentialLosingInfectionTrees = append(potentialLosingInfectionTrees, resultPair{
				Tree:          absoluteTree,
				Probabilities: probs,
			})
		}
		losingInfectionChannel <- potentialLosingInfectionTrees
	}

	doneChannel <- true

}

func processLosingInfectionTrees(losingInfectionTrees *[]resultPair, resultsMap *resultMap) {
	for _, losingInfectionTree := range *losingInfectionTrees {
		found := maybeReduceProbabilities(losingInfectionTree, resultsMap)

		if found == false {
			addLosingInfectionTree(losingInfectionTree, resultsMap)
		}
	}
}

func processNewlyInfectedTrees(newlyInfectedTrees *[]resultPair, resultsMap *resultMap) {
	for _, infectedTree := range *newlyInfectedTrees {
		found := maybeIncreaseProbabilities(infectedTree, resultsMap)

		if found == false {
			addNewlyInfectedTree(infectedTree, resultsMap)
		}
	}
}

func maybeReduceProbabilities(losingInfectionTree resultPair, resultsMap *resultMap) bool {
	probs, found := (*resultsMap)[losingInfectionTree.Tree]

	if found == false {
		return false
	}

	reduceProbabilities(probs, losingInfectionTree, resultsMap)

	return false
}

func reduceProbabilities(probs probabilities, losingInfectionTree resultPair, resultsMap *resultMap) {
	(*resultsMap)[losingInfectionTree.Tree] = probabilities{
		probs[0] - losingInfectionTree.Probabilities[0],
		probs[1] - losingInfectionTree.Probabilities[1],
		probs[2] - losingInfectionTree.Probabilities[2],
	}
}

func addLosingInfectionTree(losingInfectionTree resultPair, resultsMap *resultMap) {
	var losingInfectionProbs probabilities

	for i, probability := range losingInfectionTree.Probabilities {
		losingInfectionProbs[i] = -probability
	}

	losingInfectionTree.Probabilities = losingInfectionProbs

	(*resultsMap)[losingInfectionTree.Tree] = losingInfectionTree.Probabilities
}

func maybeIncreaseProbabilities(infectedTree resultPair, resultsMap *resultMap) bool {
	probs, found := (*resultsMap)[infectedTree.Tree]

	if found == false {
		return false
	}

	increaseProbabilities(probs, infectedTree, resultsMap)

	return true
}

func increaseProbabilities(probs probabilities, infectedTree resultPair, resultsMap *resultMap) {
	(*resultsMap)[infectedTree.Tree] = probabilities{
		infectedTree.Probabilities[0] + probs[0],
		infectedTree.Probabilities[1] + probs[1],
		infectedTree.Probabilities[2] + probs[2],
	}
}

func addNewlyInfectedTree(infectedTree resultPair, resultsMap *resultMap) {
	(*resultsMap)[infectedTree.Tree] = infectedTree.Probabilities
}
