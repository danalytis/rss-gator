package main

import (
	"context"
	"database/sql"
	"encoding/xml"
	"errors"
	"fmt"
	"html"
	"io"
	"log"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"

	"gator/internal/config"
	"gator/internal/database"
)

type state struct {
	db  *database.Queries
	cfg *config.Config
}

type command struct {
	name string
	args []string
}

type commands struct {
	commandmap map[string]func(*state, command) error
}

type RSSFeed struct {
	Channel struct {
		Title       string    `xml:"title"`
		Link        string    `xml:"link"`
		Description string    `xml:"description"`
		Item        []RSSItem `xml:"item"`
	} `xml:"channel"`
}

type RSSItem struct {
	Title       string `xml:"title"`
	Link        string `xml:"link"`
	Description string `xml:"description"`
	PubDate     string `xml:"pubDate"`
}

// Executes a registered command
func (c *commands) run(s *state, cmd command) error {
	command, ok := c.commandmap[cmd.name]
	if !ok {
		return fmt.Errorf("command does not exist")
	}
	return command(s, cmd)
}

// Adds a new command to the registry
func (c *commands) register(name string, f func(*state, command) error) {
	c.commandmap[name] = f
}

// Shows all feeds with their creators
func handlerListFeeds(s *state, cmd command) error {
	feeds, err := s.db.GetFeedsWithUserName(context.Background())
	if err != nil {
		return err
	}

	if len(feeds) == 0 {
		fmt.Println("No feeds found.")
		return nil
	}

	fmt.Println("List of Feeds:")
	for _, feed := range feeds {
		fmt.Printf("Feed Name: %s\nFeed URL: %s\nCreated by: %s\n\n", feed.FeedName, feed.FeedUrl, feed.UserName)
	}
	return nil
}

// Retrieves and parses RSS content from feed URL with HTML unescaping
// Sets a custom "gator" User-Agent header for the HTTP request.
func fetchFeed(ctx context.Context, feedURL string) (*RSSFeed, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", feedURL, nil)
	if err != nil {
		return &RSSFeed{}, err
	}
	req.Header.Set("User-Agent", "gator")
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return &RSSFeed{}, err
	}
	defer resp.Body.Close()
	res, err := io.ReadAll(resp.Body)
	if err != nil {
		return &RSSFeed{}, err
	}
	result := RSSFeed{}
	err = xml.Unmarshal(res, &result)
	if err != nil {
		return &RSSFeed{}, err
	}
	result.Channel.Title = html.UnescapeString(result.Channel.Title)
	result.Channel.Description = html.UnescapeString(result.Channel.Description)
	for i := range result.Channel.Item {
		item := &result.Channel.Item[i]
		item.Title = html.UnescapeString(item.Title)
		item.Description = html.UnescapeString(item.Description)
	}
	return &result, nil
}

// Validates and sets current user
func handlerLogin(s *state, cmd command) error {
	if len(cmd.args) == 0 {
		return fmt.Errorf("login command expects one argument\n\tUsage: login <username>")
	}
	_, err := s.db.GetUser(context.Background(), cmd.args[0])
	if err != nil {
		if err == sql.ErrNoRows {
			return fmt.Errorf("user '%s' not found", cmd.args[0])
		}
		return err
	}
	s.cfg.Current_user_name = cmd.args[0]
	fmt.Printf("User set to '%s'\n", cmd.args[0])
	if err := config.SetUser(*s.cfg, cmd.args[0]); err != nil {
		return err
	}
	return nil
}

// Creates a new user account
func handlerRegister(s *state, cmd command) error {
	if len(cmd.args) == 0 {
		return fmt.Errorf("register command expects one argument\n\tUsage: register <username>")
	}

	params := database.CreateUserParams{
		ID:        uuid.New(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Name:      cmd.args[0],
	}

	user, err := s.db.CreateUser(context.Background(), params)
	if err != nil {
		if strings.Contains(err.Error(), "duplicate") || strings.Contains(err.Error(), "unique") {
			return fmt.Errorf("user '%s' already exists\n", cmd.args[0])
		}
		return err
	}
	s.cfg.Current_user_name = cmd.args[0]
	if err := config.SetUser(*s.cfg, cmd.args[0]); err != nil {
		return err
	}

	fmt.Printf("successfully registered user: %s\n", user.Name)
	return nil
}

// Clears all users from the database
func handlerReset(s *state, cmd command) error {
	err := s.db.Reset(context.Background())
	if err != nil {
		return err
	}
	fmt.Println("database reset")
	return nil
}

// Lists all registered users
func handlerGetUsers(s *state, cmd command) error {
	users, err := s.db.GetUsers(context.Background())
	if err != nil {
		return err
	}
	for _, user := range users {
		if s.cfg.Current_user_name == user.Name {
			fmt.Printf("* %s (current)\n", user.Name)
		} else {
			fmt.Printf("* %s\n", user.Name)
		}
	}
	return nil
}

// Continuously fetches feeds at specified duration
func handlerAggregate(s *state, cmd command) error {
	if len(cmd.args) < 1 {
		return errors.New("missing required argument: time_between_reqs")
	}
	time_between_reqs := cmd.args[0]
	timeDuration, err := time.ParseDuration(time_between_reqs)
	if err != nil {
		log.Fatal("error parsing time duration")
	}
	fmt.Printf("Collecting feeds every %s\n", time_between_reqs)
	ticker := time.NewTicker(timeDuration)
	for {
		fmt.Println("collecting feeds..")
		err := scrapeFeeds(s)
		if err != nil {
			log.Fatal("error scraping feeds")
		}
		<-ticker.C
	}
}

// Create a new feed and follow it
func handlerAddFeed(s *state, cmd command, user database.User) error {
	if len(cmd.args) < 2 {
		return fmt.Errorf("addfeed requires two arguments\n\tUsage: addfeed <name> <url>")
	}
	name := cmd.args[0]
	url := cmd.args[1]

	feedParams := database.CreateFeedParams{
		ID:        uuid.New(),
		CreatedAt: sql.NullTime{Time: time.Now().UTC(), Valid: true},
		UpdatedAt: sql.NullTime{Time: time.Now().UTC(), Valid: true},
		Name:      name,
		Url:       url,
		UserID:    user.ID,
	}

	feed, err := s.db.GetFeedByURL(context.Background(), url)
	var feedID uuid.UUID
	if err == sql.ErrNoRows {
		_, err = s.db.CreateFeed(context.Background(), feedParams)
		if err != nil {
			return err
		}
		feedID = feedParams.ID
	} else if err != nil {
		return err
	} else {
		feedID = feed.ID
	}

	params := database.CreateFeedFollowParams{
		UserID: user.ID,
		FeedID: feedID,
	}

	_, err = s.db.CreateFeedFollow(context.Background(), params)
	return nil
}

// Follows an existing feed by URL
func handlerFollowFeed(s *state, cmd command, user database.User) error {
	if len(cmd.args) != 1 {
		return fmt.Errorf("follow takes one argument\n\tfollow <url>")
	}
	url := cmd.args[0]

	feed, err := s.db.GetFeedByURL(context.Background(), url)
	if err != nil {
		return err
	}
	params := database.CreateFeedFollowParams{
		UserID: user.ID,
		FeedID: feed.ID,
	}
	follow, err := s.db.CreateFeedFollow(context.Background(), params)
	if err != nil {
		return err
	}

	fmt.Printf("You (%s) are now following %s\n", follow.UserName, follow.FeedName)
	return nil
}

// Lists all followed feeds by the current user
func handlerFollowing(s *state, cmd command, user database.User) error {
	if len(cmd.args) > 0 {
		return fmt.Errorf("following takes 0 arguments")
	}

	follows, err := s.db.GetFeedFollowForUser(context.Background(), user.ID)
	if err != nil {
		return err
	}

	fmt.Printf("Feeds followed by user %s\n", user.Name)
	for _, feed := range follows {
		fmt.Printf("- %s\n", feed.FeedName)
	}
	return nil
}

// Wraps handlers to require user authentication
func middlewareLoggedIn(handler func(s *state, cmd command, user database.User) error) func(*state, command) error {
	isLogged := func(s *state, cmd command) error {
		user, err := s.db.GetUser(context.Background(), s.cfg.Current_user_name)
		if err != nil {
			if err == sql.ErrNoRows {
				return fmt.Errorf("user '%s' not found", s.cfg.Current_user_name)
			}
			return err
		}
		err = handler(s, cmd, user)
		if err != nil {
			return err
		}
		return nil
	}
	return isLogged
}

// Unfollows a feed by URL
func handlerUnfollow(s *state, cmd command, user database.User) error {
	if len(cmd.args) != 1 {
		return fmt.Errorf("unfollow takes one argument\n\tunfollow <url>\n")
	}
	url := cmd.args[0]
	s.db.DeleteFeedFollow(context.Background(), database.DeleteFeedFollowParams{
		Url:    url,
		UserID: user.ID,
	})

	fmt.Printf("%s - Feed unfollowed\n", url)
	return nil
}

// Fetches and stores posts from the next feed
func scrapeFeeds(s *state) error {
	feedToFetch, err := s.db.GenNextFeedToFetch(context.Background())
	if err != nil {
		return err
	}
	feedContent, err := fetchFeed(context.Background(), feedToFetch.Url)
	if err != nil {
		return err
	}
	err = s.db.MarkFeedFetched(context.Background(), feedToFetch.ID)
	if err != nil {
		return err
	}

	for _, content := range feedContent.Channel.Item {
		parsedTime, err := time.Parse(time.RFC1123Z, content.PubDate)
		if err != nil {
			log.Printf("Failed to parse date '%s':%v", content.PubDate, err)
			continue
		}
		err = s.db.CreatePost(context.Background(), database.CreatePostParams{
			Title: content.Title,
			Url:   content.Link,
			Description: sql.NullString{
				String: content.Description,
				Valid:  true,
			},
			PublishedAt: sql.NullTime{
				Time:  parsedTime,
				Valid: true,
			},
			FeedID: feedToFetch.ID,
		})
		if err != nil {
			var pqErr *pq.Error
			if errors.As(err, &pqErr) && pqErr.Code == "23505" {
				continue
			}
			log.Printf("Failed to create post '%s': %v", content.Title, err)
			continue
		}
	}
	return nil
}

// parsing helper
func extractCommentURL(htmlContent string) string {
	if htmlContent == "" {
		return ""
	}

	// Simple regex to extract URL from <a href="...">Comments</a> pattern
	linkRe := regexp.MustCompile(`<a[^>]*href="([^"]*)"[^>]*>(?:Comments?|Discussion)</a>`)
	match := linkRe.FindStringSubmatch(htmlContent)

	if len(match) >= 2 {
		return match[1]
	}

	return ""
}

// Helper function to clean HTML from descriptions
func stripHTMLTags(input string) string {
	re := regexp.MustCompile(`<[^>]*>`)
	stripped := re.ReplaceAllString(input, "")

	stripped = regexp.MustCompile(`\s+`).ReplaceAllString(stripped, " ")

	return strings.TrimSpace(stripped)
}

// ANSI color codes
const (
	ColorReset  = "\033[0m"
	ColorRed    = "\033[31m"
	ColorGreen  = "\033[32m"
	ColorYellow = "\033[33m"
	ColorBlue   = "\033[34m"
	ColorPurple = "\033[35m"
	ColorCyan   = "\033[36m"
	ColorWhite  = "\033[37m"
	ColorBold   = "\033[1m"
)

// Displays 'n' recent posts from followed feeds (default n=2)
func handlerBrowse(s *state, cmd command, user database.User) error {
	parsedLimit := 2

	if len(cmd.args) > 1 {
		return fmt.Errorf("browse takes one optional argument:\n\tbrowse (limit)\n")
	} else if len(cmd.args) == 1 {
		var err error
		parsedLimit, err = strconv.Atoi(cmd.args[0])
		if err != nil {
			return fmt.Errorf("invalid limit: %v", err)
		}
	}

	posts, err := s.db.GetPostsForUser(context.Background(), database.GetPostsForUserParams{
		UserID: user.ID,
		Limit:  int32(parsedLimit),
	})
	if err != nil {
		return err
	}

	if len(posts) == 0 {
		fmt.Printf("%s%s📰 No posts found. Try following some feeds first!%s\n", ColorYellow, ColorBold, ColorReset)
		return nil
	}

	fmt.Printf("\n%s%s📰 Latest Posts for %s%s\n", ColorPurple, ColorBold, user.Name, ColorReset)
	fmt.Printf("%s═══════════════════════════════════════════════════════════════%s\n", ColorPurple, ColorReset)

	for i, post := range posts {
		if i > 0 {
			fmt.Println()
		}

		// Title with cyan and bold
		fmt.Printf("%s%s📄 %s%s\n", ColorCyan, ColorBold, post.Title, ColorReset)

		// Date with green
		if post.PublishedAt.Valid {
			fmt.Printf("%s🗓️  %s%s\n", ColorGreen, post.PublishedAt.Time.Format("Jan 2, 2006 at 3:04 PM"), ColorReset)
		} else {
			fmt.Printf("%s🗓️  Published date unknown%s\n", ColorGreen, ColorReset)
		}

		// Description
		if post.Description.Valid {
			commentURL := extractCommentURL(post.Description.String)
			if commentURL != "" {
				fmt.Printf("%s💬 Comments: %s%s\n", ColorBlue, commentURL, ColorReset)
			} else {
				desc := stripHTMLTags(post.Description.String)
				if len(desc) > 150 {
					desc = desc[:147] + "..."
				}
				if strings.TrimSpace(desc) != "" {
					fmt.Printf("%s📝  %s%s\n", ColorWhite, desc, ColorReset)
				}
			}
		}

		fmt.Printf("%s🔗 %s%s\n", ColorBlue, post.Url, ColorReset)
		fmt.Printf("%s───────────────────────────────────────────────────────────────%s\n", ColorPurple, ColorReset)
	}

	fmt.Printf("\n%s%sShowing %d posts%s\n\n", ColorPurple, ColorBold, len(posts), ColorReset)
	return nil
}

func main() {
	cfg := config.Read()

	dbURL := cfg.Db_url

	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatal("could not connect to database")
	}

	dbQueries := database.New(db)
	st := &state{
		cfg: &cfg,
		db:  dbQueries,
	}

	cmds := &commands{commandmap: make(map[string]func(*state, command) error)}

	cmds.register("login", handlerLogin)
	cmds.register("register", handlerRegister)
	cmds.register("reset", handlerReset)
	cmds.register("users", handlerGetUsers)
	cmds.register("agg", handlerAggregate)
	cmds.register("addfeed", middlewareLoggedIn(handlerAddFeed))
	cmds.register("feeds", handlerListFeeds)
	cmds.register("follow", middlewareLoggedIn(handlerFollowFeed))
	cmds.register("following", middlewareLoggedIn(handlerFollowing))
	cmds.register("unfollow", middlewareLoggedIn(handlerUnfollow))
	cmds.register("browse", middlewareLoggedIn(handlerBrowse))

	args := os.Args
	if len(args) < 2 {
		log.Fatal("not enough arguments")
	}

	cmd := command{
		name: args[1],
		args: args[2:],
	}

	if err := cmds.run(st, cmd); err != nil {
		log.Fatal(err)
	}

}
