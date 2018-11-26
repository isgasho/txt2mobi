package settings

import (
	"fmt"
	"testing"
	"time"
)

func Test_ss(t *testing.T) {
	fmt.Println(fmt.Sprintf("scale_%s.png", time.Now().Format(time.RFC3339)))
}
