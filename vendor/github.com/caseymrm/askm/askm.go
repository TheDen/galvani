package askm

import (
	"log"
	"math/rand"
	"reflect"
	"time"
)

const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
const (
	letterIndexBits = 6
	letterIndexMask = 1<<letterIndexBits - 1
	letterIndexMax  = 63 / letterIndexBits
)

var randomSource = rand.NewSource(time.Now().UnixNano())

// RandomString returns a random string of the given length
func RandomString(length int) string {
	b := make([]byte, length)
	for i, cache, remain := length-1, randomSource.Int63(), letterIndexMax; i >= 0; {
		if remain == 0 {
			cache, remain = randomSource.Int63(), letterIndexMax
		}
		if index := int(cache & letterIndexMask); index < len(letterBytes) {
			b[i] = letterBytes[index]
			i--
		}
		cache >>= letterIndexBits
		remain--
	}
	return string(b)
}

// ArbitraryKeyNotInMap returns an arbitrary string key that is not yet used in the map
func ArbitraryKeyNotInMap(mapWithStringKeys interface{}) string {
	v := reflect.ValueOf(mapWithStringKeys)
	if v.Kind() != reflect.Map {
		log.Printf("Warning, %T is not a map with string keys (%v)", mapWithStringKeys, mapWithStringKeys)
		return RandomString(10)
	}
	length := 3 + v.Len()/len(letterBytes)
	key := RandomString(length)
	count := 0
	for {
		keyVal := v.MapIndex(reflect.ValueOf(key))
		if keyVal.Kind() == reflect.Invalid {
			break
		}
		key = RandomString(length)
		count++
		if count > length*len(letterBytes) {
			// This map must be quite full
			length++
		}
	}
	return key
}
