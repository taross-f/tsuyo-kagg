package kaggle

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/taross-f/tsuyo-kagg/pkg/config"
)

// APIClient handles interactions with the Kaggle API
type APIClient struct {
	Config config.Config
}

// NewAPIClient creates a new Kaggle API client
func NewAPIClient(cfg config.Config) *APIClient {
	return &APIClient{
		Config: cfg,
	}
}

// EnsureKaggleCredentials checks if Kaggle API credentials are set up
func (a *APIClient) EnsureKaggleCredentials() error {
	// Check if KAGGLE_USERNAME and KAGGLE_KEY environment variables are set
	username := os.Getenv("KAGGLE_USERNAME")
	key := os.Getenv("KAGGLE_KEY")

	if username != "" && key != "" {
		log.Println("Kaggle credentials found in environment variables")
		return nil
	}

	// Check if kaggle.json exists in the default location
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get user home directory: %w", err)
	}

	kaggleConfigPath := filepath.Join(homeDir, ".kaggle", "kaggle.json")
	if _, err := os.Stat(kaggleConfigPath); err == nil {
		log.Println("Kaggle credentials found at", kaggleConfigPath)
		return nil
	}

	return fmt.Errorf("Kaggle API credentials not found. Please set KAGGLE_USERNAME and KAGGLE_KEY environment variables or create ~/.kaggle/kaggle.json")
}

// GetUsers fetches users from Kaggle API
func (a *APIClient) GetUsers() ([]User, error) {
	// Ensure Kaggle credentials are set up
	if err := a.EnsureKaggleCredentials(); err != nil {
		return nil, err
	}

	// Use kaggle-api Python package via command line
	// This is a temporary solution until we implement direct API calls
	tempDir, err := os.MkdirTemp("", "kaggle-users")
	if err != nil {
		return nil, fmt.Errorf("failed to create temp directory: %w", err)
	}
	defer os.RemoveAll(tempDir)

	// Create a Python script to fetch users
	scriptPath := filepath.Join(tempDir, "fetch_users.py")
	script := `
import kaggle
import json
import sys

# Get list of users
users = kaggle.api.user_list()

# Filter Japanese users
japanese_users = [user for user in users if user.country == 'Japan' or user.country == 'JP' or user.country == '日本']

# Convert to JSON and print
user_data = []
for user in japanese_users:
    user_dict = {
        'displayName': user.displayName,
        'bio': user.bio if hasattr(user, 'bio') else '',
        'country': user.country,
        'kaggleUrl': f"https://www.kaggle.com/{user.username}",
        'twitterUrl': f"https://twitter.com/{user.twitterName}" if hasattr(user, 'twitterName') and user.twitterName else '',
        'linkedInUrl': user.linkedInUrl if hasattr(user, 'linkedInUrl') else '',
        'githubUrl': f"https://github.com/{user.githubName}" if hasattr(user, 'githubName') and user.githubName else '',
        'websiteUrl': user.websiteUrl if hasattr(user, 'websiteUrl') else '',
        'organization': user.organization if hasattr(user, 'organization') else '',
        'avatarUrl': user.avatarUrl if hasattr(user, 'avatarUrl') else '',
        'email': ''  # Email is not available via API
    }
    user_data.append(user_dict)

json.dump(user_data, sys.stdout)
`

	if err := os.WriteFile(scriptPath, []byte(script), 0644); err != nil {
		return nil, fmt.Errorf("failed to write Python script: %w", err)
	}

	// Run the Python script
	cmd := exec.Command("python", scriptPath)
	output, err := cmd.Output()
	if err != nil {
		var stderr string
		if exitErr, ok := err.(*exec.ExitError); ok {
			stderr = string(exitErr.Stderr)
		}
		return nil, fmt.Errorf("failed to run Python script: %w, stderr: %s", err, stderr)
	}

	// Parse the JSON output
	var users []User
	if err := json.Unmarshal(output, &users); err != nil {
		return nil, fmt.Errorf("failed to parse user data: %w", err)
	}

	log.Printf("Found %d Japanese users via Kaggle API", len(users))
	return users, nil
}

// GetUserDetails fetches details for a specific user
func (a *APIClient) GetUserDetails(username string) (*User, error) {
	// Ensure Kaggle credentials are set up
	if err := a.EnsureKaggleCredentials(); err != nil {
		return nil, err
	}

	// Extract username from URL if needed
	if strings.Contains(username, "/") {
		parts := strings.Split(username, "/")
		username = parts[len(parts)-1]
	}

	// Construct API URL
	apiURL := fmt.Sprintf("https://www.kaggle.com/api/v1/users/profile/%s", username)

	// Create request
	req, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Add authentication headers
	homeDir, _ := os.UserHomeDir()
	kaggleConfigPath := filepath.Join(homeDir, ".kaggle", "kaggle.json")
	if _, err := os.Stat(kaggleConfigPath); err == nil {
		// Read kaggle.json
		data, err := os.ReadFile(kaggleConfigPath)
		if err != nil {
			return nil, fmt.Errorf("failed to read Kaggle config: %w", err)
		}

		var config struct {
			Username string `json:"username"`
			Key      string `json:"key"`
		}
		if err := json.Unmarshal(data, &config); err != nil {
			return nil, fmt.Errorf("failed to parse Kaggle config: %w", err)
		}

		req.SetBasicAuth(config.Username, config.Key)
	} else {
		// Use environment variables
		username := os.Getenv("KAGGLE_USERNAME")
		key := os.Getenv("KAGGLE_KEY")
		req.SetBasicAuth(username, key)
	}

	// Send request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	// Check response status
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}

	// Parse response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	var userData map[string]interface{}
	if err := json.Unmarshal(body, &userData); err != nil {
		return nil, fmt.Errorf("failed to parse user data: %w", err)
	}

	// Check if user is Japanese
	country, _ := userData["country"].(string)
	if country != a.Config.TargetCountry && country != "JP" && country != "日本" {
		return nil, nil // Skip non-Japanese users
	}

	// Extract user data
	displayName, _ := userData["displayName"].(string)
	bio, _ := userData["bio"].(string)
	twitterUserName, _ := userData["twitterName"].(string)
	linkedInURL, _ := userData["linkedInUrl"].(string)
	gitHubUserName, _ := userData["githubName"].(string)
	websiteURL, _ := userData["websiteUrl"].(string)
	organization, _ := userData["organization"].(string)
	userAvatarURL, _ := userData["avatarUrl"].(string)

	user := &User{
		DisplayName:  displayName,
		Bio:          bio,
		Country:      country,
		KaggleURL:    fmt.Sprintf("https://www.kaggle.com/%s", username),
		TwitterURL:   fmt.Sprintf("https://twitter.com/%s", twitterUserName),
		LinkedInURL:  linkedInURL,
		GitHubURL:    fmt.Sprintf("https://github.com/%s", gitHubUserName),
		WebsiteURL:   websiteURL,
		Organization: organization,
		AvatarURL:    userAvatarURL,
		Email:        "", // Email is not available via API
	}

	log.Printf("Found user: %s from %s", displayName, country)
	return user, nil
}

// ExportToCSV exports users to CSV format
func (a *APIClient) ExportToCSV(users []User) string {
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
