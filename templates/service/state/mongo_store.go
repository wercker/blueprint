package state

import (
	"io"

	"github.com/pkg/errors"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

var (
	// ErrInvalidObjectID occurs when a id is supplied which is not a valid
	// ObjectId.
	ErrInvalidObjectID = errors.New("Invalid ObjectID")
)

// NewMongoStore creates a new MongoStore. Use an empty string for databaseName
// to use the database name that was provided in the connection string.
func NewMongoStore(session *mgo.Session, databaseName string) (*MongoStore, error) {
	//db := session.DB(databaseName)

	return &MongoStore{
		session: session,
		//users: db.C("users"),
	}, nil
}

type MongoStore struct {
	session *mgo.Session

	// Relavant collections objects using the same names as the collections.
	//users          *mgo.Collection
}

var _ Store = (*MongoStore)(nil)

func (s *MongoStore) Healthy() error {
	err := s.session.Ping()
	if err != nil {
		if err == io.EOF {
			s.session.Refresh()
		}

		return err
	}

	return nil
}

func (s *MongoStore) Close() error {
	s.session.Close()
	return nil
}

// ParseObjectID takes id and returns a ObjectID. First it will try to parse
// id as a Hex encoded string, otherwise it will try to parse the []byte
// representation.
// TODO(bvdberg): Move to shared library
func ParseObjectID(id string) (bson.ObjectId, error) {
	var o bson.ObjectId

	if bson.IsObjectIdHex(id) {
		o = bson.ObjectIdHex(id)
		return o, nil
	}

	o = bson.ObjectId(id)
	if o.Valid() {
		return o, nil
	}

	return o, ErrInvalidObjectID
}
