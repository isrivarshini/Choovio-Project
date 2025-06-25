package api

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/gin-gonic/gin" // Assuming you're using Gin as your router
)

// Device struct to represent a device
type Device struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Type string `json:"type"`
}

// User struct to represent a user
type User struct {
	ID        string                 `json:"id"`
	Email     string                 `json:"email"`
	Name      string                 `json:"name"`
	Password  string                 `json:"password,omitempty"` // For authentication
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
	CreatedAt time.Time              `json:"created_at"`
}

// Thing struct to represent a thing (device)
type Thing struct {
	ID       string                 `json:"id"`
	Name     string                 `json:"name"`
	Key      string                 `json:"key"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// Channel struct to represent a channel
type Channel struct {
	ID       string                 `json:"id"`
	Name     string                 `json:"name"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// UserStorage struct to manage user data persistence
type UserStorage struct {
	Users []User `json:"users"`
}

const usersFilePath = "data/users.json"

// Sample data for devices
var devices = []Device{
	{ID: "1", Name: "Device 1", Type: "Sensor"},
	{ID: "2", Name: "Device 2", Type: "Actuator"},
}

// Global variable to hold users in memory (loaded from file)
var usersData []User

// Sample data for things
var things = []Thing{
	{ID: "thing1", Name: "Temperature Sensor", Key: "temp_key_123"},
	{ID: "thing2", Name: "Humidity Sensor", Key: "humid_key_456"},
}

// Sample data for channels
var channelsData = []Channel{
	{ID: "channel1", Name: "Sensor Data Channel"},
	{ID: "channel2", Name: "Control Channel"},
}

// Initialize user storage - creates directory and loads users from file
func InitUserStorage() error {
	// Create data directory if it doesn't exist
	dataDir := filepath.Dir(usersFilePath)
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		return fmt.Errorf("failed to create data directory: %v", err)
	}

	// Load existing users or create default ones
	if err := loadUsersFromFile(); err != nil {
		fmt.Printf("No existing users file found, creating default users: %v\n", err)
		// Create default users
		usersData = []User{
			{
				ID:        "1",
				Email:     "admin@example.com",
				Name:      "Admin User",
				Password:  "admin123",
				CreatedAt: time.Now(),
			},
			{
				ID:        "2",
				Email:     "user@example.com",
				Name:      "Regular User",
				Password:  "password123",
				CreatedAt: time.Now(),
			},
		}
		// Save default users to file
		if err := saveUsersToFile(); err != nil {
			return fmt.Errorf("failed to save default users: %v", err)
		}
	}

	fmt.Printf("Loaded %d users from storage\n", len(usersData))
	return nil
}

// Load users from JSON file
func loadUsersFromFile() error {
	if _, err := os.Stat(usersFilePath); os.IsNotExist(err) {
		return fmt.Errorf("users file does not exist")
	}

	data, err := ioutil.ReadFile(usersFilePath)
	if err != nil {
		return fmt.Errorf("failed to read users file: %v", err)
	}

	var storage UserStorage
	if err := json.Unmarshal(data, &storage); err != nil {
		return fmt.Errorf("failed to unmarshal users data: %v", err)
	}

	usersData = storage.Users
	return nil
}

// Save users to JSON file
func saveUsersToFile() error {
	storage := UserStorage{Users: usersData}

	data, err := json.MarshalIndent(storage, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal users data: %v", err)
	}

	if err := ioutil.WriteFile(usersFilePath, data, 0644); err != nil {
		return fmt.Errorf("failed to write users file: %v", err)
	}

	return nil
}

// Find user by email
func findUserByEmail(email string) *User {
	for i := range usersData {
		if usersData[i].Email == email {
			return &usersData[i]
		}
	}
	return nil
}

// GetDevices handles GET requests to fetch devices
func GetDevices(c *gin.Context) {
	c.JSON(http.StatusOK, devices)
}

// GetUsers handles GET requests to fetch users
func GetUsers(c *gin.Context) {
	// Create a copy of users without passwords for security
	safeUsers := make([]User, len(usersData))
	for i, user := range usersData {
		safeUsers[i] = user
		safeUsers[i].Password = "" // Don't expose passwords
	}

	c.JSON(http.StatusOK, gin.H{
		"users": safeUsers,
		"total": len(safeUsers),
	})
}

// CreateUser handles POST requests to create a new user
func CreateUser(c *gin.Context) {
	var newUser User
	if err := c.ShouldBindJSON(&newUser); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid request"})
		return
	}

	// Check if user with email already exists
	if findUserByEmail(newUser.Email) != nil {
		c.JSON(http.StatusConflict, gin.H{"message": "User with this email already exists"})
		return
	}

	// Generate a unique ID
	newUser.ID = strconv.Itoa(len(usersData) + 1)
	newUser.CreatedAt = time.Now()

	// Set default password if not provided
	if newUser.Password == "" {
		newUser.Password = "password123" // Default password for demo
	}

	// Add user to memory
	usersData = append(usersData, newUser)

	// Save to file
	if err := saveUsersToFile(); err != nil {
		// If saving fails, remove from memory to maintain consistency
		usersData = usersData[:len(usersData)-1]
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to save user data"})
		return
	}

	// Don't return password in response for security
	responseUser := newUser
	responseUser.Password = ""

	fmt.Printf("Successfully created user: %s with email: %s\n", newUser.Name, newUser.Email)
	c.JSON(http.StatusCreated, responseUser)
}

// GetThings handles GET requests to fetch things
func GetThings(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"things": things,
		"total":  len(things),
	})
}

// CreateThing handles POST requests to create a new thing
func CreateThing(c *gin.Context) {
	var newThing Thing
	if err := c.ShouldBindJSON(&newThing); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid request"})
		return
	}

	// Generate a simple ID and key
	newThing.ID = "thing_" + strconv.Itoa(len(things)+1)
	newThing.Key = "key_" + newThing.ID
	things = append(things, newThing)

	c.JSON(http.StatusCreated, newThing)
}

// GetChannels handles GET requests to fetch channels
func GetChannels(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"channels": channelsData,
		"total":    len(channelsData),
	})
}

// CreateChannel handles POST requests to create a new channel
func CreateChannel(c *gin.Context) {
	var newChannel Channel
	if err := c.ShouldBindJSON(&newChannel); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid request"})
		return
	}

	// Generate a simple ID
	newChannel.ID = "channel_" + strconv.Itoa(len(channelsData)+1)
	channelsData = append(channelsData, newChannel)

	c.JSON(http.StatusCreated, newChannel)
}

// Health check endpoint
func HealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":       "healthy",
		"service":      "magistrala-api",
		"timestamp":    time.Now().Format(time.RFC3339),
		"users_loaded": len(usersData),
	})
}

// GetUsersData returns the users data for authentication (used by websocket.go)
func GetUsersData() []User {
	return usersData
}

// UpdateUser handles PUT requests to update an existing user
func UpdateUser(c *gin.Context) {
	userID := c.Param("id")

	var updateData User
	if err := c.ShouldBindJSON(&updateData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid request"})
		return
	}

	// Find the user by ID
	userIndex := -1
	for i, user := range usersData {
		if user.ID == userID {
			userIndex = i
			break
		}
	}

	if userIndex == -1 {
		c.JSON(http.StatusNotFound, gin.H{"message": "User not found"})
		return
	}

	// Update user fields (keep existing values if not provided)
	if updateData.Name != "" {
		usersData[userIndex].Name = updateData.Name
	}
	if updateData.Email != "" {
		// Check if new email conflicts with another user
		for i, user := range usersData {
			if i != userIndex && user.Email == updateData.Email {
				c.JSON(http.StatusConflict, gin.H{"message": "Email already exists"})
				return
			}
		}
		usersData[userIndex].Email = updateData.Email
	}
	if updateData.Password != "" {
		usersData[userIndex].Password = updateData.Password
	}
	if updateData.Metadata != nil {
		usersData[userIndex].Metadata = updateData.Metadata
	}

	// Save to file
	if err := saveUsersToFile(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to save user data"})
		return
	}

	// Return updated user without password
	responseUser := usersData[userIndex]
	responseUser.Password = ""

	fmt.Printf("Successfully updated user: %s with email: %s\n", responseUser.Name, responseUser.Email)
	c.JSON(http.StatusOK, responseUser)
}

// DeleteUser handles DELETE requests to delete a user
func DeleteUser(c *gin.Context) {
	userID := c.Param("id")

	// Find the user by ID
	userIndex := -1
	for i, user := range usersData {
		if user.ID == userID {
			userIndex = i
			break
		}
	}

	if userIndex == -1 {
		c.JSON(http.StatusNotFound, gin.H{"message": "User not found"})
		return
	}

	// Remove user from slice
	deletedUser := usersData[userIndex]
	usersData = append(usersData[:userIndex], usersData[userIndex+1:]...)

	// Save to file
	if err := saveUsersToFile(); err != nil {
		// If saving fails, restore user to maintain consistency
		usersData = append(usersData[:userIndex], append([]User{deletedUser}, usersData[userIndex:]...)...)
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to save user data"})
		return
	}

	fmt.Printf("Successfully deleted user: %s with email: %s\n", deletedUser.Name, deletedUser.Email)
	c.JSON(http.StatusOK, gin.H{"message": "User deleted successfully"})
}

// GetUserByID handles GET requests to fetch a user by ID
func GetUserByID(c *gin.Context) {
	userID := c.Param("id")

	// Find the user by ID
	for _, user := range usersData {
		if user.ID == userID {
			// Return user without password for security
			responseUser := user
			responseUser.Password = ""
			c.JSON(http.StatusOK, responseUser)
			return
		}
	}

	c.JSON(http.StatusNotFound, gin.H{"message": "User not found"})
}

// UpdateThing handles PUT requests to update an existing thing
func UpdateThing(c *gin.Context) {
	thingID := c.Param("id")

	var updateData Thing
	if err := c.ShouldBindJSON(&updateData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid request"})
		return
	}

	// Find the thing by ID
	thingIndex := -1
	for i, thing := range things {
		if thing.ID == thingID {
			thingIndex = i
			break
		}
	}

	if thingIndex == -1 {
		c.JSON(http.StatusNotFound, gin.H{"message": "Thing not found"})
		return
	}

	// Update thing fields (keep existing values if not provided)
	if updateData.Name != "" {
		things[thingIndex].Name = updateData.Name
	}
	if updateData.Key != "" {
		things[thingIndex].Key = updateData.Key
	}
	if updateData.Metadata != nil {
		things[thingIndex].Metadata = updateData.Metadata
	}

	fmt.Printf("Successfully updated thing: %s with ID: %s\n", things[thingIndex].Name, things[thingIndex].ID)
	c.JSON(http.StatusOK, things[thingIndex])
}

// DeleteThing handles DELETE requests to delete a thing
func DeleteThing(c *gin.Context) {
	thingID := c.Param("id")

	// Find the thing by ID
	thingIndex := -1
	for i, thing := range things {
		if thing.ID == thingID {
			thingIndex = i
			break
		}
	}

	if thingIndex == -1 {
		c.JSON(http.StatusNotFound, gin.H{"message": "Thing not found"})
		return
	}

	// Remove thing from slice
	deletedThing := things[thingIndex]
	things = append(things[:thingIndex], things[thingIndex+1:]...)

	fmt.Printf("Successfully deleted thing: %s with ID: %s\n", deletedThing.Name, deletedThing.ID)
	c.JSON(http.StatusOK, gin.H{"message": "Thing deleted successfully"})
}

// GetThingByID handles GET requests to fetch a thing by ID
func GetThingByID(c *gin.Context) {
	thingID := c.Param("id")

	// Find the thing by ID
	for _, thing := range things {
		if thing.ID == thingID {
			c.JSON(http.StatusOK, thing)
			return
		}
	}

	c.JSON(http.StatusNotFound, gin.H{"message": "Thing not found"})
}

// UpdateChannel handles PUT requests to update an existing channel
func UpdateChannel(c *gin.Context) {
	channelID := c.Param("id")

	var updateData Channel
	if err := c.ShouldBindJSON(&updateData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid request"})
		return
	}

	// Find the channel by ID
	channelIndex := -1
	for i, channel := range channelsData {
		if channel.ID == channelID {
			channelIndex = i
			break
		}
	}

	if channelIndex == -1 {
		c.JSON(http.StatusNotFound, gin.H{"message": "Channel not found"})
		return
	}

	// Update channel fields (keep existing values if not provided)
	if updateData.Name != "" {
		channelsData[channelIndex].Name = updateData.Name
	}
	if updateData.Metadata != nil {
		channelsData[channelIndex].Metadata = updateData.Metadata
	}

	fmt.Printf("Successfully updated channel: %s with ID: %s\n", channelsData[channelIndex].Name, channelsData[channelIndex].ID)
	c.JSON(http.StatusOK, channelsData[channelIndex])
}

// DeleteChannel handles DELETE requests to delete a channel
func DeleteChannel(c *gin.Context) {
	channelID := c.Param("id")

	// Find the channel by ID
	channelIndex := -1
	for i, channel := range channelsData {
		if channel.ID == channelID {
			channelIndex = i
			break
		}
	}

	if channelIndex == -1 {
		c.JSON(http.StatusNotFound, gin.H{"message": "Channel not found"})
		return
	}

	// Remove channel from slice
	deletedChannel := channelsData[channelIndex]
	channelsData = append(channelsData[:channelIndex], channelsData[channelIndex+1:]...)

	fmt.Printf("Successfully deleted channel: %s with ID: %s\n", deletedChannel.Name, deletedChannel.ID)
	c.JSON(http.StatusOK, gin.H{"message": "Channel deleted successfully"})
}

// GetChannelByID handles GET requests to fetch a channel by ID
func GetChannelByID(c *gin.Context) {
	channelID := c.Param("id")

	// Find the channel by ID
	for _, channel := range channelsData {
		if channel.ID == channelID {
			c.JSON(http.StatusOK, channel)
			return
		}
	}

	c.JSON(http.StatusNotFound, gin.H{"message": "Channel not found"})
}

// SetupRoutes sets up the routes for the devices API
func SetupRoutes(router *gin.Engine) {
	// Device routes
	router.GET("/devices", GetDevices)

	// User routes
	router.GET("/users", GetUsers)
	router.POST("/users", CreateUser)
	router.GET("/users/:id", GetUserByID)
	router.PUT("/users/:id", UpdateUser)
	router.DELETE("/users/:id", DeleteUser)

	// Things routes
	router.GET("/things", GetThings)
	router.POST("/things", CreateThing)
	router.GET("/things/:id", GetThingByID)
	router.PUT("/things/:id", UpdateThing)
	router.DELETE("/things/:id", DeleteThing)

	// Channel routes
	router.GET("/channels", GetChannels)
	router.POST("/channels", CreateChannel)
	router.GET("/channels/:id", GetChannelByID)
	router.PUT("/channels/:id", UpdateChannel)
	router.DELETE("/channels/:id", DeleteChannel)

	// Health check
	router.GET("/health", HealthCheck)
}
