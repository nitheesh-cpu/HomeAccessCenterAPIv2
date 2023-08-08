package api

import (
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/gin-gonic/gin"
	"github.com/gocolly/colly"
	"github.com/iancoleman/orderedmap"
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
	r.GET("/info", getInfo)
	r.GET("/averages", getAverages)
	r.GET("/classes", getClasses)
	r.GET("/reportcard", getReport)
	r.GET("/ipr", getProgressReport)
	r.GET("/transcript", getTranscript)
	r.GET("/rank", getRank)

	message := orderedmap.New()
	message.Set("title", "Welcome to the Home Access Center API!")
	message.Set("message", "This is the home page, visit the documentation at https://homeaccesscenterapi-docs.vercel.app/ for more information on how to use this API.")
	message.Set("routes", []string{"/api/name", "/api/assignments", "/api/info", "/api/averages", "/api/classes", "/api/reportcard", "/api/ipr", "api/transcript", "api/rank"})

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

	message := orderedmap.New()
	message.Set("title", "Welcome to the Home Access Center API!")
	message.Set("message", "This is the home page, visit the documentation at https://homeaccesscenterapi-docs.vercel.app/ for more information on how to use this API.")
	message.Set("routes", []string{"/api/name", "/api/assignments", "/api/info", "/api/averages", "/api/classes", "/api/reportcard", "/api/ipr", "api/transcript", "api/rank"})

	app.GET("/", func(c *gin.Context) {
		c.JSON(200, message)
	})

}

// ADD THIS SCRIPT
func Handler(w http.ResponseWriter, r *http.Request) {
	app.ServeHTTP(w, r)
}

func loginHandler(username, password string, link string) (*colly.Collector, error) {
	link = strings.TrimSuffix(link, "/")
	var loginLink string = link + "/HomeAccess/Account/LogOn"

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
			err := c.Post(loginLink, loginData)
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

	err := c.Visit(loginLink)
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
	link := c.DefaultQuery("link", "https://homeaccess.katyisd.org")
	collector, err := loginHandler(username, password, link)
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

func getAssignments(c *gin.Context) {
	username := c.Query("user")
	password := c.Query("pass")
	link := c.DefaultQuery("link", "https://homeaccess.katyisd.org")
	collector, err := loginHandler(username, password, link)
	if err != nil {
		// Handle the login error
		if err.Error() == "Invalid username or password" {
			c.JSON(401, gin.H{"error": "Invalid username or password"})
		} else {
			c.JSON(500, gin.H{"error": "Failed to log in"})
		}
		return
	}

	classes := make([]string, 0)
	averages := make([]string, 0)

	collector.OnHTML("div.AssignmentClass", func(e *colly.HTMLElement) {
		classArr := strings.Split(strings.Join(strings.Fields(e.ChildText("div.sg-header")), " "), " ")
		class := strings.Join(classArr[3:len(classArr)-3], " ")
		classes = append(classes, class)
		average := e.ChildText("span.sg-header-heading")[18:]
		averages = append(averages, average)
	})

	assignmentstable := make([][]string, 0)
	assignmentsrow := make([]string, 0)

	finaldata := make(map[string]interface{})
	finaldata["assignment"] = make([][][]string, 0)
	finaldata["categories"] = make([][][]string, 0)
	finaldata["classes"] = classes
	finaldata["averages"] = averages

	collector.OnHTML("div.AssignmentClass", func(e *colly.HTMLElement) {
		table := e.DOM.Find("table.sg-asp-table")
		if table.Length() > 0 {
			table.Each(func(_ int, j *goquery.Selection) {
				assignmentstable = nil
				j.Find("tr").Each(func(_ int, row *goquery.Selection) {
					assignmentsrow = nil
					row.Find("td").Each(func(_ int, element *goquery.Selection) {
						text := strings.TrimSpace(element.Text())
						text = strings.ReplaceAll(text, "*", "")
						assignmentsrow = append(assignmentsrow, text)
					})
					assignmentstable = append(assignmentstable, assignmentsrow)
				})

				if strings.Contains(j.AttrOr("id", ""), "CourseCategories") {
					finaldata["categories"] = append(finaldata["categories"].([][][]string), assignmentstable)
				} else if strings.Contains(j.AttrOr("id", ""), "CourseAssignments") {
					finaldata["assignment"] = append(finaldata["assignment"].([][][]string), assignmentstable)
				}
			})
		} else {
			finaldata["assignment"] = append(finaldata["assignment"].([][][]string), [][]string{})
			finaldata["categories"] = append(finaldata["categories"].([][][]string), [][]string{})
		}
	})

	collector.OnScraped(func(r *colly.Response) {
		ret := make(map[string]interface{})
		for i := 0; i < len(classes); i++ {
			average := averages[i]
			assig := finaldata["assignment"].([][][]string)[i]
			categories := finaldata["categories"].([][][]string)[i]
			l := make(map[string]interface{})
			l["average"] = average
			l["assignments"] = assig
			l["categories"] = categories
			ret[classes[i]] = l
		}

		c.JSON(200, ret)
	})

	err = collector.Visit("https://homeaccess.katyisd.org/HomeAccess/Content/Student/Assignments.aspx")
	if err != nil {
		c.JSON(500, gin.H{"error": "Failed to scrape data"})
		return
	}
}

func getInfo(c *gin.Context) {
	username := c.Query("user")
	password := c.Query("pass")
	link := c.DefaultQuery("link", "https://homeaccess.katyisd.org")
	collector, err := loginHandler(username, password, link)
	if err != nil {
		// Handle the login error
		if err.Error() == "Invalid username or password" {
			c.JSON(401, gin.H{"error": "Invalid username or password"})
		} else {
			c.JSON(500, gin.H{"error": "Failed to log in"})
		}
		return
	}

	ret := make(map[string]interface{})

	collector.OnHTML("div.sg-main-content", func(e *colly.HTMLElement) {
		name := e.ChildText("#plnMain_lblRegStudentName")
		if name != "" {
			ret["name"] = strings.TrimSpace(name)
			grade := e.ChildText("#plnMain_lblGrade")
			ret["grade"] = strings.TrimSpace(grade)
			school := e.ChildText("#plnMain_lblBuildingName")
			ret["school"] = strings.TrimSpace(school)
			dob := e.ChildText("#plnMain_lblBirthDate")
			ret["dob"] = strings.TrimSpace(dob)
			counselor := e.ChildText("#plnMain_lblCounselor")
			ret["counselor"] = strings.TrimSpace(counselor)
			language := e.ChildText("#plnMain_lblLanguage")
			ret["language"] = strings.TrimSpace(language)
			cohortYear := e.ChildText("#plnMain_lblCohortYear")
			fmt.Println(cohortYear)
			ret["cohort-year"] = strings.TrimSpace(cohortYear)
		}
	})

	collector.OnScraped(func(r *colly.Response) {
		c.JSON(200, ret)
	})

	err = collector.Visit("https://homeaccess.katyisd.org/HomeAccess/Content/Student/Registration.aspx")
	if err != nil {
		c.JSON(500, gin.H{"error": "Failed to scrape data"})
		return
	}
}

func getAverages(c *gin.Context) {
	username := c.Query("user")
	password := c.Query("pass")
	link := c.DefaultQuery("link", "https://homeaccess.katyisd.org")
	collector, err := loginHandler(username, password, link)
	if err != nil {
		// Handle the login error
		if err.Error() == "Invalid username or password" {
			c.JSON(401, gin.H{"error": "Invalid username or password"})
		} else {
			c.JSON(500, gin.H{"error": "Failed to log in"})
		}
		return
	}

	classes := make([]string, 0)
	averages := make([]string, 0)

	collector.OnHTML("div.AssignmentClass", func(e *colly.HTMLElement) {
		classArr := strings.Split(strings.Join(strings.Fields(e.ChildText("div.sg-header")), " "), " ")
		class := strings.Join(classArr[3:len(classArr)-3], " ")
		classes = append(classes, class)
		average := e.ChildText("span.sg-header-heading")[18:]
		averages = append(averages, average)
	})

	collector.OnScraped(func(r *colly.Response) {
		ret := orderedmap.New()
		for i := 0; i < len(classes); i++ {
			average := averages[i]
			ret.Set(classes[i], average)
		}

		c.JSON(200, ret)
	})

	err = collector.Visit("https://homeaccess.katyisd.org/HomeAccess/Content/Student/Assignments.aspx")
	if err != nil {
		c.JSON(500, gin.H{"error": "Failed to scrape data"})
		return
	}
}

func getClasses(c *gin.Context) {
	username := c.Query("user")
	password := c.Query("pass")
	link := c.DefaultQuery("link", "https://homeaccess.katyisd.org")
	collector, err := loginHandler(username, password, link)
	if err != nil {
		// Handle the login error
		if err.Error() == "Invalid username or password" {
			c.JSON(401, gin.H{"error": "Invalid username or password"})
		} else {
			c.JSON(500, gin.H{"error": "Failed to log in"})
		}
		return
	}

	classes := make([]string, 0)

	collector.OnHTML("div.AssignmentClass", func(e *colly.HTMLElement) {
		classArr := strings.Split(strings.Join(strings.Fields(e.ChildText("div.sg-header")), " "), " ")
		class := strings.Join(classArr[3:len(classArr)-3], " ")
		classes = append(classes, class)
	})

	collector.OnScraped(func(r *colly.Response) {
		c.JSON(200, classes)
	})

	err = collector.Visit("https://homeaccess.katyisd.org/HomeAccess/Content/Student/Assignments.aspx")
	if err != nil {
		c.JSON(500, gin.H{"error": "Failed to scrape data"})
		return
	}
}

func getReport(c *gin.Context) {
	username := c.Query("user")
	password := c.Query("pass")
	link := c.DefaultQuery("link", "https://homeaccess.katyisd.org")
	collector, err := loginHandler(username, password, link)
	if err != nil {
		// Handle the login error
		if err.Error() == "Invalid username or password" {
			c.JSON(401, gin.H{"error": "Invalid username or password"})
		} else {
			c.JSON(500, gin.H{"error": "Failed to log in"})
		}
		return
	}

	finalData := orderedmap.New()
	headers := []string{"Course", "Description", "Period", "Teacher", "Room", "1st", "2nd", "3rd", "Exam1", "Sem1", "4th", "5th", "6th", "Exam2", "Sem2", "CND1", "CND2", "CND3", "CND4", "CND5", "CND6"}
	finalData.Set("headers", headers)

	var row []string
	var data [][]string
	counter := 0

	collector.OnHTML("td", func(e *colly.HTMLElement) {
		counter++
		if counter > 32 {
			row = append(row, strings.TrimSpace(e.Text))
		}
		if len(row)%32 == 0 && counter > 32 {
			data = append(data, row)
			row = nil
		}
	})

	collector.OnScraped(func(r *colly.Response) {
		for i := range data {
			data[i] = append(data[i][:23], data[i][32:]...)
			data[i] = append(data[i][:5], data[i][7:]...)
		}
		finalData.Set("data", data)

		c.JSON(200, finalData)
	})

	err = collector.Visit("https://homeaccess.katyisd.org/HomeAccess/Content/Student/ReportCards.aspx")
	if err != nil {
		c.JSON(500, gin.H{"error": "Failed to scrape data"})
		return
	}
}

func getProgressReport(c *gin.Context) {
	username := c.Query("user")
	password := c.Query("pass")
	link := c.DefaultQuery("link", "https://homeaccess.katyisd.org")
	collector, err := loginHandler(username, password, link)
	if err != nil {
		// Handle the login error
		if err.Error() == "Invalid username or password" {
			c.JSON(401, gin.H{"error": "Invalid username or password"})
		} else {
			c.JSON(500, gin.H{"error": "Failed to log in"})
		}
		return
	}

	finalData := orderedmap.New()
	var headers []string
	var row []string
	var data [][]string

	collector.OnHTML("tr", func(e *colly.HTMLElement) {
		e.ForEach("td", func(_ int, el *colly.HTMLElement) {
			row = append(row, strings.TrimSpace(el.Text))
		})
		fmt.Println(row)
		data = append(data, row)
		row = nil
	})

	err = collector.Visit("https://homeaccess.katyisd.org/HomeAccess/Content/Student/InterimProgress.aspx")
	if err != nil {
		c.JSON(500, gin.H{"error": "Failed to scrape data"})
		return
	}

	if len(data) == 0 {
		c.JSON(200, nil)
		return
	}

	headers = data[0]
	data = data[1:]

	finalData.Set("headers", headers)
	finalData.Set("data", data)

	c.JSON(200, finalData)
}

func getTranscript(c *gin.Context){
	username := c.Query("user")
	password := c.Query("pass")
	link := c.DefaultQuery("link", "https://homeaccess.katyisd.org")
	collector, err := loginHandler(username, password, link)
	if err != nil {
		// Handle the login error
		if err.Error() == "Invalid username or password" {
			c.JSON(401, gin.H{"error": "Invalid username or password"})
		} else {
			c.JSON(500, gin.H{"error": "Failed to log in"})
		}
		return
	}

	transcript := orderedmap.New()

	collector.OnHTML("td.sg-transcript-group", func(e *colly.HTMLElement) {
		semester := orderedmap.New()

		// First table
		e.ForEach("table > tbody > tr > td > span", func(_ int, el *colly.HTMLElement) {
			if strings.Contains(el.Attr("id"), "YearValue") {
				semester.Set("year", el.Text)
			} else if strings.Contains(el.Attr("id"), "GroupValue") {
				semester.Set("semester", el.Text)
			} else if strings.Contains(el.Attr("id"), "GradeValue") {
				semester.Set("grade", el.Text)
			} else if strings.Contains(el.Attr("id"), "BuildingValue") {
				semester.Set("school", el.Text)
			}
		})

		
		finaldata := make([][]string, 0)
		data := make([]string, 0)

		// Second table
		e.ForEach("table:nth-child(2) > tbody > tr", func(_ int, el *colly.HTMLElement) {
			if strings.Contains(el.Attr("class"), "sg-asp-table-header-row") ||
				strings.Contains(el.Attr("class"), "sg-asp-table-data-row") {
				el.ForEach("td", func(_ int, el2 *colly.HTMLElement) {
					data = append(data, el2.Text)
				})
				finaldata = append(finaldata, data)
				semester.Set("data", finaldata)
				data = make([]string, 0)
			}
		})

		// Third table
		e.ForEach("table:nth-child(3) > tbody > tr > td > label", func(_ int, el *colly.HTMLElement) {
			if strings.Contains(el.Attr("id"), "CreditValue") {
				semester.Set("credits", el.Text)
			}
		})
		
		year, _ := semester.Get("year")
		semesterNum, _ := semester.Get("semester")
		title := year.(string) + " - Semester " + semesterNum.(string)
		fmt.Println(title)

		transcript.Set(title, semester)
	})

	collector.OnHTML("table#plnMain_rpTranscriptGroup_tblCumGPAInfo", func(e *colly.HTMLElement) {
		e.ForEach("tbody > tr.sg-asp-table-data-row", func(_ int, el *colly.HTMLElement) {
			var text string
			var value string
			el.ForEach(" td > span", func(_ int, el2 *colly.HTMLElement) {
				if strings.Contains(el2.Attr("id"), "GPADescr") {
					text = el2.Text
				}
				if strings.Contains(el2.Attr("id"), "GPACum") {
					value = el2.Text
				}
				if strings.Contains(el2.Attr("id"), "GPARank") {
					transcript.Set("rank", el2.Text)
				}
				if strings.Contains(el2.Attr("id"), "GPAQuartile") {
					transcript.Set("quartile", el2.Text)
				}
				
			})
			transcript.Set(text, value)
		})
	})

	collector.OnScraped(func(r *colly.Response) {
		c.JSON(200, transcript)
	})

	err = collector.Visit(link+"/HomeAccess/Content/Student/Transcript.aspx")
	if err != nil {
		c.JSON(500, gin.H{"error": "Failed to scrape data"})
		return
	}
}

func getRank(c *gin.Context){
	username := c.Query("user")
	password := c.Query("pass")
	link := c.DefaultQuery("link", "https://homeaccess.katyisd.org")
	collector, err := loginHandler(username, password, link)
	if err != nil {
		// Handle the login error
		if err.Error() == "Invalid username or password" {
			c.JSON(401, gin.H{"error": "Invalid username or password"})
		} else {
			c.JSON(500, gin.H{"error": "Failed to log in"})
		}
		return
	}

	transcript := orderedmap.New()

	collector.OnHTML("table#plnMain_rpTranscriptGroup_tblCumGPAInfo", func(e *colly.HTMLElement) {
		e.ForEach("tbody > tr.sg-asp-table-data-row", func(_ int, el *colly.HTMLElement) {
			var text string
			var value string
			el.ForEach(" td > span", func(_ int, el2 *colly.HTMLElement) {
				if strings.Contains(el2.Attr("id"), "GPADescr") {
					text = el2.Text
				}
				if strings.Contains(el2.Attr("id"), "GPACum") {
					value = el2.Text
				}
				if strings.Contains(el2.Attr("id"), "GPARank") {
					transcript.Set("rank", el2.Text)
				}
				if strings.Contains(el2.Attr("id"), "GPAQuartile") {
					transcript.Set("quartile", el2.Text)
				}
				
			})
			transcript.Set(text, value)
		})
	})

	collector.OnScraped(func(r *colly.Response) {
		c.JSON(200, transcript)
	})

	err = collector.Visit(link+"/HomeAccess/Content/Student/Transcript.aspx")
	if err != nil {
		c.JSON(500, gin.H{"error": "Failed to scrape data"})
		return
	}
}