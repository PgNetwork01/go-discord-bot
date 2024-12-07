package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"sync"

	"github.com/bwmarrin/discordgo"
	"github.com/joho/godotenv"
)

type Webhook struct {
	ID    string `json:"id"`
	Token string `json:"token"`
}

var (
	webhooks          map[string]Webhook
	manager           *ShardManager
	colors            = map[string]int{"normal": 0x00FF00}
	totalShards       = 1 // Set to the desired number of shards
	version           = "1.0.0"
	repoReleaseAPI    = "https://api.github.com/repos/CorwinDev/Discord-Bot/releases/latest"
	configWebhookPath = "./config/webhooks.json"
	configBotPath     = "./config/bot.go"
)

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Fatal("Error loading .env file:", err)
	}

	// Load webhooks configuration
	file, err := os.Open(configWebhookPath)
	if err != nil {
		log.Fatal("Error loading webhook configuration:", err)
	}
	defer file.Close()
	if err := json.NewDecoder(file).Decode(&webhooks); err != nil {
		log.Fatal("Error parsing webhook configuration:", err)
	}

	// Check for updates
	go checkForUpdates()

	// Start shards
	manager = NewShardManager(totalShards)
	manager.StartShards()
}

// Check if bot is up to date
func checkForUpdates() {
	resp, err := http.Get(repoReleaseAPI)
	if err != nil {
		log.Println("Failed to check if bot is up to date!")
		return
	}
	defer resp.Body.Close()

	var release struct {
		TagName string `json:"tag_name"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		log.Println("Error decoding release data:", err)
		return
	}

	if release.TagName != version {
		fmt.Printf("Your bot is not up to date! Please update to the latest version: %s -> %s\n", version, release.TagName)
	}
}

// ShardManager handles shard operations
type ShardManager struct {
	Shards []*discordgo.Session
	mu     sync.Mutex
}

func NewShardManager(totalShards int) *ShardManager {
	return &ShardManager{
		Shards: make([]*discordgo.Session, totalShards),
	}
}

func (sm *ShardManager) StartShards() {
	for i := 0; i < len(sm.Shards); i++ {
		go sm.launchShard(i)
	}
}

func (sm *ShardManager) launchShard(shardID int) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	// Initialize Discord session
	token := os.Getenv("DISCORD_TOKEN")
	session, err := discordgo.New("Bot " + token)
	if err != nil {
		log.Fatalf("Error creating Discord session for shard %d: %v", shardID, err)
	}
	sm.Shards[shardID] = session

	session.AddHandler(func(s *discordgo.Session, r *discordgo.Ready) {
		fmt.Printf("Shard #%d is ready\n", shardID+1)
	})

	if err := session.Open(); err != nil {
		log.Fatalf("Error opening connection for shard %d: %v", shardID, err)
	}
	defer session.Close()

	fmt.Printf("Shard #%d launched successfully\n", shardID+1)
}
