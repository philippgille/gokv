package combiner_test

import (
	"fmt"

	"github.com/philippgille/gokv/combiner"
	"github.com/philippgille/gokv/file"
	"github.com/philippgille/gokv/gomap"
)

func ExampleNewStore() {
	st, err := combiner.NewStore(combiner.Options{
		Parallel: true,
		Backends: []combiner.Backend{
			combiner.MustBackend(gomap.NewStore(gomap.DefaultOptions)),
			combiner.NewBackend(file.NewStore(file.DefaultOptions)),
		},
	})
	if err != nil {
		panic(err)
	}
	defer st.Close()

	st.Set("something", 123)

	var x int
	_, err = st.Get("something", &x)
	if err != nil {
		panic(err)
	}

	fmt.Println(x) // should print '123'
}
