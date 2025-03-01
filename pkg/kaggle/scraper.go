package kaggle

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/taross-f/tsuyo-kagg/pkg/config"
)

// User represents a Kaggle user
type User struct {
	DisplayName  string `json:"displayName"`
	Bio          string `json:"bio"`
	Country      string `json:"country"`
	KaggleURL    string `json:"kaggleUrl"`
	TwitterURL   string `json:"twitterUrl"`
	LinkedInURL  string `json:"linkedInUrl"`
	GitHubURL    string `json:"githubUrl"`
	WebsiteURL   string `json:"websiteUrl"`
	Organization string `json:"organization"`
	AvatarURL    string `json:"avatarUrl"`
	Email        string `json:"email"`
}

// Scraper handles scraping Kaggle data
type Scraper struct {
	Config config.Config
}

// NewScraper creates a new Kaggle scraper
func NewScraper(cfg config.Config) *Scraper {
	return &Scraper{
		Config: cfg,
	}
}

// GetRankings fetches user rankings from Kaggle
func (s *Scraper) GetRankings() ([]User, error) {
	rand.Seed(time.Now().UnixNano())
	var users []User

	for i := 0; i < s.Config.MaxPages; i++ {
		log.Printf("Fetching page %d of %d", i+1, s.Config.MaxPages)

		rankingURL := fmt.Sprintf("%s/rankings?group=competitions&page=%d&pageSize=%d",
			s.Config.KaggleBaseURL, i+1, s.Config.PageSize)

		splashURL := fmt.Sprintf("%s/render.html?url=%s&timeout=%d&wait=%d",
			s.Config.SplashURL, url.QueryEscape(rankingURL), s.Config.RequestTimeout, s.Config.WaitTime)

		// Fetch the page using Splash
		resp, err := http.Get(splashURL)
		if err != nil {
			log.Printf("Error fetching ranking page %d: %v", i+1, err)
			continue
		}
		defer resp.Body.Close()

		// Parse the HTML document
		doc, err := goquery.NewDocumentFromReader(resp.Body)
		if err != nil {
			log.Printf("Error parsing HTML from page %d: %v", i+1, err)
			continue
		}

		// Find user links in the rankings page
		doc.Find("a.block.UserEntity-link").Each(func(j int, selection *goquery.Selection) {
			userURL, exists := selection.Attr("href")
			if !exists {
				return
			}

			// Make sure it's a user profile link
			if !strings.HasPrefix(userURL, "/") {
				return
			}

			fullUserURL := fmt.Sprintf("%s%s", s.Config.KaggleBaseURL, userURL)
			user, err := s.GetUserDetails(fullUserURL)
			if err != nil {
				log.Printf("Error fetching user details: %v", err)
				return
			}

			if user != nil {
				users = append(users, *user)
			}

			// Random delay between requests
			delay := time.Duration(rand.Intn(s.Config.MaxDelay-s.Config.MinDelay+1)+s.Config.MinDelay) * time.Second
			time.Sleep(delay)
		})

		// If we didn't find any users on this page, we might be at the end
		if len(users) == 0 && i > 0 {
			log.Printf("No users found on page %d, might be at the end of rankings", i+1)
		}
	}

	return users, nil
}

// GetUserDetails fetches details for a specific user
func (s *Scraper) GetUserDetails(userURL string) (*User, error) {
	splashURL := fmt.Sprintf("%s/render.html?url=%s&timeout=%d&wait=%d",
		s.Config.SplashURL, url.QueryEscape(userURL), s.Config.RequestTimeout, s.Config.WaitTime)

	resp, err := http.Get(splashURL)
	if err != nil {
		return nil, fmt.Errorf("error fetching user page: %w", err)
	}
	defer resp.Body.Close()

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error parsing user page HTML: %w", err)
	}

	scriptText := doc.Find("body > main > div > div.site-layout__main-content > script").Text()

	if len(strings.Split(scriptText, "Kaggle.State.push(")) < 2 {
		return nil, fmt.Errorf("insufficient data in user page")
	}

	jsonPart := strings.Split(strings.Split(scriptText, "Kaggle.State.push(")[1], ");")[0]

	var userData map[string]interface{}
	if err := json.Unmarshal([]byte(jsonPart), &userData); err != nil {
		return nil, fmt.Errorf("error parsing user JSON: %w", err)
	}

	country, _ := userData["country"].(string)
	if country != s.Config.TargetCountry && country != "JP" && country != "日本" {
		return nil, nil // Skip non-Japanese users
	}

	displayName, _ := userData["displayName"].(string)
	bio, _ := userData["bio"].(string)
	twitterUserName, _ := userData["twitterUserName"].(string)
	linkedInURL, _ := userData["linkedInUrl"].(string)
	gitHubUserName, _ := userData["gitHubUserName"].(string)
	websiteURL, _ := userData["websiteUrl"].(string)
	organization, _ := userData["organization"].(string)
	userAvatarURL, _ := userData["userAvatarUrl"].(string)
	email, _ := userData["email"].(string)

	user := &User{
		DisplayName:  displayName,
		Bio:          bio,
		Country:      country,
		KaggleURL:    userURL,
		TwitterURL:   fmt.Sprintf("https://twitter.com/%s", twitterUserName),
		LinkedInURL:  linkedInURL,
		GitHubURL:    fmt.Sprintf("https://github.com/%s", gitHubUserName),
		WebsiteURL:   websiteURL,
		Organization: organization,
		AvatarURL:    userAvatarURL,
		Email:        email,
	}

	log.Printf("Found user: %s from %s", displayName, country)

	return user, nil
}

// ExportToCSV exports users to CSV format
func (s *Scraper) ExportToCSV(users []User) string {
	header := "name,bio,country,Kaggle,Twitter,LinkedIn,Github,Blog"
	lines := []string{header}

	for _, user := range users {
		// Replace newlines and commas in bio to avoid breaking CSV format
		cleanBio := strings.Replace(strings.Replace(user.Bio, "\n", " ", -1), ",", ";", -1)

		line := fmt.Sprintf("%s,\"%s\",%s,%s,%s,%s,%s,%s",
			user.DisplayName,
			cleanBio,
			user.Country,
			user.KaggleURL,
			user.TwitterURL,
			user.LinkedInURL,
			user.GitHubURL,
			user.WebsiteURL)

		lines = append(lines, line)
	}

	return strings.Join(lines, "\n")
}
