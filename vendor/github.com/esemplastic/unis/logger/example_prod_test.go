package logger

func ExampleNewProd() {
	logger := NewProd()

	logger("It shouldn't print a thing!")

	// Output:
}
