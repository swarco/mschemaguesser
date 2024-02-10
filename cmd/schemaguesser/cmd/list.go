package cmd

import (
	"fmt"

	"okieoth/schemaguesser/internal/pkg/mongoHelper"

	"github.com/spf13/cobra"
)

var databaseName string

var collectionName string

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "Discovery the database you are connected to",
	Long:  "Provides information about existing databases, collections and their indexes",
	Run: func(cmd *cobra.Command, args []string) {
		// Logic for the greet command
		fmt.Println("To learn about the possible options type:\n`schemaguesser list --help`")
	},
}

var databasesCmd = &cobra.Command{
	Use:   "databases",
	Short: "Names of databases available in the connected server",
	Long:  "Prints a list of available databases i the connected server",
	Run: func(cmd *cobra.Command, args1 []string) {
		dbs, err := mongoHelper.ListDatabases(mongoHelper.ConStr)
		if err != nil {
			msg := fmt.Sprintf("Error while reading existing databases: \n%v\n", err)
			panic(msg)
		}
		for _, s := range dbs {
			fmt.Println(s)
		}
	},
}

var collectionsCmd = &cobra.Command{
	Use:   "collections",
	Short: "Collections available in the specified database",
	Long:  "Provides information about existing collections in the specified database",
	Run: func(cmd *cobra.Command, args []string) {
		collections, err := mongoHelper.ListCollections(mongoHelper.ConStr, databaseName)
		if err != nil {
			msg := fmt.Sprintf("Error while reading collections for database (%s): \n%v\n", databaseName, err)
			panic(msg)
		}
		for _, s := range collections {
			fmt.Println(s)
		}
	},
}

var indexesCmd = &cobra.Command{
	Use:   "indexes",
	Short: "Indexes to a given collection",
	Long:  "Provides information about indexes of a given collection",
	Run: func(cmd *cobra.Command, args []string) {
		indexes, err := mongoHelper.ListIndexes(mongoHelper.ConStr, databaseName, collectionName)
		if err != nil {
			msg := fmt.Sprintf("Error while reading indexes for collection (%s.%s): \n%v\n", databaseName, collectionName, err)
			panic(msg)
		}
		for _, s := range indexes {
			fmt.Println(s)
		}
	},
}

func init() {
	listCmd.AddCommand(databasesCmd)
	listCmd.AddCommand(collectionsCmd)
	listCmd.AddCommand(indexesCmd)

	collectionsCmd.Flags().StringVar(&databaseName, "database", "", "Database to query existing collections")
	collectionsCmd.MarkFlagRequired("database")

	indexesCmd.Flags().StringVar(&databaseName, "database", "", "Database to query existing collections")
	indexesCmd.MarkFlagRequired("database")

	indexesCmd.Flags().StringVar(&collectionName, "collection", "", "Name of the collection to show the indexes")
	indexesCmd.MarkFlagRequired("collection")
}
