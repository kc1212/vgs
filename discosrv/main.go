package main

// import "github.com/kc1212/vgs/model"

type StringSet struct {
	set map[string]bool
}

func (s *StringSet) Add(v string) bool {
	_, found := s.set[v]
	if !found {
		s.set[v] = true
		return true
	}
	return false
}

func (s *StringSet) Delete(v string) bool {
	_, found := s.set[v]
	if !found {
		return false
	}
	delete(s.set, v)
	return true
}

type DiscoSrv struct {
	gss StringSet
	rms StringSet
}

func main() {
}
