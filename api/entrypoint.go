package api

import (
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"github.com/gin-gonic/gin"
	"github.com/iancoleman/orderedmap"
	"net/http"
	"net/http/cookiejar"
	"strings"
	"errors"
)

var (
	app *gin.Engine
)

// CREATE ENDPOIND
func myRoute(r *gin.RouterGroup) {
	r.GET("/admin", func(c *gin.Context) {
		c.String(http.StatusOK, "Hello from golang in vercel")
	})

	r.GET("/name", getName)
	r.GET("/assignments", getAssignments)
	// r.GET("/info", getInfo)
	// r.GET("/averages", getAverages)
	// r.GET("/classes", getClasses)
	// r.GET("/reportcard", getReport)
	// r.GET("/ipr", getProgressReport)

	message := orderedmap.New()
	message.Set("title", "Welcome to the Home Access Center API!")
	message.Set("message", "This is the home page, visit the documentation at https://homeaccesscenterapi-docs.vercel.app/ for more information on how to use this API.")
	message.Set("routes", []string{"/api/name", "/api/assignments", "/api/info", "/api/averages", "/api/classes", "/api/reportcard", "/api/ipr"})

	r.GET("/help", func(c *gin.Context) {
		c.JSON(200, message)
	})

	r.GET("/", func(c *gin.Context) {
		c.JSON(200, message)
	})

}

func init() {
	app = gin.New()
	r := app.Group("/api")
	myRoute(r)

}

// ADD THIS SCRIPT
func Handler(w http.ResponseWriter, r *http.Request) {
	app.ServeHTTP(w, r)
}

func loginHandler(username, password string) (*http.Client, error) {
	jar, err := cookiejar.New(nil)
	if err != nil {
		return nil, err
	}

	client := &http.Client{
		Jar: jar,
	}

	loginURL := "https://homeaccess.katyisd.org/HomeAccess/Account/LogOn"
	loginData := fmt.Sprintf("__RequestVerificationToken=&SCKTY00328510CustomEnabled=True&SCKTY00436568CustomEnabled=True&Database=10&VerificationOption=UsernamePassword&LogOnDetails.UserName=%s&tempUN=&tempPW=&LogOnDetails.Password=%s", username, password)
	req, err := http.NewRequest("POST", loginURL, strings.NewReader(loginData))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		return client, nil
	}

	return nil, errors.New("Invalid username or password")
}

func getName(c *gin.Context) {
	username := c.Query("user")
	password := c.Query("pass")
	client, err := loginHandler(username, password)
	if err != nil {
		if err.Error() == "Invalid username or password" {
			c.JSON(401, gin.H{"error": "Invalid username or password"})
		} else {
			c.JSON(500, gin.H{"error": "Failed to log in"})
		}
		return
	}

	resp, err := client.Get("https://homeaccess.katyisd.org/HomeAccess/Classes/Classwork")
	if err != nil {
		c.JSON(500, gin.H{"error": "Failed to scrape data"})
		return
	}
	defer resp.Body.Close()

	// Parse the HTML document
	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		c.JSON(500, gin.H{"error": "Failed to parse HTML"})
		return
	}

	// Find the student's name
	name := doc.Find("#plnMain_studentNameLabel").Text()

	c.JSON(200, gin.H{"name": name})
}

func getAssignments(c *gin.Context) {
	username := c.Query("user")
	password := c.Query("pass")
	client, err := loginHandler(username, password)
	if err != nil {
		if err.Error() == "Invalid username or password" {
			c.JSON(401, gin.H{"error": "Invalid username or password"})
		} else {
			c.JSON(500, gin.H{"error": "Failed to log in"})
		}
		return
	}

	resp, err := client.Get("https://homeaccess.katyisd.org/HomeAccess/Content/Student/Assignments.aspx")
	if err != nil {
		c.JSON(500, gin.H{"error": "Failed to scrape data"})
		return
	}
	defer resp.Body.Close()

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		c.JSON(500, gin.H{"error": "Failed to parse HTML"})
		return
	}

	// Find the assignment table rows
	var assignments []map[string]string
	doc.Find("#plnMain_rgAssigments_ctl00__0 tbody tr").Each(func(i int, s *goquery.Selection) {
		assignment := make(map[string]string)
		assignment["class"] = s.Find(".AssignmentClass").Text()
		assignment["description"] = s.Find(".AssignmentDescription").Text()
		assignment["dueDate"] = s.Find(".AssignmentDueDate").Text()
		assignment["score"] = s.Find(".AssignmentScore").Text()
		assignment["category"] = s.Find(".AssignmentCategory").Text()
		assignments = append(assignments, assignment)
	})

	c.JSON(200, gin.H{"assignments": assignments})
}

// Rest of the code...
