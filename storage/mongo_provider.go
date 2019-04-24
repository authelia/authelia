package storage

import (
	"context"
	"time"

	"github.com/clems4ever/authelia/configuration/schema"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/clems4ever/authelia/models"
)

const (
	prefered2FAMethodCollection        = "prefered_2fa_method"
	identityValidationTokensCollection = "identity_validation_tokens"
	authenticationLogsCollection       = "authentication_logs"
	u2fRegistrationsCollection         = "u2f_devices"
	totpSecretsCollection              = "totp_secrets"
)

// MongoProvider is a storage provider persisting data in a SQLite database.
type MongoProvider struct {
	configuration schema.MongoStorageConfiguration
}

// NewMongoProvider construct a mongo provider.
func NewMongoProvider(configuration schema.MongoStorageConfiguration) *MongoProvider {
	return &MongoProvider{configuration}
}

func (p *MongoProvider) connect() (*mongo.Client, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	clientOptions := options.Client().ApplyURI(p.configuration.URL)

	if p.configuration.Auth.Username != "" && p.configuration.Auth.Password != "" {
		credentials := options.Credential{
			Username: p.configuration.Auth.Username,
			Password: p.configuration.Auth.Password,
		}
		clientOptions.SetAuth(credentials)
	}
	return mongo.Connect(ctx, clientOptions)
}

type prefered2FAMethodDocument struct {
	UserID string `bson:"userId"`
	Method string `bson:"method"`
}

// LoadPrefered2FAMethod load the prefered method for 2FA from sqlite db.
func (p *MongoProvider) LoadPrefered2FAMethod(username string) (string, error) {
	client, err := p.connect()
	if err != nil {
		return "", nil
	}
	defer client.Disconnect(context.Background())

	collection := client.
		Database(p.configuration.Database).
		Collection(prefered2FAMethodCollection)

	res := prefered2FAMethodDocument{}
	err = collection.FindOne(context.Background(),
		bson.M{"userId": username}).
		Decode(&res)

	if err != nil {
		if err == mongo.ErrNoDocuments {
			return "", nil
		}
		return "", err
	}

	return res.Method, nil
}

// SavePrefered2FAMethod save the prefered method for 2FA in sqlite db.
func (p *MongoProvider) SavePrefered2FAMethod(username string, method string) error {
	client, err := p.connect()
	if err != nil {
		return nil
	}
	defer client.Disconnect(context.Background())

	collection := client.
		Database(p.configuration.Database).
		Collection(prefered2FAMethodCollection)

	updateOptions := options.ReplaceOptions{}
	updateOptions.SetUpsert(true)
	_, err = collection.ReplaceOne(context.Background(),
		bson.M{"userId": username},
		bson.M{"userId": username, "method": method},
		&updateOptions)

	if err != nil {
		return err
	}

	return nil
}

// IdentityTokenDocument model for the identiy token documents.
type IdentityTokenDocument struct {
	Token string `bson:"token"`
}

// FindIdentityVerificationToken look for an identity verification token in DB.
func (p *MongoProvider) FindIdentityVerificationToken(token string) (bool, error) {
	client, err := p.connect()
	if err != nil {
		return false, nil
	}
	defer client.Disconnect(context.Background())

	collection := client.
		Database(p.configuration.Database).
		Collection(identityValidationTokensCollection)

	res := IdentityTokenDocument{}
	err = collection.FindOne(context.Background(),
		bson.M{"token": token}).Decode(&res)

	if err != nil {
		if err == mongo.ErrNoDocuments {
			return false, nil
		}
		return false, err
	}

	return true, nil
}

// SaveIdentityVerificationToken save an identity verification token in DB.
func (p *MongoProvider) SaveIdentityVerificationToken(token string) error {
	client, err := p.connect()
	if err != nil {
		return nil
	}
	defer client.Disconnect(context.Background())

	collection := client.
		Database(p.configuration.Database).
		Collection(identityValidationTokensCollection)

	options := options.InsertOneOptions{}
	_, err = collection.InsertOne(context.Background(),
		bson.M{"token": token},
		&options)

	if err != nil {
		return err
	}

	return nil
}

// RemoveIdentityVerificationToken remove an identity verification token from the DB.
func (p *MongoProvider) RemoveIdentityVerificationToken(token string) error {
	client, err := p.connect()
	if err != nil {
		return nil
	}
	defer client.Disconnect(context.Background())

	collection := client.
		Database(p.configuration.Database).
		Collection(identityValidationTokensCollection)

	options := options.DeleteOptions{}
	_, err = collection.DeleteOne(context.Background(),
		bson.M{"token": token},
		&options)

	if err != nil {
		return err
	}
	return nil
}

// TOTPSecretDocument model of document storing TOTP secrets
type TOTPSecretDocument struct {
	UserID string `bson:"userId"`
	Secret string `bson:"secret"`
}

// SaveTOTPSecret save a TOTP secret of a given user.
func (p *MongoProvider) SaveTOTPSecret(username string, secret string) error {
	client, err := p.connect()
	if err != nil {
		return nil
	}
	defer client.Disconnect(context.Background())

	collection := client.
		Database(p.configuration.Database).
		Collection(totpSecretsCollection)

	options := options.ReplaceOptions{}
	options.SetUpsert(true)
	_, err = collection.ReplaceOne(context.Background(),
		bson.M{"userId": username},
		bson.M{"userId": username, "secret": secret},
		&options)

	if err != nil {
		return err
	}

	return nil
}

// LoadTOTPSecret load a TOTP secret given a username.
func (p *MongoProvider) LoadTOTPSecret(username string) (string, error) {
	client, err := p.connect()
	if err != nil {
		return "", nil
	}
	defer client.Disconnect(context.Background())

	collection := client.
		Database(p.configuration.Database).
		Collection(totpSecretsCollection)

	res := TOTPSecretDocument{}
	err = collection.FindOne(context.Background(),
		bson.M{"userId": username}).Decode(&res)

	if err != nil {
		if err == mongo.ErrNoDocuments {
			return "", nil
		}
		return "", err
	}
	return res.Secret, nil
}

// U2FDeviceDocument model of document storing U2F device
type U2FDeviceDocument struct {
	UserID       string `bson:"userId"`
	DeviceHandle []byte `bson:"deviceHandle"`
}

// SaveU2FDeviceHandle save a registered U2F device registration blob.
func (p *MongoProvider) SaveU2FDeviceHandle(username string, deviceBytes []byte) error {
	client, err := p.connect()
	if err != nil {
		return nil
	}
	defer client.Disconnect(context.Background())

	collection := client.
		Database(p.configuration.Database).
		Collection(u2fRegistrationsCollection)

	options := options.ReplaceOptions{}
	options.SetUpsert(true)

	_, err = collection.ReplaceOne(context.Background(),
		bson.M{"userId": username},
		bson.M{"userId": username, "deviceHandle": deviceBytes},
		&options)

	if err != nil {
		return err
	}

	return nil
}

// LoadU2FDeviceHandle load a U2F device registration blob for a given username.
func (p *MongoProvider) LoadU2FDeviceHandle(username string) ([]byte, error) {
	client, err := p.connect()
	if err != nil {
		return nil, nil
	}
	defer client.Disconnect(context.Background())

	collection := client.
		Database(p.configuration.Database).
		Collection(u2fRegistrationsCollection)

	res := U2FDeviceDocument{}
	err = collection.FindOne(context.Background(),
		bson.M{"userId": username}).Decode(&res)

	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, ErrNoU2FDeviceHandle
		}
		return nil, err
	}

	return res.DeviceHandle, nil
}

// AuthenticationLogDocument model of document storing authentication logs
type AuthenticationLogDocument struct {
	UserID  string    `bson:"userId"`
	Time    time.Time `bson:"time"`
	Success bool      `bson:"success"`
}

// AppendAuthenticationLog append a mark to the authentication log.
func (p *MongoProvider) AppendAuthenticationLog(attempt models.AuthenticationAttempt) error {
	client, err := p.connect()
	if err != nil {
		return nil
	}
	defer client.Disconnect(context.Background())

	collection := client.
		Database(p.configuration.Database).
		Collection(authenticationLogsCollection)

	options := options.InsertOneOptions{}
	_, err = collection.InsertOne(context.Background(),
		bson.M{
			"userId":  attempt.Username,
			"time":    attempt.Time,
			"success": attempt.Successful,
		},
		&options)

	if err != nil {
		return err
	}

	return nil
}

// LoadLatestAuthenticationLogs retrieve the latest marks from the authentication log.
func (p *MongoProvider) LoadLatestAuthenticationLogs(username string, fromDate time.Time) ([]models.AuthenticationAttempt, error) {
	client, err := p.connect()
	if err != nil {
		return nil, nil
	}
	defer client.Disconnect(context.Background())

	collection := client.
		Database(p.configuration.Database).
		Collection(authenticationLogsCollection)

	options := options.FindOptions{}
	options.SetSort(bson.M{"time": -1})
	cursor, err := collection.Find(context.Background(),
		bson.M{
			"$and": bson.M{
				"userId": username,
				"time":   bson.M{"$gt": fromDate},
			},
		})

	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, err
	}

	res := []AuthenticationLogDocument{}
	cursor.All(context.Background(), &res)

	attempts := []models.AuthenticationAttempt{}
	for _, r := range res {
		attempt := models.AuthenticationAttempt{
			Username:   r.UserID,
			Time:       r.Time,
			Successful: r.Success,
		}
		attempts = append(attempts, attempt)
	}

	return attempts, nil
}
