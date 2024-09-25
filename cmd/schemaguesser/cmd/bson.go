package cmd

import (
	"fmt"
	"log"
	"slices"
	"sync"
	"time"

	"go.mongodb.org/mongo-driver/mongo"

	"okieoth/schemaguesser/internal/pkg/mongoHelper"
	"okieoth/schemaguesser/internal/pkg/progressbar"

	"okieoth/schemaguesser/internal/pkg/utils"

	"github.com/spf13/cobra"
)

var bsonCmd = &cobra.Command{
	Use:   "bson",
	Short: "dump raw bson content",
	Long:  "With this command you can dump raw content of one or more mongodb collections",
	Run: func(cmd *cobra.Command, args []string) {
		client, err := mongoHelper.Connect(mongoHelper.ConStr)
		if err != nil {
			msg := fmt.Sprintf("Failed to connect to db: %v", err)
			panic(msg)
		}
		defer mongoHelper.CloseConnection(client)

		if databaseName == "all" {
			bsonForAllDatabases(client, true)
		} else {
			if collectionName == "all" {
				bsonForAllCollections(client, databaseName, true)
			} else {
				bsonForOneCollection(client, databaseName, collectionName, false, true)
			}
		}

	},
}

func bsonForOneCollection(client *mongo.Client, dbName string, collName string, doRecover bool, initProgressBar bool) {
	defer func() {
		if doRecover {
			if r := recover(); r != nil {
				log.Printf("Recovered while handling collection (db: %s, collection: %s): %v", dbName, collName, r)
			}
		}
	}()
	if initProgressBar {
		descr := fmt.Sprintf("BSON export of %s:%s", dbName, collName)
		progressbar.Init(1, descr)
	}
	//bsonRaw, err := mongoHelper.QueryCollectionWithAggregation(client, dbName, collName, int(itemCount))
	bsonRaw, err := mongoHelper.QueryCollection(client, dbName, collName, int(itemCount), useAggregation, mongoV44)

	if err != nil {
		msg := fmt.Sprintf("Error while reading data for collection (%s.%s): \n%v\n", dbName, collName, err)
		panic(msg)
	}
	outputFile, err := utils.CreateOutputFile(outputDir, "bson", dbName, collName)
	if err != nil {
		panic(err)
	}
	defer outputFile.Close()

	startTime := time.Now()
	for _, b := range bsonRaw {
		utils.DumpBsonCollectionData(b, outputFile)
		utils.DumpBsonCollectionData([]byte("\n"), outputFile)
	}
	log.Printf("[%s:%s] BSON exported for collection in %v\n", dbName, collName, time.Since(startTime))
	if initProgressBar {
		progressbar.ProgressOne()
	}
}

func bsonForAllCollections(client *mongo.Client, dbName string, initProgressBar bool) {
	collections := mongoHelper.ReadCollectionsOrPanic(client, dbName)
	var wg sync.WaitGroup
	if initProgressBar {
		progressbar.Init(int64(len(*collections)), "BSON export for all collections")
	}

	for _, coll := range *collections {
		if slices.Contains(blacklist, coll) {
			log.Printf("[%s:%s] skip blacklisted collection\n", dbName, coll)
			continue
		}
		wg.Add(1)
		go func(s string) {
			startTime := time.Now()
			defer func() {
				log.Printf("[%s:%s] BSON export of collection in %v\n", dbName, s, time.Since(startTime))
				wg.Done()
				if initProgressBar {
					progressbar.ProgressOne()
				}
			}()
			bsonForOneCollection(client, dbName, s, true, false)
		}(coll)
	}
	wg.Wait()
}

func bsonForAllDatabases(client *mongo.Client, initProgressBar bool) {
	dbs := mongoHelper.ReadDatabasesOrPanic(client)
	var wg sync.WaitGroup
	if initProgressBar {
		progressbar.Init(int64(len(*dbs)), "BSON export for all databases")
	}
	for _, db := range *dbs {
		if slices.Contains(blacklist, db) {
			log.Printf("[%s] skip blacklisted DB\n", db)
			continue
		}
		wg.Add(1)
		go func(s string) {
			startTime := time.Now()
			defer func() {
				log.Printf("[%s] BSON exported from DB in %v\n", s, time.Since(startTime))
				wg.Done()
				if initProgressBar {
					progressbar.ProgressOne()
				}
			}()
			bsonForAllCollections(client, s, false)
		}(db)
	}
	wg.Wait()
}
