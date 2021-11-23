package p

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"database/sql"

	"github.com/mmcdole/gofeed"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
	"github.com/uptrace/bun/driver/pgdriver"
)

const (
	host                   = "localhost"
	port                   = 5432
	dbUser                 = "postgres"
	dbPwd                  = "password"
	dbName                 = "postgres"
	instanceConnectionName = "instance:connection:name"
)

// PubSubMessage is the payload of a Pub/Sub event. Please refer to the docs for
// additional information regarding Pub/Sub events.
type PubSubMessage struct {
	Data []byte `json:"data"`
}

// HelloPubSub consumes a Pub/Sub message.
func HelloPubSub(ctx context.Context, m PubSubMessage) error {

	log.Println(string(m.Data))

	socketDir, isSet := os.LookupEnv("DB_SOCKET_DIR")
	if !isSet {
		socketDir = "/cloudsql"
	}
	psqlInfo := fmt.Sprintf("unix://%s:%s@%s%s/%s/.s.PGSQL.5432?sslmode=disable", dbUser, dbPwd, dbName, socketDir, instanceConnectionName)
	pgdb := sql.OpenDB(pgdriver.NewConnector(pgdriver.WithDSN(psqlInfo)))
	db := bun.NewDB(pgdb, pgdialect.New())

	feeds := make([]Feed, 0)
	if err := db.NewSelect().
		Model(&feeds).
		Scan(ctx); err != nil {
		panic(err)
	}

	for _, dbFeed := range feeds {
		log.Printf("Updating feed %s", dbFeed.Title)
		fp := gofeed.NewParser()
		feed, err := fp.ParseURL(dbFeed.Link)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Could not read feed [%s]: %v\n", dbFeed.Title, err)
			dbFeed.Valid = false
			_, err = db.NewUpdate().
				Model(&dbFeed).
				Column("valid").
				Where("id = ?", dbFeed.ID).
				Exec(ctx)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Could not update feed [%s]: %v\n", dbFeed.Title, err)
			}
		} else {
			dbFeed.Valid = true
			dbFeed.Updated = time.Now()
			_, err = db.NewUpdate().
				Model(&dbFeed).
				Column("valid").
				Column("updated").
				Where("id = ?", dbFeed.ID).
				Exec(ctx)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Could not update feed [%s]: %v\n", dbFeed.Title, err)
			}
			for _, item := range feed.Items {
				exists, err := db.NewSelect().
					Model((*Entry)(nil)).
					Where("feed_id = ?", dbFeed.ID).
					Where("link = ?", item.Link).
					Exists(ctx)
				if err != nil {
					fmt.Fprintf(os.Stderr, "Could not check entry [%s]: %v\n", item.Link, err)
				}
				if !exists {
					dbEntry := new(Entry)
					dbEntry.FeedID = dbFeed.ID
					dbEntry.Title = item.Title
					dbEntry.Link = item.Link
					dbEntry.ToRead = dbFeed.MarkNewToRead
					if item.Author != nil {
						dbEntry.Author = item.Author.Name
					}
					published := item.PublishedParsed
					if published != nil {
						dbEntry.Published = *published
					}
					updated := item.UpdatedParsed
					if updated != nil {
						dbEntry.Updated = *updated
					} else {
						dbEntry.Updated = time.Now()
					}
					dbEntry.Content = item.Content
					dbEntry.Description = item.Description
					_, err = db.NewInsert().
						Model(dbEntry).
						Exec(ctx)
					if err != nil {
						fmt.Fprintf(os.Stderr, "Could not insert entry [%s]: %v\n", item.Link, err)
					}
				}
			}
		}
	}
	return nil
}

type Feed struct {
	bun.BaseModel `bun:"rss_feed,alias:f"`
	ID            int64
	Link          string
	BaseLink      string
	Title         string
	Description   string
	Updated       time.Time
	ToUpdate      bool
	MarkNewToRead bool
	CategoryID    int64
	Valid         bool
	Viewframe     bool
	CatOrder      int
}

type Entry struct {
	bun.BaseModel `bun:"rss_entry,alias:e"`
	ID            int64
	Published     time.Time
	Updated       time.Time
	Link          string
	Title         string
	Description   string
	Author        string
	Read          bool
	Content       string
	FeedID        int64
	Favourite     bool
	ToRead        bool
}
