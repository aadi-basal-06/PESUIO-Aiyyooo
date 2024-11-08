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
	Questions []Question `json:"QUESTIONS"`
}

type Question struct {
	Question string `json:"Q"`
	Answer   string `json:"A"`
}

type RunPythonRequest struct {
	Filename string `json:"FILENAME"`
}

var users = []User{}

func signup(c *gin.Context) {
	var newuser User
	if c.BindJSON(&newuser) == nil {
		for _, user := range users {
			if user.Username == newuser.Username {
				c.JSON(http.StatusConflict, gin.H{"ERROR": "Username already exists, Try Again"})
				return
			}
		}
		users = append(users, newuser)
		c.JSON(http.StatusCreated, gin.H{"RECEIVED": "User registered successfully"})
	} else {
		c.JSON(http.StatusBadRequest, gin.H{"ERROR": "Invalid input"})
	}
}

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

func askQuestion(c *gin.Context) {
	var req struct {
		Username string `json:"USERNAME"`
		Question string `json:"QUESTION"`
	}

	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"ERROR": "Invalid input"})
		return
	}

	for _, user := range users {
		if user.Username == req.Username {
			for _, q := range user.Questions {
				if q.Question == req.Question {
					c.JSON(http.StatusOK, gin.H{"ANSWER": q.Answer})
					return
				}
			}
			c.JSON(http.StatusNotFound, gin.H{"ERROR": "Question not found"})
			return
		}
	}
	c.JSON(http.StatusNotFound, gin.H{"ERROR": "User not found"})
}

func addQuestion(c *gin.Context) {
	var req struct {
		Username string `json:"USERNAME"`
		Question string `json:"Q"`
		Answer   string `json:"A"`
	}

	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"ERROR": "Invalid input"})
		return
	}

	for i, user := range users {
		if user.Username == req.Username {
			users[i].Questions = append(users[i].Questions, Question{
				Question: req.Question,
				Answer:   req.Answer,
			})
			c.JSON(http.StatusOK, gin.H{"RECEIVED": "Question and answer added successfully"})
			return
		}
	}
	c.JSON(http.StatusNotFound, gin.H{"ERROR": "User not found"})
}

func runPython(c *gin.Context) {
	var req RunPythonRequest

	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"ERROR": "Invalid input, please provide a filename"})
		return
	}

	if _, err := os.Stat(req.Filename); os.IsNotExist(err) {
		c.JSON(http.StatusNotFound, gin.H{"ERROR": fmt.Sprintf("File '%s' not found", req.Filename)})
		return
	}

	code, err := ioutil.ReadFile(req.Filename)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"ERROR": "Failed to read file"})
		return
	}

	tmpfile, err := ioutil.TempFile("", "temp-*.py")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"ERROR": "Failed to create temp file"})
		return
	}
	defer os.Remove(tmpfile.Name())
	defer tmpfile.Close()

	if _, err := tmpfile.Write(code); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"ERROR": "Failed to write to temp file"})
		return
	}

	cmd := exec.Command("python", tmpfile.Name())
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out
	err = cmd.Run()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"ERROR": "Failed to execute Python code", "DETAILS": out.String()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"OUTPUT": out.String()})
}


func main() {
	router := gin.Default()
	router.POST("/signup", signup)
	router.POST("/signin", signin)
	router.POST("/askquestion", askQuestion) 
	router.POST("/addquestion", addQuestion) 
	router.POST("/runpython", runPython)
	router.Run(":8080")
}
