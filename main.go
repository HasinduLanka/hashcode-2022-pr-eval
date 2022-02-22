package main

import (
	"encoding/json"
	"os"
	"strconv"
	"strings"
)

type TestCase struct {
	Customers *[]Customer
}

type Customer struct {
	Likes    map[string]struct{} // Hash Set
	Dislikes map[string]struct{} // Hash Set
}

var TestCases = map[int]TestCase{}

func main() {
	for i := 1; i < 6; i++ {
		B, _ := os.ReadFile("input_data/" + strconv.Itoa(i) + ".txt")
		Case := string(B)
		TestCases[i] = ParseTest(Case)
		// Log(TestCases[i])

		// println("---------------------------------------------")

	}

	println("Test case output file A evaluated: " + strconv.Itoa(TestCases[1].EvaluateFile("outputs/A.txt")))
	println("Test case output file B evaluated: " + strconv.Itoa(TestCases[2].EvaluateFile("outputs/B.txt")))
	println("Test case output file C evaluated: " + strconv.Itoa(TestCases[3].EvaluateFile("outputs/C.txt")))
	println("Test case output file D evaluated: " + strconv.Itoa(TestCases[4].EvaluateFile("outputs/D.txt")))
	println("Test case output file E evaluated: " + strconv.Itoa(TestCases[5].EvaluateFile("outputs/E.txt")))

}

func (TC TestCase) Evaluate(Recipe map[string]struct{}) int {
	score := 0
	for _, customer := range *TC.Customers {
		ok := true
		for like := range customer.Likes {
			if _, contains := Recipe[like]; !contains {
				ok = false
				break
			}
		}
		if ok {
			for dislike := range customer.Dislikes {
				if _, contains := Recipe[dislike]; contains {
					ok = false
					break
				}
			}

			if ok {
				score++
			}
		}
	}

	return score
}

func (TC TestCase) EvaluateFile(filename string) int {
	B, _ := os.ReadFile(filename)
	S := string(B)
	return TC.Evaluate(ParseRecipe(S))
}

func ParseTest(Case string) TestCase {
	customerFeedbacks := strings.Split(Case, "\n")[1:]

	lencustomerFeedbacks := len(customerFeedbacks)

	Customers := (make([]Customer, lencustomerFeedbacks/2))

	iC := 0
	for i := 0; i < lencustomerFeedbacks; i += 1 {
		if i%2 == 0 {
			// Likes
			likes := strings.Split(customerFeedbacks[i], " ")
			if len(likes) > 1 {
				// Customers[iC].Likes = likes[1:]
				Customers[iC].Likes = make(map[string]struct{})
				for _, like := range likes[1:] {
					Customers[iC].Likes[like] = struct{}{}
				}
			}

		} else {
			// Dislikes
			dislikes := strings.Split(customerFeedbacks[i], " ")
			if len(dislikes) > 1 {
				// Customers[iC].Dislikes = dislikes[1:]
				Customers[iC].Dislikes = make(map[string]struct{})
				for _, dislike := range dislikes[1:] {
					Customers[iC].Dislikes[dislike] = struct{}{}
				}
			}

			iC++
		}
	}

	tc := TestCase{}
	tc.Customers = &Customers

	return tc
}

func ParseRecipe(S string) map[string]struct{} {
	arr := strings.Split(S, " ")
	if len(arr) == 1 {
		return map[string]struct{}{}
	} else {
		Recipe := make(map[string]struct{})
		for _, item := range arr[1:] {
			Recipe[item] = struct{}{}
		}
		return Recipe
	}
}

func Log(O interface{}) {
	B, ee := json.MarshalIndent(O, "", "\t")

	if ee == nil {
		println(string(B))
	}
}
