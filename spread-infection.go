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
type resultPair struct {
	Tree          coordinatePair
	Probabilities probabilities
}

func main() {
	// sampleTree := coordinatePair{0, -0.5}
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
	var ringList [][]coordinatePair
	err = json.Unmarshal(ringListJSON, &ringList)
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
	var potentialTrees []coordinatePair
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
	var treesNewlyInfected []coordinatePair
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
	var treesLosingInfection []coordinatePair
	err = json.Unmarshal(treesLosingInfectionJSON, &treesLosingInfection)
	if err != nil {
		fmt.Println(err)
		return
	}

	newlyInfectedChannel := make(chan resultPair)
	losingInfectionChannel := make(chan resultPair)
	doneChannel := make(chan bool)

	processRing := func(ringIndex int) {
		ring := ringList[ringIndex]

		var probSC float64
		if ringIndex < len(probSCList) {
			probSC = probSCList[ringIndex]
		} else {
			probSC = 0
		}

		probPF := probPFList[ringIndex]
		probHN := probHNList[ringIndex]

		probs := probabilities{probSC, probPF, probHN}

		processTree := func(treeChannel chan resultPair, tree coordinatePair, probs probabilities) {
			for _, potentialTree := range potentialTrees {
				if potentialTree == tree {
					treeChannel <- resultPair{Tree: tree, Probabilities: probs}
				}
			}
		}

		for _, neighbor := range ring {
			for _, infectedTree := range treesNewlyInfected {
				tree := coordinatePair{neighbor[0] + infectedTree[0], neighbor[1] + infectedTree[1]}

				if tree[0] < 0 {
					tree[0] = tree[0] + worldWidth
				}

				if tree[0] > worldWidth - 1 {
					tree[0] = tree[0] - worldWidth
				}

				if tree[1] < -0.5 {
					tree[1] = tree[1] + worldWidth
				}

				if tree[1] > worldWidth - 1 {
					tree[1] = tree[1] - worldWidth
				}

				// if tree == sampleTree {
				// 	fmt.Println("probs of sample tree", probs)
				// 	fmt.Println("in ring ", ringIndex)
				// }

				processTree(newlyInfectedChannel, tree, probs)
			}
			for _, losingInfectionTree := range treesLosingInfection {
				tree := coordinatePair{neighbor[0] + losingInfectionTree[0], neighbor[1] + losingInfectionTree[1]}

				if tree[0] < 0 {
					tree[0] = tree[0] + worldWidth
				}

				if tree[0] > worldWidth - 1 {
					tree[0] = tree[0] - worldWidth
				}

				if tree[1] < -0.5 {
					tree[1] = tree[1] + worldWidth
				}

				if tree[1] > worldWidth - 1 {
					tree[1] = tree[1] - worldWidth
				}

				processTree(losingInfectionChannel, tree, probs)
			}
		}

		doneChannel <- true
	}

	for ringIndex := range ringList {
		go processRing(ringIndex)
	}

	ringCount := len(ringList)
	processedRings := 0
	resultsSlice := make([]resultPair, 0)
	
ResultProcessLoop:
	for {
		select {
		case infectedTree := <-newlyInfectedChannel:

			// if infectedTree.Tree == sampleTree {
			// 	fmt.Println("looking at sample tree in infected", infectedTree.Probabilities)
			// }

			found := false

			for i, result := range resultsSlice {
				if result.Tree == infectedTree.Tree {
					result.Probabilities = probabilities{
						infectedTree.Probabilities[0] + result.Probabilities[0],
						infectedTree.Probabilities[1] + result.Probabilities[1],
						infectedTree.Probabilities[2] + result.Probabilities[2],
					}

					resultsSlice[i] = result

					found = true
				}
			}

			if found == false {
				resultsSlice = append(resultsSlice, infectedTree)
			}

		case losingInfectionTree := <-losingInfectionChannel:
			found := false

			// if losingInfectionTree.Tree == sampleTree {
			// 	fmt.Println("looking at sample tree in uninfected", losingInfectionTree.Probabilities)
			// }

			for i, result := range resultsSlice {
				if result.Tree == losingInfectionTree.Tree {
					result.Probabilities = probabilities{
						result.Probabilities[0] - losingInfectionTree.Probabilities[0],
						result.Probabilities[1] - losingInfectionTree.Probabilities[1],
						result.Probabilities[2] - losingInfectionTree.Probabilities[2],
					}

					resultsSlice[i] = result

					found = true
				}
			}

			if found == false {
				var losingInfectionProbs probabilities

				for i, probability := range losingInfectionTree.Probabilities {
					losingInfectionProbs[i] = -probability
				}

				losingInfectionTree.Probabilities = losingInfectionProbs

				resultsSlice = append(resultsSlice, losingInfectionTree)
			}
		case <-doneChannel:
			processedRings++

			if processedRings == ringCount {
				break ResultProcessLoop
			}
		}
	}

	type exportPair struct {
		Tree          []coordinatePair
		Probabilities probabilities
	}

	exportsSlice := make([]exportPair, 0, len(resultsSlice))

	for _, resultPair := range resultsSlice {
		exportsSlice = append(exportsSlice, exportPair{
			Tree:          []coordinatePair{resultPair.Tree},
			Probabilities: resultPair.Probabilities,
		})
	}

	resultsJSON, err := json.Marshal(exportsSlice)
	os.Stdout.WriteString(string(resultsJSON))
}
