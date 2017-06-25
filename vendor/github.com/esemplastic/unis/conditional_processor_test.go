package unis

import (
	"strings"
	"testing"
)

type staticPathResolver struct {
	paramStart    byte
	wildcardStart byte
}

func newStaticPathResolver(paramStartSymbol, wildcardStartParamSymbol byte) Processor {
	return staticPathResolver{
		paramStart:    paramStartSymbol,
		wildcardStart: wildcardStartParamSymbol,
	}
}

func (s staticPathResolver) Process(original string) (result string) {
	i := strings.IndexByte(original, s.paramStart)
	v := strings.IndexByte(original, s.wildcardStart)

	return NewConditional(NewRangeEnd(i),
		NewRangeEnd(v)).Process(original)
}

var resolveStaticPath = newStaticPathResolver(':', '*')

func TestConditional(t *testing.T) {
	tests := []originalAgainstResult{
		{"/api/users/:id", "/api/users/"},
		{"/public/assets/*file", "/public/assets/"},
		{"/profile/:id/files/*file", "/profile/"},
	}

	testOriginalAgainstResult(resolveStaticPath, tests, t)
}
