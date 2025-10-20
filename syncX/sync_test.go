package syncX

import (
	"errors"
	"sync"
	"testing"
)

func TestMap_Load(t *testing.T) {
	t.Run("key exists", func(t *testing.T) {
		m := &Map[string, int]{}
		m.Store("key1", 100)

		val, ok := m.Load("key1")
		if !ok {
			t.Error("Expected key to exist")
		}
		if val != 100 {
			t.Errorf("Expected value 100, got %d", val)
		}
	})

	t.Run("key does not exist", func(t *testing.T) {
		m := &Map[string, int]{}

		val, ok := m.Load("nonexistent")
		if ok {
			t.Error("Expected key to not exist")
		}
		if val != 0 {
			t.Errorf("Expected zero value 0, got %d", val)
		}
	})

	t.Run("nil value", func(t *testing.T) {
		m := &Map[string, *int]{}
		m.Store("nil_key", nil)

		val, ok := m.Load("nil_key")
		if !ok {
			t.Error("Expected key to exist")
		}
		if val != nil {
			t.Errorf("Expected nil value, got %v", val)
		}
	})
}

func TestMap_Store(t *testing.T) {
	t.Run("store and retrieve", func(t *testing.T) {
		m := &Map[int, string]{}

		m.Store(1, "value1")
		m.Store(2, "value2")

		val, ok := m.Load(1)
		if !ok || val != "value1" {
			t.Errorf("Expected value1, got %v", val)
		}

		val, ok = m.Load(2)
		if !ok || val != "value2" {
			t.Errorf("Expected value2, got %v", val)
		}
	})

	t.Run("overwrite existing key", func(t *testing.T) {
		m := &Map[string, string]{}

		m.Store("key", "old_value")
		m.Store("key", "new_value")

		val, ok := m.Load("key")
		if !ok || val != "new_value" {
			t.Errorf("Expected new_value, got %v", val)
		}
	})
}

func TestMap_LoadOrStore(t *testing.T) {
	t.Run("load existing", func(t *testing.T) {
		m := &Map[string, int]{}
		m.Store("key", 100)

		actual, loaded := m.LoadOrStore("key", 200)
		if !loaded {
			t.Error("Expected to load existing value")
		}
		if actual != 100 {
			t.Errorf("Expected 100, got %d", actual)
		}
	})

	t.Run("store new", func(t *testing.T) {
		m := &Map[string, int]{}

		actual, loaded := m.LoadOrStore("new_key", 300)
		if loaded {
			t.Error("Expected to store new value")
		}
		if actual != 300 {
			t.Errorf("Expected 300, got %d", actual)
		}

		// Verify the value was actually stored
		val, ok := m.Load("new_key")
		if !ok || val != 300 {
			t.Error("Value was not stored correctly")
		}
	})
}

func TestMap_LoadOrStoreFunc(t *testing.T) {
	t.Run("load existing", func(t *testing.T) {
		m := &Map[string, string]{}
		m.Store("key", "existing")

		called := false
		actual, loaded, err := m.LoadOrStoreFunc("key", func() (string, error) {
			called = true
			return "new", nil
		})

		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
		if !loaded {
			t.Error("Expected to load existing value")
		}
		if actual != "existing" {
			t.Errorf("Expected 'existing', got %s", actual)
		}
		if called {
			t.Error("Factory function should not be called for existing key")
		}
	})

	t.Run("store new", func(t *testing.T) {
		m := &Map[string, string]{}

		called := false
		actual, loaded, err := m.LoadOrStoreFunc("new_key", func() (string, error) {
			called = true
			return "created", nil
		})

		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
		if loaded {
			t.Error("Expected to store new value")
		}
		if actual != "created" {
			t.Errorf("Expected 'created', got %s", actual)
		}
		if !called {
			t.Error("Factory function should be called for new key")
		}

		// Verify the value was stored
		val, ok := m.Load("new_key")
		if !ok || val != "created" {
			t.Error("Value was not stored correctly")
		}
	})

	t.Run("factory returns error", func(t *testing.T) {
		m := &Map[string, string]{}

		expectedErr := errors.New("factory error")
		_, _, err := m.LoadOrStoreFunc("key", func() (string, error) {
			return "", expectedErr
		})

		if err != expectedErr {
			t.Errorf("Expected error %v, got %v", expectedErr, err)
		}

		// Verify no value was stored
		_, ok := m.Load("key")
		if ok {
			t.Error("Value should not be stored when factory returns error")
		}
	})

	t.Run("concurrent LoadOrStoreFunc", func(t *testing.T) {
		m := &Map[string, int]{}
		const goroutines = 10
		var wg sync.WaitGroup

		counter := 0
		var mu sync.Mutex

		for i := 0; i < goroutines; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				_, _, err := m.LoadOrStoreFunc("key", func() (int, error) {
					mu.Lock()
					counter++
					val := counter
					mu.Unlock()
					return val, nil
				})
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
			}()
		}

		wg.Wait()

		// Only one goroutine should have called the factory
		if counter != 1 {
			t.Errorf("Expected factory to be called once, got %d times", counter)
		}

		val, ok := m.Load("key")
		if !ok || val != 1 {
			t.Error("Value was not stored correctly")
		}
	})
}

func TestMap_LoadAndDelete(t *testing.T) {
	t.Run("delete existing", func(t *testing.T) {
		m := &Map[string, int]{}
		m.Store("key", 100)

		val, loaded := m.LoadAndDelete("key")
		if !loaded {
			t.Error("Expected to load and delete existing value")
		}
		if val != 100 {
			t.Errorf("Expected 100, got %d", val)
		}

		// Verify the key was deleted
		_, ok := m.Load("key")
		if ok {
			t.Error("Key should be deleted")
		}
	})

	t.Run("delete non-existing", func(t *testing.T) {
		m := &Map[string, int]{}

		val, loaded := m.LoadAndDelete("nonexistent")
		if loaded {
			t.Error("Expected no value to be loaded")
		}
		if val != 0 {
			t.Errorf("Expected zero value, got %d", val)
		}
	})
}

func TestMap_Delete(t *testing.T) {
	m := &Map[string, int]{}
	m.Store("key", 100)

	m.Delete("key")

	_, ok := m.Load("key")
	if ok {
		t.Error("Key should be deleted")
	}

	// Delete non-existing key (should not panic)
	m.Delete("nonexistent")
}

func TestMap_Range(t *testing.T) {
	t.Run("range over all entries", func(t *testing.T) {
		m := &Map[int, string]{}
		expected := map[int]string{
			1: "one",
			2: "two",
			3: "three",
		}

		for k, v := range expected {
			m.Store(k, v)
		}

		found := make(map[int]string)
		m.Range(func(key int, value string) bool {
			found[key] = value
			return true
		})

		if len(found) != len(expected) {
			t.Errorf("Expected %d entries, found %d", len(expected), len(found))
		}

		for k, expectedV := range expected {
			if foundV, ok := found[k]; !ok || foundV != expectedV {
				t.Errorf("Key %d: expected %s, got %s", k, expectedV, foundV)
			}
		}
	})

	t.Run("early termination", func(t *testing.T) {
		m := &Map[int, string]{}
		for i := 0; i < 10; i++ {
			m.Store(i, "value")
		}

		count := 0
		m.Range(func(key int, value string) bool {
			count++
			return count < 5 // Stop after 5 iterations
		})

		if count != 5 {
			t.Errorf("Expected to iterate 5 times, got %d", count)
		}
	})

	t.Run("empty map", func(t *testing.T) {
		m := &Map[string, int]{}

		called := false
		m.Range(func(key string, value int) bool {
			called = true
			return true
		})

		if called {
			t.Error("Range should not call function on empty map")
		}
	})
}

func TestMap_WithPointerKeys(t *testing.T) {
	t.Run("pointer keys comparison", func(t *testing.T) {
		m := &Map[*int, string]{}

		key1 := new(int)
		*key1 = 100
		key2 := new(int)
		*key2 = 100 // Same value, different pointer

		m.Store(key1, "value1")

		// Same pointer should work
		val, ok := m.Load(key1)
		if !ok || val != "value1" {
			t.Error("Failed to load with same pointer")
		}

		// Different pointer with same value should not work
		val, ok = m.Load(key2)
		if ok {
			t.Error("Different pointer should not match")
		}
	})
}

func TestMap_Concurrent(t *testing.T) {
	m := &Map[int, int]{}
	const goroutines = 100
	const operations = 1000
	var wg sync.WaitGroup

	// Concurrent writers
	for i := 0; i < goroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < operations; j++ {
				key := id*operations + j
				m.Store(key, key*2)
				m.Load(key)
				if j%10 == 0 {
					m.Delete(key)
				}
			}
		}(i)
	}

	// Concurrent reader
	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; i < operations; i++ {
			m.Range(func(key int, value int) bool {
				return true
			})
		}
	}()

	wg.Wait()

	// Verify no panics occurred and basic functionality works
	m.Store(999, 1998)
	val, ok := m.Load(999)
	if !ok || val != 1998 {
		t.Error("Basic functionality broken after concurrent access")
	}
}

func TestMap_ZeroValues(t *testing.T) {
	t.Run("distinguish between non-existent and zero value", func(t *testing.T) {
		m := &Map[string, int]{}

		// Store zero value
		m.Store("zero", 0)

		// Should be able to distinguish
		val, ok := m.Load("zero")
		if !ok {
			t.Error("Key with zero value should exist")
		}
		if val != 0 {
			t.Errorf("Expected 0, got %d", val)
		}

		// Non-existent key
		val, ok = m.Load("nonexistent")
		if ok {
			t.Error("Non-existent key should not be found")
		}
		if val != 0 {
			t.Errorf("Expected 0 for non-existent key, got %d", val)
		}
	})
}
