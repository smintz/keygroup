package keygroup

import (
	"context"
	"fmt"
	"strconv"
	"time"
)

func ExampleKeyGroup() {
	f := func(ctx context.Context, key string) {
		intVar, err := strconv.Atoi(key)
		if err == nil {
			time.Sleep(time.Millisecond * time.Duration(intVar))
		}
		fmt.Println("starting", key)
		<-ctx.Done()
		time.Sleep(time.Millisecond * time.Duration(intVar))
		fmt.Println("stopping", key)
	}
	kg := NewKeyGroup(f)
	kg.Add("1", "2", "3", "4", "5")
	time.Sleep(time.Second)
	kg.Update([]string{"1", "2", "4", "5", "6"})
	time.Sleep(time.Second)
	kg.StopKeys("2")
	time.Sleep(time.Second)
	kg.CancelWait()
	// Output:
	// starting 1
	// starting 2
	// starting 3
	// starting 4
	// starting 5
	// stopping 3
	// starting 6
	// stopping 2
	// stopping 1
	// stopping 4
	// stopping 5
	// stopping 6
}
