package logger

func ExampleNewDev() {
	logger := NewDev()

	logger("It should print to the console")

	// Output:
	// It should print to the console
	//
}
