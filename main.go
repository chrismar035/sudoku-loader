package main

import (
	"bufio"
	"bytes"
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/chrismar035/sudoku-solver"

	"gopkg.in/redis.v3"
)

type Sudoku struct {
	Id       string  `json:"id"`
	Puzzle   [81]int `json:"puzzle"`
	Solution [81]int `json:"solution"`
	Name     string  `json:"name"`
}

func main() {
	client := redis.NewClient(&redis.Options{
		Addr:     os.Getenv("REDIS_ADDR"),
		Password: "", // no password set
		DB:       0,  // use default DB
	})

	_, err := client.Ping().Result()
	if err != nil {
		panic(err)
	}

	key := "puzzles"

	file, err := os.Open("puzzles.txt")
	if err != nil {
		fmt.Println("Error opening puzzles", err)
	}
	scanner := bufio.NewScanner(file)
	lineNumber := 0
	for scanner.Scan() {
		lineNumber++
		line := scanner.Text()
		if len(line) != 81 {
			fmt.Println("Puzzle improperly sized on line", line, err)
		}

		var source bytes.Buffer
		var grid solver.Grid
		for i, char := range strings.Split(line, "") {
			source.WriteString(char)
			grid[i], err = strconv.Atoi(char)
			if err != nil {
				fmt.Println("Invalid puzzle line", lineNumber, "char", i)
			}
		}

		puzzle := solver.Puzzle{Initial: grid}
		solver := solver.NewSolver()
		puzzle.Solution = solver.Solve(grid)

		for _, number := range puzzle.Solution {
			source.WriteByte(byte(number))
		}
		hasher := sha1.New()
		hasher.Write(source.Bytes())
		sha := hex.EncodeToString(hasher.Sum(nil))

		sudoku := Sudoku{Id: sha, Puzzle: puzzle.Initial, Solution: puzzle.Solution}

		var contentBuf bytes.Buffer
		if err = json.NewEncoder(&contentBuf).Encode(sudoku); err != nil {
			panic(err)
		}
		client.SAdd(key, contentBuf.String())
		fmt.Println("Added", key, sha, contentBuf.String())
	}
}
