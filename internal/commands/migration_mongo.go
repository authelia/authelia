package commands

import (
	"context"
	"log"
	"time"

	"github.com/spf13/cobra"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/authelia/authelia/internal/models"
	"github.com/authelia/authelia/internal/storage"
)

var mongoURL string
var mongoDatabase string

// MigrateMongoCmd migration command
var MigrateMongoCmd = &cobra.Command{
	Use:   "mongo",
	Short: "Migrate data from v3 mongo database into database configured in v4 configuration file",
	Run:   migrateMongo,
}

func init() {
	MigrateMongoCmd.PersistentFlags().StringVar(&mongoURL, "url", "", "The address to the mongo server")
	MigrateMongoCmd.MarkPersistentFlagRequired("url")

	MigrateMongoCmd.PersistentFlags().StringVar(&mongoDatabase, "database", "", "The mongo database")
	MigrateMongoCmd.MarkPersistentFlagRequired("database")

	MigrateMongoCmd.PersistentFlags().StringVarP(&configurationPath, "config", "c", "", "The configuration file of Authelia v4")
	MigrateMongoCmd.MarkPersistentFlagRequired("config")
}

func migrateMongo(cmd *cobra.Command, args []string) {
	dbProvider := createDBProvider(configurationPath)
	client, err := mongo.NewClient(options.Client().ApplyURI(mongoURL))

	if err != nil {
		log.Fatal(err)
	}

	err = client.Connect(context.Background())

	if err != nil {
		log.Fatal(err)
	}

	db := client.Database(mongoDatabase)

	migrateMongoU2FDevices(db, dbProvider)
	migrateMongoTOTPDevices(db, dbProvider)
	migrateMongoPreferences(db, dbProvider)

	log.Println("Migration done!")
}

func migrateMongoU2FDevices(db *mongo.Database, dbProvider storage.Provider) {
	u2fCollection := db.Collection("u2f_registrations")

	cur, err := u2fCollection.Find(context.Background(), bson.D{})
	if err != nil {
		log.Fatal(err)
	}
	defer cur.Close(context.Background())

	for cur.Next(context.Background()) {
		var result U2FDeviceHandleV3
		err := cur.Decode(&result)
		if err != nil {
			log.Fatal(err)
		}

		kH, err := decodeWebsafeBase64(result.Registration.KeyHandle)

		if err != nil {
			log.Fatal(err)
		}

		pK, err := decodeWebsafeBase64(result.Registration.PublicKey)

		if err != nil {
			log.Fatal(err)
		}

		err = dbProvider.SaveU2FDeviceHandle(result.UserID, kH, pK)

		if err != nil {
			log.Fatal(err)
		}
	}
	if err := cur.Err(); err != nil {
		log.Fatal(err)
	}
}

func migrateMongoTOTPDevices(db *mongo.Database, dbProvider storage.Provider) {
	u2fCollection := db.Collection("totp_secrets")

	cur, err := u2fCollection.Find(context.Background(), bson.D{})
	if err != nil {
		log.Fatal(err)
	}
	defer cur.Close(context.Background())

	for cur.Next(context.Background()) {
		var result TOTPSecretsV3
		err := cur.Decode(&result)
		if err != nil {
			log.Fatal(err)
		}

		err = dbProvider.SaveTOTPSecret(result.UserID, result.Secret.Base32)

		if err != nil {
			log.Fatal(err)
		}
	}
	if err := cur.Err(); err != nil {
		log.Fatal(err)
	}
}

func migrateMongoPreferences(db *mongo.Database, dbProvider storage.Provider) {
	u2fCollection := db.Collection("prefered_2fa_method")

	cur, err := u2fCollection.Find(context.Background(), bson.D{})
	if err != nil {
		log.Fatal(err)
	}
	defer cur.Close(context.Background())

	for cur.Next(context.Background()) {
		var result PreferencesV3
		err := cur.Decode(&result)
		if err != nil {
			log.Fatal(err)
		}

		err = dbProvider.SavePreferred2FAMethod(result.UserID, result.Method)

		if err != nil {
			log.Fatal(err)
		}
	}
	if err := cur.Err(); err != nil {
		log.Fatal(err)
	}
}

func migrateMongoAuthenticationTraces(db *mongo.Database, dbProvider storage.Provider) {
	u2fCollection := db.Collection("authentication_traces")

	cur, err := u2fCollection.Find(context.Background(), bson.D{})
	if err != nil {
		log.Fatal(err)
	}
	defer cur.Close(context.Background())

	for cur.Next(context.Background()) {
		var result AuthenticationTraceV3
		err := cur.Decode(&result)
		if err != nil {
			log.Fatal(err)
		}

		attempt := models.AuthenticationAttempt{
			Username:   result.UserID,
			Successful: result.Successful,
			Time:       time.Unix(result.Date.Date/1000.0, 0),
		}

		err = dbProvider.AppendAuthenticationLog(attempt)

		if err != nil {
			log.Fatal(err)
		}
	}
	if err := cur.Err(); err != nil {
		log.Fatal(err)
	}
}
