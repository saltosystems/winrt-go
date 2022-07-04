package winrt

import "sync"

// RefCount represents a reference counter.
type RefCount struct {
	sync.Mutex
	refs uint64
}

// NewRefCount creates a new reference counter instance with the counter set to zero.
func NewRefCount() *RefCount {
	return &RefCount{}
}

// AddRef increments the reference counter by one
func (r *RefCount) AddRef() uint64 {
	r.Lock()
	defer r.Unlock()
	r.refs++
	return r.refs
}

// Release decrements the reference counter by one. If it was already zero, it will just return zero.
func (r *RefCount) Release() uint64 {
	r.Lock()
	defer r.Unlock()

	if r.refs > 0 {
		r.refs--
	}

	return r.refs
}
