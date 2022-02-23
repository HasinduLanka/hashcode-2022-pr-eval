package main

import (
	"encoding/json"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"
)

type TestCase struct {
	Customers       *[]Customer
	Ingredients     map[string]struct{}
	IngredientsArr  []string
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

var FitterIndex = 0 // Used to terminate older fit goroutines
const MaxFitters = 4

func main() {
	tcLetter := "E"
	tcIndex := 5

	if len(os.Args) == 3 {
		Log(os.Args)
		tcLetter = os.Args[1]
		tcIndexArg, err := strconv.Atoi(os.Args[2])

		if err != nil {
			panic(err)
		}

		tcIndex = tcIndexArg
	} else if len(os.Args) != 1 {
		println("Usage: go run . [Test Case Letter] [Test Case Index]")
		println("Example: go run . D 4")
		return
	}

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

	RA := ParseRecipeFromFile("outputs/" + tcLetter + ".txt")

	TCC := TestCases[tcIndex]
	// Log(TC3)

	TCC.Evaluate(&RA)

	go func() {
		maxScore := 0
		var bestRecipe *Recipe = nil

		go func() {
			runningMaxScore := 0
			for {
				if bestRecipe != nil {
					if runningMaxScore < bestRecipe.Score {
						runningMaxScore = bestRecipe.Score
						println("\nRunning Max Score: " + strconv.Itoa(runningMaxScore))

						bestRecipeClone := bestRecipe.Clone()
						go TCC.FitAdd(&bestRecipeClone)
					}
				}
				time.Sleep(time.Second * 10)
			}
		}()

		for {
			for recipe := range BestRecipes {
				if recipe.Score > maxScore {
					maxScore = recipe.Score
					println("New Best Score: " + strconv.Itoa(maxScore))

					recipeClone := recipe.Clone()
					bestRecipe = &recipeClone
					recipe.Save("outputs/" + tcLetter + strconv.Itoa(maxScore) + ".txt")

				}
			}
			time.Sleep(time.Second * 4)
		}
	}()

	// Start the first fit
	TCC.FitAdd(&RA)

	// Wait until sun burns out
	for {
		time.Sleep(time.Second * 60)
	}

}

func (TC *TestCase) FitAdd(origRecipe *Recipe) *Recipe {

	maxFitterIndex := FitterIndex + MaxFitters
	FitterIndex++

	candidateIngs := TC.CloneIngredients()
	for ingredient := range origRecipe.Ingredients {
		delete(candidateIngs, ingredient)
	}

	ingredientsArr := MakeIngredientsArr(candidateIngs)

	// counter := 0

	addFn := func(indexArr []int) bool {
		// LogLine(indexArr)
		recipe := origRecipe.Clone()

		for _, ing := range indexArr {
			recipe.Ingredients[ingredientsArr[ing]] = struct{}{}
		}

		TC.Evaluate(&recipe)
		// Log(recipe)

		BestRecipes <- recipe

		// Return true to terminate the index maker
		return maxFitterIndex < FitterIndex

		// counter++
	}

	// go func() {
	// 	for {
	// 		println(counter)
	// 		// LogLine(recipe)
	// 		time.Sleep(time.Second)
	// 	}
	// }()

	ingCount := 2

	for ingCount = 2; ingCount < 4; ingCount++ {
		go IndexMaker(make([]int, ingCount), 0, 0, len(ingredientsArr), addFn)

	}

	IndexMaker(make([]int, ingCount), 0, 0, len(ingredientsArr), addFn)

	return nil
}

func IndexMaker(arr []int, index int, begin int, length int, fn func([]int) bool) bool {

	newIndex := index + 1

	edge := (newIndex == len(arr))

	for i := begin; i < length; i++ {
		arr[index] = i

		if edge {
			if fn(arr) {
				println("Terminating ", index, begin, length)
				return true
			}

		} else {
			if IndexMaker(arr, newIndex, i+1, length, fn) {
				return true
			}
		}

	}
	return false
}

func (TC *TestCase) Evaluate(recipe *Recipe) int {
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

func (TC *TestCase) CloneIngredients() map[string]struct{} {
	Ingredients := make(map[string]struct{}, len(TC.Ingredients))
	for ingredient := range TC.Ingredients {
		Ingredients[ingredient] = struct{}{}
	}
	return Ingredients
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
			likes := strings.Split(strings.TrimSpace(customerFeedbacks[i]), " ")
			if len(likes) > 1 {
				// Customers[iC].Likes = likes[1:]
				Customers[iC].Likes = make(map[string]struct{})
				for _, like := range likes[1:] {
					Customers[iC].Likes[like] = struct{}{}
					Ingredients[like] = struct{}{}
				}

				delete(Customers[iC].Likes, "")
			}

		} else {
			// Dislikes
			dislikes := strings.Split(strings.TrimSpace(customerFeedbacks[i]), " ")
			if len(dislikes) > 1 {
				// Customers[iC].Dislikes = dislikes[1:]
				Customers[iC].Dislikes = make(map[string]struct{})
				for _, dislike := range dislikes[1:] {
					Customers[iC].Dislikes[dislike] = struct{}{}
					Ingredients[dislike] = struct{}{}
				}
			}

			delete(Customers[iC].Dislikes, "")

			iC++
		}
	}

	delete(Ingredients, "")

	tc := TestCase{}
	tc.Customers = &Customers
	tc.Ingredients = Ingredients
	tc.IngredientLimit = len(Ingredients)
	tc.IngredientsArr = MakeIngredientsArr(Ingredients)

	return tc
}

func MakeIngredientsArr(Ingredients map[string]struct{}) []string {
	IngredientsArr := make([]string, 0, len(Ingredients))

	for ingredient := range Ingredients {
		IngredientsArr = append(IngredientsArr, ingredient)
	}

	sort.Slice(IngredientsArr, func(i, j int) bool {
		return IngredientsArr[i] < IngredientsArr[j]
	})

	return IngredientsArr
}

func ParseRecipe(S string) Recipe {
	arr := strings.Split(strings.TrimSpace(S), " ")
	if len(arr) == 1 {
		return Recipe{Ingredients: make(map[string]struct{}), Score: -1}
	} else {
		recipe := make(map[string]struct{})
		for _, item := range arr[1:] {
			recipe[item] = struct{}{}
		}
		delete(recipe, "")
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

func (R Recipe) Save(filename string) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	file.WriteString(strconv.Itoa(len(R.Ingredients)) + " ")

	for ingredient := range R.Ingredients {
		_, err := file.WriteString(ingredient + " ")
		if err != nil {
			return err
		}
	}

	return nil
}

func Log(O interface{}) {
	// B, ee := json.Marshal(O)
	B, ee := json.MarshalIndent(O, "", "\t")

	if ee == nil {
		println(string(B))
	}
	println("")
}
func LogLine(O interface{}) {
	B, ee := json.Marshal(O)
	// B, ee := json.MarshalIndent(O, "", "\t")

	if ee == nil {
		println(string(B))
	}
	println("")
}

// func (TC TestCase) FitAdd(recipe *Recipe) *Recipe {
// 	// oldScore := TC.Evaluate(recipe)
// 	betterScores := make([]Recipe, 0, len(TC.Ingredients))

// 	for ingredient := range TC.Ingredients {

// 		newRecipe := recipe.Clone()
// 		newRecipe.Ingredients[ingredient] = struct{}{}
// 		newScore := TC.Evaluate(&newRecipe)
// 		if newScore >= recipe.Score {
// 			betterScores = append(betterScores, newRecipe)
// 		}
// 	}

// 	// Log(recipe)
// 	// Log(betterScores)
// 	// println("+++++++++++++++++++++++++++++++++++++++++++++")

// 	if len(betterScores) == 0 {
// 		print(".", len(recipe.Ingredients), ". ")
// 		return nil

// 	} else if len(betterScores) == 1 {
// 		BestRecipes <- betterScores[0]
// 		print("*", len(recipe.Ingredients), "* ")
// 		return &betterScores[0]

// 	} else {

// 		if len(recipe.Ingredients) > TC.IngredientLimit {
// 			betterMax := betterScores[0]
// 			for _, R := range betterScores {
// 				if R.Score > betterMax.Score {
// 					betterMax = R
// 				}
// 			}
// 			BestRecipes <- betterMax
// 			return &betterMax
// 		}

// 		// print("-", len(recipe.Ingredients), "- ")

// 		recBetter := make([]Recipe, 0, len(betterScores)/2)
// 		for _, R := range betterScores {
// 			recResult := TC.FitAdd(&R)
// 			if recResult != nil {
// 				recBetter = append(recBetter, *recResult)
// 			}
// 		}

// 		if len(recBetter) == 0 {
// 			betterMax := betterScores[0]
// 			for _, R := range betterScores {
// 				if R.Score > betterMax.Score {
// 					betterMax = R
// 				}
// 			}
// 			BestRecipes <- betterMax
// 			return &betterMax

// 		} else if len(recBetter) == 1 {
// 			BestRecipes <- recBetter[0]
// 			return &recBetter[0]

// 		} else {
// 			recMax := recBetter[0]
// 			for _, R := range recBetter {
// 				if R.Score > recMax.Score {
// 					recMax = R
// 				}
// 			}
// 			BestRecipes <- recMax
// 			return &recMax
// 		}
// 	}

// }
