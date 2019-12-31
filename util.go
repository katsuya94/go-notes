package main

func OrDie(err error) {
	if err != nil {
		panic(err)
	}
}
