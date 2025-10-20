package syncX

import (
	"sync"
	"sync/atomic"
	"testing"
)

func TestNewPool(t *testing.T) {
	t.Run("create pool with factory", func(t *testing.T) {
		called := false
		factory := func() string {
			called = true
			return "new_value"
		}

		pool := NewPool(factory)
		if pool == nil {
			t.Fatal("NewPool returned nil")
		}

		// Verify factory is called
		val := pool.Get()
		if !called {
			t.Error("Factory function was not called")
		}
		if val != "new_value" {
			t.Errorf("Expected 'new_value', got %s", val)
		}
	})

	t.Run("factory returns non-zero value", func(t *testing.T) {
		type complexStruct struct {
			ID   int
			Name string
			Data []byte
		}

		pool := NewPool(func() complexStruct {
			return complexStruct{
				ID:   1,
				Name: "test",
				Data: make([]byte, 10),
			}
		})

		obj := pool.Get()
		if obj.ID != 1 || obj.Name != "test" || len(obj.Data) != 10 {
			t.Error("Factory did not return expected struct")
		}
	})
}

func TestPool_Get(t *testing.T) {
	t.Run("get new instance when pool is empty", func(t *testing.T) {
		counter := 0
		pool := NewPool(func() int {
			counter++
			return counter
		})

		val := pool.Get()
		if val != 1 {
			t.Errorf("Expected 1, got %d", val)
		}

		val = pool.Get()
		if val != 2 {
			t.Errorf("Expected 2, got %d", val)
		}
	})

	t.Run("get reused instance", func(t *testing.T) {
		pool := NewPool(func() string {
			return "new"
		})

		// Put an instance back
		pool.Put("reused")

		val := pool.Get()
		if val == "reused" {
			// This is expected behavior - should get the reused instance
			t.Log("Correctly retrieved reused instance")
		}
	})

	t.Run("type safety", func(t *testing.T) {
		type myType struct {
			value int
		}

		pool := NewPool(func() myType {
			return myType{value: 42}
		})

		obj := pool.Get()
		if obj.value != 42 {
			t.Errorf("Expected value 42, got %d", obj.value)
		}

		// This should compile and work correctly with the correct type
		obj.value = 100
		pool.Put(obj)

		obj2 := pool.Get()
		// Note: we can't guarantee we get the same instance back due to pool semantics,
		// but if we do, it should have the modified value
		if obj2.value == 100 {
			t.Log("Retrieved modified instance (pool reuse working)")
		}
	})
}

func TestPool_Put(t *testing.T) {
	t.Run("put and retrieve", func(t *testing.T) {
		creationCount := 0
		pool := NewPool(func() []byte {
			creationCount++
			return make([]byte, 1024)
		})

		// Get a new instance
		buf := pool.Get()
		if creationCount != 1 {
			t.Error("Factory should have been called once")
		}

		// Put it back
		pool.Put(buf)

		// Get again - might get the same instance back
		buf2 := pool.Get()
		if creationCount == 1 {
			t.Log("Pool reused instance (creation count unchanged)")
		}
		_ = buf2 // Use the variable
	})

	t.Run("put nil should not panic", func(t *testing.T) {
		pool := NewPool(func() *int {
			return new(int)
		})

		// This should not panic, even though the comment says not to put nil
		// We're testing the behavior with zero values
		var zero *int
		pool.Put(zero)

		// Get should still work
		val := pool.Get()
		if val == nil {
			t.Error("Get returned nil from factory")
		}
	})

	t.Run("put different values", func(t *testing.T) {
		pool := NewPool(func() int {
			return 0
		})

		pool.Put(100)
		pool.Put(200)
		pool.Put(300)

		// We can't guarantee the order of retrieval, but we should get one of the values back
		val := pool.Get()
		if val == 100 || val == 200 || val == 300 {
			t.Logf("Retrieved one of the put values: %d", val)
		} else if val == 0 {
			t.Log("Retrieved new instance from factory")
		}
	})
}

func TestPool_Concurrent(t *testing.T) {
	t.Run("concurrent get and put", func(t *testing.T) {
		const goroutines = 100
		const operations = 1000

		var creationCount int32
		pool := NewPool(func() int {
			atomic.AddInt32(&creationCount, 1)
			return 0
		})

		var wg sync.WaitGroup
		for i := 0; i < goroutines; i++ {
			wg.Add(1)
			go func(id int) {
				defer wg.Done()
				for j := 0; j < operations; j++ {
					// Mix of Get and Put operations
					val := pool.Get()
					if j%2 == 0 {
						pool.Put(val + 1)
					}
				}
			}(i)
		}

		wg.Wait()

		t.Logf("Total creations: %d", creationCount)
		// The main test is that no panic occurred during concurrent access
	})

	t.Run("stress test", func(t *testing.T) {
		type testStruct struct {
			data [64]byte
			idx  int
		}

		var counter int32
		pool := NewPool(func() testStruct {
			atomic.AddInt32(&counter, 1)
			return testStruct{idx: int(atomic.LoadInt32(&counter))}
		})

		var wg sync.WaitGroup
		const workers = 50
		const iterations = 200

		for i := 0; i < workers; i++ {
			wg.Add(1)
			go func(workerID int) {
				defer wg.Done()
				items := make([]testStruct, 0, 10)
				for j := 0; j < iterations; j++ {
					// Get some items
					for k := 0; k < 5; k++ {
						item := pool.Get()
						item.data[0] = byte(workerID)
						items = append(items, item)
					}

					// Put some items back
					for k := 0; k < 3 && len(items) > 0; k++ {
						item := items[len(items)-1]
						items = items[:len(items)-1]
						pool.Put(item)
					}
				}

				// Put remaining items back
				for _, item := range items {
					pool.Put(item)
				}
			}(i)
		}

		wg.Wait()

		t.Logf("Total objects created: %d", atomic.LoadInt32(&counter))
		// Success: no panic during concurrent access
	})
}

func TestPool_ZeroValue(t *testing.T) {
	t.Run("factory returns zero value", func(t *testing.T) {
		pool := NewPool(func() int {
			return 0 // zero value
		})

		val := pool.Get()
		if val != 0 {
			t.Errorf("Expected 0, got %d", val)
		}

		pool.Put(42)
		val = pool.Get()
		// Could get either 42 (reused) or 0 (new)
		if val != 42 && val != 0 {
			t.Errorf("Expected 42 or 0, got %d", val)
		}
	})

	t.Run("struct with zero values", func(t *testing.T) {
		type data struct {
			value    int
			text     string
			slice    []int
			ptr      *int
			disabled bool
		}

		pool := NewPool(func() data {
			return data{} // zero value struct
		})

		item := pool.Get()
		if item.value != 0 || item.text != "" || item.slice != nil || item.ptr != nil || item.disabled != false {
			t.Error("Expected zero value struct")
		}
	})
}

func TestPool_MemoryReuse(t *testing.T) {
	t.Run("verify reuse with pointer tracking", func(t *testing.T) {
		type tracked struct {
			id int
		}

		created := make(map[*tracked]bool)
		var mu sync.Mutex
		var nextID int

		pool := NewPool(func() *tracked {
			mu.Lock()
			defer mu.Unlock()
			obj := &tracked{id: nextID}
			nextID++
			created[obj] = true
			return obj
		})

		// Get and put some objects
		obj1 := pool.Get()
		pool.Put(obj1)

		obj2 := pool.Get()
		pool.Put(obj2)

		obj3 := pool.Get()

		mu.Lock()
		creationCount := len(created)
		mu.Unlock()

		t.Logf("Objects created: %d", creationCount)

		// Clean up references
		_ = obj1
		_ = obj2
		_ = obj3
	})
}

// Benchmark to demonstrate performance characteristics
func BenchmarkPool(b *testing.B) {
	type dataStruct struct {
		values [100]int64
	}

	b.Run("with pool", func(b *testing.B) {
		pool := NewPool(func() dataStruct {
			return dataStruct{}
		})

		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				obj := pool.Get()
				// Simulate some work
				for i := range obj.values {
					obj.values[i] = int64(i)
				}
				pool.Put(obj)
			}
		})
	})

	b.Run("without pool", func(b *testing.B) {
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				obj := dataStruct{}
				// Simulate some work
				for i := range obj.values {
					obj.values[i] = int64(i)
				}
				_ = obj // Use the variable
			}
		})
	})
}

func ExamplePool() {
	// Create a pool for byte slices
	pool := NewPool(func() []byte {
		return make([]byte, 1024)
	})

	// Get a byte slice from the pool
	buf := pool.Get()
	// Use the buffer
	buf[0] = 1
	// Return it to the pool when done
	pool.Put(buf)

	// Output:
}
