package main

import (
	"encoding/json"
	"os"
	"strconv"
	"strings"
	"time"
)

type TestCase struct {
	Customers       *[]Customer
	Ingredients     map[string]struct{}
	IngredientLimit int
}

type Customer struct {
	Likes    map[string]struct{} // Hash Set
	Dislikes map[string]struct{} // Hash Set
}

type Recipe struct {
	Ingredients map[string]struct{}
	Score       int
}

var TestCases = map[int]TestCase{}
var BestRecipes = make(chan Recipe, 10000)

func main() {
	for i := 1; i < 6; i++ {
		B, _ := os.ReadFile("input_data/" + strconv.Itoa(i) + ".txt")
		Case := string(B)
		TestCases[i] = ParseTest(Case)
		// Log(TestCases[i])

		// println("---------------------------------------------")

	}

	go func() {
		maxScore := 0
		for {
			for recipe := range BestRecipes {
				if recipe.Score > maxScore {
					maxScore = recipe.Score
					println("\nNew Best Score: " + strconv.Itoa(maxScore))
					Log(recipe)
				} else {
					print(recipe.Score, " ")
				}
			}
			time.Sleep(time.Second * 4)
		}
	}()

	println("Test case output file A evaluated: " + strconv.Itoa(TestCases[1].EvaluateFile("outputs/A.txt")))
	println("Test case output file B evaluated: " + strconv.Itoa(TestCases[2].EvaluateFile("outputs/B.txt")))
	println("Test case output file C evaluated: " + strconv.Itoa(TestCases[3].EvaluateFile("outputs/C.txt")))
	println("Test case output file D evaluated: " + strconv.Itoa(TestCases[4].EvaluateFile("outputs/D.txt")))
	println("Test case output file E evaluated: " + strconv.Itoa(TestCases[5].EvaluateFile("outputs/E.txt")))

	RA := ParseRecipeFromFile("outputs/C.txt")

	TC3 := TestCases[3]
	// TC3.IngredientLimit = 5
	Log(TC3)

	TC3.Evaluate(&RA)
	TC3.FitAdd(&RA)

	time.Sleep(time.Second * 3)

	// Wait until chaned is closed
	for len(BestRecipes) > 0 {
		time.Sleep(time.Second * 4)
	}

}

func (TC TestCase) FitAdd(recipe *Recipe) *Recipe {
	// oldScore := TC.Evaluate(recipe)
	betterScores := make([]Recipe, 0, len(TC.Ingredients))

	for ingredient := range TC.Ingredients {

		newRecipe := recipe.Clone()
		newRecipe.Ingredients[ingredient] = struct{}{}
		newScore := TC.Evaluate(&newRecipe)
		if newScore >= recipe.Score {
			betterScores = append(betterScores, newRecipe)
		}
	}

	// Log(recipe)
	// Log(betterScores)
	// println("+++++++++++++++++++++++++++++++++++++++++++++")

	if len(betterScores) == 0 {
		print(".", len(recipe.Ingredients), ". ")
		return nil

	} else if len(betterScores) == 1 {
		BestRecipes <- betterScores[0]
		print("*", len(recipe.Ingredients), "* ")
		return &betterScores[0]

	} else {

		if len(recipe.Ingredients) > TC.IngredientLimit {
			betterMax := betterScores[0]
			for _, R := range betterScores {
				if R.Score > betterMax.Score {
					betterMax = R
				}
			}
			BestRecipes <- betterMax
			return &betterMax
		}

		// print("-", len(recipe.Ingredients), "- ")

		recBetter := make([]Recipe, 0, len(betterScores)/2)
		for _, R := range betterScores {
			recResult := TC.FitAdd(&R)
			if recResult != nil {
				recBetter = append(recBetter, *recResult)
			}
		}

		if len(recBetter) == 0 {
			betterMax := betterScores[0]
			for _, R := range betterScores {
				if R.Score > betterMax.Score {
					betterMax = R
				}
			}
			BestRecipes <- betterMax
			return &betterMax

		} else if len(recBetter) == 1 {
			BestRecipes <- recBetter[0]
			return &recBetter[0]

		} else {
			recMax := recBetter[0]
			for _, R := range recBetter {
				if R.Score > recMax.Score {
					recMax = R
				}
			}
			BestRecipes <- recMax
			return &recMax
		}
	}

}

func (TC TestCase) Evaluate(recipe *Recipe) int {
	score := 0
	for _, customer := range *TC.Customers {
		ok := true
		for like := range customer.Likes {
			if _, contains := recipe.Ingredients[like]; !contains {
				ok = false
				break
			}
		}
		if ok {
			for dislike := range customer.Dislikes {
				if _, contains := recipe.Ingredients[dislike]; contains {
					ok = false
					break
				}
			}

			if ok {
				score++
			}
		}
	}
	recipe.Score = score
	return score
}

func (TC TestCase) EvaluateFile(filename string) int {
	B, _ := os.ReadFile(filename)
	S := string(B)
	recipe := ParseRecipe(S)
	return TC.Evaluate(&recipe)
}

func ParseTest(Case string) TestCase {
	customerFeedbacks := strings.Split(Case, "\n")[1:]

	lencustomerFeedbacks := len(customerFeedbacks)

	Customers := (make([]Customer, lencustomerFeedbacks/2))
	Ingredients := make(map[string]struct{})

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
					Ingredients[like] = struct{}{}
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
					Ingredients[dislike] = struct{}{}
				}
			}

			iC++
		}
	}

	tc := TestCase{}
	tc.Customers = &Customers
	tc.Ingredients = Ingredients
	tc.IngredientLimit = len(Ingredients)

	return tc
}

func ParseRecipe(S string) Recipe {
	arr := strings.Split(S, " ")
	if len(arr) == 1 {
		return Recipe{Ingredients: make(map[string]struct{}), Score: -1}
	} else {
		recipe := make(map[string]struct{})
		for _, item := range arr[1:] {
			recipe[item] = struct{}{}
		}
		return Recipe{Ingredients: recipe, Score: -1}
	}
}

func ParseRecipeFromFile(file string) Recipe {
	B, _ := os.ReadFile(file)
	S := string(B)
	return ParseRecipe(S)
}

func (R Recipe) Clone() Recipe {
	recipe := Recipe{Ingredients: make(map[string]struct{}, len(R.Ingredients)), Score: R.Score}
	for ingredient := range R.Ingredients {
		recipe.Ingredients[ingredient] = struct{}{}
	}
	return recipe
}

func Log(O interface{}) {
	// B, ee := json.Marshal(O)
	B, ee := json.MarshalIndent(O, "", "\t")

	if ee == nil {
		println(string(B))
	}
	println("")
}
