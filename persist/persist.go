package persist

import (
	"log"
	"os"
	"time"

	"github.com/eikeon/dynamodb"
	"github.com/eikeon/marvin/nog"
)

var messageTableName string = "MarvinMessage"

func init() {
	if hostname, err := os.Hostname(); err == nil {
		messageTableName = messageTableName + "-" + hostname
	} else {
		log.Println("error getting hostname:", err)
	}
}

func (p *Persist) initDB() dynamodb.DynamoDB {
	db := dynamodb.NewDynamoDB()
	if db != nil {
		t, err := db.Register(messageTableName, (*nog.Message)(nil))
		if err != nil {
			panic(err)
		}
		pt := dynamodb.ProvisionedThroughput{ReadCapacityUnits: 1, WriteCapacityUnits: 1}
		if _, err := db.CreateTable(t.TableName, t.AttributeDefinitions, t.KeySchema, pt, nil); err != nil {
			log.Println("CreateTable:", err)
		}
		for {
			if description, err := db.DescribeTable(messageTableName, nil); err != nil {
				log.Println("DescribeTable err:", err)
			} else {
				log.Println(description.Table.TableStatus)
				if description.Table.TableStatus == "ACTIVE" {
					break
				}
			}
			time.Sleep(time.Second)
		}
	}
	return db
}

type Persist struct {
	db dynamodb.DynamoDB
}

func (p *Persist) Run(in <-chan nog.Message, out chan<- nog.Message) {
	for {
		select {
		case m := <-in:
			if p.db == nil {
				if db := p.initDB(); db != nil {
					p.db = db
				} else {
					log.Println("WARNING: could not create database to persist message.")
					goto DONE
				}
			}
			p.db.PutItem(messageTableName, p.db.ToItem(&m), nil)
		}
	}
DONE:
}

func (p *Persist) Log() (messages []*nog.Message) {
	if p.db != nil {
		when := time.Now().Format(time.RFC3339Nano)
		hash := when[0:10]
		conditions := dynamodb.KeyConditions{"Hash": {[]dynamodb.AttributeValue{{"S": hash}}, "EQ"}}
		if sr, err := p.db.Query(messageTableName, &dynamodb.QueryOptions{KeyConditions: conditions}); err == nil {
			for i := 0; i < sr.Count; i++ {
				messages = append(messages, p.db.FromItem(messageTableName, sr.Items[i]).(*nog.Message))
			}
		} else {
			log.Println("scan error:", err)
		}
	}
	return
}
