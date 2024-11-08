package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"

	"github.com/gin-gonic/gin"
)

type User struct {
	Username  string     `json:"USERNAME"`
	Password  string     `json:"PASSWORD"`
	Questions []Question `json:"QUESTIONS"` // Store questions and answers directly in the user struct
}

type Question struct {
	Question string `json:"Q"`
	Answer   string `json:"A"`
}

type RunPythonRequest struct {
	Filename string `json:"FILENAME"`
}

// Store users in a slice instead of a map
var users = []User{}

// Signup route
func signup(c *gin.Context) {
	var newuser User
	if c.BindJSON(&newuser) == nil {
		// Check if the username already exists
		for _, user := range users {
			if user.Username == newuser.Username {
				c.JSON(http.StatusConflict, gin.H{"ERROR": "Username already exists, Try Again"})
				return
			}
		}
		// Append the new user to the users slice
		users = append(users, newuser)
		c.JSON(http.StatusCreated, gin.H{"RECEIVED": "User registered successfully"})
	} else {
		c.JSON(http.StatusBadRequest, gin.H{"ERROR": "Invalid input"})
	}
}

// Signin route
func signin(c *gin.Context) {
	var loginDetails User
	if c.BindJSON(&loginDetails) == nil {
		for _, user := range users {
			if user.Username == loginDetails.Username {
				if user.Password == loginDetails.Password {
					c.JSON(http.StatusOK, gin.H{"RECEIVED": "You are logged in successfully"})
				} else {
					c.JSON(http.StatusForbidden, gin.H{"ERROR": "Incorrect password, Try Again"})
				}
				return
			}
		}
		c.JSON(http.StatusNotFound, gin.H{"ERROR": "Invalid Username"})
	} else {
		c.JSON(http.StatusBadRequest, gin.H{"ERROR": "Invalid input"})
	}
}

// AskQuestion handles asking a question and retrieving an answer
func askQuestion(c *gin.Context) {
	var req struct {
		Username string `json:"USERNAME"`
		Question string `json:"QUESTION"`
	}

	// Bind the incoming JSON data
	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"ERROR": "Invalid input"})
		return
	}

	// Search for the user and their question
	for _, user := range users {
		if user.Username == req.Username {
			// If user found, loop through their questions
			for _, q := range user.Questions {
				if q.Question == req.Question {
					c.JSON(http.StatusOK, gin.H{"ANSWER": q.Answer})
					return
				}
			}
			// If the question isn't found for the user
			c.JSON(http.StatusNotFound, gin.H{"ERROR": "Question not found"})
			return
		}
	}
	// If user doesn't exist
	c.JSON(http.StatusNotFound, gin.H{"ERROR": "User not found"})
}

// AddQuestion allows a user to add a question and its answer
func addQuestion(c *gin.Context) {
	var req struct {
		Username string `json:"USERNAME"`
		Question string `json:"Q"`
		Answer   string `json:"A"`
	}

	// Bind the incoming JSON data
	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"ERROR": "Invalid input"})
		return
	}

	// Search for the user and add the question
	for i, user := range users {
		if user.Username == req.Username {
			// If user exists, append the new question and answer
			users[i].Questions = append(users[i].Questions, Question{
				Question: req.Question,
				Answer:   req.Answer,
			})
			c.JSON(http.StatusOK, gin.H{"RECEIVED": "Question and answer added successfully"})
			return
		}
	}
	// If user doesn't exist
	c.JSON(http.StatusNotFound, gin.H{"ERROR": "User not found"})
}

// RunPython handles execution of Python code from a specified file
func runPython(c *gin.Context) {
	var req RunPythonRequest

	// Bind JSON to get the filename from the request body
	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"ERROR": "Invalid input, please provide a filename"})
		return
	}

	// Check if the file exists in the current directory
	if _, err := os.Stat(req.Filename); os.IsNotExist(err) {
		c.JSON(http.StatusNotFound, gin.H{"ERROR": fmt.Sprintf("File '%s' not found", req.Filename)})
		return
	}

	// Read the content of the specified file
	code, err := ioutil.ReadFile(req.Filename)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"ERROR": "Failed to read file"})
		return
	}

	// Save Python code to a temporary file
	tmpfile, err := ioutil.TempFile("", "script-*.py")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"ERROR": "Failed to create temp file"})
		return
	}
	defer os.Remove(tmpfile.Name()) // Delete the temp file after execution
	defer tmpfile.Close()

	// Write the code to the temporary file
	if _, err := tmpfile.Write(code); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"ERROR": "Failed to write to temp file"})
		return
	}

	// Execute the Python script and capture output
	cmd := exec.Command("python", tmpfile.Name())
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out
	err = cmd.Run()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"ERROR": "Failed to execute Python code", "DETAILS": out.String()})
		return
	}

	// Return the actual output of the executed Python script
	c.JSON(http.StatusOK, gin.H{"OUTPUT": out.String()})
}


func main() {
	router := gin.Default()
	router.POST("/signup", signup)
	router.POST("/signin", signin)
	router.POST("/askquestion", askQuestion) // For asking questions
	router.POST("/addquestion", addQuestion) // For adding questions and answers
	router.POST("/runpython", runPython)
	router.Run(":8080")
}
