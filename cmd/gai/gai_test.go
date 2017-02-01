package main

// func TestResolvePkgSpec(t *testing.T) {
// 	res := resolvePkgSpec([]string{"github.com/myitcv/..."})

// 	found := false

// 	for _, v := range res {
// 		if v == "github.com/myitcv/gogenerate" {
// 			found = true
// 		}
// 	}

// 	if !found {
// 		t.Fatal("could not find github.com/myitcv/gogenerate package")
// 	}
// }

// func TestGoDo(t *testing.T) {
// 	gg := "github.com/myitcv/gg"
// 	res := goDo([]string{gg}, "go", "install")

// 	if len(res) != 1 {
// 		t.Fatalf("expected 1 result got %v", len(res))
// 	}

// 	if res[0] != gg {
// 		t.Fatalf("expected output %q, got %q", gg, res[0])
// 	}
// }
