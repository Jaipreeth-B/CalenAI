package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/glebarez/sqlite"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// --- Models ---

type Task struct {
	ID            uint           `gorm:"primaryKey" json:"id"`
	Title         string         `json:"title"`
	Desc          string         `json:"desc"`
	IsCompleted   bool           `json:"is_completed"`
	StartDateTime *time.Time     `json:"start_date_time"`
	EndDateTime   *time.Time     `json:"end_date_time"`
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`
	DeletedAt     gorm.DeletedAt `gorm:"index" json:"-"`
}

type Session struct {
	ID        string        `gorm:"primaryKey" json:"id"`
	Title     string        `json:"title"`
	Messages  []ChatMessage `json:"messages"`
	CreatedAt time.Time     `json:"created_at"`
}

type ChatMessage struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	SessionID string    `gorm:"index" json:"session_id"`
	Role      string    `json:"role"` // 'user', 'assistant', 'system', 'tool'
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"created_at"`
}

var db *gorm.DB

func initDB() {
	var err error
	db, err = gorm.Open(sqlite.Open("calenai.db"), &gorm.Config{})
	if err != nil {
		log.Fatal("failed to connect database")
	}

	db.AutoMigrate(&Task{}, &Session{}, &ChatMessage{})
}

func main() {
	initDB()

	r := gin.Default()

	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:5173", "http://localhost:4173", "http://127.0.0.1:5173", "http://127.0.0.1:4173"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	api := r.Group("/api")
	{
		// Tasks
		api.GET("/tasks", getTasks)
		api.POST("/tasks", createTask)
		api.PUT("/tasks/:id", updateTask)
		api.DELETE("/tasks/:id", deleteTask)

		// Sessions (Chat History)
		api.GET("/sessions", getSessions)
		api.POST("/sessions", createSession)
		api.GET("/sessions/:id", getSessionMessages)
		api.DELETE("/sessions/:id", deleteSession)

		// Chat
		api.POST("/chat/:session_id", handleChat)

		// Context Helper
		api.GET("/context", getAIContext)
	}

	fmt.Println("Server running on :9090")
	r.Run(":9090")
}

// --- Handlers: Tasks ---

func getTasks(c *gin.Context) {
	var tasks []Task
	db.Find(&tasks)
	c.JSON(http.StatusOK, tasks)
}

func createTask(c *gin.Context) {
	var task Task
	if err := c.ShouldBindJSON(&task); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	db.Create(&task)
	c.JSON(http.StatusCreated, task)
}

func updateTask(c *gin.Context) {
	id := c.Param("id")
	var task Task
	if err := db.First(&task, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Task not found"})
		return
	}
	if err := c.ShouldBindJSON(&task); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	db.Save(&task)
	c.JSON(http.StatusOK, task)
}

func deleteTask(c *gin.Context) {
	id := c.Param("id")
	db.Delete(&Task{}, id)
	c.JSON(http.StatusOK, gin.H{"message": "Task deleted"})
}

// --- Handlers: Sessions ---

func getSessions(c *gin.Context) {
	var sessions []Session
	db.Order("created_at desc").Find(&sessions)
	c.JSON(http.StatusOK, sessions)
}

func createSession(c *gin.Context) {
	session := Session{
		ID:    uuid.New().String(),
		Title: "New Chat " + time.Now().Format("Jan 02, 15:04"),
	}
	db.Create(&session)
	c.JSON(http.StatusCreated, session)
}

func getSessionMessages(c *gin.Context) {
	sessionID := c.Param("id")
	var messages []ChatMessage
	db.Where("session_id = ?", sessionID).Order("created_at asc").Find(&messages)
	c.JSON(http.StatusOK, messages)
}

func deleteSession(c *gin.Context) {
	sessionID := c.Param("id")
	db.Where("session_id = ?", sessionID).Delete(&ChatMessage{})
	db.Delete(&Session{}, "id = ?", sessionID)
	c.JSON(http.StatusOK, gin.H{"message": "Session deleted"})
}

// --- AI Chat Logic (Multi-Think) ---

type ChatRequest struct {
	Message string `json:"message"`
}

type OllamaMessage struct {
	Role      string           `json:"role"`
	Content   string           `json:"content"`
	ToolCalls []OllamaToolCall `json:"tool_calls,omitempty"`
}

type OllamaToolCall struct {
	Function struct {
		Name      string                 `json:"name"`
		Arguments map[string]interface{} `json:"arguments"`
	} `json:"function"`
}

func handleChat(c *gin.Context) {
	sessionID := c.Param("session_id")
	var req ChatRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Persist User Message
	db.Create(&ChatMessage{SessionID: sessionID, Role: "user", Content: req.Message})

	// 1. Context Preparation
	var history []ChatMessage
	db.Where("session_id = ?", sessionID).Order("created_at asc").Find(&history)

	systemPrompt := fmt.Sprintf(
		"You are CalenAI, a sophisticated minimalist assistant. "+
			"Current Date/Time: %s (%s). "+
			"You MUST resolve relative dates (e.g., 'next Tuesday', 'tomorrow', 'in two weeks') YOURSELF using the Current Date provided above. DO NOT ask the user to provide dates you can calculate. "+
			"For example, if today is Tuesday, Jun 02, 'next Tuesday' is Jun 09. "+
			"If a user mentions an event without a date (e.g., 'my dentist appointment'), use 'get_tasks' to see if it exists in their schedule. "+
			"NEVER assume a time (hour/minute) if missing, but ALWAYS calculate the date yourself. "+
			"For recurring tasks, call 'add_task' once for each specific date. "+
			"Keep responses brief and focus on the action.",
		time.Now().Format("Monday, January 02, 2006"), time.Now().Format("15:04"),
	)

	var messages []OllamaMessage
	messages = append(messages, OllamaMessage{Role: "system", Content: systemPrompt})
	for _, m := range history {
		messages = append(messages, OllamaMessage{Role: m.Role, Content: m.Content})
	}

	// 2. Reasoning Loop (Multi-Think)
	response, err := callOllamaWithTools(messages)
	if err != nil {
		log.Printf("Ollama error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "AI failure"})
		return
	}

	// Persist Assistant Response
	db.Create(&ChatMessage{SessionID: sessionID, Role: "assistant", Content: response})

	c.JSON(http.StatusOK, gin.H{"response": response})
}

func callOllamaWithTools(messages []OllamaMessage) (string, error) {
	ollamaURL := "http://localhost:11434/api/chat"

	tools := []map[string]interface{}{
		{
			"type": "function",
			"function": map[string]interface{}{
				"name":        "add_task",
				"description": "Add a SINGLE task or event. For recurring tasks, call this tool multiple times. Example format: '2026-06-02T17:00:00'.",
				"parameters": map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"title":           map[string]interface{}{"type": "string"},
						"desc":            map[string]interface{}{"type": "string"},
						"start_date_time": map[string]interface{}{"type": "string", "description": "ISO8601 format (e.g., 2026-06-02T17:00:00)."},
						"end_date_time":   map[string]interface{}{"type": "string", "description": "ISO8601 format (e.g., 2026-06-02T18:00:00)."},
					},
					"required": []string{"title", "start_date_time"},
				},
			},
		},
		{
			"type": "function",
			"function": map[string]interface{}{
				"name":        "get_tasks",
				"description": "Retrieve all tasks and events to see current schedule and task IDs.",
				"parameters":  map[string]interface{}{"type": "object", "properties": map[string]interface{}{}},
			},
		},
		{
			"type": "function",
			"function": map[string]interface{}{
				"name":        "update_task",
				"description": "Update an existing task's properties.",
				"parameters": map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"id":              map[string]interface{}{"type": "integer"},
						"title":           map[string]interface{}{"type": "string"},
						"desc":            map[string]interface{}{"type": "string"},
						"is_completed":    map[string]interface{}{"type": "boolean"},
						"start_date_time": map[string]interface{}{"type": "string"},
					},
					"required": []string{"id"},
				},
			},
		},
		{
			"type": "function",
			"function": map[string]interface{}{
				"name":        "delete_task",
				"description": "Delete a task by its ID.",
				"parameters": map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"id": map[string]interface{}{"type": "integer"},
					},
					"required": []string{"id"},
				},
			},
		},
	}

	payload := map[string]interface{}{
		"model":    "qwen3.5:9b",
		"messages": messages,
		"tools":    tools,
		"stream":   false,
	}

	jsonData, _ := json.Marshal(payload)
	resp, err := http.Post(ollamaURL, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	var ollamaResp struct {
		Message OllamaMessage `json:"message"`
	}
	json.Unmarshal(body, &ollamaResp)

	if len(ollamaResp.Message.ToolCalls) > 0 {
		for _, tc := range ollamaResp.Message.ToolCalls {
			log.Printf("AI thinking... Executing tool: %s", tc.Function.Name)
			res := executeTool(tc)
			messages = append(messages, ollamaResp.Message)
			messages = append(messages, OllamaMessage{Role: "tool", Content: res})
		}
		return callOllamaWithTools(messages)
	}

	return ollamaResp.Message.Content, nil
}

func parseFlexibleDate(s string) (time.Time, error) {
	formats := []string{
		time.RFC3339,
		"2006-01-02T15:04:05Z07:00",
		"2006-01-02T15:04:05",
		"2006-01-02 15:04:05",
		"2006-01-02",
	}
	for _, f := range formats {
		if t, err := time.Parse(f, s); err == nil {
			// Force the parsed clock time into the Local timezone
			// This prevents AI-generated 'Z' suffixes from shifting the time by +5:30 etc.
			return time.Date(t.Year(), t.Month(), t.Day(), t.Hour(), t.Minute(), t.Second(), t.Nanosecond(), time.Local), nil
		}
	}
	return time.Time{}, fmt.Errorf("invalid date format")
}

func executeTool(tc OllamaToolCall) string {
	switch tc.Function.Name {
	case "add_task":
		title, _ := tc.Function.Arguments["title"].(string)
		desc, _ := tc.Function.Arguments["desc"].(string)
		startStr, _ := tc.Function.Arguments["start_date_time"].(string)

		if startStr == "" {
			return "Error: start_date_time is required."
		}

		tStart, err := parseFlexibleDate(startStr)
		if err != nil {
			return "Error: Invalid start date format. Please use ISO8601 (e.g., 2026-06-02T17:00:00)."
		}

		task := Task{Title: title, Desc: desc, StartDateTime: &tStart}

		// Handle End Date Time
		if endStr, ok := tc.Function.Arguments["end_date_time"].(string); ok && endStr != "" {
			tEnd, err := parseFlexibleDate(endStr)
			if err == nil {
				task.EndDateTime = &tEnd
			}
		}

		db.Create(&task)
		return fmt.Sprintf("Task '%s' added successfully for %s.", title, tStart.Format("Jan 02, 15:04"))
	case "get_tasks":
		var tasks []Task
		db.Find(&tasks)
		data, _ := json.Marshal(tasks)
		return string(data)
	case "update_task":
		idFloat, _ := tc.Function.Arguments["id"].(float64)
		id := uint(idFloat)
		var task Task
		if err := db.First(&task, id).Error; err != nil {
			return "Error: Task not found."
		}

		if title, ok := tc.Function.Arguments["title"].(string); ok {
			task.Title = title
		}
		if desc, ok := tc.Function.Arguments["desc"].(string); ok {
			task.Desc = desc
		}
		if completed, ok := tc.Function.Arguments["is_completed"].(bool); ok {
			task.IsCompleted = completed
		}
		if startStr, ok := tc.Function.Arguments["start_date_time"].(string); ok {
			t, _ := time.Parse(time.RFC3339, startStr)
			task.StartDateTime = &t
		}

		db.Save(&task)
		return "Task updated successfully."
	case "delete_task":
		idFloat, _ := tc.Function.Arguments["id"].(float64)
		id := uint(idFloat)
		var task Task
		if err := db.First(&task, id).Error; err != nil {
			log.Printf("AI tried to delete non-existent task ID: %d", id)
			return fmt.Sprintf("Error: Task with ID %d not found.", id)
		}
		db.Delete(&task)
		log.Printf("AI successfully deleted task: %s (ID: %d)", task.Title, id)
		return fmt.Sprintf("Successfully deleted task: %s", task.Title)
	default:
		return "Error: Unknown tool."
	}
}

func getAIContext(c *gin.Context) {
	var tasks []Task
	db.Find(&tasks)
	c.JSON(http.StatusOK, gin.H{
		"current_time": time.Now(),
		"day":          time.Now().Format("Monday"),
		"task_count":   len(tasks),
	})
}
