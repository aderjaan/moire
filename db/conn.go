package db

import (
	"errors"
	"sync"
	"time"

	"github.com/bulletind/moire/logger"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

var Conn *MConn
var log = logger.Logger

func EpochHours() int64 {
	now := time.Now()
	return 3600 * (now.Unix() / 3600)
}

func EpochNow() int64 {
	now := time.Now()
	return now.UnixNano() / int64(time.Millisecond) //Convert to Milliseconds
}

type M bson.M

func Convert(doc M, out interface{}) {
	stream, err := bson.Marshal(doc)
	if err == nil {
		bson.Unmarshal(stream, out)
	} else {
		panic(err)
	}
}

type MConn struct {
	Session *mgo.Session
	Dbname  string
}

func (self *MConn) getCursor(session *mgo.Session, table string,
	query M) *mgo.Query {

	fields, err1 := query["fields"].(M)
	delete(query, "fields")
	if !err1 {
		fields = M{}
	}

	sort, err2 := query["sort"].(string)
	delete(query, "sort")
	if !err2 {
		sort = "$natural"
	}

	skip, err3 := query["skip"].(int)
	delete(query, "skip")
	if !err3 {
		skip = 0
	}

	limit, err4 := query["limit"].(int)
	delete(query, "limit")
	if !err4 {
		limit = 0
	}

	cursor := self.GetCursor(session, table, query)
	return cursor.Limit(limit).Skip(skip).Sort(sort).Select(fields)
}

type MapReduce mgo.MapReduce

func (self *MConn) MapReduce(session *mgo.Session, table string,
	query M, result interface{}, job *MapReduce) (*mgo.MapReduceInfo, error) {
	db := session.DB(self.Dbname)

	coll := db.C(table)
	realJob := mgo.MapReduce{Map: job.Map, Reduce: job.Reduce,
		Finalize: job.Finalize, Scope: job.Scope, Verbose: true}
	return coll.Find(query).MapReduce(&realJob, result)
}

func (self *MConn) DropIndex(table string, key ...string) error {
	//Create a Session Copy and be responsible for Closing it.
	session := self.Session.Copy()
	db := session.DB(self.Dbname)
	defer session.Close()

	coll := db.C(table)
	return coll.DropIndex(key...)
}

func (self *MConn) DropIndices(table string) error {
	//Create a Session Copy and be responsible for Closing it.
	session := self.Session.Copy()
	db := session.DB(self.Dbname)
	defer session.Close()

	collection := db.C(table)
	indexes, err := collection.Indexes()
	if err == nil {
		for _, index := range indexes {
			err = collection.DropIndex(index.Key...)
			if err != nil {
				return err
			}
		}
	}

	if err != nil {
		panic(err)
	}
	return nil
}

func (self *MConn) findAndApply(table string, query M,
	change mgo.Change, result interface{}) error {
	//Create a Session Copy and be responsible for Closing it.
	session := self.Session.Copy()
	db := session.DB(self.Dbname)
	defer session.Close()

	change.ReturnNew = true

	coll := db.C(table)
	log.Debug("FindAndApply", "query", query)
	_, err := coll.Find(query).Apply(change, result)
	if err != nil {
		log.Error("FindAndApply", "table", table, "error", err)
	}
	return err
}

func (self *MConn) FindAndUpsert(table string, query M,
	doc M, result interface{}) error {
	change := mgo.Change{
		Update:    doc,
		Upsert:    true,
		ReturnNew: true,
	}
	return self.findAndApply(table, query, change, result)
}

func (self *MConn) FindAndUpdate(table string, query M,
	doc M, result interface{}) error {
	change := mgo.Change{
		Update:    doc,
		Upsert:    false,
		ReturnNew: true,
	}
	return self.findAndApply(table, query, change, result)
}

func (self *MConn) EnsureIndex(table string, index mgo.Index) error {
	//Create a Session Copy and be responsible for Closing it.
	session := self.Session.Copy()
	db := session.DB(self.Dbname)
	defer session.Close()

	coll := db.C(table)
	return coll.EnsureIndex(index)
}

func (self *MConn) GetCursor(session *mgo.Session, table string,
	query M) *mgo.Query {
	db := session.DB(self.Dbname)

	coll := db.C(table)
	out := coll.Find(query)

	return out
}

func (self *MConn) Get(session *mgo.Session, table string,
	query M) *mgo.Iter {
	return self.getCursor(session, table, query).Iter()
}

func (self *MConn) HintedGetOne(table string, query M,
	result interface{}, hint string) error {
	//Create a Session Copy and be responsible for Closing it.
	session := self.Session.Copy()
	defer session.Close()

	cursor := self.getCursor(session, table, query).Hint(hint)
	err := cursor.One(result)
	if err != nil {
		log.Error("HintedGetOne", "table", table, "error", err)
	}

	return err
}

func (self *MConn) GetOne(table string, query M,
	result interface{}) error {
	//Create a Session Copy and be responsible for Closing it.
	session := self.Session.Copy()
	defer session.Close()

	cursor := self.getCursor(session, table, query)
	err := cursor.One(result)
	if err != nil {
		log.Error("GetOne", "table", table, "error", err)
	}

	return err
}

func (self *MConn) HintedCount(table string, query M, hint string) int {
	//Create a Session Copy and be responsible for Closing it.
	session := self.Session.Copy()
	defer session.Close()

	cursor := self.getCursor(session, table, query).
		Select(M{"_id": 1}).Hint(hint)
	count, err := cursor.Count()
	if err != nil {
		log.Error("HintedCount", "table", table, "error", err)
	}

	return count
}

func (self *MConn) Count(table string, query M) int {
	//Create a Session Copy and be responsible for Closing it.
	session := self.Session.Copy()
	defer session.Close()

	cursor := self.getCursor(session, table, query).Select(M{"_id": 1})
	count, err := cursor.Count()
	if err != nil {
		log.Error("Count", "table", table, "error", err)
	}

	return count
}

func (self *MConn) Upsert(table string, query M, doc M) error {

	//Create a Session Copy and be responsible for Closing it.
	session := self.Session.Copy()
	db := session.DB(self.Dbname)
	defer session.Close()

	var err error
	if len(doc) == 0 {
		err = errors.New(
			"Empty upsert is blocked. Refer to " +
				"https://github.com/Simversity/blackjack/issues/1051",
		)
	} else {
		coll := db.C(table)
		_, err = coll.Upsert(query, doc)
	}

	if err != nil {
		log.Error("Upsert", "table", table, "error", err)
	}
	return err
}

func AlterDoc(doc *M, operator string, operation M) {
	spec := *doc
	if spec[operator] != nil {
		op, _ := spec[operator].(M)
		for key, value := range op {
			operation[key] = value
		}
	}
	spec[operator] = operation
}

func (self *MConn) Update(table string, query M, doc M) error {

	//Create a Session Copy and be responsible for Closing it.
	session := self.Session.Copy()
	db := session.DB(self.Dbname)
	defer session.Close()

	coll := db.C(table)
	var update_err error
	if len(doc) == 0 {
		update_err = errors.New(
			"Empty Update is blocked. Refer to " +
				"https://github.com/Simversity/blackjack/issues/1051",
		)
	} else {
		AlterDoc(&doc, "$set", M{"updated_on": EpochNow()})
		_, update_err = coll.UpdateAll(query, doc)
	}

	if update_err != nil {
		log.Error("Update", "table", table, "error", update_err)
	}
	return update_err
}

func (self *MConn) Delete(table string, query M) error {
	//Create a Session Copy and be responsible for Closing it.
	session := self.Session.Copy()
	db := session.DB(self.Dbname)
	defer session.Close()

	var delete_err error

	coll := db.C(table)

	_, delete_err = coll.RemoveAll(query)

	if delete_err != nil {
		log.Error("Delete", "table", table, "error", delete_err)
	}

	return delete_err
}

func InArray(key string, arrays ...[]string) bool {
	for _, val := range arrays {
		for _, one := range val {
			if key == one {
				return true
			}
		}
	}
	return false
}

func (self *MConn) Insert(table string, doc interface{}) {
	//Create a Session Copy and be responsible for Closing it.
	session := self.Session.Copy()
	db := session.DB(self.Dbname)
	defer session.Close()

	session.SetSafe(&mgo.Safe{})

	coll := db.C(table)

	err := coll.Insert(doc)
	if err != nil {
		panic(err)
	}

	return
}

func (self *MConn) Aggregate(session *mgo.Session, table string,
	doc []M) *mgo.Pipe {
	//Create a Session Copy and be responsible for Closing it.
	db := session.DB(self.Dbname)

	coll := db.C(table)
	return coll.Pipe(doc)
}

var cached = struct {
	sync.RWMutex
	sessions map[string]*mgo.Session
}{sessions: map[string]*mgo.Session{}}

func GetConn(db_name string, address string, creds ...string) *MConn {
	//Check if the connection has been stored already.
	var session *mgo.Session
	var ok bool

	cached.RLock()
	session, ok = cached.sessions[db_name]
	cached.RUnlock()

	if !ok {
		var username, password string

		if len(creds) > 0 {
			username = creds[0]
			if len(creds) > 1 {
				password = creds[1]
			}
		}

		// Timeout is the amount of time to wait for a server to respond when
		// first connecting and on follow up operations in the session. If
		// timeout is zero, the call may block forever waiting for a connection
		// to be established.

		// FailFast will cause connection and query attempts to fail faster when
		// the server is unavailable, instead of retrying until the configured
		// timeout period. Note that an unavailable server may silently drop
		// packets instead of rejecting them, in which case it's impossible to
		// distinguish it from a slow server, so the timeout stays relevant.

		info := mgo.DialInfo{
			Addrs:    []string{address},
			Timeout:  60 * time.Second,
			FailFast: true,
			Database: db_name,
			Username: username,
			Password: password,
		}

		var err error
		session, err = mgo.DialWithInfo(&info)
		if err != nil {
			panic(err)
		}

		//Save the Session for Later use.

		cached.Lock()
		cached.sessions[db_name] = session
		cached.Unlock()
	}

	//Return only a Session & the name. Let the Consumer make a Session.Copy()
	//to ensure that database state is resumed.

	return &MConn{session, db_name}
}
