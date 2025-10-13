package atomicx

import "fmt"

func ExampleNewValue() {
	val := NewValue[int]()
	data := val.Load()
	fmt.Println(data)
	// Output:
	// 0
}

func ExampleNewValueOf() {
	val := NewValueOf[int](123)
	data := val.Load()
	fmt.Println(data)
	// Output:
	// 123
}

func ExampleValue_Load() {
	val := NewValueOf[int](123)
	data := val.Load()
	fmt.Println(data)
	// Output:
	// 123
}

func ExampleValue_Store() {
	val := NewValueOf[int](123)
	data := val.Load()
	fmt.Println(data)
	val.Store(456)
	data = val.Load()
	fmt.Println(data)
	// Output:
	// 123
	// 456
}

func ExampleValue_Swap() {
	val := NewValueOf[int](123)
	oldVal := val.Swap(456)
	newVal := val.Load()
	fmt.Printf("old: %d, new: %d", oldVal, newVal)
	// Output:
	// old: 123, new: 456
}

func ExampleValue_CompareAndSwap() {
	val := NewValueOf[int](123)
	swapped := val.CompareAndSwap(123, 456)
	fmt.Println(swapped)

	swapped = val.CompareAndSwap(455, 459)
	fmt.Println(swapped)
	// Output:
	// true
	// false
}
