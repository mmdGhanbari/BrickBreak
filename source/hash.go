package brick_break

import (
  "math/rand"
  "time"
)

const charset = "abcdefghABCDEFGH01234567890"
const length int = 6

var seededRand *rand.Rand = rand.New(
  rand.NewSource(time.Now().UnixNano()))

func StringWithCharset(charset string) string {
  b := make([]byte, length)
  for i := range b {
    b[i] = charset[seededRand.Intn(len(charset))]
  }
  return string(b)
}

func GetRandomHash() string {
  return StringWithCharset(charset)
}
