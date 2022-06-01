package storer

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"

	"github.com/digitalocean-apps/sample-with-database/pkg/model"
	"github.com/pkg/errors"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

const (
	defaultDatabase   = "notes"
	defaultCollection = "notes"
)

// Mongo is a mongo storer implementation
type Mongo struct {
	ctx context.Context
	DB  *mongo.Client
}

// NewMongoClient creates a mongo client
func NewMongoClient(connection string, ca string) (*Mongo, error) {
	opts := options.Client()
	opts.ApplyURI(connection)

	if ca != "" {
		roots := x509.NewCertPool()
		ok := roots.AppendCertsFromPEM([]byte(ca))
		if !ok {
			return nil, fmt.Errorf("appending certs from pem")
		}
		opts.SetTLSConfig(&tls.Config{
			RootCAs: roots,
		})
	}

	client, err := mongo.NewClient(opts)
	if err != nil {
		return nil, errors.Wrap(err, "client creation failed")
	}
	ctx := context.Background()
	err = client.Connect(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "connection failed")
	}

	err = client.Ping(ctx, readpref.Primary())
	if err != nil {
		return nil, errors.Wrap(err, "ping failed")
	}

	return &Mongo{
		ctx: context.Background(),
		DB:  client,
	}, nil
}

// Get gets a Note from the DB
func (m *Mongo) Get(id string) (*model.Note, error) {
	var note model.Note
	col := m.DB.Database(defaultDatabase).Collection(defaultCollection)
	err := col.FindOne(m.ctx, bson.M{"uuid": id}).Decode(&note)
	if err != nil {
		return nil, errors.Wrap(err, "finding note")
	}

	return &note, nil
}

// Create creates a note
func (m *Mongo) Create(note *model.Note) error {
	col := m.DB.Database(defaultDatabase).Collection(defaultCollection)
	_, err := col.InsertOne(m.ctx, note)
	if err != nil {
		return errors.Wrap(err, "creating note")
	}
	return nil
}

// Close the connection
func (m *Mongo) Close() error {
	return m.DB.Disconnect(context.Background())
}
