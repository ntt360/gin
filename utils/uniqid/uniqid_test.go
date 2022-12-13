package uniqid

import (
	"fmt"
	"log"
	"testing"
)

func TestFrom10To62(t *testing.T) {
	println(From10To62(920000000))
}

func TestSnowflake62(t *testing.T) {
	for i := 0; i < 10; i++ {
		hashID, err := Snowflake62()
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("hashID: %s, shuffle: %s \n", hashID, Shuffle(hashID))
	}
}
