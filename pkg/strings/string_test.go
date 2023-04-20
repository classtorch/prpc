package strings

import "testing"

func TestString(t *testing.T)  {
	t.Log(Split2("consul://127.0.0.1:8500/uclass-account", "://"))
}
