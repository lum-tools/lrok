package names

import (
	"crypto/md5"
	"fmt"
	"math/rand"
	"time"
)

var adjectives = []string{
	"happy", "brave", "clever", "gentle", "jolly", "kind", "lively", "merry",
	"nice", "polite", "proud", "silly", "wise", "witty", "zealous", "calm",
	"cool", "eager", "fierce", "friendly", "quick", "quiet", "rapid", "sharp",
	"swift", "active", "agile", "alert", "bright", "busy", "caring", "daring",
	"dynamic", "elegant", "expert", "fancy", "fresh", "generous", "graceful",
	"grateful", "honest", "humble", "inspired", "joyful", "keen", "loyal", "mindful",
	"noble", "optimistic", "passionate", "peaceful", "playful", "powerful", "radiant",
	"reliable", "serene", "sincere", "smart", "smooth", "spirited", "steady", "strong",
	"sunny", "talented", "trusty", "vibrant", "warm", "wonderful", "zesty", "bold",
	"curious", "dazzling", "electric", "epic", "fearless", "genuine", "glorious",
	"heroic", "inventive", "magnificent", "magical", "mighty", "mystic",
	"perfect", "stellar", "supreme", "terrific", "ultimate", "valiant", "vivid",
}

var animals = []string{
	"panda", "tiger", "lion", "eagle", "dolphin", "wolf", "fox", "bear",
	"hawk", "owl", "shark", "whale", "leopard", "jaguar", "cheetah", "panther",
	"falcon", "raven", "phoenix", "dragon", "unicorn", "griffin", "pegasus",
	"otter", "badger", "raccoon", "squirrel", "rabbit", "deer", "moose", "elk",
	"bison", "buffalo", "rhino", "elephant", "giraffe", "zebra", "gazelle",
	"antelope", "cougar", "lynx", "bobcat", "coyote", "jackal", "hyena",
	"mongoose", "meerkat", "ferret", "weasel", "marten", "seal", "walrus",
	"penguin", "albatross", "pelican", "flamingo", "crane", "heron", "stork",
	"peacock", "parrot", "cockatoo", "macaw", "toucan", "kingfisher", "woodpecker",
	"sparrow", "robin", "cardinal", "bluejay", "finch", "canary", "swallow",
	"swift", "hummingbird", "condor", "vulture", "kite", "buzzard", "osprey",
	"kestrel", "merlin", "goshawk", "harrier", "caracara", "secretary-bird",
}

// Generate creates a random readable tunnel name like "happy-dolphin"
func Generate() string {
	rand.Seed(time.Now().UnixNano())
	adjective := adjectives[rand.Intn(len(adjectives))]
	animal := animals[rand.Intn(len(animals))]
	return fmt.Sprintf("%s-%s", adjective, animal)
}

// GenerateSeeded creates a deterministic name based on a seed
func GenerateSeeded(seed string) string {
	hash := md5.Sum([]byte(seed))
	
	// Use first byte for adjective, second byte for animal
	adjIdx := int(hash[0]) % len(adjectives)
	animalIdx := int(hash[1]) % len(animals)
	
	return fmt.Sprintf("%s-%s", adjectives[adjIdx], animals[animalIdx])
}

// GenerateUnique generates a unique name not in the existing set
func GenerateUnique(existing map[string]bool, maxAttempts int) string {
	if maxAttempts == 0 {
		maxAttempts = 100
	}
	
	for i := 0; i < maxAttempts; i++ {
		name := Generate()
		if !existing[name] {
			return name
		}
	}
	
	// Fallback: add random suffix
	baseName := Generate()
	suffix := rand.Intn(9000) + 1000
	return fmt.Sprintf("%s-%d", baseName, suffix)
}

