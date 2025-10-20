package syncX

import (
	"sync"
	"sync/atomic"
	"testing"
)

func TestNewLimitPool(t *testing.T) {
	t.Run("create with max tokens", func(t *testing.T) {
		factory := func() string {
			return "new"
		}

		pool := NewLimitPool(5, factory)
		if pool == nil {
			t.Fatal("NewLimitPool returned nil")
		}

		// Verify we can get tokens
		val, ok := pool.Get()
		if !ok {
			t.Error("Should get token successfully")
		}
		if val != "new" {
			t.Errorf("Expected 'new', got %s", val)
		}
	})

	t.Run("zero max tokens", func(t *testing.T) {
		pool := NewLimitPool(0, func() int {
			return 42
		})

		val, ok := pool.Get()
		if ok {
			t.Error("Should not get token with zero max tokens")
		}
		if val != 0 {
			t.Errorf("Expected zero value, got %d", val)
		}
	})
}

func TestLimitPool_Get(t *testing.T) {
	t.Run("get with available tokens", func(t *testing.T) {
		callCount := 0
		pool := NewLimitPool(3, func() int {
			callCount++
			return callCount
		})

		// First three gets should succeed
		for i := 0; i < 3; i++ {
			val, ok := pool.Get()
			if !ok {
				t.Errorf("Get %d: expected to succeed", i+1)
			}
			if val != i+1 {
				t.Errorf("Get %d: expected %d, got %d", i+1, i+1, val)
			}
		}

		// Fourth get should fail (no tokens left)
		val, ok := pool.Get()
		if ok {
			t.Error("Fourth get should fail (no tokens)")
		}
		if val != 0 {
			t.Errorf("Expected zero value, got %d", val)
		}
	})

	t.Run("return value indicates source", func(t *testing.T) {
		pool := NewLimitPool(1, func() string {
			return "from_factory"
		})

		// First get: from pool (with token)
		val1, ok1 := pool.Get()
		if !ok1 {
			t.Error("First get should succeed")
		}
		if val1 != "from_factory" {
			t.Errorf("Expected 'from_factory', got %s", val1)
		}

		// Second get: no token, returns zero value
		val2, ok2 := pool.Get()
		if ok2 {
			t.Error("Second get should fail (no token)")
		}
		if val2 != "" {
			t.Errorf("Expected empty string, got %s", val2)
		}
	})
}

func TestLimitPool_Put(t *testing.T) {
	t.Run("put returns token", func(t *testing.T) {
		pool := NewLimitPool(2, func() int {
			return 100
		})

		// Use both tokens
		val1, _ := pool.Get()
		val2, _ := pool.Get()

		// Try to get third - should fail
		_, ok := pool.Get()
		if ok {
			t.Error("Should not get third token")
		}

		// Put one back
		pool.Put(val1)

		// Should be able to get one now
		val3, ok := pool.Get()
		if !ok {
			t.Error("Should get token after Put")
		}
		if val3 != 100 {
			t.Errorf("Expected 100, got %d", val3)
		}

		_ = val2 // Use variable
	})

	t.Run("put increases token count", func(t *testing.T) {
		pool := NewLimitPool(1, func() string {
			return "value"
		})

		// Use the only token
		val, ok := pool.Get()
		if !ok {
			t.Error("Should get token")
		}

		// Verify no more tokens
		_, ok = pool.Get()
		if ok {
			t.Error("Should not get second token")
		}

		// Put back
		pool.Put(val)

		// Should be able to get again
		val2, ok := pool.Get()
		if !ok {
			t.Error("Should get token after Put")
		}
		if val2 != "value" {
			t.Errorf("Expected 'value', got %s", val2)
		}
	})

	t.Run("put beyond initial max tokens", func(t *testing.T) {
		pool := NewLimitPool(1, func() int {
			return 1
		})

		// Get and put multiple times beyond initial limit
		for i := 0; i < 5; i++ {
			val, ok := pool.Get()
			if !ok {
				t.Errorf("Iteration %d: should get token", i)
			}
			pool.Put(val + 1) // Put back a different value
		}

		// Should still be able to get
		val, ok := pool.Get()
		if !ok {
			t.Error("Should get token after multiple put cycles")
		}
		if val != 1 {
			t.Logf("Got value %d (might be from pool reuse)", val)
		}
	})
}

func TestLimitPool_TokenCounting(t *testing.T) {
	t.Run("token count accuracy", func(t *testing.T) {
		const maxTokens = 5
		pool := NewLimitPool(maxTokens, func() int {
			return 0
		})

		// Use some tokens
		for i := 0; i < 3; i++ {
			_, ok := pool.Get()
			if !ok {
				t.Errorf("Get %d should succeed", i)
			}
		}

		// Put one back
		pool.Put(42)

		// Should have maxTokens - 3 + 1 = 3 tokens available now
		successCount := 0
		for i := 0; i < maxTokens; i++ {
			_, ok := pool.Get()
			if ok {
				successCount++
			}
		}

		if successCount != 3 {
			t.Errorf("Expected 3 successful gets, got %d", successCount)
		}
	})
}

func TestLimitPool_Concurrent(t *testing.T) {
	t.Run("concurrent get and put", func(t *testing.T) {
		const maxTokens = 10
		const goroutines = 20
		const operations = 100

		var createdCount int32
		pool := NewLimitPool(maxTokens, func() []byte {
			atomic.AddInt32(&createdCount, 1)
			return make([]byte, 1024)
		})

		var wg sync.WaitGroup
		var successfulGets int32
		var failedGets int32

		for i := 0; i < goroutines; i++ {
			wg.Add(1)
			go func(id int) {
				defer wg.Done()
				localSuccess := 0
				localFail := 0

				for j := 0; j < operations; j++ {
					buf, ok := pool.Get()
					if ok {
						localSuccess++
						// Do some work
						if len(buf) > 0 {
							buf[0] = byte(id)
						}
						// Return most of them, but not all
						if j%3 != 0 {
							pool.Put(buf)
						}
					} else {
						localFail++
					}
				}

				atomic.AddInt32(&successfulGets, int32(localSuccess))
				atomic.AddInt32(&failedGets, int32(localFail))
			}(i)
		}

		wg.Wait()

		t.Logf("Successful gets: %d, Failed gets: %d, Created: %d",
			successfulGets, failedGets, createdCount)

		// The total successful gets should not exceed maxTokens * (operations + some margin)
		// due to token limiting, but it's hard to predict exactly due to concurrency
		if successfulGets == 0 {
			t.Error("Should have some successful gets")
		}

		// At least some gets should have failed due to token limiting
		if failedGets == 0 {
			t.Log("Note: No gets failed (possible but unlikely with these parameters)")
		}
	})

	t.Run("token count consistency under concurrency", func(t *testing.T) {
		const maxTokens = 5
		const goroutines = 10
		const operations = 50

		pool := NewLimitPool(maxTokens, func() int {
			return 0
		})

		var wg sync.WaitGroup
		results := make(chan bool, goroutines*operations) // track successful gets

		for i := 0; i < goroutines; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				for j := 0; j < operations; j++ {
					val, ok := pool.Get()
					if ok {
						// Simulate work
						_ = val + 1
						// Return most items
						if j%4 != 0 {
							pool.Put(val)
						}
					}
					results <- ok
				}
			}()
		}

		wg.Wait()
		close(results)

		// Count successful operations
		successCount := 0
		for ok := range results {
			if ok {
				successCount++
			}
		}

		t.Logf("Total successful operations: %d", successCount)

		// We can't predict the exact count due to concurrency,
		// but it should be roughly bounded by the token mechanism
		if successCount > maxTokens*operations {
			t.Error("Token limiting seems ineffective")
		}
	})
}

func TestLimitPool_EdgeCases(t *testing.T) {
	t.Run("negative tokens handled correctly", func(t *testing.T) {
		pool := NewLimitPool(1, func() string {
			return "test"
		})

		// Use the only token
		val, ok := pool.Get()
		if !ok {
			t.Error("Should get token")
		}

		// Try to get another - should fail
		_, ok = pool.Get()
		if ok {
			t.Error("Should not get second token")
		}

		// Put back - should restore token
		pool.Put(val)

		// Should be able to get again
		_, ok = pool.Get()
		if !ok {
			t.Error("Should get token after Put")
		}
	})

	t.Run("zero value handling", func(t *testing.T) {
		type testStruct struct {
			value int
		}

		pool := NewLimitPool(1, func() testStruct {
			return testStruct{value: 42}
		})

		// Successful get
		val, ok := pool.Get()
		if !ok || val.value != 42 {
			t.Error("Should get valid value")
		}

		// Failed get returns zero value
		zeroVal, ok := pool.Get()
		if ok {
			t.Error("Should not get token")
		}
		if zeroVal.value != 0 {
			t.Errorf("Expected zero value struct, got %v", zeroVal)
		}
	})

	t.Run("put without prior get", func(t *testing.T) {
		pool := NewLimitPool(2, func() int {
			return 1
		})

		// Put without getting first - should increase token count beyond initial
		pool.Put(100)
		pool.Put(200)
		pool.Put(300)

		// Should be able to get more than initial max tokens
		successCount := 0
		for i := 0; i < 5; i++ {
			_, ok := pool.Get()
			if ok {
				successCount++
			}
		}

		if successCount < 3 {
			t.Errorf("Should get at least 3 values after multiple puts, got %d", successCount)
		}
	})
}

func TestLimitPool_Integration(t *testing.T) {
	t.Run("complete usage cycle", func(t *testing.T) {
		const maxTokens = 3
		var creationCounter int

		pool := NewLimitPool(maxTokens, func() *[]byte {
			creationCounter++
			data := make([]byte, 0, 1024)
			return &data
		})

		// Phase 1: Initial allocation
		items := make([]*[]byte, 0, maxTokens)
		for i := 0; i < maxTokens; i++ {
			item, ok := pool.Get()
			if !ok {
				t.Errorf("Should get token %d", i)
			}
			items = append(items, item)
		}

		if creationCounter != maxTokens {
			t.Errorf("Expected %d creations, got %d", maxTokens, creationCounter)
		}

		// Phase 2: Token exhaustion
		_, ok := pool.Get()
		if ok {
			t.Error("Should not get token beyond max")
		}

		// Phase 3: Return and reuse
		for _, item := range items {
			pool.Put(item)
		}

		// Should be able to get all again
		reuseCount := 0
		for i := 0; i < maxTokens; i++ {
			_, ok := pool.Get()
			if ok {
				reuseCount++
			}
		}

		if reuseCount != maxTokens {
			t.Errorf("Should reuse all %d items, got %d", maxTokens, reuseCount)
		}

		// Creation counter should not increase during reuse phase
		if creationCounter != maxTokens {
			t.Errorf("Expected creation counter to remain %d, got %d", maxTokens, creationCounter)
		}
	})
}

func BenchmarkLimitPool(b *testing.B) {
	b.Run("with tokens", func(b *testing.B) {
		pool := NewLimitPool(100, func() [128]byte {
			return [128]byte{}
		})

		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				data, ok := pool.Get()
				if ok {
					// Simulate work
					data[0] = 1
					pool.Put(data)
				}
			}
		})
	})

	b.Run("token exhaustion", func(b *testing.B) {
		pool := NewLimitPool(1, func() int {
			return 1
		})

		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				val, ok := pool.Get()
				if ok {
					pool.Put(val)
				}
			}
		})
	})
}
