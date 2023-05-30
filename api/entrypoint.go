package api

import (
	"net/http"
	"fmt"
	"errors"
	"strings"

	"github.com/gocolly/colly"
	"github.com/gin-gonic/gin"
)

var(
	app *gin.Engine

)

// CREATE ENDPOIND

func myRoute(r *gin.RouterGroup){
	r.GET("/admin",func(c *gin.Context){
		c.String(http.StatusOK,"Hello from golang in vercel")
	})

	r.GET("/name", getName)


}

func init(){
	app = gin.New()
	r := app.Group("/api")
	myRoute(r)

}

// ADD THIS SCRIPT
func Handler(w http.ResponseWriter , r *http.Request){
	app.ServeHTTP(w,r)
}

func loginHandler(username, password string) (*colly.Collector, error) {
	c := colly.NewCollector()
	c.UserAgent = "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_9_4) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/36.0.1985.125 Safari/537.36"

	// Create a login data map to store login data
	loginData := map[string]string{
		"__RequestVerificationToken": "",
		"SCKTY00328510CustomEnabled": "True",
		"SCKTY00436568CustomEnabled": "True",
		"Database":                   "10",
		"VerificationOption":         "UsernamePassword",
		"LogOnDetails.UserName":      "",
		"tempUN":                     "",
		"tempPW":                     "",
		"LogOnDetails.Password":      "",
	}

	loginData["LogOnDetails.UserName"] = username
	loginData["LogOnDetails.Password"] = password

	loginSuccessful := false
	loginAttempted := false

	// Find input with name="__RequestVerificationToken" and get the value
	c.OnHTML("input[name='__RequestVerificationToken']", func(e *colly.HTMLElement) {
		if !loginAttempted {
			loginAttempted = true

			requestVerificationToken := e.Attr("value")

			// Update the request payload with the received token
			loginData["__RequestVerificationToken"] = requestVerificationToken

			// post login data and print response
			err := c.Post("https://homeaccess.katyisd.org/HomeAccess/Account/LogOn", loginData)
			if err != nil {
				fmt.Println(loginData)
			}
		}
	})

	// Attach callback after login request
	c.OnResponse(func(r *colly.Response) {
		// Check response URL to verify successful login
		fmt.Print(r.Request.URL.String())
		if strings.Contains(r.Request.URL.String(), "LogOn") {
			loginSuccessful = false
		} else {
			loginSuccessful = true
		}
	})

	err := c.Visit("https://homeaccess.katyisd.org/HomeAccess/Account/LogOn")
	if err != nil {
		return nil, err
	}

	if loginSuccessful {
		return c, nil
	} else {
		return nil, errors.New("Invalid username or password")
	}
}

func getName(c *gin.Context) {
	// Call the loginHandler to perform the login process
	username := c.Query("user")
	password := c.Query("pass")
	collector, err := loginHandler(username, password)
	if err != nil {
		// Handle the login error
		if err.Error() == "Invalid username or password" {
			c.JSON(401, gin.H{"error": "Invalid username or password"})
		} else {
			c.JSON(500, gin.H{"error": "Failed to log in"})
		}
		return
	}

	// Attach the callback to extract the name after successful login
	collector.OnHTML("div.sg-banner-menu-container", func(e *colly.HTMLElement) {
		// Extract the name from the element and return it as the response
		name := e.ChildText("span")
		c.JSON(200, gin.H{"name": name})
	})

	// Perform the scraping operation
	err = collector.Visit("https://homeaccess.katyisd.org/HomeAccess/Classes/Classwork")
	if err != nil {
		// Handle the scraping error
		c.JSON(500, gin.H{"error": "Failed to scrape data"})
		return
	}
}