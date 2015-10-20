package nonsentence

import (
	"encoding/json"
	"fmt"
	"github.com/boltdb/bolt"
	"log"
	"math/rand"
	"strings"
	"time"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

type Nonsentence struct {
	// The bolt database containing all nonsentences
	db *bolt.DB
}

// Create a nonsentence instance using the given database file; will only return once the database
// file has been opened.
func New(file string) (*Nonsentence, error) {
	db, err := bolt.Open(file, 0600, nil)
	if err != nil {
		return nil, err
	}
	
	// Ensure buckets exist in DB
	if err := db.Update(func(tx *bolt.Tx) error {
		if _, err := tx.CreateBucketIfNotExists([]byte("words")); err != nil {
			return err
		}
		if _, err := tx.CreateBucketIfNotExists([]byte("starts")); err != nil {
			return err
		}
		return nil
	}); err != nil {
		return nil, err
	}
	
	return &Nonsentence{
		db: db,
	}, nil
}

// Close the nonsentence database; must always be called on program exit!
func (ns *Nonsentence) Close() error {
	return ns.db.Close()
}

// Add a sentence to to the database
func (ns *Nonsentence) Add(sentence string) error {
	return ns.db.Update(func(tx *bolt.Tx) error {
		wordsBucket := tx.Bucket([]byte("words"))
		startsBucket := tx.Bucket([]byte("starts"))
		if (wordsBucket == nil) || (startsBucket == nil) {
			return fmt.Errorf("Buckets not found")
		}
		
		// Split sentence on whitespace
		var words = strings.Fields(sentence)
		
		if len(words) < 3 {
			log.Printf("Ignoring small sentence: %v", sentence)
			return nil
		}
		
		// Store words in wordsBucket
		for i := 2; i < len(words); i++ {
			if err := storeWords(wordsBucket, words[i-2], words[i-1], words[i]); err != nil {
				return err
			}
		}
		if err := storeWords(wordsBucket, words[len(words)-2], words[len(words)-1], ""); err != nil {
			return err
		}
		
		// Store starts in startsBucket
		key := []byte(words[0] + " " + words[1])
		if err := startsBucket.Put(key, []byte{}); err != nil {
			return err
		}
		
		return nil
	})
}

// Store a word,word -> word sequence
func storeWords(bucket *bolt.Bucket, word1, word2, word3 string) error {
	key := []byte(word1 + " " + word2)
	
	// Get value from bucket and decode it
	rawValue := bucket.Get(key)
	var value []string
	if rawValue == nil {
		value = make([]string, 0, 1)
	} else {
		if err := json.Unmarshal(rawValue, &value); err != nil {
			log.Printf("Cannot decode raw value for key '%v': %v; starting new empty key; old value is: %v", string(key), string(rawValue))
			value = make([]string, 0, 1)
		}
	}
	
	// Add new word to value
	value = append(value, word3)
	
	// Encode value and store it in bucket
	rawValue, err := json.Marshal(value)
	if err != nil {
		return err
	}
	
	if err := bucket.Put(key, rawValue); err != nil {
		return err
	}
	
	// All done
	return nil
}

// Make a new sentence using the database
func (ns *Nonsentence) Make() (string, error) {
	var sentence string
	return sentence, ns.db.View(func(tx *bolt.Tx) error {
		wordsBucket := tx.Bucket([]byte("words"))
		startsBucket := tx.Bucket([]byte("starts"))
		if (wordsBucket == nil) || (startsBucket == nil) {
			return fmt.Errorf("Buckets not found")
		}
		
		word1, word2, err := getStart(startsBucket)
		if err != nil {
			return err
		}
		
		sentence = word1 + " " + word2
		for {
			word, err := getWord(wordsBucket, word1, word2)
			if err != nil {
				return err
			}
			if word == "" {
				break
			}
			sentence = sentence + " " + word
			word1 = word2
			word2 = word
		}
		
		return nil
	})
}

// Get 2 random start words
func getStart(bucket *bolt.Bucket) (string, string, error) {
	var start []byte
	
	var total int = 0
	if err := bucket.ForEach(func(key, _ []byte) error {
		total++
		if rand.Intn(total) == 0 {
			start = key
		}
		return nil
	}); err != nil {
		return "", "", err
	}
	
	var words = strings.Fields(string(start))
	if len(words) != 2 {
		return "", "", fmt.Errorf("Start '%v' does not contain exactly 2 words", string(start))
	}
	
	return words[0], words[1], nil
}

// Get a random word following the given 2 words; empty for stop or none found
func getWord(bucket *bolt.Bucket, word1, word2 string) (string, error) {
	key := []byte(word1 + " " + word2)
	
	// Get value from bucket and decode it
	rawValue := bucket.Get(key)
	var value []string
	if rawValue == nil {
		// None found
		return "", nil
	} else {
		if err := json.Unmarshal(rawValue, &value); err != nil {
			log.Printf("Cannot decode raw value for key '%v': %v; assuming none found; value is: %v", string(key), string(rawValue))
			return "", nil
		}
	}
	
	// Return a random word
	return value[rand.Intn(len(value))], nil
}
