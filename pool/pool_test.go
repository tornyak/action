package pool
import (
	"log"
	"io"
	"sync/atomic"
	"testing"
	"sync"
	"time"
"math/rand"
)

const (
	maxGoroutines = 25
	pooledResources = 2
)

var idCounter int32

//dbConnetion simulates a resource to share
type dbConnection struct {
	ID int32
}

func (dbConn *dbConnection) Close() error {
	log.Println("Close: Connection", dbConn.ID)
	return nil
}

func createConnection() (io.Closer, error) {
	id := atomic.AddInt32(&idCounter, 1)
	log.Println("Create: New Connection", id)

	return &dbConnection{id}, nil
}

func TestOk(t *testing.T) {
	var wg sync.WaitGroup
	wg.Add(maxGoroutines)

	// Create the pool to mannage our connections.
	p, err := New(createConnection, pooledResources)
	if err != nil {
		log.Println(err)
	}

	for query := 0; query < maxGoroutines; query++ {
		// Each go routine needs its own copy of the query value
		// else they will all be sharing the same query variable
		go func(q int){
			performQueries(q, p)
			wg.Done()
		}(query)
	}

	wg.Wait()

	log.Println("Shutdown Program.")
	p.Close()
}

func performQueries(query int, p *Pool) {
	conn, err := p.Acquire()
	if err != nil {
		log.Println(err)
		return
	}

	defer p.Release(conn)

	time.Sleep(time.Duration(rand.Intn(1000)) * time.Millisecond)
	log.Printf("QID[%d] CID[%d]\n", query, conn.(*dbConnection).ID)
}



