package pool_test

import (
	"testing"

	"github.com/theHamdiz/it/pool"
)

// TestNewPool ensures that a pool is correctly initialized and that new objects are created
func TestNewPool(t *testing.T) {
	pool_ := pool.NewPool(func() int { return 42 })

	if pool_ == nil {
		t.Fatal("Expected Pool instance, got nil")
	}

	obj := pool_.Get()
	if obj != 42 {
		t.Errorf("Expected initial object to be 42, got %v", obj)
	}
}

// TestPoolGetPut ensures that objects are correctly retrieved and reused
func TestPoolGetPut(t *testing.T) {
	pool_ := pool.NewPool(func() string { return "new-object" })

	obj1 := pool_.Get()
	if obj1 != "new-object" {
		t.Errorf("Expected 'new-object', got %v", obj1)
	}

	obj1 = "reused-object"
	pool_.Put(obj1)

	// Get it again (should be reused)
	obj2 := pool_.Get()
	if obj2 != "reused-object" {
		t.Errorf("Expected 'reused-object', got %v", obj2)
	}
}

// TestPoolConcurrentAccess ensures that the pool works correctly in concurrent scenarios
func TestPoolConcurrentAccess(t *testing.T) {
	pool_ := pool.NewPool(func() *int {
		val := 0
		return &val
	})

	const workers = 100
	results := make(map[*int]int)
	ch := make(chan *int, workers)

	for i := 0; i < workers; i++ {
		go func() {
			obj := pool_.Get()
			*obj += 1
			pool_.Put(obj)
			ch <- obj
		}()
	}

	// Collect results
	for i := 0; i < workers; i++ {
		obj := <-ch
		results[obj]++
	}

	// Ensure we are reusing objects
	if len(results) > 10 {
		t.Errorf("Expected object reuse, got %d unique objects", len(results))
	}
}

// TetPoolNewObjectCreation ensures a new object is created if the pool is empty
func TestPoolNewObjectCreation(t *testing.T) {
	pool_ := pool.NewPool(func() int { return 99 })

	obj := pool_.Get()
	if obj != 99 {
		t.Errorf("Expected new object to be 99, got %v", obj)
	}
}

// TestPool_PutNilValue ensures putting a nil value doesn't break the pool (only applicable for pointer types)
func TestPoolPutNilValue(t *testing.T) {
	pool_ := pool.NewPool(func() *string {
		str := "initialized"
		return &str
	})

	// Put a nil pointer into the pool
	pool_.Put(nil)

	// Get an object; should still be initialized correctly
	obj := pool_.Get()
	if obj == nil || *obj != "initialized" {
		t.Errorf("Expected 'initialized', got nil or unexpected value")
	}
}
