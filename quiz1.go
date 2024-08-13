package main

import (
	"bufio"
	"encoding/csv"
	"flag"
	"fmt"
	"io"
	"log"
	"math/rand"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

var (
	flagFilePath string
	flagDuration int
	flagRandom   bool
	wg           sync.WaitGroup
)

func init() {
	flag.StringVar(&flagFilePath, "FilePath", "problems.csv", "The file path of the problems")
	flag.IntVar(&flagDuration, "Timeout", 10, "Duration of quiz")
	flag.BoolVar(&flagRandom, "IsRandom", true, "Questions should be random or not")
	flag.Parse()
}

func main() {

	csvPath, err := filepath.Abs(flagFilePath)
	if err != nil {
		log.Fatalln("Unable to parse the given file path")
	}

	file, err := os.Open(csvPath)
	if err != nil {
		log.Fatalln("Unable to open the file")
	}

	csvReader := csv.NewReader(file)
	data, err := csvReader.ReadAll()
	if err != nil {
		log.Fatalln("Unable to read the csv data")
	}

	totalquestions := len(data)
	questions := make(map[int]string, totalquestions)
	answers := make(map[int]string, totalquestions)
	responses := make(map[int]string, totalquestions)

	for i, data := range data {
		questions[i] = data[0]
		answers[i] = data[1]
	}

	respondTo := make(chan string)

	fmt.Println("Please press enter to start the quiz")
	bufio.NewScanner(os.Stdout).Scan()

	var rnd *rand.Rand
	if flagRandom {
		source := rand.NewSource(time.Now().UTC().UnixNano())
		rnd = rand.New(source)

	}

	rnd.Shuffle(totalquestions, func(i, j int) {
		questions[i], questions[j] = questions[j], questions[i]
		answers[i], answers[j] = answers[j], answers[i]
	})


	timeUp := time.After(time.Duration(flagDuration) * time.Second)
	wg.Add(1)

	go func() {
	label:
		for i := 0; i < totalquestions; i++ {
			//index := randomPoolofQuestions[i]
			go SendQuestion(os.Stdout, os.Stdin, questions[i], respondTo)
			select {
			case <-timeUp:
				fmt.Fprintln(os.Stderr, "\n Time's Up")
				break label
			case ans, ok := <-respondTo:
				if ok {
					responses[i] = ans
				} else {
					break label
				}
			}
		}
		wg.Done()
	}()
	wg.Wait()

	correct := 0
	for i := 0; i < totalquestions; i++ {
		if CheckAnswers(answers[i], responses[i]) {
			correct++
		}
	}
	summary(correct, totalquestions)

}

func SendQuestion(w io.Writer, r io.Reader, question string, response chan string) {
	reader := bufio.NewReader(r)
	fmt.Fprintln(w, "Question : ", question)
	fmt.Fprint(w, "Answer : ")
	answer, err := reader.ReadString('\n')
	if err != nil {
		close(response)
		if err == io.EOF {
			return
		}
		log.Fatalln(err)
	}
	response <- answer
}

func CheckAnswers(actualanswer string, givenanswer string) bool {
	return strings.EqualFold(strings.TrimSpace(actualanswer), strings.TrimSpace(givenanswer))
}

func summary(correctanswers int, totalquestions int) {
	fmt.Fprintf(os.Stdout, "Score (%d/%d) ", correctanswers, totalquestions)
}
