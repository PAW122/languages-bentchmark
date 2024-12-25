package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

// A structure describing one task
type Task struct {
	TaskName string  `json:"taskName"`
	MatrixA  [][]int `json:"matrixA"`
	MatrixB  [][]int `json:"matrixB"`
}

// Structure to load/save tasks.json
type TasksFile struct {
	Tasks []Task `json:"tasks"`
}

// Test result recording structure (for the tester)
type TestResult struct {
	TaskName     string
	ElapsedMs    int64
	ServerStatus string
}

// -----------------------------------------
// Helper functions
// -----------------------------------------

// Generates a random matrix of rows x cols (values ​​0..9)
func generateRandomMatrix(rows, cols int) [][]int {
	m := make([][]int, rows)
	for i := 0; i < rows; i++ {
		rowData := make([]int, cols)
		for j := 0; j < cols; j++ {
			rowData[j] = rand.Intn(10)
		}
		m[i] = rowData
	}
	return m
}

// Convert JSON matrix ([][]interface{}) -> [][]int
func convertMatrix(matrixData []interface{}) [][]int {
	var result [][]int

	for _, row := range matrixData {
		rowSlice, ok := row.([]interface{})
		if !ok {
			continue
		}
		intRow := make([]int, len(rowSlice))
		for i, val := range rowSlice {
			floatVal, ok := val.(float64) // JSON unmarshal converts numbers to float64
			if !ok {
				intRow[i] = 0
			} else {
				intRow[i] = int(floatVal)
			}
		}
		result = append(result, intRow)
	}

	return result
}

// Simple verification of matrix multiplication results
func verifyMatrixMultiplication(A, B, result [][]int) bool {
	n := len(A)
	if n == 0 {
		return false
	}
	m := len(A[0]) // number of columns in A
	p := len(B[0]) // number of columns in B

	// We check if result has dimensions n x p
	if len(result) != n || len(result[0]) != p {
		return false
	}

	// Checking the numbers
	for i := 0; i < n; i++ {
		for j := 0; j < p; j++ {
			expectedVal := 0
			for k := 0; k < m; k++ {
				expectedVal += A[i][k] * B[k][j]
			}
			if result[i][j] != expectedVal {
				return false
			}
		}
	}
	return true
}

// -----------------------------------------
// Main function
// -----------------------------------------

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run main.go <komenda> [argumenty...]")
		os.Exit(1)
	}

	// Initialization of the pseudorandom number generator
	rand.Seed(time.Now().UnixNano())

	command := os.Args[1]

	switch command {
	case "run_tests":
		runTests()
	case "test_server_output":
		if len(os.Args) < 3 {
			fmt.Println("Usage: go run main.go test_server_output <sciezka_do_pliku>")
			os.Exit(1)
		}
		testServerOutput(os.Args[2])
	case "add_task":
		// Example: go run main.go add_task matrix_multiplication 100x100
		if len(os.Args) < 4 {
			fmt.Println("Usage: go run main.go add_task <taskName> <sizeNxM>")
			os.Exit(1)
		}
		addTask(os.Args[2], os.Args[3])
	default:
		fmt.Printf("Unknown command: %s\n", command)
	}
}

// -----------------------------------------
// Command implementations
// -----------------------------------------

// 1) run_tests
func runTests() {
	// 1) We load the tasks.json file
	tasksFileData, err := os.ReadFile("tasks.json")
	if err != nil {
		log.Fatalf("Unable to read tasks.json: %v", err)
	}

	var tf TasksFile
	if err := json.Unmarshal(tasksFileData, &tf); err != nil {
		log.Fatalf("Error parsing tasks.json: %v", err)
	}

	// 2) For each task – send POST to the server and measure the time
	results := []TestResult{}
	for _, task := range tf.Tasks {
		payloadBytes, err := json.Marshal(task)
		if err != nil {
			log.Println("Task serialization error:", err)
			continue
		}

		start := time.Now()
		resp, err := http.Post("http://localhost:3000", "application/json", bytes.NewBuffer(payloadBytes))
		if err != nil {
			log.Println("Error sending request:", err)
			continue
		}
		defer resp.Body.Close()

		elapsed := time.Since(start).Milliseconds()
		results = append(results, TestResult{
			TaskName:     task.TaskName,
			ElapsedMs:    elapsed,
			ServerStatus: resp.Status,
		})

		// bodyBytes, _ := ioutil.ReadAll(resp.Body)
		// fmt.Println("Server response:", string(bodyBytes))
	}

	// 3) Display / Save test results (tester)
	for _, r := range results {
		fmt.Printf("Task: %s, Time: %d ms, Status: %s\n", r.TaskName, r.ElapsedMs, r.ServerStatus)
	}
}

// 2) test_server_output
func testServerOutput(filePath string) {
	// Read the results file (written by the Node.js server)
	content, err := os.ReadFile(filePath)
	if err != nil {
		log.Fatalf("Error reading results file: %v", err)
	}

	var results []map[string]interface{}
	if err := json.Unmarshal(content, &results); err != nil {
		log.Fatalf("Error parsing results file: %v", err)
	}

	// Example verification of each entry
	for i, entry := range results {
		taskName, _ := entry["taskName"].(string)
		matrixA, _ := entry["matrixA"].([]interface{})
		matrixB, _ := entry["matrixB"].([]interface{})
		matrixR, _ := entry["result"].([]interface{})

		convA := convertMatrix(matrixA)
		convB := convertMatrix(matrixB)
		convR := convertMatrix(matrixR)

		if taskName == "matrix_multiplication" {
			ok := verifyMatrixMultiplication(convA, convB, convR)
			if ok {
				fmt.Printf("Entry #%d: The result of matrix multiplication is correct.\n", i)
			} else {
				fmt.Printf("Entry #%d: ERROR - the result of multiplication is incorrect.\n", i)
			}
		} else {
			fmt.Printf("Entry #%d: Unknown task type (%s)\n", i, taskName)
		}
	}
}

// 3) add_task
func addTask(taskName, sizeArg string) {
	// separate e.g. "100x100" -> ["100", "100"]
	parts := strings.Split(sizeArg, "x")
	if len(parts) != 2 {
		fmt.Println("Incorrect size format. Use e.g. 100x100.")
		os.Exit(1)
	}

	rowsA, err := strconv.Atoi(parts[0])
	if err != nil {
		fmt.Printf("Error parsing size (rows): %v\n", err)
		os.Exit(1)
	}
	colsA, err := strconv.Atoi(parts[1])
	if err != nil {
		fmt.Printf("Error parsing size (column): %v\n", err)
		os.Exit(1)
	}

	// If this is a 'matrix_multiplication' task, let's create NxN matrices A and B
	// (assuming we want to multiply 100x100 by 100x100)
	if taskName == "matrix_multiplication" {
		// A = rowsA x colsA
		// B = colsA x colsB (for classic multiplication).
		// To make the result 100x100, let's make B = colsA x rowsA.
		// Or we simplify and give A and B the same NxN:
		matrixA := generateRandomMatrix(rowsA, colsA)
		matrixB := generateRandomMatrix(colsA, rowsA) // -> wynik: (rowsA x rowsA)
		// You can also give B = rowsA x colsA if you want NxN and NxN -> NxN.
		// Then B = rowsA x colsA -> the result will be rowsA x colsA.

		// read the tasks.json file (if it doesn't exist, create an empty structure)
		tf := loadTasksFile("tasks.json")

		// Add new task
		newTask := Task{
			TaskName: "matrix_multiplication",
			MatrixA:  matrixA,
			MatrixB:  matrixB,
		}
		tf.Tasks = append(tf.Tasks, newTask)

		// Save result
		saveTasksFile("tasks.json", tf)
		fmt.Printf("Added task '%s' with matrices of dimensions [%dx%d] i [%dx%d]\n",
			taskName, rowsA, colsA, colsA, rowsA)
	} else {
		// Support for other tasks in the future
		fmt.Printf("Unknown task: %s\n", taskName)
	}
}

// -----------------------------------------
// Helper functions for reading/writing tasks.json
// -----------------------------------------

func loadTasksFile(path string) TasksFile {
	var tf TasksFile

	if _, err := os.Stat(path); os.IsNotExist(err) {
		// The file does not exist, we return an empty structure
		return tf
	}

	data, err := os.ReadFile(path)
	if err != nil {
		log.Fatalf("Cannot read file %s: %v", path, err)
	}
	if len(data) == 0 {
		// Empty file
		return tf
	}

	if err := json.Unmarshal(data, &tf); err != nil {
		log.Fatalf("Error parsing file %s: %v", path, err)
	}
	return tf
}

func saveTasksFile(path string, tf TasksFile) {
	data, err := json.MarshalIndent(tf, "", "  ")
	if err != nil {
		log.Fatalf("Error serializing tasks.json: %v", err)
	}
	if err := os.WriteFile(path, data, 0644); err != nil {
		log.Fatalf("Error writing to tasks.json file: %v", err)
	}
}
